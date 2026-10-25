package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// errAborted は確認を断ったときのエラーです。3箇所から同じ文言を返すので
// まとめてあります。
func errAborted() error {
	return errors.New(i18n.T("中止しました", "aborted"))
}

// runOperation は生成されたコマンド1つ分の実行です。
func (a *App) runOperation(cmd *cobra.Command, sc *spec.Command, args []string) error {
	r, err := a.resolve()
	if err != nil {
		return err
	}

	canonical := canonicalPathParams(sc)
	pathValues := map[string]string{}
	for _, p := range canonical {
		if v, ok := flagValue(cmd, p.Flag); ok {
			pathValues[p.Flag] = v
		}
	}
	// 位置引数はパスパラメータの並び順に対応します。フラグより後に読むので、
	// 両方指定されたときは位置引数が勝ちます（打った通りに動く方を採る）。
	for i, arg := range args {
		if arg == "" || i >= len(canonical) {
			continue
		}
		pathValues[canonical[i].Flag] = arg
	}
	// プロファイルの既定サイトは、まだ決まっていないときだけ効かせます。
	if r.SiteID != "" {
		for _, p := range canonical {
			if p.Name == "site_id" && pathValues[p.Flag] == "" {
				pathValues[p.Flag] = r.SiteID
			}
		}
	}

	op := selectOperation(sc, pathValues)
	if op == nil {
		return missingPathParams(sc, pathValues)
	}

	req, err := a.buildRequest(cmd, op, pathValues, r.SiteID)
	if err != nil {
		return err
	}

	if s, ok := isSweeping(op); ok && !a.g.dryRun {
		given, _ := flagValue(cmd, confirmFlag)
		if err := a.verifySweeping(context.Background(), s, pathValues["site-id"], given); err != nil {
			return err
		}
	}

	dir, uploadingDir := flagValue(cmd, dirFlag)
	if uploadingDir {
		if req.Upload != nil {
			return usagef("%s", i18n.T(
				"--file と --dir は同時に使えません",
				"--file and --dir cannot be used together"))
		}
		if a.g.dryRun {
			return a.printDryRunDir(op, *req, dir)
		}
	}

	if a.g.dryRun {
		return a.printDryRun(op, req)
	}
	if isDestructive(op) {
		ok, err := a.confirm(fmt.Sprintf(i18n.T(
			"%sに対して %s %s を実行します。よろしいですか？",
			"On %s, about to run %s %s. Continue?"), a.target(), op.Method, req.Path))
		if err != nil {
			return err
		}
		if !ok {
			return errAborted()
		}
	}

	client, err := a.client()
	if err != nil {
		return err
	}

	// 差し替えはアセットを消します。--dir と組み合わせると何十件も
	// 消えうるので、始める前に1度だけ確かめます。
	var replacer *assetReplacer
	if wantsReplace, _ := cmd.Flags().GetBool(replaceFlag); wantsReplace {
		ok, err := a.confirm(fmt.Sprintf(i18n.T(
			"%sに対して、送り先にある同名のアセットを消してから送ります。よろしいですか？",
			"On %s, any asset with the same name at the destination will be deleted first. Continue?"), a.target()))
		if err != nil {
			return err
		}
		if !ok {
			return errAborted()
		}
		if replacer, err = a.newAssetReplacer(context.Background(), client, req.Form.Get("site_id")); err != nil {
			return err
		}
	}

	all, _ := cmd.Flags().GetBool(allFlag)

	// 書き出しの失敗はログにしか残りません。送る前の「一番新しいエラー」を
	// 控えておき、後で増えた分だけを見ます（時刻ではなく ID で区切るのは、
	// サーバとこちらの時計がずれるためです）。
	var logMark int64
	verifying := a.g.verifyPublish && isPublishing(op)
	if verifying {
		siteID := pathValues["site-id"]
		if siteID == "" {
			verifying = false
		} else if mark, markErr := a.newestErrorLogID(context.Background(), client, siteID); markErr != nil {
			fmt.Fprintf(a.env.Err, i18n.T(
				"--%s: ログを読めなかったので、書き出せたかどうかは確かめられません（%v）\n",
				"--%s: could not read the log, so whether it publishes is unverified (%v)\n"),
				verifyPublishFlag, markErr)
			verifying = false
		} else {
			logMark = mark
		}
	}

	wait, _ := cmd.Flags().GetBool(waitFlag)

	// ID を渡したプレビューで本体を省いたときは、保存済みの中身を添えます。
	// MT は本体も必須なので、そのままだと 400 になります。
	if bodyParam, ok := previewBodyParam(op); ok {
		if err := a.fillPreviewBody(context.Background(), client, op, req, bodyParam); err != nil {
			return a.describeError(err, nil)
		}
	}

	var resp *api.Response
	if uploadingDir {
		resp, err = a.uploadDir(context.Background(), client, op, *req, dir, replacer)
	} else if all {
		size := 0
		if v, ok := flagValue(cmd, "limit"); ok {
			size, _ = strconv.Atoi(v)
		}
		resp, err = a.fetchAll(context.Background(), client, *req, size)
	} else if wait && isRebuild(op) {
		resp, err = a.waitRebuild(context.Background(), client, *req)
	} else {
		if replacer != nil && req.Upload != nil {
			if err := replacer.remove(context.Background(), req.Form.Get("path"), req.Upload.Filename); err != nil {
				return err
			}
		}
		resp, err = client.Do(context.Background(), *req)
	}
	// 再構築の足踏みは publishError で返ります。応答は出したうえで、
	// 終了コードで区別できるようにします。
	var stalled *publishError
	if errors.As(err, &stalled) {
		err = nil
	}
	if err != nil {
		return a.describeError(err, resp)
	}

	if verifying {
		if verr := a.verifyPublished(context.Background(), client, pathValues["site-id"], logMark); verr != nil {
			// 応答を捨てると、作成したもののIDが失われます。先に出します。
			a.writeOutput(cmd, sc, resp)
			return verr
		}
	}
	if stalled != nil {
		a.writeOutput(cmd, sc, resp)
		return stalled
	}
	return a.writeOutput(cmd, sc, resp)
}

