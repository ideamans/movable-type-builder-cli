package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// verifyPublishFlag は「保存できたか」ではなく「書き出せたか」まで確かめる
// フラグです。
//
// MT は記事を保存すると同期で再構築しますが、**そこで失敗しても API は 200 を
// 返します**。しかも失敗の原因は、その記事とは関係のないテンプレートのことが
// あります（テーマ由来のアーカイブテンプレートが1つ壊れていると、記事の
// 個別アーカイブごと書き出されない）。手がかりは MT のログにしか残らないので、
// 探しに行かないかぎり「成功したのにサイトが変わらない」に気づけません。
//
// このフラグを付けると、送信の前後でサイトのエラーログを見比べて、増えていれば
// 標準エラーに出し、終了コード 8 で終わります。
const verifyPublishFlag = "verify-publish"

// publishingOps は、成功してもファイルの書き出しに失敗しうる操作です。
//
// 「サイトの下の非 GET すべて」にしないのは、カテゴリやカスタムフィールドの
// 更新まで毎回2リクエスト増えるからです。再構築が走るものだけを挙げます。
var publishingOps = map[string]bool{
	"POST /sites/{site_id}/entries":                                                 true,
	"PUT /sites/{site_id}/entries/{entry_id}":                                       true,
	"DELETE /sites/{site_id}/entries/{entry_id}":                                    true,
	"POST /sites/{site_id}/pages":                                                   true,
	"PUT /sites/{site_id}/pages/{page_id}":                                          true,
	"DELETE /sites/{site_id}/pages/{page_id}":                                       true,
	"POST /sites/{site_id}/templates/{template_id}/publish":                         true,
	"POST /sites/{site_id}/contentTypes/{content_type_id}/data":                     true,
	"PUT /sites/{site_id}/contentTypes/{content_type_id}/data/{content_data_id}":    true,
	"DELETE /sites/{site_id}/contentTypes/{content_type_id}/data/{content_data_id}": true,
}

// isPublishing は再構築を伴う操作かを返します。
func isPublishing(op *spec.Operation) bool {
	return publishingOps[op.Method+" "+op.Path]
}

// publishError は「保存はできたが書き出せなかった」ことを表します。
// 終了コードを 8 に分けるためだけの型です。
type publishError struct{ err error }

func (e *publishError) Error() string { return e.err.Error() }
func (e *publishError) Unwrap() error { return e.err }

// logRecord はログ1件のうち、この機能が見る部分だけです。
type logRecord struct {
	ID      json.Number `json:"id"`
	Level   string      `json:"level"`
	Message string      `json:"message"`
}

// logPage はログ一覧の応答です。
type logPage struct {
	Items []logRecord `json:"items"`
}

// newestErrorLogID は、いまサイトにある一番新しいエラーログの ID を返します。
//
// **時刻ではなく ID で区切ります。** ログの日時はサーバの時計で、こちらの
// 時計とはずれます。数秒の差でも取りこぼしたり余計なものを拾ったりするので、
// 単調に増える ID を目印にします。
func (a *App) newestErrorLogID(ctx context.Context, client *api.Client, siteID string) (int64, error) {
	page, err := a.fetchErrorLogs(ctx, client, siteID, 1)
	if err != nil {
		return 0, err
	}
	if len(page.Items) == 0 {
		return 0, nil
	}
	return logID(page.Items[0]), nil
}

// fetchErrorLogs は新しい順にエラーログを取ります。
func (a *App) fetchErrorLogs(ctx context.Context, client *api.Client, siteID string, limit int) (*logPage, error) {
	q := url.Values{}
	q.Set("level", "error")
	q.Set("sortBy", "created_on")
	q.Set("sortOrder", "descend")
	q.Set("limit", strconv.Itoa(limit))
	q.Set("fields", "id,level,message")
	resp, err := client.Do(ctx, api.Request{
		Method: "GET",
		Path:   "/sites/" + siteID + "/logs",
		Query:  q,
	})
	if err != nil {
		return nil, err
	}
	var page logPage
	if err := json.Unmarshal(resp.Body, &page); err != nil {
		return nil, err
	}
	return &page, nil
}

// logID は ID を数値にします。MT は文字列で返すことも数値で返すこともあります。
func logID(r logRecord) int64 {
	n, err := r.ID.Int64()
	if err != nil {
		return 0
	}
	return n
}

// verifyLimit は書き出しの失敗を探すときに見るログの件数です。記事1件の
// 保存でカテゴリ別・年別など複数のアーカイブが失敗しうるので、少し多めに見ます。
const verifyLimit = 50

// verifyPublished は since より後のエラーログを探し、あればその内容で
// publishError を返します。
//
// ログが読めないとき（権限が足りないなど）は、確かめられなかったことを
// 伝えるだけにします。**検査の失敗で、成功した保存をエラーにはしません。**
func (a *App) verifyPublished(ctx context.Context, client *api.Client, siteID string, since int64) error {
	page, err := a.fetchErrorLogs(ctx, client, siteID, verifyLimit)
	if err != nil {
		fmt.Fprintf(a.env.Err, i18n.T(
			"--%s: ログを読めなかったので、書き出せたかどうかは確かめられませんでした（%v）\n",
			"--%s: could not read the log, so whether it published is unverified (%v)\n"),
			verifyPublishFlag, err)
		return nil
	}

	var found []logRecord
	for _, r := range page.Items {
		if logID(r) > since {
			found = append(found, r)
		}
	}
	if len(found) == 0 {
		return nil
	}

	// 新しい順に来るので、起きた順に並べ直します。
	var b strings.Builder
	b.WriteString(fmt.Sprintf(i18n.T(
		"保存はできましたが、静的ファイルの書き出しに失敗しています（MT のログに %d 件のエラー）。",
		"the save went through, but publishing the static files failed (%d errors in the MT log)."), len(found)))
	for i := len(found) - 1; i >= 0; i-- {
		b.WriteString("\n  " + oneLine(found[i].Message))
	}
	b.WriteString(i18n.T(
		"\n再構築は関係のないテンプレートの誤りでも失敗します。"+
			"mtb sites templates preview <サイトID> で本文を渡すと、テンプレートを保存せずに試せます。",
		"\nA rebuild fails even on an error in an unrelated template. "+
			"mtb sites templates preview <site-id> renders a body without saving it, which is the quickest way to find it."))
	return &publishError{fmt.Errorf("%s", b.String())}
}
