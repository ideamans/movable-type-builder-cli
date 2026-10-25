// Package api は Movable Type Data API への1接続分のクライアントです。
//
// 外部依存は Doer（HTTP送信）と TokenStore（認証情報の出し入れ）と Now
// （時刻）の3つだけで、いずれも差し替えられます。テストは net/http/httptest
// すら使わずに、応答を組み立てる Doer を渡して検証します。
package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// Doer は HTTP リクエストの送信です。*http.Client がそのまま満たします。
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// TokenStore は1プロファイル分の認証情報の出し入れです。
type TokenStore interface {
	Credential() (*config.Credential, error)
	Save(c *config.Credential) error
}

// MemoryTokenStore はメモリ上に持つだけの TokenStore です。
// 認証情報を保存しない使い方（MTB_ACCESS_TOKEN だけで叩く）とテストで使います.
type MemoryTokenStore struct {
	Cred *config.Credential
}

// Credential は保持している認証情報を返します。
func (m *MemoryTokenStore) Credential() (*config.Credential, error) {
	if m.Cred == nil {
		m.Cred = &config.Credential{}
	}
	return m.Cred, nil
}

// Save は認証情報を差し替えます。
func (m *MemoryTokenStore) Save(c *config.Credential) error {
	m.Cred = c
	return nil
}

// Client は1つの接続先に対するクライアントです。
type Client struct {
	// BaseURL は mt-data-api.cgi までの URL です（末尾に /vN を含みません）。
	BaseURL string
	// APIVersion は "v7" などのエンドポイントバージョンです。
	APIVersion string
	// ClientID は認証時に名乗るIDです。
	ClientID string
	// HTTP は送信手段です。
	HTTP Doer
	// Tokens は認証情報の置き場です。nil ならメモリ上に持ちます。
	Tokens TokenStore
	// Now は現在時刻です。nil なら time.Now です。
	Now func() time.Time
	// MethodOverride が真なら PUT / DELETE を POST + __method で送ります。
	// PUT と DELETE を通さない前段（一部のレンタルサーバや WAF）向けの逃げ道で、
	// MT 側が公式に受け付ける形です。
	MethodOverride bool
	// Trace が非 nil なら、送信したリクエストの要約を書き出します。
	Trace io.Writer
	// UserAgent は名乗る User-Agent です（mtb/<版>）。空なら Go の既定のまま。
	// 既定の Go-http-client では、Web サーバのログから mtb の呼び出しを見分けられません。
	UserAgent string
	// AuditNote は「何の作業か」の説明です。X-MTB-Audit-Note ヘッダで送り、
	// サーバ側で監査ログを取っていれば（MtbPro など）一緒に残ります。
	AuditNote string
}

// AuditNoteHeader は作業の説明を送るヘッダです。
const AuditNoteHeader = "X-MTB-Audit-Note"

// decorate は全リクエストに共通のヘッダを付けます。
//
// 作業の説明は UTF-8 をパーセントエンコードして送ります。HTTP のヘッダは
// ASCII しか確実に通らず、日本語をそのまま入れると途中で化けるか落とされます。
func (c *Client) decorate(req *http.Request) {
	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}
	if c.AuditNote != "" {
		req.Header.Set(AuditNoteHeader, url.PathEscape(c.AuditNote))
	}
}

// NewHTTPClient は実際に通信する Doer を作ります。
func NewHTTPClient(insecure bool, timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // 検証中サーバ向けに明示的に選ぶ逃げ道
	}
	return &http.Client{Transport: transport, Timeout: timeout}
}

// Upload はファイル1つの multipart 送信です。
type Upload struct {
	Field    string
	Filename string
	Content  []byte
	// ContentType は省略するとファイル名の拡張子から決めます。
	// Movable Type はこの値をアセットの mimeType にそのまま保存するため、
	// 何も付けないと画像が application/octet-stream として記録されます。
	ContentType string
}