// writeOutput は --fields の絞り込みを含めた出力です。
func (a *App) writeOutput(cmd *cobra.Command, sc *spec.Command, resp *api.Response) error {
	if fields, ok := flagValue(cmd, fieldsFlag); ok {
		if !hasServerFields(sc) {
			return a.writeResponseFields(resp, fields)
		}
		// サーバ側で絞ったぶんは手元で絞り直しません。ただし1つも
		// 一致しないと {} が返るだけで、綴り違いなのか対象が空なのかが
		// 区別できないので、そこだけは知らせます。
		a.warnEmptyFields(resp, fields)
	}
	return a.writeResponse(resp)
}

// selectOperation は与えられたパスパラメータで成立する最初のオペレーションを
// 返します。Ops はパスパラメータの多い順に並んでいるので、より具体的な方が
// 優先されます。
func selectOperation(sc *spec.Command, values map[string]string) *spec.Operation {
	for _, op := range sc.Ops {
		ok := true
		for _, p := range op.PathParams {
			if values[p.Flag] == "" {
				ok = false
				break
			}
		}
		if ok {
			return op
		}
	}
	return nil
}

func missingPathParams(sc *spec.Command, values map[string]string) error {
	op := sc.Ops[len(sc.Ops)-1]
	var missing []string
	for _, p := range op.PathParams {
		if values[p.Flag] == "" {
			missing = append(missing, "--"+p.Flag)
		}
	}
	hint := ""
	for _, p := range op.PathParams {
		if p.Name == "site_id" && values[p.Flag] == "" {
			hint = i18n.T(
				"\nサイトIDを毎回打ちたくなければ mtb profile set --site-id <ID> で既定にできます。",
				"\nTo avoid typing the site ID every time, set a default with mtb profile set --site-id <ID>.")
		}
	}
	return fmt.Errorf(i18n.T("%s が要ります（%s）%s", "%s is required (%s)%s"), strings.Join(missing, " "), op.Path, hint)
}

