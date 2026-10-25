package cmd

import (
	"net/url"
	"strings"
	"testing"
)

func TestAuthLoginStoresSessionNotPassword(t *testing.T) {
	h := newHarness(t, withEnvVars(map[string]string{"MTB_PASSWORD": "web-service-password"}))
	if err := h.store.Save(mustProfile("prod", "https://example.com/mt/mt-data-api.cgi", "")); err != nil {
		t.Fatal(err)
	}
	h.reply(okJSON(`{"accessToken":"token-1","sessionId":"session-1","expiresIn":3600,"remember":true}`))

	out := h.mustRun("auth", "login", "--username", "alice")

	values, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("username") != "alice" || values.Get("password") != "web-service-password" {
		t.Errorf("送信内容が違います: %v", values)
	}
	if values.Get("remember") != "1" {
		t.Errorf("remember = %q（既定では永続セッションを作るべきです）", values.Get("remember"))
	}

	cred, err := h.store.Credential("prod")
	if err != nil {
		t.Fatal(err)
	}
	if cred.SessionID != "session-1" || cred.AccessToken != "token-1" {
		t.Errorf("認証情報が保存されていません: %+v", cred)
	}
	if cred.ExpiresAt != testNow+3600 {
		t.Errorf("失効時刻 = %d", cred.ExpiresAt)
	}

	// パスワードはどこにも残さない
	for _, name := range h.mem.Names() {
		buf, err := h.mem.Read(name)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(buf), "web-service-password") {
			t.Errorf("%s にパスワードが保存されています", name)
		}
	}

	// 次回の入力を省けるようユーザー名だけプロファイルに控える
	f, err := h.store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if f.Profiles["prod"].Username != "alice" {
		t.Errorf("ユーザー名が控えられていません: %+v", f.Profiles["prod"])
	}
	if !strings.Contains(out, "サインインしました") {
		t.Errorf("出力 = %s", out)
	}
}

func TestAuthLoginNonInteractiveWithoutPassword(t *testing.T) {
	h := newHarness(t)
	err := h.run("auth", "login", "--username", "alice")
	if err == nil {
		t.Fatal("パスワードが無いのに成功しています")
	}
	// 対話できない場面で止まったなら、渡し方まで教えないと先へ進めない。
	for _, want := range []string{"--password-stdin", "MTB_PASSWORD"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("%s の案内がありません: %v", want, err)
		}
	}
}

// パイプで渡せること。--password はプロセス一覧に残るので、
// 自動化するならこちらが本命になる。
func TestAuthLoginPasswordStdin(t *testing.T) {
	h := newHarness(t)
	// echo が付ける改行は落とす。
	h.env.In = strings.NewReader("web-service-password\n")
	h.reply(okJSON(`{"accessToken":"token-1","sessionId":"session-1","expiresIn":3600}`))

	h.mustRun("auth", "login", "--username", "alice", "--password-stdin")

	values, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("password") != "web-service-password" {
		t.Errorf("パスワード = %q", values.Get("password"))
	}
}

// 標準入力を読むと決めたなら、環境変数に落ちて別のパスワードで
// 通ってしまうより、空だと言って止まる方がよい。
func TestAuthLoginPasswordStdinEmpty(t *testing.T) {
	h := newHarness(t, withEnvVars(map[string]string{"MTB_PASSWORD": "from-env"}))
	h.env.In = strings.NewReader("")

	err := h.run("auth", "login", "--username", "alice", "--password-stdin")
	if err == nil {
		t.Fatal("空の標準入力で成功しています")
	}
	if len(h.doer.sent) != 0 {
		t.Error("送信しています")
	}
}

// 2つ渡されたらどちらを使うべきか決められない。黙って片方を選ばない。
func TestAuthLoginPasswordStdinConflictsWithFlag(t *testing.T) {
	h := newHarness(t)
	h.env.In = strings.NewReader("from-stdin\n")

	err := h.run("auth", "login", "--username", "alice", "--password", "from-flag", "--password-stdin")
	if err == nil {
		t.Fatal("両方指定して成功しています")
	}
	if code := ExitCode(err); code != ExitUsage {
		t.Errorf("exit = %d, want %d（呼び出し方の誤り）", code, ExitUsage)
	}
	if len(h.doer.sent) != 0 {
		t.Error("送信しています")
	}
}

func TestAuthLoginPromptsWhenInteractive(t *testing.T) {
	h := newHarness(t)
	h.prompt.NotInteractive = false
	h.prompt.Lines = []string{"bob"}
	h.prompt.Secrets = []string{"pw"}
	h.reply(okJSON(`{"accessToken":"t","sessionId":"s","expiresIn":3600}`))

	h.mustRun("auth", "login")

	values, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("username") != "bob" || values.Get("password") != "pw" {
		t.Errorf("対話入力が使われていません: %v", values)
	}
	// パスワードはエコーしない入力で聞くこと
	if !contains(h.prompt.Asked, "Webサービスパスワード") {
		t.Errorf("聞いた項目 = %v", h.prompt.Asked)
	}
}

