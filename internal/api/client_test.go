package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/ideamans/movable-type-builder-cli/internal/api"
	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// recorded は fakeDoer が記録した1リクエストです。本文は読み切ったものを持ちます。
type recorded struct {
	Method string
	URL    string
	Header http.Header
	Body   string
}

// fakeDoer は応答を順に返す HTTP 送信のスタブです。
type fakeDoer struct {
	responses []stub
	sent      []recorded
	err       error
}

type stub struct {
	status      int
	body        string
	contentType string
}

func jsonStub(status int, body string) stub {
	return stub{status: status, body: body, contentType: "application/json"}
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
	f.sent = append(f.sent, recorded{
		Method: req.Method,
		URL:    req.URL.String(),
		Header: req.Header.Clone(),
		Body:   body,
	})
	if f.err != nil {
		return nil, f.err
	}
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

func (f *fakeDoer) last() recorded { return f.sent[len(f.sent)-1] }

const fixedNow = 1_700_000_000

func newClient(doer *fakeDoer, cred *config.Credential) (*api.Client, *api.MemoryTokenStore) {
	tokens := &api.MemoryTokenStore{Cred: cred}
	return &api.Client{
		BaseURL:    "https://example.com/mt/mt-data-api.cgi",
		APIVersion: "v7",
		ClientID:   "mtb",
		HTTP:       doer,
		Tokens:     tokens,
		Now:        func() time.Time { return time.Unix(fixedNow, 0) },
	}, tokens
}

func TestURLBuilding(t *testing.T) {
	c, _ := newClient(&fakeDoer{}, nil)
	got := c.URL("/sites/1/entries", url.Values{"limit": {"5"}})
	want := "https://example.com/mt/mt-data-api.cgi/v7/sites/1/entries?limit=5"
	if got != want {
		t.Errorf("URL = %q, want %q", got, want)
	}
	if c.Endpoint() != "https://example.com/mt/mt-data-api.cgi/v7" {
		t.Errorf("Endpoint = %q", c.Endpoint())
	}
}

func TestDoSendsAccessToken(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(200, `{"totalResults":0,"items":[]}`)}}
	c, _ := newClient(doer, &config.Credential{
		AccessToken: "token-1",
		ExpiresAt:   fixedNow + 3600,
	})

	resp, err := c.Do(context.Background(), api.Request{
		Method: http.MethodGet,
		Path:   "/sites/1/entries",
		Query:  url.Values{"limit": {"5"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != 200 || !resp.IsJSON() {
		t.Errorf("resp = %d %s", resp.Status, resp.ContentType())
	}
	sent := doer.last()
	if got := sent.Header.Get("X-MT-Authorization"); got != "MTAuth accessToken=token-1" {
		t.Errorf("認証ヘッダ = %q", got)
	}
	if !strings.HasSuffix(sent.URL, "/v7/sites/1/entries?limit=5") {
		t.Errorf("URL = %q", sent.URL)
	}
}

// 失効したトークンは、送る前にセッションから取り直す。ここが「気軽な認証」の要。
func TestDoRefreshesExpiredTokenBeforeSending(t *testing.T) {
	doer := &fakeDoer{responses: []stub{
		jsonStub(200, `{"accessToken":"token-2","expiresIn":3600}`),
		jsonStub(200, `{"ok":true}`),
	}}
	c, tokens := newClient(doer, &config.Credential{
		SessionID:   "session-1",
		AccessToken: "token-1",
		ExpiresAt:   fixedNow - 10,
	})

	if _, err := c.Do(context.Background(), api.Request{Method: http.MethodGet, Path: "/sites"}); err != nil {
		t.Fatal(err)
	}
	if len(doer.sent) != 2 {
		t.Fatalf("送信回数 = %d, want 2", len(doer.sent))
	}
	refresh := doer.sent[0]
	if !strings.HasSuffix(refresh.URL, "/v7/token") || refresh.Method != http.MethodPost {
		t.Errorf("更新リクエストが違います: %s %s", refresh.Method, refresh.URL)
	}
	if got := refresh.Header.Get("X-MT-Authorization"); got != "MTAuth sessionId=session-1" {
		t.Errorf("更新はセッションIDで認証すべきです: %q", got)
	}
	if got := doer.sent[1].Header.Get("X-MT-Authorization"); got != "MTAuth accessToken=token-2" {
		t.Errorf("新しいトークンで送っていません: %q", got)
	}
	if tokens.Cred.AccessToken != "token-2" {
		t.Errorf("保存されたトークン = %q", tokens.Cred.AccessToken)
	}
	if want := int64(fixedNow + 3600); tokens.Cred.ExpiresAt != want {
		t.Errorf("失効時刻 = %d, want %d", tokens.Cred.ExpiresAt, want)
	}
	if tokens.Cred.SessionID != "session-1" {
		t.Errorf("セッションIDが失われています: %q", tokens.Cred.SessionID)
	}
}

// サーバ側でトークンが先に無効化されていた場合、401 を見てから1度だけ取り直す。
func TestDoRetriesOnceAfterUnauthorized(t *testing.T) {
	doer := &fakeDoer{responses: []stub{
		jsonStub(401, `{"error":{"code":401,"message":"Invalid access token"}}`),
		jsonStub(200, `{"accessToken":"token-2","expiresIn":3600}`),
		jsonStub(200, `{"ok":true}`),
	}}
	c, _ := newClient(doer, &config.Credential{
		SessionID:   "session-1",
		AccessToken: "token-1",
		ExpiresAt:   fixedNow + 3600,
	})

	resp, err := c.Do(context.Background(), api.Request{Method: http.MethodGet, Path: "/sites"})
	if err != nil {
		t.Fatalf("再試行で成功するはずです: %v", err)
	}
	if resp.Status != 200 {
		t.Errorf("status = %d", resp.Status)
	}
	if len(doer.sent) != 3 {
		t.Fatalf("送信回数 = %d, want 3", len(doer.sent))
	}
}

// セッションが無ければ 401 を握りつぶさず、そのままエラーにする。
func TestDoDoesNotRetryWithoutSession(t *testing.T) {
	doer := &fakeDoer{responses: []stub{
		jsonStub(401, `{"error":{"code":401,"message":"Invalid access token"}}`),
	}}
	c, _ := newClient(doer, &config.Credential{AccessToken: "token-1", ExpiresAt: fixedNow + 3600})

	_, err := c.Do(context.Background(), api.Request{Method: http.MethodGet, Path: "/sites"})
	apiErr, ok := api.AsError(err)
	if !ok {
		t.Fatalf("err = %v, want *api.Error", err)
	}
	if !apiErr.Unauthorized() || apiErr.Message != "Invalid access token" {
		t.Errorf("エラーの中身が違います: %+v", apiErr)
	}
	if apiErr.Hint() == "" {
		t.Error("401 には次の手が示されるべきです")
	}
	if len(doer.sent) != 1 {
		t.Errorf("送信回数 = %d, want 1", len(doer.sent))
	}
}

func TestDoWithoutCredentialsFails(t *testing.T) {
	doer := &fakeDoer{}
	c, _ := newClient(doer, &config.Credential{})
	_, err := c.Do(context.Background(), api.Request{Method: http.MethodGet, Path: "/sites"})
	if !errors.Is(err, api.ErrNotAuthenticated) {
		t.Fatalf("err = %v, want ErrNotAuthenticated", err)
	}
	if len(doer.sent) != 0 {
		t.Error("認証情報が無いのに送信しています")
	}
}

// 公開エンドポイントは認証なしで叩ける。
func TestDoNoAuthSkipsToken(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(200, `{"version":"9.1.0"}`)}}
	c, _ := newClient(doer, &config.Credential{})
	if _, err := c.Do(context.Background(), api.Request{Method: http.MethodGet, Path: "/version", NoAuth: true}); err != nil {
		t.Fatal(err)
	}
	if got := doer.last().Header.Get("X-MT-Authorization"); got != "" {
		t.Errorf("認証ヘッダが付いています: %q", got)
	}
}

func TestMethodOverride(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(200, `{}`)}}
	c, _ := newClient(doer, &config.Credential{AccessToken: "t", ExpiresAt: fixedNow + 3600})
	c.MethodOverride = true

	if _, err := c.Do(context.Background(), api.Request{
		Method: http.MethodDelete,
		Path:   "/sites/1/entries/2",
	}); err != nil {
		t.Fatal(err)
	}
	sent := doer.last()
	if sent.Method != http.MethodPost {
		t.Errorf("method = %s, want POST", sent.Method)
	}
	values, err := url.ParseQuery(sent.Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("__method") != "DELETE" {
		t.Errorf("__method = %q", values.Get("__method"))
	}
}