// buildRequest はフラグから API リクエストを組み立てます。
func (a *App) buildRequest(cmd *cobra.Command, op *spec.Operation, pathValues map[string]string, defaultSiteID string) (*api.Request, error) {
	path := op.Path
	for _, p := range op.PathParams {
		value := pathValues[p.Flag]
		// パスに入る値が数値でないと、URL は組み立てられてしまい 404 として
		// 返ってくる。原因が値の間違いだと分かりにくいので手元で弾く。
		if p.Type == "integer" && !isInteger(value) {
			return nil, usagef(i18n.T("--%s には数値が要ります: %q", "--%s takes a number: %q"), p.Flag, value)
		}
		path = strings.ReplaceAll(path, "{"+p.Name+"}", url.PathEscape(value))
	}

	req := &api.Request{Method: op.Method, Path: path, Query: url.Values{}, Form: url.Values{}}

	for _, p := range op.Query {
		if v, ok := flagValue(cmd, p.Flag); ok {
			a.warnIfLooksLikeFile(p.Flag, v)
			req.Query.Set(p.Name, v)
		}
	}
	if err := a.appendPairs(cmd, extraQueryFlag, req.Query); err != nil {
		return nil, err
	}

	// リソース本体は --set で部品から組めます。明示の JSON があれば
	// それを土台にして、上に重ねます。
	bodies := map[string]string{}
	for _, p := range op.Body {
		// 繰り返して送る配列は JSON ではないので、ここでは扱いません。
		if p.Repeated || (p.Type != "object" && p.Type != "array") {
			continue
		}
		if v, ok := flagValue(cmd, p.Flag); ok {
			payload, err := a.readJSONValue(p.Flag, v)
			if err != nil {
				return nil, err
			}
			bodies[p.Flag] = payload
		}
	}
	if err := a.applySets(cmd, op, bodies); err != nil {
		return nil, err
	}

	for _, p := range op.Body {
		v, ok := flagValue(cmd, p.Flag)
		if _, built := bodies[p.Flag]; built {
			ok = true
		}
		if !ok {
			// 本文の site_id もプロファイルの既定サイトで埋めます
			// （/assets/upload のようにパスではなく本文で受けるものがあるため）。
			if p.Name == "site_id" && defaultSiteID != "" {
				req.Form.Set(p.Name, defaultSiteID)
			}
			continue
		}
		switch {
		case p.Repeated:
			// 選択肢の並びは同じ名前を繰り返して送ります。カンマ区切りでも
			// JSON の配列でも受け取ります（打つ側がどちらを思いつくか
			// 分からないので、両方通します）。
			values, err := splitRepeated(p.Flag, v)
			if err != nil {
				return nil, err
			}
			for _, item := range values {
				req.Form.Add(p.Name, item)
			}
			continue
		}
		switch p.Type {
		case "file":
			content, err := a.env.ReadFile(v)
			if err != nil {
				return nil, usagef("--%s: %v", p.Flag, err)
			}
			name := filepath.Base(v)
			if v == "-" {
				name = "stdin"
			}
			req.Upload = &api.Upload{Field: p.Name, Filename: name, Content: content}
		case "object", "array":
			payload := bodies[p.Flag]
			// 空の更新は「何もしない」ではない。MT は指定しなかった関連
			// （フォルダ・カテゴリなど）を外すことがある。再公開のつもりで
			// 送って中身が変わる事故が起きるので、気づけるようにする。
			if op.Method == http.MethodPut && isEmptyJSONObject(payload) {
				fmt.Fprintf(a.env.Err, i18n.T(
					"警告: --%s が空のオブジェクトです。MT は指定しなかった関連（フォルダ・カテゴリなど）を外すことがあります\n",
					"Warning: --%s is an empty object. MT may clear the associations you left out (folders, categories and so on)\n"),
					p.Flag)
			}
			req.Form.Set(p.Name, payload)
		default:
			a.warnIfLooksLikeFile(p.Flag, v)
			req.Form.Set(p.Name, v)
		}
	}
	if err := a.appendPairs(cmd, extraFormFlag, req.Form); err != nil {
		return nil, err
	}

	if len(req.Query) == 0 {
		req.Query = nil
	}
	if len(req.Form) == 0 {
		req.Form = nil
	}
	return req, nil
}