func TestAuthLoginNoRemember(t *testing.T) {
	h := newHarness(t, withEnvVars(map[string]string{"MTB_PASSWORD": "pw"}))
	if err := h.store.Save(mustProfile("prod", "https://example.com/mt/mt-data-api.cgi", "")); err != nil {
		t.Fatal(err)
	}
	h.reply(okJSON(`{"accessToken":"t","sessionId":"s","expiresIn":3600}`))
	h.mustRun("auth", "login", "--username", "alice", "--no-remember")

	values, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("remember") != "0" {
		t.Errorf("remember = %q", values.Get("remember"))
	}
}

func TestAuthStatus(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("auth", "status")
	for _, want := range []string{"prod", "https://example.com/mt/mt-data-api.cgi/v7", "alice", "セッション: あり"} {
		if !strings.Contains(out, want) {
			t.Errorf("出力に %q がありません:\n%s", want, out)
		}
	}
	// トークンそのものは表示しない
	if strings.Contains(out, "token-1") {
		t.Errorf("アクセストークンが表示されています:\n%s", out)
	}
}

func TestAuthStatusWhenLoggedOut(t *testing.T) {
	h := newHarness(t)
	if err := h.store.DeleteCredential("prod"); err != nil {
		t.Fatal(err)
	}
	out := h.mustRun("auth", "status")
	if !strings.Contains(out, "未サインイン") {
		t.Errorf("出力 = %s", out)
	}
}

func TestAuthStatusWithEnvToken(t *testing.T) {
	h := newHarness(t, withEnvVars(map[string]string{
		"MTB_BASE_URL":     "https://ci.example.com/mt-data-api.cgi",
		"MTB_ACCESS_TOKEN": "ci-token",
	}), withoutProfile())
	out := h.mustRun("auth", "status")
	if !strings.Contains(out, "MTB_ACCESS_TOKEN") {
		t.Errorf("出力 = %s", out)
	}
}

// 環境変数のトークンだけで叩けること（CI やコンテナからの経路）。
func TestEnvAccessTokenIsUsedWithoutConfigFile(t *testing.T) {
	h := newHarness(t, withEnvVars(map[string]string{
		"MTB_BASE_URL":     "https://ci.example.com/mt-data-api.cgi",
		"MTB_ACCESS_TOKEN": "ci-token",
	}), withoutProfile())
	h.reply(okJSON(`{"totalResults":0,"items":[]}`))
	h.mustRun("sites", "entries", "list", "--site-id", "1")

	if got := h.last().Header.Get("X-MT-Authorization"); got != "MTAuth accessToken=ci-token" {
		t.Errorf("認証ヘッダ = %q", got)
	}
	// 設定ファイルには何も書かない
	for _, name := range h.mem.Names() {
		if name == "credentials.json" {
			t.Error("環境変数のトークンが保存されています")
		}
	}
}

func TestAuthRefresh(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"accessToken":"token-2","expiresIn":7200}`))
	out := h.mustRun("auth", "refresh")

	if !strings.HasSuffix(pathOf(t, h.last().URL), "/v7/token") {
		t.Errorf("URL = %s", h.last().URL)
	}
	cred, err := h.store.Credential("prod")
	if err != nil {
		t.Fatal(err)
	}
	if cred.AccessToken != "token-2" || cred.ExpiresAt != testNow+7200 {
		t.Errorf("更新されていません: %+v", cred)
	}
	if !strings.Contains(out, "更新しました") {
		t.Errorf("出力 = %s", out)
	}
}

func TestAuthLogoutClearsStoredCredential(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"status":"success"}`))
	h.mustRun("auth", "logout")

	if h.last().Method != "DELETE" {
		t.Errorf("method = %s", h.last().Method)
	}
	cred, err := h.store.Credential("prod")
	if err != nil {
		t.Fatal(err)
	}
	if cred.SessionID != "" || cred.AccessToken != "" {
		t.Errorf("認証情報が残っています: %+v", cred)
	}
}

// 期限切れのトークンは、コマンド実行時に黙って取り直される。
func TestExpiredTokenIsRefreshedTransparently(t *testing.T) {
	h := newHarness(t)
	if err := h.store.PutCredential("prod", expiredCredential()); err != nil {
		t.Fatal(err)
	}
	h.reply(
		okJSON(`{"accessToken":"token-2","expiresIn":3600}`),
		okJSON(`{"totalResults":0,"items":[]}`),
	)
	h.mustRun("sites", "entries", "list", "--site-id", "1")

	if len(h.doer.sent) != 2 {
		t.Fatalf("送信回数 = %d, want 2", len(h.doer.sent))
	}
	if got := h.last().Header.Get("X-MT-Authorization"); got != "MTAuth accessToken=token-2" {
		t.Errorf("新しいトークンで送っていません: %q", got)
	}
	// 取り直したトークンは保存され、次回の更新を省ける
	cred, err := h.store.Credential("prod")
	if err != nil {
		t.Fatal(err)
	}
	if cred.AccessToken != "token-2" {
		t.Errorf("保存されていません: %+v", cred)
	}
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if strings.Contains(s, want) {
			return true
		}
	}
	return false
}
