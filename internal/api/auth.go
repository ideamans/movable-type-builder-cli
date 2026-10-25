package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ideamans/movable-type-builder-cli/internal/config"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

// authHeader は MT がアクセストークンとセッションIDの両方を受け取るヘッダです。
const authHeader = "X-MT-Authorization"

// tokenResponse は /authentication と /token の応答です。
type tokenResponse struct {
	AccessToken string `json:"accessToken"`
	SessionID   string `json:"sessionId"`
	ExpiresIn   int64  `json:"expiresIn"`
	Remember    bool   `json:"remember"`
}

// Login はユーザー名とパスワードでセッションとアクセストークンを作ります。
//
// remember を立てると MT 側が永続セッションを作るので、セッションIDだけを
// 保存しておけば以後パスワードなしでアクセストークンを取り直せます。
// これが「気軽な認証」の実体で、CLI が保存するのはトークンとセッションIDだけ、
// パスワードは保存しません。
func (c *Client) Login(ctx context.Context, username, password string, remember bool) (*config.Credential, error) {
	if username == "" || password == "" {
		return nil, errors.New(i18n.T("ユーザー名とパスワードが要ります", "a username and a password are required"))
	}
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	form.Set("clientId", c.ClientID)
	if remember {
		form.Set("remember", "1")
	} else {
		form.Set("remember", "0")
	}

	resp, err := c.send(ctx, Request{
		Method: http.MethodPost,
		Path:   "/authentication",
		Form:   form,
		NoAuth: true,
	}, true)
	if err != nil {
		return nil, err
	}
	if resp.Status >= 400 {
		e := newError(http.MethodPost, c.URL("/authentication", nil), resp)
		if e.Unauthorized() {
			return nil, fmt.Errorf(i18n.T(
				"%w（Webサービスパスワードを使っているか確認してください。MT のサインインパスワードとは別です）",
				"%w (check that you are using the Web Services Password, which is not the Movable Type sign-in password)"), e)
		}
		return nil, e
	}

	var tr tokenResponse
	if err := decodeJSON(resp, &tr); err != nil {
		return nil, err
	}
	if tr.AccessToken == "" {
		return nil, errors.New(i18n.T(
			"アクセストークンが返りませんでした。clientId と MT のバージョンを確認してください",
			"no access token was returned; check the clientId and the Movable Type version"))
	}

	cred := &config.Credential{
		SessionID:   tr.SessionID,
		AccessToken: tr.AccessToken,
		Remember:    tr.Remember || remember,
		Username:    username,
	}
	if tr.ExpiresIn > 0 {
		cred.ExpiresAt = c.now().Add(time.Duration(tr.ExpiresIn) * time.Second).Unix()
	}
	if err := c.tokens().Save(cred); err != nil {
		return nil, err
	}
	return cred, nil
}

// Refresh はセッションIDから新しいアクセストークンを取り直します。
func (c *Client) Refresh(ctx context.Context) (*config.Credential, error) {
	cred, err := c.tokens().Credential()
	if err != nil {
		return nil, err
	}
	if cred.SessionID == "" {
		return nil, ErrNotAuthenticated
	}

	form := url.Values{}
	form.Set("clientId", c.ClientID)

	req, err := c.newSessionRequest(ctx, http.MethodPost, "/token", form, cred.SessionID)
	if err != nil {
		return nil, err
	}
	resp, err := c.roundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.Status >= 400 {
		e := newError(http.MethodPost, c.URL("/token", nil), resp)
		if e.Unauthorized() {
			return nil, fmt.Errorf(i18n.T(
				"セッションが失効しています。mtb auth login をやり直してください: %w",
				"the session has expired; sign in again with mtb auth login: %w"), e)
		}
		return nil, e
	}

	var tr tokenResponse
	if err := decodeJSON(resp, &tr); err != nil {
		return nil, err
	}
	if tr.AccessToken == "" {
		return nil, errors.New(i18n.T("アクセストークンが返りませんでした", "no access token was returned"))
	}

	updated := *cred
	updated.AccessToken = tr.AccessToken
	if tr.SessionID != "" {
		updated.SessionID = tr.SessionID
	}
	updated.ExpiresAt = 0
	if tr.ExpiresIn > 0 {
		updated.ExpiresAt = c.now().Add(time.Duration(tr.ExpiresIn) * time.Second).Unix()
	}
	if err := c.tokens().Save(&updated); err != nil {
		return nil, err
	}
	return &updated, nil
}

// Logout はセッションごと無効にします。他の端末で取ったアクセストークンも
// まとめて無効になります。
func (c *Client) Logout(ctx context.Context) error {
	cred, err := c.tokens().Credential()
	if err != nil {
		return err
	}
	if cred.SessionID == "" && cred.AccessToken == "" {
		return nil
	}

	if cred.SessionID != "" {
		form := url.Values{}
		form.Set("clientId", c.ClientID)
		req, err := c.newSessionRequest(ctx, http.MethodDelete, "/authentication", form, cred.SessionID)
		if err != nil {
			return err
		}
		resp, err := c.roundTrip(req)
		if err != nil {
			return err
		}
		// 既に切れているセッションを消そうとした 401 は成功と同じ扱いにします。
		if resp.Status >= 400 && resp.Status != http.StatusUnauthorized {
			return newError(http.MethodDelete, c.URL("/authentication", nil), resp)
		}
	}
	return c.tokens().Save(&config.Credential{})
}

// newSessionRequest は「アクセストークンではなくセッションID」で認証する
// リクエストを組み立てます。トークンの取り直しとサインアウトだけがこの形です。
func (c *Client) newSessionRequest(ctx context.Context, method, path string, form url.Values, sessionID string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.URL(path, nil), formBody(form))
	if err != nil {
		return nil, err
	}
	if len(form) > 0 {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Accept", "application/json")
	c.decorate(req)
	req.Header.Set(authHeader, "MTAuth sessionId="+sessionID)
	return req, nil
}