// readJSONValue は JSON を受け取ります。@ で始まればファイル、"-" なら標準入力。
// ここで構文を検査しておくのは、MT の 400 より手元のエラーの方が直しやすい
// からです。
func (a *App) readJSONValue(flag, value string) (string, error) {
	raw := value
	if strings.HasPrefix(value, "@") || value == "-" {
		name := strings.TrimPrefix(value, "@")
		buf, err := a.env.ReadFile(name)
		if err != nil {
			return "", usagef("--%s: %v", flag, err)
		}
		raw = string(buf)
	}
	if !json.Valid([]byte(raw)) {
		return "", usagef(i18n.T("--%s に渡した値が JSON として不正です", "the value passed to --%s is not valid JSON"), flag)
	}
	return raw, nil
}

// appendPairs は key=value の繰り返しフラグを url.Values に足します。
// 値は mtb api の --form と同じ読み方で、@ファイル名 と - を受け付けます。
func (a *App) appendPairs(cmd *cobra.Command, name string, into url.Values) error {
	pairs, err := cmd.Flags().GetStringArray(name)
	if err != nil {
		return nil // そのコマンドに無ければ何もしない
	}
	return a.addPairs(pairs, name, into)
}

// addPairs は key=value の並びを url.Values に足します。値が @ で始まれば
// ファイルから、- なら標準入力から読みます。長い JSON をシェルに直接書かずに
// 済ませるための逃げ道で、生成コマンドと mtb api で挙動を揃えてあります。
func (a *App) addPairs(pairs []string, flag string, into url.Values) error {
	for _, pair := range pairs {
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			return usagef(i18n.T("--%s は key=value の形で指定してください: %q", "--%s takes key=value: %q"), flag, pair)
		}
		value, err := a.readValue(flag+" "+k, v)
		if err != nil {
			return err
		}
		into.Add(k, value)
	}
	return nil
}

// readValue は @ファイル名 と - をファイル・標準入力からの読み込みに解決します。
// それ以外はそのままの値です。
func (a *App) readValue(label, value string) (string, error) {
	if !strings.HasPrefix(value, "@") && value != "-" {
		return value, nil
	}
	name := strings.TrimPrefix(value, "@")
	buf, err := a.env.ReadFile(name)
	if err != nil {
		return "", usagef("--%s: %v", label, err)
	}
	return string(buf), nil
}

// flagValue は明示的に指定されたフラグの値を返します。既定値のままなら
// 「指定なし」として扱い、API には送りません。MT 側の既定に任せるためです。
func flagValue(cmd *cobra.Command, name string) (string, bool) {
	f := cmd.Flags().Lookup(name)
	if f == nil || !f.Changed {
		return "", false
	}
	return f.Value.String(), true
}

// target は「いまどこに向いているか」を1行で表します。
//
// 確認のときに接続先が見えないのが一番危ない。prod と stg の両方に
// サインインしたまま作業していると、打ったコマンドは同じなので、
// 出ている文字列だけが取り違えに気づける手がかりになります。
func (a *App) target() string {
	r, err := a.resolve()
	if err != nil {
		return "?"
	}
	return fmt.Sprintf(i18n.T("%s（%s/%s）", "%s (%s/%s)"), r.Name, r.BaseURL, r.APIVersion)
}

// isDestructive は確認を挟むべき操作かどうかです。DELETE だけを対象にします。
func isDestructive(op *spec.Operation) bool {
	return strings.EqualFold(op.Method, "DELETE")
}