func TestFormEncoding(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(200, `{}`)}}
	c, _ := newClient(doer, &config.Credential{AccessToken: "t", ExpiresAt: fixedNow + 3600})

	if _, err := c.Do(context.Background(), api.Request{
		Method: http.MethodPost,
		Path:   "/sites/1/entries",
		Form:   url.Values{"entry": {`{"title":"日本語のタイトル"}`}, "publish": {"0"}},
	}); err != nil {
		t.Fatal(err)
	}
	sent := doer.last()
	if got := sent.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q", got)
	}
	values, err := url.ParseQuery(sent.Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("entry") != `{"title":"日本語のタイトル"}` {
		t.Errorf("entry = %q", values.Get("entry"))
	}
	if values.Get("publish") != "0" {
		t.Errorf("publish = %q", values.Get("publish"))
	}
}

func TestMultipartUpload(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(200, `{}`)}}
	c, _ := newClient(doer, &config.Credential{AccessToken: "t", ExpiresAt: fixedNow + 3600})

	if _, err := c.Do(context.Background(), api.Request{
		Method: http.MethodPost,
		Path:   "/sites/1/assets/upload",
		Form:   url.Values{"path": {"/2026"}},
		Upload: &api.Upload{Field: "file", Filename: "photo.jpg", Content: []byte("binary-data")},
	}); err != nil {
		t.Fatal(err)
	}

	sent := doer.last()
	mediaType, params, err := mime.ParseMediaType(sent.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	if mediaType != "multipart/form-data" {
		t.Fatalf("Content-Type = %q", mediaType)
	}

	reader := multipart.NewReader(strings.NewReader(sent.Body), params["boundary"])
	fields := map[string]string{}
	var filename string
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		buf, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		fields[part.FormName()] = string(buf)
		if part.FormName() == "file" {
			filename = part.FileName()
		}
	}
	if fields["file"] != "binary-data" {
		t.Errorf("file = %q", fields["file"])
	}
	if filename != "photo.jpg" {
		t.Errorf("filename = %q", filename)
	}
	if fields["path"] != "/2026" {
		t.Errorf("path = %q", fields["path"])
	}
}

