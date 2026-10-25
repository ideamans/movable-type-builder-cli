package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/cmd"
	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/prompt"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// testEnv は外に一切触れない Env です。main の配線だけを見ます。
func testEnv(t *testing.T, respond func(*http.Request) (*http.Response, error)) (*cmd.Env, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	store := config.New(config.NewMemStorage(), config.MapEnv(map[string]string{
		"MTB_BASE_URL":     "https://example.com/mt/mt-data-api.cgi",
		"MTB_ACCESS_TOKEN": "token",
		// 文言の検査は日本語で行います（英語は internal/cmd の language_test.go）。
		"MTB_LANG": "ja",
	}))
	return &cmd.Env{
		In: strings.NewReader(""), Out: out, Err: errOut,
		Store: store, Prompt: &prompt.Scripted{NotInteractive: true},
		Now:  func() time.Time { return time.Unix(1_700_000_000, 0) },
		Spec: spec.MustLoad(),
		NewDoer: func(bool, time.Duration) api.Doer {
			return doerFunc(respond)
		},
		ReadFile:  func(string) ([]byte, error) { return nil, errors.New("no files") },
		WriteFile: func(string, []byte) error { return nil },
	}, out, errOut
}

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(status int, body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func TestRunSuccess(t *testing.T) {
	env, out, errOut := testEnv(t, func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"totalResults":0,"items":[]}`)
	})
	if code := run(env, []string{"sites", "list", "--fields", "id"}); code != cmd.ExitOK {
		t.Errorf("exit = %d, want %d (%s)", code, cmd.ExitOK, errOut.String())
	}
	if !strings.Contains(out.String(), "totalResults") {
		t.Errorf("stdout = %q", out.String())
	}
}

// API のエラーが、そのまま終了コードに変換されて返ること。
func TestRunMapsAPIErrorToExitCode(t *testing.T) {
	for _, tc := range []struct {
		status int
		want   int
	}{
		{http.StatusForbidden, cmd.ExitForbidden},
		{http.StatusNotFound, cmd.ExitNotFound},
		{http.StatusInternalServerError, cmd.ExitServer},
	} {
		env, _, errOut := testEnv(t, func(*http.Request) (*http.Response, error) {
			return jsonResponse(tc.status, `{"error":{"code":1,"message":"x"}}`)
		})
		if code := run(env, []string{"sites", "list", "--fields", "id"}); code != tc.want {
			t.Errorf("%d のとき exit = %d, want %d", tc.status, code, tc.want)
		}
		if !strings.Contains(errOut.String(), "エラー:") {
			t.Errorf("%d のとき標準エラーに出ていません: %q", tc.status, errOut.String())
		}
	}
}

// 呼び出し方の誤りは、送信せずに ExitUsage で返ること。
func TestRunUsageError(t *testing.T) {
	sent := 0
	env, _, errOut := testEnv(t, func(*http.Request) (*http.Response, error) {
		sent++
		return jsonResponse(200, `{}`)
	})
	if code := run(env, []string{"sites", "entries", "get", "not-a-number", "1"}); code != cmd.ExitUsage {
		t.Errorf("exit = %d, want %d", code, cmd.ExitUsage)
	}
	if sent != 0 {
		t.Error("送信しています")
	}
	if !strings.Contains(errOut.String(), "--site-id") {
		t.Errorf("標準エラー = %q", errOut.String())
	}
}

// 知らないフラグも ExitUsage（cobra のエラーを包んでいる）。
func TestRunUnknownFlag(t *testing.T) {
	env, _, _ := testEnv(t, func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{}`)
	})
	if code := run(env, []string{"sites", "list", "--nonexistent"}); code != cmd.ExitUsage {
		t.Errorf("exit = %d, want %d", code, cmd.ExitUsage)
	}
}

// 非推奨の --llm は cobra に入る前に処理され、どの位置でも効く。
func TestRunHandlesLegacyLLMFlag(t *testing.T) {
	for _, args := range [][]string{{"--llm"}, {"sites", "entries", "list", "--llm"}} {
		env, out, _ := testEnv(t, func(*http.Request) (*http.Response, error) {
			t.Errorf("--llm なのに送信しています")
			return jsonResponse(200, `{}`)
		})
		if code := run(env, args); code != cmd.ExitOK {
			t.Errorf("%v: exit = %d", args, code)
		}
		if !strings.Contains(out.String(), "AIエージェント向けリファレンス") {
			t.Errorf("%v: 出力が違います", args)
		}
	}
}

// llm サブコマンドも配線されていること。
func TestRunLLMSubcommand(t *testing.T) {
	env, out, _ := testEnv(t, func(*http.Request) (*http.Response, error) {
		return jsonResponse(200, `{}`)
	})
	if code := run(env, []string{"llm"}); code != cmd.ExitOK {
		t.Fatalf("exit = %d", code)
	}
	if !strings.Contains(out.String(), "Webサービスパスワード") {
		t.Errorf("出力が違います")
	}
}

// 接続できないときは ExitUnreachable。
func TestRunUnreachable(t *testing.T) {
	env, _, _ := testEnv(t, func(*http.Request) (*http.Response, error) {
		return nil, &net.DNSError{Err: "no such host", Name: "example.com"}
	})
	if code := run(env, []string{"sites", "list"}); code != cmd.ExitUnreachable {
		t.Errorf("exit = %d, want %d", code, cmd.ExitUnreachable)
	}
}