// uploadTypes は拡張子と Content-Type の対応です。mime.TypeByExtension は
// システムの /etc/mime.types を見るので環境で結果が変わります。Web で扱う
// ものだけは、どの機械でも同じになるようここで決め打ちします。
var uploadTypes = map[string]string{
	".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
	".gif": "image/gif", ".webp": "image/webp", ".avif": "image/avif",
	".svg": "image/svg+xml", ".ico": "image/vnd.microsoft.icon",
	".bmp": "image/bmp", ".tif": "image/tiff", ".tiff": "image/tiff",
	".pdf": "application/pdf", ".txt": "text/plain", ".csv": "text/csv",
	".json": "application/json", ".xml": "application/xml",
	".zip": "application/zip", ".mp4": "video/mp4", ".mp3": "audio/mpeg",
}

// contentType はファイル名から Content-Type を決めます。
func (u *Upload) contentType() string {
	if u.ContentType != "" {
		return u.ContentType
	}
	ext := strings.ToLower(filepath.Ext(u.Filename))
	if t, ok := uploadTypes[ext]; ok {
		return t
	}
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return "application/octet-stream"
}

// Request は API 呼び出し1回分です。
type Request struct {
	Method string
	// Path は "/sites/1/entries" のようにバージョンより後ろの部分です。
	Path  string
	Query url.Values
	Form  url.Values
	// Upload があれば multipart/form-data で送ります。Form の値も同送します。
	Upload *Upload
	// NoAuth を立てると認証ヘッダを付けず、トークンの更新も試みません。
	NoAuth bool
}

// Response は API の応答です。本文はそのまま持ちます。JSON でない応答
// （エクスポート、バックアップ、プレビューHTML）があるためです。
type Response struct {
	Status int
	Header http.Header
	Body   []byte
}

// ContentType は Content-Type のメディアタイプ部分です。
func (r *Response) ContentType() string {
	ct := r.Header.Get("Content-Type")
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	return strings.TrimSpace(ct)
}

// IsJSON は本文が JSON かどうかです。
func (r *Response) IsJSON() bool {
	ct := r.ContentType()
	return ct == "application/json" || strings.HasSuffix(ct, "+json")
}

func (c *Client) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func (c *Client) tokens() TokenStore {
	if c.Tokens == nil {
		c.Tokens = &MemoryTokenStore{}
	}
	return c.Tokens
}

// Endpoint はバージョンまで含めた URL の前半です。
func (c *Client) Endpoint() string {
	return strings.TrimRight(c.BaseURL, "/") + "/" + c.APIVersion
}

// URL は path に対応する完全な URL を返します。
func (c *Client) URL(path string, query url.Values) string {
	u := c.Endpoint() + "/" + strings.TrimPrefix(path, "/")
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	return u
}