func TestLogin(t *testing.T) {
	doer := &fakeDoer{responses: []stub{
		jsonStub(200, `{"accessToken":"token-1","sessionId":"session-1","expiresIn":3600,"remember":true}`),
	}}
	c, tokens := newClient(doer, nil)

	cred, err := c.Login(context.Background(), "alice", "web-service-password", true)
	if err != nil {
		t.Fatal(err)
	}

	sent := doer.last()
	if !strings.HasSuffix(sent.URL, "/v7/authentication") || sent.Method != http.MethodPost {
		t.Errorf("%s %s", sent.Method, sent.URL)
	}
	if sent.Header.Get("X-MT-Authorization") != "" {
		t.Error("サインイン時に認証ヘッダを付けています")
	}
	values, err := url.ParseQuery(sent.Body)
	if err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]string{
		"username": "alice",
		"password": "web-service-password",
		"clientId": "mtb",
		"remember": "1",
	} {
		if values.Get(k) != want {
			t.Errorf("%s = %q, want %q", k, values.Get(k), want)
		}
	}

	if cred.SessionID != "session-1" || cred.AccessToken != "token-1" {
		t.Errorf("cred = %+v", cred)
	}
	if cred.Username != "alice" {
		t.Errorf("Username = %q", cred.Username)
	}
	if tokens.Cred.AccessToken != "token-1" {
		t.Error("認証情報が保存されていません")
	}
	// パスワードは保存しない
	buf, err := json.Marshal(tokens.Cred)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(buf), "web-service-password") {
		t.Error("パスワードが保存されています")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	doer := &fakeDoer{}
	c, _ := newClient(doer, nil)

	// Webサービスパスワードの取り違えは最も多い詰まり方なので、
	// どちらの言語でも必ず触れること。
	for _, tc := range []struct {
		lang i18n.Lang
		want string
	}{
		{i18n.JA, "Webサービスパスワード"},
		{i18n.EN, "Web Services Password"},
	} {
		i18n.SetLang(tc.lang)
		doer.responses = []stub{jsonStub(401, `{"error":{"code":401,"message":"Invalid login"}}`)}
		_, err := c.Login(context.Background(), "alice", "wrong", true)
		if err == nil {
			t.Fatal("エラーになるはずです")
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: 助言が足りません: %v", tc.lang, err)
		}
	}
	i18n.SetLang(i18n.EN)
}

