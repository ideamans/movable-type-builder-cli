package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// writeResponse は応答を出力します。
//
// 既定は auto で、JSON なら読める形に整形し、そうでなければ本文をそのまま
// 出します。エクスポート・バックアップ・プレビューHTMLのように JSON を
// 返さないエンドポイントがあるためです。
func (a *App) writeResponse(resp *api.Response) error {
	var body []byte
	var err error
	if a.g.extract != "" {
		body, err = a.extractField(resp)
	} else {
		body, err = a.formatBody(resp)
	}
	if err != nil {
		return err
	}
	if a.g.outFile != "" {
		if a.env.WriteFile == nil {
			return errors.New(i18n.T("--out はこの実行環境では使えません", "--out is not available in this environment"))
		}
		if err := a.env.WriteFile(a.g.outFile, body); err != nil {
			return err
		}
		fmt.Fprintf(a.env.Err, i18n.T("%s に %d バイト書き出しました\n", "%s: wrote %d bytes\n"), a.g.outFile, len(body))
		return nil
	}
	if _, err := a.env.Out.Write(body); err != nil {
		return err
	}
	if len(body) > 0 && body[len(body)-1] != '\n' {
		_, err = fmt.Fprintln(a.env.Out)
	}
	return err
}

func (a *App) formatBody(resp *api.Response) ([]byte, error) {
	switch a.g.format {
	case "raw":
		// raw は無加工。バイト列そのものが要るときの逃げ道です。
		return resp.Body, nil
	case "json", "auto", "":
		if a.g.format == "auto" && !resp.IsJSON() {
			return resp.Body, nil
		}
		if !json.Valid(resp.Body) {
			// JSON を期待したのに違うものが返るのは、ベースURLの間違いで
			// MT の管理画面 HTML を掴んでいるときによく起きます。
			return resp.Body, nil
		}
		out := &bytes.Buffer{}
		if a.g.compact {
			if err := json.Compact(out, resp.Body); err != nil {
				return nil, err
			}
		} else if err := json.Indent(out, resp.Body, "", "  "); err != nil {
			return nil, err
		}
		return unescapeUnicode(out.Bytes()), nil
	default:
		return nil, usagef(i18n.T("--format は auto / json / raw のいずれかです: %q", "--format takes auto, json or raw: %q"), a.g.format)
	}
}

