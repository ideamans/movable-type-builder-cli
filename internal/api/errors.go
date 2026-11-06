package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// ErrNotAuthenticated は認証情報がまったく無いときのエラーです。
var ErrNotAuthenticated = i18n.Error(
	"認証されていません。mtb auth login を実行してください",
	"not authenticated; run: mtb auth login")

// Error は API がエラーを返したときのエラーです。
type Error struct {
	// Status は HTTP のステータスコードです。
	Status int
	// Code は MT が本文に入れてくるコードです（多くは Status と同じ）。
	Code int
	// Message は MT のエラーメッセージです。
	Message string
	// Body は応答本文そのままです。
	Body []byte
	// Method と URL は何を叩いたかです。
	Method string
	URL    string
}

func (e *Error) Error() string {
	msg := e.Message
	if msg == "" {
		msg = strings.TrimSpace(string(e.Body))
		if len(msg) > 200 {
			msg = msg[:200] + "…"
		}
	}
	if msg == "" {
		msg = http.StatusText(e.Status)
	}
	return fmt.Sprintf("%s %s: %d %s", e.Method, e.URL, e.Status, msg)
}

// Unauthorized は認証切れ・認証失敗かどうかです。
func (e *Error) Unauthorized() bool { return e.Status == http.StatusUnauthorized }

// Forbidden は権限不足かどうかです。
func (e *Error) Forbidden() bool { return e.Status == http.StatusForbidden }

// NotFound は対象が無いかどうかです。
func (e *Error) NotFound() bool { return e.Status == http.StatusNotFound }

// UnknownEndpoint は「そのパスのエンドポイントが MT に無い」404 かどうかです。
// 対象が無い 404（Site not found など）とは助言が違います。
func (e *Error) UnknownEndpoint() bool {
	return e.NotFound() && (e.Message == "" || e.Message == "Unknown endpoint")
}

// Hint は次に何をすればよいかの助言です。無ければ空文字を返します。
//
// MT のメッセージで出し分けます。同じ 403 でも、中身の無い "Forbidden" は
// たいていサイトの Data API が無効なせいで、ロールを見直しても直りません。
func (e *Error) Hint() string {
	switch e.Status {
	case http.StatusUnauthorized:
		return i18n.T(
			"セッションが切れている可能性があります。mtb auth login をやり直してください。",
			"The session may have expired. Sign in again with: mtb auth login")
	case http.StatusForbidden:
		switch {
		case e.Message == "" || e.Message == "Forbidden":
			return i18n.T(
				"そのサイトの Data API が無効になっている可能性があります（新しく作ったサイトは無効で始まり、システム管理者だけが素通りします）。管理画面のサイトの設定「Webサービス」で有効にしてください（mtb-pro があれば mtb sites data-api update <サイトID> --enabled true）。有効なら、このユーザーのロールを確認してください。",
				"The site's Data API may be disabled (new sites start disabled and only system administrators bypass it). Enable it in the site's Web Services settings in the admin screen (with mtb-pro: mtb sites data-api update <site-id> --enabled true). If it is enabled, check this user's role.")
		case strings.Contains(e.Message, "turned off"):
			// サーバの設定でわざと閉じてある（mtb-pro の MtbPro の運用時の設定など）。
			// 文面に「システム管理者が開けられる」とあっても、下の管理者の助言を出さない
			// （管理者のプロファイルでも通らないのに、エージェントが持ち替えて遠回りする）
			return i18n.T(
				"サーバの設定でこの操作は止められています。権限の問題ではないので、プロファイルやロールを変えても通りません。開けるかどうかは MT の管理者が決めることなので、利用者に伝えてください。",
				"The server has this operation turned off on purpose. It is not a permission problem: another profile or role will not get through. Whether to turn it on is the MT administrator's decision, so tell the user.")
		case strings.Contains(e.Message, "mt-config.cgi"):
			// MT の設定で無効になっている（mtb-pro の MtbPro のファイル送信 MtbProAllowFileUpload など）。
			// 文面がどの設定かを言っているので、ロールの助言を出さない（見当違いの方へ行かせない）
			return i18n.T(
				"MT のサーバの設定（mt-config.cgi）で無効になっています。権限の問題ではありません。メッセージにある設定を足せるのは MT のサーバの管理者なので、利用者に伝えてください。",
				"It is disabled in the MT server's configuration (mt-config.cgi). It is not a permission problem. Only the MT server's administrator can add the setting named in the message, so tell the user.")
		case strings.Contains(e.Message, "Only a system administrator"):
			return i18n.T(
				"システム管理者だけができる操作です。ロールでは足りないので、システム管理者のプロファイルで実行してください。",
				"Only a system administrator can do this; no role is enough. Run it with a system administrator's profile.")
		}
		return i18n.T(
			"このユーザーには権限がありません。MT の管理画面でロールを確認してください。",
			"This user lacks the permission. Check their role in the Movable Type admin screen.")
	case http.StatusMethodNotAllowed:
		return i18n.T(
			"PUT / DELETE が前段で止められている可能性があります。--method-override を付けて試してください。",
			"Something in front may be blocking PUT/DELETE. Try again with --method-override.")
	case http.StatusNotFound:
		if e.UnknownEndpoint() {
			return i18n.T(
				"MT にそのエンドポイントがありません。ベースURL（mt-data-api.cgi まで含めるか）とパスを確認してください。",
				"MT has no such endpoint. Check the base URL (it must reach mt-data-api.cgi) and the path.")
		}
		return i18n.T(
			"指定した対象がありません。ID・パス・サイトの取り違えを確認してください。",
			"The target does not exist. Check the IDs, the path and that it is the right site.")
	}
	return ""
}

func newError(method, url string, resp *Response) *Error {
	e := &Error{Status: resp.Status, Method: method, URL: url, Body: resp.Body}
	var payload struct {
		Error struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp.Body, &payload); err == nil {
		e.Code = payload.Error.Code
		e.Message = payload.Error.Message
	}
	return e
}

// AsError はエラー連鎖から *Error を取り出します。
func AsError(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