func TestRefreshWithoutSession(t *testing.T) {
	c, _ := newClient(&fakeDoer{}, &config.Credential{AccessToken: "t"})
	if _, err := c.Refresh(context.Background()); !errors.Is(err, api.ErrNotAuthenticated) {
		t.Fatalf("err = %v", err)
	}
}

func TestLogoutClearsCredentials(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(200, `{"status":"success"}`)}}
	c, tokens := newClient(doer, &config.Credential{SessionID: "session-1", AccessToken: "token-1"})

	if err := c.Logout(context.Background()); err != nil {
		t.Fatal(err)
	}
	sent := doer.last()
	if sent.Method != http.MethodDelete || !strings.HasSuffix(sent.URL, "/v7/authentication") {
		t.Errorf("%s %s", sent.Method, sent.URL)
	}
	if got := sent.Header.Get("X-MT-Authorization"); got != "MTAuth sessionId=session-1" {
		t.Errorf("セッションIDで認証すべきです: %q", got)
	}
	if tokens.Cred.SessionID != "" || tokens.Cred.AccessToken != "" {
		t.Errorf("認証情報が残っています: %+v", tokens.Cred)
	}
}

// すでにサーバ側で切れているセッションを消そうとしても、手元は片付ける。
func TestLogoutToleratesExpiredSession(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(401, `{"error":{"code":401,"message":"Invalid session"}}`)}}
	c, tokens := newClient(doer, &config.Credential{SessionID: "old", AccessToken: "old"})
	if err := c.Logout(context.Background()); err != nil {
		t.Fatalf("err = %v", err)
	}
	if tokens.Cred.SessionID != "" {
		t.Error("手元の認証情報が消えていません")
	}
}

func TestNonJSONResponsePassesThrough(t *testing.T) {
	doer := &fakeDoer{responses: []stub{{status: 200, body: "id,title\n1,hello\n", contentType: "text/csv"}}}
	c, _ := newClient(doer, &config.Credential{AccessToken: "t", ExpiresAt: fixedNow + 3600})

	resp, err := c.Do(context.Background(), api.Request{Method: http.MethodGet, Path: "/sites/1/entries/export"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.IsJSON() {
		t.Error("text/csv が JSON と判定されています")
	}
	if string(resp.Body) != "id,title\n1,hello\n" {
		t.Errorf("body = %q", resp.Body)
	}
}

func TestErrorMessageIncludesMethodAndURL(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(404, `{"error":{"code":404,"message":"Entry not found"}}`)}}
	c, _ := newClient(doer, &config.Credential{AccessToken: "t", ExpiresAt: fixedNow + 3600})

	_, err := c.Do(context.Background(), api.Request{Method: http.MethodGet, Path: "/sites/1/entries/999"})
	e, ok := api.AsError(err)
	if !ok {
		t.Fatalf("err = %v", err)
	}
	if !e.NotFound() {
		t.Error("NotFound が false です")
	}
	for _, want := range []string{"GET", "/v7/sites/1/entries/999", "Entry not found"} {
		if !strings.Contains(e.Error(), want) {
			t.Errorf("エラー文に %q がありません: %s", want, e.Error())
		}
	}
}

