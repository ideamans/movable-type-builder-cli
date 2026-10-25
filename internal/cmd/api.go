package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// newAPICmd は生のリクエストを送る逃げ道です。
//
// 埋め込んだ定義は取り込んだ時点の v7 なので、プラグインが足したエンドポイントや
// MT 側の新しい追加はここを通ります。生成コマンドで足りるならそちらを使うこと
// （--help に説明とパラメータが出るため）。
func (a *App) newAPICmd() *cobra.Command {
	var (
		queries []string
		forms   []string
		upload  string
		noAuth  bool
	)
	cmd := &cobra.Command{
		Use:   i18n.T("api <メソッド> <パス>", "api <method> <path>"),
		Short: i18n.T("任意のエンドポイントを直接叩く", "Call any endpoint directly"),
		Long: i18n.T(
			"パスはバージョンより後ろの部分です（先頭のスラッシュはあってもなくても構いません）。\n\n"+
				"  mtb api GET /sites/1/entries --query limit=5\n"+
				"  mtb api POST /sites/1/entries --form 'entry={\"title\":\"新しい記事\"}'\n"+
				"  mtb api POST /sites/1/assets/upload --file ./photo.jpg --form path=/2026\n\n"+
				"生成済みのコマンドで足りる操作は、そちらを使ってください（--help に\n"+
				"パラメータと、公式リファレンスの該当箇所が出ます）。\n\n"+
				"**プラグインが足したエンドポイントはここから叩きます。** 定義には\n"+
				"載らないのでコマンドは生えませんが（拡張があれば生えます）、\n"+
				"--all・--fields・--set は生成コマンドと同じように使えます。\n\n"+
				"  mtb endpoints list --fields component,verb,route   # 何があるか\n",
			"The path is the part after the version (a leading slash is optional).\n\n"+
				"  mtb api GET /sites/1/entries --query limit=5\n"+
				"  mtb api POST /sites/1/entries --form 'entry={\"title\":\"A new entry\"}'\n"+
				"  mtb api POST /sites/1/assets/upload --file ./photo.jpg --form path=/2026\n\n"+
				"Where a generated command already covers the operation, prefer it: its\n"+
				"--help lists the parameters and links to the official reference.\n\n"+
				"**Endpoints added by plugins are reached from here.** They are not in the\n"+
				"definition so no command is generated for them (an extension can add one),\n"+
				"but --all, --fields and --set work just as they do on the generated commands.\n\n"+
				"  mtb endpoints list --fields component,verb,route   # what this MT has\n"),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			method := strings.ToUpper(args[0])
			path := args[1]
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			req := api.Request{Method: method, Path: path, Query: url.Values{}, Form: url.Values{}, NoAuth: noAuth}
			if err := a.addPairs(queries, "query", req.Query); err != nil {
				return err
			}
			if err := a.addPairs(forms, "form", req.Form); err != nil {
				return err
			}
			// --set は定義を持たないので、名前は検査せず打たれたとおりに組みます。
			bodies := map[string]string{}
			for k := range req.Form {
				bodies[k] = req.Form.Get(k)
			}
			if err := a.applySetsTo(cmd, bodies, func(string) bool { return true }); err != nil {
				return err
			}
			for k, v := range bodies {
				req.Form.Set(k, v)
			}
			if upload != "" {
				buf, err := a.env.ReadFile(upload)
				if err != nil {
					return usagef("--file: %v", err)
				}
				req.Upload = &api.Upload{Field: "file", Filename: baseName(upload), Content: buf}
			}
			if len(req.Query) == 0 {
				req.Query = nil
			}
			if len(req.Form) == 0 {
				req.Form = nil
			}

			if a.g.dryRun {
				client, err := a.client()
				if err != nil {
					return err
				}
				fmt.Fprintf(a.env.Out, "%s %s\n", req.Method, client.URL(req.Path, req.Query))
				for _, k := range sortedFormKeys(req.Form) {
					for _, v := range req.Form[k] {
						fmt.Fprintf(a.env.Out, "  %s=%s\n", k, v)
					}
				}
				return nil
			}
			if method == "DELETE" {
				ok, err := a.confirm(fmt.Sprintf(i18n.T(
					"%sに対して %s %s を実行します。よろしいですか？",
					"On %s, about to run %s %s. Continue?"), a.target(), method, path))
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
			var resp *api.Response
			if all, _ := cmd.Flags().GetBool(allFlag); all {
				resp, err = a.fetchAll(context.Background(), client, req, 0)
			} else {
				resp, err = client.Do(context.Background(), req)
			}
			if err != nil {
				return a.describeError(err, resp)
			}
			if fields, ok := flagValue(cmd, fieldsFlag); ok {
				return a.writeResponseFields(resp, fields)
			}
			return a.writeResponse(resp)
		},
	}
	f := cmd.Flags()
	f.StringArrayVar(&queries, "query", nil, i18n.T("クエリパラメータ（key=value、値の @ファイル名 でファイルから読む、繰り返し可）", "query parameter (key=value; a value of @filename is read from that file; repeatable)"))
	f.StringArrayVar(&forms, "form", nil, i18n.T("本文パラメータ（key=value、値の @ファイル名 でファイルから読む、繰り返し可）", "body parameter (key=value; a value of @filename is read from that file; repeatable)"))
	f.StringVar(&upload, "file", "", i18n.T("multipart で送るファイル（フィールド名は file）", "file to send as multipart (the field name is file)"))
	f.BoolVar(&noAuth, "no-auth", false, i18n.T("認証ヘッダを付けない（公開エンドポイントの確認用）", "omit the authentication header (for checking public endpoints)"))
	// 生成コマンドと同じ道具をここでも使えるようにします。プラグインが
	// 足したエンドポイントこそ、ページ繰りや本体の組み立てが要ります。
	f.Bool(allFlag, false, i18n.T(
		"全ページを取得して1つの応答にまとめる（limit と offset は自動で動かします）",
		"fetch every page and merge them into one response (limit and offset are driven automatically)"))
	f.String(fieldsFlag, "", i18n.T(
		"応答から取り出す属性をカンマ区切りで指定（手元で絞ります）",
		"comma-separated fields to keep from the response (filtered locally)"))
	f.StringArray(setFlag, nil, i18n.T(
		"本文を項目ごとに組み立てる（<名前>.<項目>=<値>、値の @ファイル名 でファイルから読む、繰り返し可）",
		"build a body parameter field by field (<name>.<field>=<value>; a value of @filename is read from that file; repeatable)"))
	f.StringArray(setJSONFlag, nil, i18n.T(
		"--set と同じだが値を JSON として解釈する（数値・真偽値・入れ子）",
		"like --set but the value is parsed as JSON (numbers, booleans, nested objects)"))
	return cmd
}

func baseName(path string) string {
	if i := strings.LastIndexAny(path, "/\\"); i >= 0 {
		return path[i+1:]
	}
	return path
}
