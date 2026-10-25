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

// Hint は次に何をすればよいかの助言です。無ければ空文字を返します。
func (e *Error) Hint() string {
	switch e.Status {
	case http.StatusUnauthorized:
		return i18n.T(
			"セッションが切れている可能性があります。mtb auth login をやり直してください。",
			"The session may have expired. Sign in again with: mtb auth login")
	case http.StatusForbidden:
		return i18n.T(
			"このユーザーには権限がありません。MT の管理画面でロールを確認してください。",
			"This user lacks the permission. Check their role in the Movable Type admin screen.")
	case http.StatusMethodNotAllowed:
		return i18n.T(
			"PUT / DELETE が前段で止められている可能性があります。--method-override を付けて試してください。",
			"Something in front may be blocking PUT/DELETE. Try again with --method-override.")
	case http.StatusNotFound:
		return i18n.T(
			"パスとIDを確認してください。ベースURLの間違い（mt-data-api.cgi まで含めるか）も 404 になります。",
			"Check the path and the IDs. A wrong base URL (it must reach mt-data-api.cgi) also returns 404.")
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