// トレース出力にパスワードを残さない。
func TestTraceRedactsPassword(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(200, `{"accessToken":"t","sessionId":"s","expiresIn":3600}`)}}
	trace := &strings.Builder{}
	c, _ := newClient(doer, nil)
	c.Trace = trace

	if _, err := c.Login(context.Background(), "alice", "secret-password", true); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(trace.String(), "secret-password") {
		t.Errorf("トレースにパスワードが出ています: %s", trace.String())
	}
	if !strings.Contains(trace.String(), "alice") {
		t.Errorf("トレースが出ていません: %s", trace.String())
	}
}

// Movable Type は multipart のパートに付いた Content-Type を、そのまま
// アセットの mimeType に保存する。何も付けないと画像が
// application/octet-stream として記録されてしまう。
func TestUploadSendsContentTypeByExtension(t *testing.T) {
	tests := map[string]string{
		"photo.jpg":   "image/jpeg",
		"logo.PNG":    "image/png",
		"small.gif":   "image/gif",
		"banner.webp": "image/webp",
		"diagram.svg": "image/svg+xml",
		"doc.pdf":     "application/pdf",
		"note.txt":    "text/plain",
		"data.bin":    "application/octet-stream",
		"noext":       "application/octet-stream",
	}
	for filename, want := range tests {
		doer := &fakeDoer{responses: []stub{jsonStub(200, `{}`)}}
		c, _ := newClient(doer, &config.Credential{AccessToken: "t", ExpiresAt: fixedNow + 3600})
		if _, err := c.Do(context.Background(), api.Request{
			Method: http.MethodPost,
			Path:   "/sites/1/assets/upload",
			Upload: &api.Upload{Field: "file", Filename: filename, Content: []byte("x")},
		}); err != nil {
			t.Fatal(err)
		}
		got := partContentType(t, doer.last(), "file")
		if got != want {
			t.Errorf("%s の Content-Type = %q, want %q", filename, got, want)
		}
	}
}

// 明示された Content-Type は拡張子より優先する。
func TestUploadHonoursExplicitContentType(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(200, `{}`)}}
	c, _ := newClient(doer, &config.Credential{AccessToken: "t", ExpiresAt: fixedNow + 3600})
	if _, err := c.Do(context.Background(), api.Request{
		Method: http.MethodPost, Path: "/sites/1/assets/upload",
		Upload: &api.Upload{Field: "file", Filename: "a.bin", Content: []byte("x"), ContentType: "image/png"},
	}); err != nil {
		t.Fatal(err)
	}
	if got := partContentType(t, doer.last(), "file"); got != "image/png" {
		t.Errorf("Content-Type = %q", got)
	}
}

// 日本語のファイル名がそのまま届くこと。
func TestUploadKeepsNonASCIIFilename(t *testing.T) {
	doer := &fakeDoer{responses: []stub{jsonStub(200, `{}`)}}
	c, _ := newClient(doer, &config.Credential{AccessToken: "t", ExpiresAt: fixedNow + 3600})
	if _, err := c.Do(context.Background(), api.Request{
		Method: http.MethodPost, Path: "/sites/1/assets/upload",
		Upload: &api.Upload{Field: "file", Filename: "会社ロゴ.png", Content: []byte("x")},
	}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(doer.last().Body, `filename="会社ロゴ.png"`) {
		t.Errorf("ファイル名が壊れています: %s", doer.last().Body[:200])
	}
}

// partContentType は multipart のうち name が一致するパートの Content-Type です。
func partContentType(t *testing.T, sent recorded, field string) string {
	t.Helper()
	_, params, err := mime.ParseMediaType(sent.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	r := multipart.NewReader(strings.NewReader(sent.Body), params["boundary"])
	for {
		part, err := r.NextPart()
		if errors.Is(err, io.EOF) {
			return ""
		}
		if err != nil {
			t.Fatal(err)
		}
		if part.FormName() == field {
			return part.Header.Get("Content-Type")
		}
	}
}
