package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// dirFlag はディレクトリを丸ごと送るフラグです。
//
// Data API に一括アップロードは無いので、mtb が1ファイルずつ送ります。
// ページを繰る --all と同じ立場のもので、API の都合をそのまま利用者に
// 渡さない、というのがこの CLI の立場です。
//
// デザインのアセットは1枚ずつ上げるものではありません。テーマの静的
// ファイルは theme-images / theme-css のようなフォルダ単位で扱うと
// MT のマニュアルにも書かれていて、その単位で送れないと手でループを
// 書くことになります。
const dirFlag = "dir"

// hasFileParam はファイルを受け取るオペレーションを持つかを返します。
func hasFileParam(sc *spec.Command) bool {
	for _, op := range sc.Ops {
		if fileParam(op) != nil {
			return true
		}
	}
	return false
}

// fileParam はファイル本体のパラメータです。無ければ nil。
func fileParam(op *spec.Operation) *spec.Param {
	for _, p := range op.Body {
		if p.Type == "file" {
			return p
		}
	}
	return nil
}

// hasPathParam は送り先のパスを受け取るかを返します。アップロード先の
// ディレクトリ構成を保つのに使います。
func hasPathParam(op *spec.Operation) bool {
	for _, p := range op.Body {
		if p.Name == "path" {
			return true
		}
	}
	return false
}

// uploadDir は dir 以下のファイルを1つずつ送ります。
//
// **途中で失敗したらそこで止めます。** 残りを送り続けても、どこまで
// 通ったのかが分からない中途半端な状態が広がるだけです。何件目で
// 止まったかを出すので、直して同じコマンドをもう一度叩けます
// （同名のファイルは上書きされるため、やり直しは安全です）。
func (a *App) uploadDir(ctx context.Context, client *api.Client, op *spec.Operation, base api.Request, dir string, replacer *assetReplacer) (*api.Response, error) {
	if a.env.ListFiles == nil {
		return nil, usagef("%s", i18n.T("--dir はこの実行環境では使えません", "--dir is not available in this environment"))
	}
	names, err := a.env.ListFiles(dir)
	if err != nil {
		return nil, usagef("--%s: %v", dirFlag, err)
	}
	names = uploadable(names)
	if len(names) == 0 {
		return nil, usagef(i18n.T("--%s に送れるファイルがありません: %s", "--%s has no files to send: %s"), dirFlag, dir)
	}

	field := fileParam(op).Name
	basePath := base.Form.Get("path")
	keepPath := hasPathParam(op)

	items := make([]json.RawMessage, 0, len(names))
	var last *api.Response
	for i, name := range names {
		content, err := a.env.ReadFile(path.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf(i18n.T("%d/%d %s: %w", "%d/%d %s: %w"), i+1, len(names), name, err)
		}

		req := base
		req.Form = cloneValues(base.Form)
		if keepPath {
			// 空のまま送ると「サイトルート」と「指定なし」の区別が
			// つかないので、値があるときだけ載せます。
			if p := remotePath(basePath, name); p != "" {
				req.Form.Set("path", p)
			} else {
				req.Form.Del("path")
			}
		}
		req.Upload = &api.Upload{Field: field, Filename: path.Base(name), Content: content}

		if a.g.verbose {
			fmt.Fprintf(a.env.Err, "%d/%d %s\n", i+1, len(names), name)
		}
		if replacer != nil {
			if err := replacer.remove(ctx, req.Form.Get("path"), req.Upload.Filename); err != nil {
				return nil, err
			}
		}
		resp, err := client.Do(ctx, req)
		if err != nil {
			// 何件通ったのかを言わないと、やり直しの判断ができません。
			// エラーそのものは包まずに返します——包むと describeError が
			// 中の API エラーを取り出すときに 401/403 の助言が落ちるためです。
			fmt.Fprintf(a.env.Err, i18n.T(
				"%d/%d %s の送信で止まりました（%d 件は送信済み）\n",
				"stopped while sending %d/%d %s (%d already sent)\n"),
				i+1, len(names), name, len(items))
			return resp, err
		}
		last = resp
		items = append(items, json.RawMessage(resp.Body))
	}

	body, err := json.Marshal(map[string]any{"items": items, "totalResults": len(items)})
	if err != nil {
		return nil, err
	}
	merged := &api.Response{Status: 200, Body: body}
	if last != nil {
		merged.Status = last.Status
		merged.Header = last.Header
	}
	return merged, nil
}

// cloneValues は url.Values を複製します。1回分の送信でパスを差し替えても
// 次のファイルに残らないようにするためです。
func cloneValues(v url.Values) url.Values {
	out := make(url.Values, len(v))
	for k, vals := range v {
		out[k] = append([]string(nil), vals...)
	}
	return out
}

// printDryRunDir は --dry-run のときに、何をどこへ送るつもりかを並べます。
func (a *App) printDryRunDir(op *spec.Operation, base api.Request, dir string) error {
	client, err := a.client()
	if err != nil {
		return err
	}
	if a.env.ListFiles == nil {
		return usagef("%s", i18n.T("--dir はこの実行環境では使えません", "--dir is not available in this environment"))
	}
	names, err := a.env.ListFiles(dir)
	if err != nil {
		return usagef("--%s: %v", dirFlag, err)
	}
	names = uploadable(names)
	basePath := base.Form.Get("path")
	keepPath := hasPathParam(op)
	for _, name := range names {
		fmt.Fprintf(a.env.Out, "%s %s\n", base.Method, client.URL(base.Path, base.Query))
		if p := remotePath(basePath, name); keepPath && p != "" {
			fmt.Fprintf(a.env.Out, "  path=%s\n", p)
		}
		fmt.Fprintf(a.env.Out, "  %s=@%s\n", fileParam(op).Name, path.Join(dir, name))
	}
	fmt.Fprintf(a.env.Out, i18n.T("  （%d ファイル）\n", "  (%d files)\n"), len(names))
	return nil
}

// uploadable は隠しファイルを除きます。.DS_Store や .gitkeep をサイトの
// 公開領域へ配ってしまうのは、まず意図ではありません。
func uploadable(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		hidden := false
		for _, part := range strings.Split(name, "/") {
			if strings.HasPrefix(part, ".") {
				hidden = true
				break
			}
		}
		if !hidden {
			out = append(out, name)
		}
	}
	return out
}

// remotePath は送り先を組み立てます。ローカルのディレクトリ構成を
// --path の下にそのまま写します。
func remotePath(base, name string) string {
	sub := path.Dir(name)
	if sub == "." {
		return base
	}
	if base == "" {
		return "/" + sub
	}
	return strings.TrimSuffix(base, "/") + "/" + sub
}