// Do はリクエストを送ります。認証が要るなら、必要に応じてアクセストークンを
// 取り直してから送り、それでも 401 なら1度だけ取り直して送り直します。
//
// HTTP のステータスが 400 以上のときは *Error を返しますが、Response も
// 一緒に返すので、呼び出し側は本文をそのまま出すこともできます。
func (c *Client) Do(ctx context.Context, req Request) (*Response, error) {
	if !req.NoAuth {
		if err := c.ensureToken(ctx); err != nil {
			return nil, err
		}
	}

	resp, err := c.send(ctx, req, req.NoAuth)
	if err != nil {
		return nil, err
	}

	if resp.Status == http.StatusUnauthorized && !req.NoAuth {
		cred, cerr := c.tokens().Credential()
		if cerr == nil && cred.SessionID != "" {
			if _, rerr := c.Refresh(ctx); rerr == nil {
				resp, err = c.send(ctx, req, false)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if resp.Status >= 400 {
		return resp, newError(req.Method, c.URL(req.Path, req.Query), resp)
	}
	return resp, nil
}

// ensureToken はアクセストークンが失効していればセッションから取り直します。
func (c *Client) ensureToken(ctx context.Context) error {
	cred, err := c.tokens().Credential()
	if err != nil {
		return err
	}
	if !cred.Expired(c.now()) {
		return nil
	}
	if cred.SessionID == "" {
		if cred.AccessToken != "" {
			// 期限だけ分からないトークンを渡されている場合はそのまま使う。
			return nil
		}
		return ErrNotAuthenticated
	}
	_, err = c.Refresh(ctx)
	return err
}

func (c *Client) send(ctx context.Context, req Request, noAuth bool) (*Response, error) {
	method := strings.ToUpper(req.Method)
	form := url.Values{}
	for k, vs := range req.Form {
		form[k] = append([]string(nil), vs...)
	}

	// MT は PUT / DELETE を POST + __method で受け付けます。
	if c.MethodOverride && (method == http.MethodPut || method == http.MethodDelete) {
		form.Set("__method", method)
		method = http.MethodPost
	}

	var (
		body        io.Reader
		contentType string
	)
	switch {
	case req.Upload != nil:
		buf := &bytes.Buffer{}
		w := multipart.NewWriter(buf)
		for k, vs := range form {
			for _, v := range vs {
				if err := w.WriteField(k, v); err != nil {
					return nil, err
				}
			}
		}
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name=%q; filename=%q`,
				escapeMultipart(req.Upload.Field), escapeMultipart(req.Upload.Filename)))
		header.Set("Content-Type", req.Upload.contentType())
		part, err := w.CreatePart(header)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(req.Upload.Content); err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		body, contentType = buf, w.FormDataContentType()
	case len(form) > 0:
		body = strings.NewReader(form.Encode())
		contentType = "application/x-www-form-urlencoded"
	}

	httpReq, err := http.NewRequestWithContext(ctx, method, c.URL(req.Path, req.Query), body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		httpReq.Header.Set("Content-Type", contentType)
	}
	httpReq.Header.Set("Accept", "application/json")
	c.decorate(httpReq)
	if !noAuth {
		cred, err := c.tokens().Credential()
		if err != nil {
			return nil, err
		}
		if cred.AccessToken != "" {
			httpReq.Header.Set(authHeader, "MTAuth accessToken="+cred.AccessToken)
		}
	}

	if c.Trace != nil && len(form) > 0 {
		fmt.Fprintf(c.Trace, "> body: %s\n", redactForm(form).Encode())
	}
	return c.roundTrip(httpReq)
}

// roundTrip は組み立て済みのリクエストを送り、本文を読み切って返します。
func (c *Client) roundTrip(httpReq *http.Request) (*Response, error) {
	if c.Trace != nil {
		fmt.Fprintf(c.Trace, "> %s %s\n", httpReq.Method, httpReq.URL.String())
	}

	httpResp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", httpReq.Method, httpReq.URL.Redacted(), err)
	}
	defer httpResp.Body.Close()

	buf, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf(i18n.T("応答を読めません: %w", "cannot read the response: %w"), err)
	}
	if c.Trace != nil {
		fmt.Fprintf(c.Trace, "< %d %s (%d bytes)\n", httpResp.StatusCode, httpResp.Header.Get("Content-Type"), len(buf))
	}
	return &Response{Status: httpResp.StatusCode, Header: httpResp.Header, Body: buf}, nil
}

// formBody はフォーム値の本文を返します。空なら nil（本文なし）です。
func formBody(form url.Values) io.Reader {
	if len(form) == 0 {
		return nil
	}
	return strings.NewReader(form.Encode())
}

// escapeMultipart は multipart のヘッダに入れられない文字を逃がします。
// mime/multipart の CreateFormFile と同じ扱いにしてあります。
func escapeMultipart(s string) string {
	return multipartEscaper.Replace(s)
}

var multipartEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

// redactForm はトレース出力からパスワードを伏せます。
func redactForm(form url.Values) url.Values {
	out := url.Values{}
	for k, vs := range form {
		if k == "password" {
			out.Set(k, "********")
			continue
		}
		out[k] = vs
	}
	return out
}

// decodeJSON は応答本文を v に読み込みます。
func decodeJSON(resp *Response, v any) error {
	if err := json.Unmarshal(resp.Body, v); err != nil {
		return fmt.Errorf(i18n.T(
			"応答を JSON として読めません（%d %s）: %w",
			"the response is not valid JSON (%d %s): %w"), resp.Status, resp.ContentType(), err)
	}
	return nil
}
