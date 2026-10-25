package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// allFlag は「全部取る」フラグです。
//
// MT の一覧は既定10件で、全件ほしければ limit と offset を動かしながら
// 何度も呼ぶことになります。毎回そのループを書かせるのは無駄なので、
// limit と offset の両方を持つコマンドにだけこのフラグを生やします。
const allFlag = "all"

// defaultPageSize は --all のときの1ページあたりの件数です。
// MT 側に上限は無いようですが、応答が大きくなりすぎないところで刻みます。
const defaultPageSize = 100

// maxPages は暴走よけです。ここに達したら打ち切って、その旨を伝えます。
const maxPages = 500

// hasPagination は limit と offset を両方持つ（＝ページを繰れる）かを返します。
func hasPagination(sc *spec.Command) bool {
	var limit, offset bool
	for _, op := range sc.Ops {
		for _, p := range op.Query {
			switch p.Name {
			case "limit":
				limit = true
			case "offset":
				offset = true
			}
		}
	}
	return limit && offset
}

// listPage は一覧応答のうち、繰るのに要る部分だけを見たものです。
type listPage struct {
	Items        []json.RawMessage `json:"items"`
	TotalResults int               `json:"totalResults"`
}

// fetchAll は limit と offset を動かしながら全ページを取り、1つの応答に畳みます。
//
// 打ち切りの条件は3つ。totalResults に達した、空のページが返った、maxPages に
// 達した。件数が途中で変わっても止まるように、必ず空ページでも抜けます。
func (a *App) fetchAll(ctx context.Context, client *api.Client, req api.Request, pageSize int) (*api.Response, error) {
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	var (
		items []json.RawMessage
		total int
		last  *api.Response
		// extra は items と totalResults 以外の上位のキー（最初のページのもの）。
		// 拡張の一覧は解釈した条件などを添えることがあり、畳むときに落とさない。
		extra map[string]json.RawMessage
	)
	for page := 0; ; page++ {
		if page >= maxPages {
			fmt.Fprintf(a.env.Err, i18n.T(
				"警告: %d ページで打ち切りました。取得できたのは %d 件です\n",
				"Warning: stopped after %d pages; %d items were retrieved\n"), maxPages, len(items))
			break
		}

		q := url.Values{}
		for k, v := range req.Query {
			q[k] = append([]string(nil), v...)
		}
		q.Set("limit", strconv.Itoa(pageSize))
		q.Set("offset", strconv.Itoa(len(items)))

		attempt := req
		attempt.Query = q
		resp, err := client.Do(ctx, attempt)
		if err != nil {
			return resp, err
		}
		last = resp

		var pageDoc listPage
		if err := json.Unmarshal(resp.Body, &pageDoc); err != nil {
			// 一覧の形をしていない応答は、そのまま返すのが正しい。
			if page == 0 {
				return resp, nil
			}
			break
		}
		if page == 0 {
			total = pageDoc.TotalResults
			_ = json.Unmarshal(resp.Body, &extra)
			delete(extra, "items")
			delete(extra, "totalResults")
		}
		if len(pageDoc.Items) == 0 {
			break
		}
		items = append(items, pageDoc.Items...)
		if total > 0 && len(items) >= total {
			break
		}
	}

	if items == nil {
		items = []json.RawMessage{}
	}
	// 集めきった件数を totalResults にします。サーバの申告をそのまま載せると、
	// /search のように「limit で打ち切られると 0 を返す」エンドポイントで
	// 実際に取れた件数と食い違います。
	if total > len(items) {
		fmt.Fprintf(a.env.Err, i18n.T(
			"警告: サーバは %d 件と申告していますが、取得できたのは %d 件です\n",
			"Warning: the server reports %d items but only %d were retrieved\n"), total, len(items))
	}
	doc := map[string]any{"items": items, "totalResults": len(items)}
	for k, v := range extra {
		doc[k] = v
	}
	body, err := json.Marshal(doc)
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
