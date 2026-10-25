package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/prompt"
)

// 終了コード。エージェントやスクリプトが「何が起きたか」で分岐できるように、
// よくある失敗を区別します。0 と 1 以外を足すときは README と
// internal/llmdocs/00-guide.md の表も直すこと。
const (
	// ExitOK は成功です。
	ExitOK = 0
	// ExitError は分類できないエラーです。
	ExitError = 1
	// ExitAuth は未認証・認証切れ（401）です。mtb auth login をやり直す。
	ExitAuth = 2
	// ExitForbidden は権限不足（403）です。リトライしても直りません。
	ExitForbidden = 3
	// ExitNotFound は対象が無い（404）です。
	ExitNotFound = 4
	// ExitUsage は呼び出し方の誤りです（フラグ・JSON・key=value・対話不可）。
	ExitUsage = 5
	// ExitUnreachable は接続できない・タイムアウトです。リトライの余地があります。
	ExitUnreachable = 6
	// ExitServer はサーバ側のエラー（5xx）です。
	ExitServer = 7
	// ExitPublishFailed は「保存はできたが、静的ファイルの書き出しに失敗した」です。
	// --verify-publish を付けたときだけ返します。API は 200 を返すので、
	// これを分けないと「成功したのにサイトが変わらない」に気づけません。
	ExitPublishFailed = 8
)

// usageError は「呼び出し方が違う」ことを表します。送信前に気づいた誤りは
// これに包み、ExitUsage で終わらせます。
type usageError struct{ err error }

func (e *usageError) Error() string { return e.err.Error() }
func (e *usageError) Unwrap() error { return e.err }

// usagef は書式付きで usageError を作ります。
func usagef(format string, args ...any) error {
	return &usageError{fmt.Errorf(format, args...)}
}

// ExitCode はエラーに対応する終了コードを返します。
func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}

	var pe *publishError
	if errors.As(err, &pe) {
		return ExitPublishFailed
	}

	var ue *usageError
	if errors.As(err, &ue) {
		return ExitUsage
	}
	if errors.Is(err, prompt.ErrNotInteractive) {
		return ExitUsage
	}
	if errors.Is(err, api.ErrNotAuthenticated) {
		return ExitAuth
	}
	if errors.Is(err, config.ErrNoProfile) || errors.Is(err, config.ErrNoSuchProfile) {
		return ExitUsage
	}

	if apiErr, ok := api.AsError(err); ok {
		switch {
		case apiErr.Status == http.StatusUnauthorized:
			return ExitAuth
		case apiErr.Status == http.StatusForbidden:
			return ExitForbidden
		case apiErr.Status == http.StatusNotFound:
			return ExitNotFound
		case apiErr.Status >= 500:
			return ExitServer
		case apiErr.Status >= 400:
			return ExitUsage
		}
	}

	// 到達できない・時間切れはリトライの余地があるので分けます。
	var netErr net.Error
	if errors.As(err, &netErr) || errors.Is(err, context.DeadlineExceeded) {
		return ExitUnreachable
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return ExitUnreachable
	}
	if errors.Is(err, os.ErrNotExist) {
		return ExitUsage
	}
	return ExitError
}