// extractField は応答から1つの項目を取り出します。
//
// 文字列ならそのまま（引用符もエスケープも外して）、それ以外は JSON で
// 書きます。jq -r と同じ振る舞いで、テンプレートの text をファイルへ
// 落とすのに要ります。これが無いと、mtb を使うのに jq が要ります。
func (a *App) extractField(resp *api.Response) ([]byte, error) {
	if a.g.format == "raw" {
		return nil, usagef("%s", i18n.T(
			"--format raw と --extract は同時に使えません（raw は無加工という意味です）",
			"--format raw and --extract cannot be used together (raw means untouched)"))
	}
	if !json.Valid(resp.Body) {
		return nil, fmt.Errorf(i18n.T(
			"--extract %s: 応答が JSON ではありません",
			"--extract %s: the response is not JSON"), a.g.extract)
	}
	var doc any
	if err := json.Unmarshal(resp.Body, &doc); err != nil {
		return nil, err
	}
	value, err := walkPath(doc, strings.Split(a.g.extract, "."))
	if err != nil {
		return nil, err
	}
	if s, ok := value.(string); ok {
		return []byte(s), nil
	}
	out := &bytes.Buffer{}
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	if !a.g.compact {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(value); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// walkPath は text や items.0.name のような道をたどります。
// 数字は配列の添字として扱います。
func walkPath(doc any, path []string) (any, error) {
	cur := doc
	for i, key := range path {
		where := strings.Join(path[:i+1], ".")
		switch node := cur.(type) {
		case map[string]any:
			v, ok := node[key]
			if !ok {
				return nil, fmt.Errorf(i18n.T(
					"--extract %s: 応答にありません。ここにあるのは %s",
					"--extract %s: not in the response. What is here: %s"), where, joinKeys(node))
			}
			cur = v
		case []any:
			n, err := strconv.Atoi(key)
			if err != nil {
				// 配列に名前で入ろうとしている。件数の話をしても伝わらない。
				return nil, fmt.Errorf(i18n.T(
					"--extract %s: ここは %d 件の配列なので、%s ではなく数字の添字が要ります（items.0.name のように）",
					"--extract %s: this is an array of %d, so it needs a numeric index, not %s (as in items.0.name)"),
					where, len(node), key)
			}
			if n < 0 || n >= len(node) {
				return nil, fmt.Errorf(i18n.T(
					"--extract %s: 配列の %d 件目は取れません（%d 件あります）",
					"--extract %s: cannot take item %d of the array (it has %d)"), where, n, len(node))
			}
			cur = node[n]
		default:
			return nil, fmt.Errorf(i18n.T(
				"--extract %s: これ以上たどれません",
				"--extract %s: cannot go deeper here"), where)
		}
	}
	return cur, nil
}

// joinKeys は「ここにあるのは」の一覧です。綴り違いはよくあるので、
// 何が取れるのかをその場で見せます。
func joinKeys(node map[string]any) string {
	keys := make([]string, 0, len(node))
	for k := range node {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return strings.Join(keys, ", ")
}

// describeError は API のエラーを、次に何をすればよいかまで含めた形にします。
func (a *App) describeError(err error, resp *api.Response) error {
	e, ok := api.AsError(err)
	if !ok {
		return err
	}
	if hint := e.Hint(); hint != "" {
		return fmt.Errorf("%w\n%s", e, hint)
	}
	return e
}

func sortStrings(s []string) { sort.Strings(s) }

// writeResponseFields は応答から指定した属性だけを残して出力します。
//
// MT は GET にしか fields クエリを用意していないため、作成・更新の応答は
// 全属性が返ってきます。同じ --fields が使えたり使えなかったりする方が
// 混乱するので、サーバが対応していないぶんはここで絞ります。
func (a *App) writeResponseFields(resp *api.Response, fields string) error {
	if !resp.IsJSON() || !json.Valid(resp.Body) {
		// 絞りようがない応答（エクスポートなど）は素通しします。
		return a.writeResponse(resp)
	}
	wanted := splitFields(fields)
	filtered, available, err := pickFields(resp.Body, wanted)
	if err != nil {
		return err
	}
	// 1つも一致しないと {} が出るだけで、何が起きたのか分からない。
	// 属性名の綴り違いはよくあるので、その応答が持っている名前を教える。
	if len(available) > 0 {
		fmt.Fprintf(a.env.Err, i18n.T(
			"--fields の %s はこの応答にありません。使える属性: %s\n",
			"--fields: %s is not in this response. Available fields: %s\n"),
			strings.Join(wanted, ","), strings.Join(available, ", "))
	}
	return a.writeResponse(&api.Response{Status: resp.Status, Header: resp.Header, Body: filtered})
}

func splitFields(s string) []string {
	var out []string
	for _, f := range strings.Split(s, ",") {
		if f = strings.TrimSpace(f); f != "" {
			out = append(out, f)
		}
	}
	return out
}

// pickFields は JSON から指定した属性だけを残します。items を持つ一覧の
// 応答では、items の各要素を絞り、totalResults は残します。
//
// 2つ目の戻り値は「1つも一致しなかったときに、その応答が実際に持っている
// 属性名」です。一致したときは空を返します。
func pickFields(body []byte, fields []string) ([]byte, []string, error) {
	if len(fields) == 0 {
		return body, nil, nil
	}
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		// 配列やスカラーが返る応答は絞らずそのまま返します。
		return body, nil, nil
	}

	if items, ok := doc["items"].([]any); ok {
		picked := make([]any, 0, len(items))
		matched := false
		var sample map[string]any
		for _, item := range items {
			obj, ok := item.(map[string]any)
			if !ok {
				picked = append(picked, item)
				matched = true
				continue
			}
			if sample == nil {
				sample = obj
			}
			got := pickFrom(obj, fields)
			if len(got) > 0 {
				matched = true
			}
			picked = append(picked, got)
		}
		out := map[string]any{"items": picked}
		if total, ok := doc["totalResults"]; ok {
			out["totalResults"] = total
		}
		buf, err := json.Marshal(out)
		return buf, missReport(matched, sample), err
	}

	got := pickFrom(doc, fields)
	buf, err := json.Marshal(got)
	return buf, missReport(len(got) > 0, doc), err
}

// missReport は一致が無かったときだけ、実際にある属性名を名前順で返します。
func missReport(matched bool, sample map[string]any) []string {
	if matched || len(sample) == 0 {
		return nil
	}
	names := make([]string, 0, len(sample))
	for k := range sample {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func pickFrom(obj map[string]any, fields []string) map[string]any {
	out := make(map[string]any, len(fields))
	for _, f := range fields {
		if v, ok := obj[f]; ok {
			out[f] = v
		}
	}
	return out
}

// warnEmptyFields は、サーバ側で絞った結果が空だったときに知らせます。
//
// --fields の綴りを間違えると MT は {} を返します。作成直後に --fields id で
// {} が返ると、**作成が失敗したのか属性名を間違えたのか区別が付きません**。
// 手元で絞る経路には既に案内がある（writeResponseFields）ので、サーバ側で
// 絞る経路の穴を埋めます。
func (a *App) warnEmptyFields(resp *api.Response, fields string) {
	if !resp.IsJSON() || !json.Valid(resp.Body) {
		return
	}
	var doc map[string]any
	if err := json.Unmarshal(resp.Body, &doc); err != nil {
		return
	}
	empty := false
	if items, ok := doc["items"].([]any); ok {
		if len(items) == 0 {
			// 対象が0件なだけです。--fields のせいではありません。
			return
		}
		empty = true
		for _, item := range items {
			obj, ok := item.(map[string]any)
			if !ok || len(obj) > 0 {
				empty = false
				break
			}
		}
	} else {
		empty = len(doc) == 0
	}
	if !empty {
		return
	}
	fmt.Fprintf(a.env.Err, i18n.T(
		"--fields の %s はこの応答に1つも一致しませんでした。"+
			"このエンドポイントはサーバ側で絞るので、使える属性名は --fields を外すと分かります\n",
		"--fields: none of %s matched this response. "+
			"This endpoint filters on the server, so drop --fields to see what it has\n"), fields)
}
