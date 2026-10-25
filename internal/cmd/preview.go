package cmd

import (
	"context"
	"net/url"
	"strings"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// プレビューは「保存せずに描く」ためのエンドポイントです。テンプレートを
// 試すのに捨て用のテンプレートもサンドボックスのサイトも要りません。
//
// ただし **ID を渡しても本体を省くと 400 になります**
// （`A resource "template" is required.`）。既存のものを描いてみたいだけなのに
// 中身を自分で取り直して渡す必要があり、素直な呼び方が通りません。
// mtb が保存済みの中身を取ってきて添えます。
//
// 本体を渡したときはそちらが勝ちます（差し替えて試すのが本来の使い方なので）。

// previewOps は「ID を取るプレビュー」と、その本体パラメータ名です。
// 取得先は末尾の /preview を落としたパスで、4種とも同じ形をしています。
var previewOps = map[string]string{
	"POST /sites/{site_id}/templates/{template_id}/preview":                               "template",
	"POST /sites/{site_id}/entries/{entry_id}/preview":                                    "entry",
	"POST /sites/{site_id}/pages/{page_id}/preview":                                       "page",
	"POST /sites/{site_id}/contentTypes/{content_type_id}/data/{content_data_id}/preview": "content_data",
}

// previewBodyParam は ID を取るプレビューなら、その本体パラメータ名を返します。
func previewBodyParam(op *spec.Operation) (string, bool) {
	name, ok := previewOps[op.Method+" "+op.Path]
	return name, ok
}

// fillPreviewBody は本体が指定されていないプレビューに、保存済みの中身を添えます。
//
// **原稿のまま取ります。** 記事とページの GET は既定で変換後の HTML を返すので、
// そのままプレビューに渡すと二重に変換されます（Markdown の記事が顕著）。
// noTextFilter=1 を付けられるエンドポイントでは必ず付けます。
func (a *App) fillPreviewBody(ctx context.Context, client *api.Client, op *spec.Operation, req *api.Request, bodyParam string) error {
	if req.Form.Get(bodyParam) != "" {
		return nil
	}
	get := api.Request{
		Method: "GET",
		Path:   strings.TrimSuffix(req.Path, "/preview"),
	}
	if a.supportsNoTextFilter(get.Path) {
		get.Query = url.Values{"noTextFilter": {"1"}}
	}
	resp, err := client.Do(ctx, get)
	if err != nil {
		return err
	}
	if req.Form == nil {
		req.Form = url.Values{}
	}
	req.Form.Set(bodyParam, string(resp.Body))
	return nil
}

// supportsNoTextFilter は、その GET が noTextFilter を受け付けるかを定義から調べます。
// パスの ID は実際の値に置き換わっているので、形で突き合わせます。
func (a *App) supportsNoTextFilter(path string) bool {
	for _, c := range a.env.Spec.Commands {
		for _, op := range c.Ops {
			if op.Method != "GET" || !samePathShape(op.Path, path) {
				continue
			}
			for _, p := range op.Query {
				if p.Name == "noTextFilter" {
					return true
				}
			}
		}
	}
	return false
}

// samePathShape は定義のパス（{site_id} 入り）と実際のパスが同じ形かを見ます。
func samePathShape(pattern, path string) bool {
	ps := strings.Split(strings.Trim(pattern, "/"), "/")
	as := strings.Split(strings.Trim(path, "/"), "/")
	if len(ps) != len(as) {
		return false
	}
	for i := range ps {
		if strings.HasPrefix(ps[i], "{") {
			continue
		}
		if ps[i] != as[i] {
			return false
		}
	}
	return true
}