// confirm は破壊的操作の確認です。--yes があれば聞きません。対話できない
// 環境（エージェントや CI）では黙って進めず、--yes を要求します。
func (a *App) confirm(message string) (bool, error) {
	if a.g.assumeYes {
		return true, nil
	}
	if a.env.Prompt == nil || !a.env.Prompt.Interactive() {
		return false, fmt.Errorf(i18n.T(
			"確認が必要な操作です。意図した操作なら --yes を付けてください: %s",
			"this operation needs confirmation; add --yes if it is what you intend: %s"), message)
	}
	return a.env.Prompt.Confirm(message, false)
}

func (a *App) printDryRun(op *spec.Operation, req *api.Request) error {
	client, err := a.client()
	if err != nil {
		return err
	}
	fmt.Fprintf(a.env.Out, "%s %s\n", req.Method, client.URL(req.Path, req.Query))
	if len(req.Form) > 0 {
		for _, k := range sortedFormKeys(req.Form) {
			for _, v := range req.Form[k] {
				fmt.Fprintf(a.env.Out, "  %s=%s\n", k, v)
			}
		}
	}
	if req.Upload != nil {
		fmt.Fprintf(a.env.Out, "  %s=@%s (%d bytes)\n", req.Upload.Field, req.Upload.Filename, len(req.Upload.Content))
	}
	if isDestructive(op) {
		fmt.Fprintln(a.env.Out, i18n.T(
			"  （破壊的な操作です。実行するには --yes が要ります）",
			"  (destructive operation; running it for real needs --yes)"))
	}
	return nil
}

// warnIfLooksLikeFile は、文字列の項目に @ファイル名 を渡したときに教えます。
//
// @ を展開するのは JSON を取る項目（object / array）とファイル添付だけで、
// 文字列の項目では値そのものとして送られます。MT には @SITE_LOGO のような
// タグ記法があるので一律に展開はできませんが、本当にそのファイルがあるなら
// ほぼ確実に渡し方の間違いです。黙って送ると、届いた先で初めて分かります。
func (a *App) warnIfLooksLikeFile(flag, value string) {
	if !strings.HasPrefix(value, "@") || !a.fileExists(strings.TrimPrefix(value, "@")) {
		return
	}
	fmt.Fprintf(a.env.Err, i18n.T(
		"警告: --%s は文字列の項目なので %q はファイル名ではなく値として送られます\n",
		"Warning: --%s is a string parameter, so %q is sent as the value itself, not as a filename\n"), flag, value)
}

// fileExists は読めるファイルがあるかを返します。渡し方の取り違えを
// 見つけるためだけに使うので、読めなければ false でよい。
func (a *App) fileExists(name string) bool {
	if a.env.ReadFile == nil || name == "" || name == "-" {
		return false
	}
	_, err := a.env.ReadFile(name)
	return err == nil
}

// isInteger は符号つき整数として読めるかを返します。
func isInteger(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '-' || s[0] == '+' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isEmptyJSONObject は {} かどうかを返します（空白は無視）。
func isEmptyJSONObject(payload string) bool {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal([]byte(payload), &obj); err != nil {
		return false
	}
	return len(obj) == 0
}

func sortedFormKeys(v url.Values) []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

// splitRepeated は繰り返し送る配列の値を分解します。JSON の配列と
// カンマ区切りの両方を受け取ります。
func splitRepeated(flag, value string) ([]string, error) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "[") {
		var items []any
		if err := json.Unmarshal([]byte(value), &items); err != nil {
			return nil, usagef(i18n.T("--%s に渡した値が JSON として不正です", "the value passed to --%s is not valid JSON"), flag)
		}
		out := make([]string, 0, len(items))
		for _, item := range items {
			out = append(out, fmt.Sprint(item))
		}
		return out, nil
	}
	var out []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	if len(out) == 0 {
		return nil, usagef(i18n.T("--%s が空です", "--%s is empty"), flag)
	}
	return out, nil
}
