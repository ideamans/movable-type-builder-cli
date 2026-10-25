package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/prompt"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

const testNow = 1_700_000_000

// stub は fakeDoer が返す応答1つ分です。
type stub struct {
	status      int
	body        string
	contentType string
}

func okJSON(body string) stub { return stub{status: 200, body: body, contentType: "application/json"} }

// recorded は送信されたリクエストの記録です。
type recorded struct {
	Method string
	URL    string
	Header http.Header
	Body   string
}

type fakeDoer struct {
	responses []stub
	sent      []recorded
}

func (f *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	body := ""
	if req.Body != nil {
		buf, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		body = string(buf)
	}
	f.sent = append(f.sent, recorded{Method: req.Method, URL: req.URL.String(), Header: req.Header.Clone(), Body: body})
	if len(f.responses) == 0 {
		return nil, errors.New("fakeDoer: 応答が足りません")
	}
	s := f.responses[0]
	f.responses = f.responses[1:]
	ct := s.contentType
	if ct == "" {
		ct = "text/plain"
	}
	return &http.Response{
		StatusCode: s.status,
		Header:     http.Header{"Content-Type": []string{ct}},
		Body:       io.NopCloser(strings.NewReader(s.body)),
	}, nil
}

// harness は外部依存をすべて差し替えた CLI 一式です。
type harness struct {
	t       *testing.T
	out     *bytes.Buffer
	errOut  *bytes.Buffer
	mem     *config.MemStorage
	store   *config.Store
	doer    *fakeDoer
	prompt  *prompt.Scripted
	env     *Env
	files   map[string]string
	written map[string][]byte
	removed []string
	envVars map[string]string

	lastInsecure bool
	lastTimeout  time.Duration
	slept        []time.Duration
}

type harnessOption func(*harness)

// withEnvVars は環境変数を足します。表示言語（MTB_LANG）は上書きしない
// 限り日本語のままです。
func withEnvVars(vars map[string]string) harnessOption {
	return func(h *harness) {
		for k, v := range vars {
			h.envVars[k] = v
		}
	}
}

// withLang は表示言語を差し替えます。
func withLang(lang string) harnessOption {
	return func(h *harness) { h.envVars["MTB_LANG"] = lang }
}

// withoutProfile は既定のプロファイルを登録しません。
func withoutProfile() harnessOption {
	return func(h *harness) { h.files["__no_profile__"] = "" }
}

func newHarness(t *testing.T, opts ...harnessOption) *harness {
	t.Helper()
	h := &harness{
		t:       t,
		out:     &bytes.Buffer{},
		errOut:  &bytes.Buffer{},
		mem:     config.NewMemStorage(),
		doer:    &fakeDoer{},
		prompt:  &prompt.Scripted{NotInteractive: true},
		files:   map[string]string{},
		written: map[string][]byte{},
		// テストの検査対象は日本語の文言です。英語は language_test.go で
		// 別に確かめます。
		envVars: map[string]string{"MTB_LANG": "ja"},
	}
	for _, opt := range opts {
		opt(h)
	}
	h.store = config.New(h.mem, config.MapEnv(h.envVars))

	if _, skip := h.files["__no_profile__"]; !skip {
		if err := h.store.Save(&config.File{
			Current:  "prod",
			Profiles: map[string]*config.Profile{"prod": {BaseURL: "https://example.com/mt/mt-data-api.cgi", APIVersion: "v7"}},
		}); err != nil {
			t.Fatal(err)
		}
		if err := h.store.PutCredential("prod", &config.Credential{
			SessionID:   "session-1",
			AccessToken: "token-1",
			ExpiresAt:   testNow + 3600,
			Username:    "alice",
		}); err != nil {
			t.Fatal(err)
		}
	}

	h.env = &Env{
		In:     strings.NewReader(""),
		Out:    h.out,
		Err:    h.errOut,
		Store:  h.store,
		Prompt: h.prompt,
		Now:    func() time.Time { return time.Unix(testNow, 0) },
		// 再構築の呼び直しは待たずに回します。
		Sleep: func(d time.Duration) { h.slept = append(h.slept, d) },
		NewDoer: func(insecure bool, timeout time.Duration) api.Doer {
			h.lastInsecure = insecure
			h.lastTimeout = timeout
			return h.doer
		},
		Spec: spec.MustLoad(),
		ReadFile: func(name string) ([]byte, error) {
			content, ok := h.files[name]
			if !ok {
				return nil, fmt.Errorf("そのようなファイルはありません: %s: %w", name, fs.ErrNotExist)
			}
			return []byte(content), nil
		},
		WriteFile: func(name string, data []byte) error {
			h.written[name] = data
			return nil
		},
		ExtensionsDir: "/config/extensions",
		RemoveAll: func(path string) error {
			h.removed = append(h.removed, path)
			return nil
		},
		ListFiles: func(dir string) ([]string, error) {
			prefix := strings.TrimSuffix(dir, "/") + "/"
			var out []string
			for name := range h.files {
				if strings.HasPrefix(name, prefix) {
					out = append(out, strings.TrimPrefix(name, prefix))
				}
			}
			if len(out) == 0 {
				return nil, errors.New("そのようなディレクトリはありません: " + dir)
			}
			sort.Strings(out)
			return out, nil
		},
	}
	return h
}

// run はコマンドを1回実行します。App は毎回作り直します（フラグの状態が
// 実行をまたいで残らないようにするため）。
func (h *harness) run(args ...string) error {
	h.t.Helper()
	h.out.Reset()
	h.errOut.Reset()
	return New(h.env, args...).Execute(args)
}

// mustRun は失敗したらテストを止めます。
func (h *harness) mustRun(args ...string) string {
	h.t.Helper()
	if err := h.run(args...); err != nil {
		h.t.Fatalf("mtb %s: %v", strings.Join(args, " "), err)
	}
	return h.out.String()
}

// reply は次の応答を積みます。
func (h *harness) reply(s ...stub) *harness {
	h.doer.responses = append(h.doer.responses, s...)
	return h
}

// last は最後に送信されたリクエストです。
func (h *harness) last() recorded {
	h.t.Helper()
	if len(h.doer.sent) == 0 {
		h.t.Fatal("リクエストが1件も送信されていません")
	}
	return h.doer.sent[len(h.doer.sent)-1]
}

// mustProfile はテスト用のプロファイル設定を作ります。
func mustProfile(name, baseURL, siteID string) *config.File {
	return &config.File{
		Current: name,
		Profiles: map[string]*config.Profile{
			name: {BaseURL: baseURL, APIVersion: "v7", DefaultSiteID: siteID},
		},
	}
}

// expiredCredential は失効済みのアクセストークンとセッションIDを持つ認証情報です。
func expiredCredential() *config.Credential {
	return &config.Credential{
		SessionID:   "session-1",
		AccessToken: "token-1",
		ExpiresAt:   testNow - 10,
		Username:    "alice",
	}
}

// asciiEscape は非 ASCII 文字を \uXXXX に変換します。Movable Type が返す
// JSON の形をテストで再現するためのもので、テストのソースに生の
// エスケープ列を書かずに済ませます。
func asciiEscape(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r < 0x80 {
			b.WriteRune(r)
			continue
		}
		for _, u := range utf16.Encode([]rune{r}) {
			fmt.Fprintf(&b, "\\u%04x", u)
		}
	}
	return b.String()
}
