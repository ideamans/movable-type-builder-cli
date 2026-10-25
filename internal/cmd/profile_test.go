package cmd

import (
	"strings"
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/config"
)

func TestProfileAddNormalizesURLAndBecomesCurrent(t *testing.T) {
	h := newHarness(t, withoutProfile())
	// リファレンスから /v7 ごと貼っても通ること
	out := h.mustRun("profile", "add", "prod", "--base-url", "https://example.com/cgi-bin/mt/mt-data-api.cgi/v7/")

	f, err := h.store.Load()
	if err != nil {
		t.Fatal(err)
	}
	p, ok := f.Profiles["prod"]
	if !ok {
		t.Fatal("プロファイルが保存されていません")
	}
	if p.BaseURL != "https://example.com/cgi-bin/mt/mt-data-api.cgi" || p.APIVersion != "v7" {
		t.Errorf("正規化されていません: %+v", p)
	}
	if f.Current != "prod" {
		t.Errorf("最初のプロファイルは既定になるべきです: %q", f.Current)
	}
	if !strings.Contains(out, "mtb auth login") {
		t.Errorf("次の手順が案内されていません: %s", out)
	}
}

func TestProfileAddRejectsDuplicate(t *testing.T) {
	h := newHarness(t)
	err := h.run("profile", "add", "prod", "--base-url", "https://other.example.com/api.cgi")
	if err == nil || !strings.Contains(err.Error(), "すでにあります") {
		t.Fatalf("err = %v", err)
	}
}

func TestProfileAddRequiresBaseURL(t *testing.T) {
	h := newHarness(t, withoutProfile())
	err := h.run("profile", "add", "prod")
	if err == nil || !strings.Contains(err.Error(), "--base-url") {
		t.Fatalf("err = %v", err)
	}
}

func TestProfileAddPromptsWhenInteractive(t *testing.T) {
	h := newHarness(t, withoutProfile())
	h.prompt.NotInteractive = false
	h.prompt.Lines = []string{"https://asked.example.com/mt-data-api.cgi"}
	h.prompt.Confirms = []bool{false} // サインインは別のテストで見る

	h.mustRun("profile", "add", "prod")

	f, err := h.store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if f.Profiles["prod"].BaseURL != "https://asked.example.com/mt-data-api.cgi" {
		t.Errorf("入力が使われていません: %+v", f.Profiles["prod"])
	}
}

// 聞くときは例を見せる。「URL」とだけ聞かれても、サイトのトップと
// mt-data-api.cgi までのどちらを入れるのか分からない。
func TestProfileAddPromptShowsExample(t *testing.T) {
	h := newHarness(t, withoutProfile())
	h.prompt.NotInteractive = false
	h.prompt.Lines = []string{"https://asked.example.com/mt-data-api.cgi"}
	h.prompt.Confirms = []bool{false} // サインインは別のテストで見る

	h.mustRun("profile", "add", "prod")

	asked := strings.Join(h.prompt.Asked, "\n")
	if !strings.Contains(asked, baseURLExample) {
		t.Errorf("聞くときに例が出ていません: %v", h.prompt.Asked)
	}
}

// 名前を省いたら、いまあるプロファイルを見せてから聞く。
//
// 一覧を先に出すのは名前を思いつくためだけではない。「もう prod がある」
// 「いま向いているのは stg だ」が同時に分かるのが効く。
func TestProfileAddWithoutNameListsAndAsks(t *testing.T) {
	h := newHarness(t)
	h.prompt.NotInteractive = false
	h.prompt.Lines = []string{"stg", "https://stg.example.com/mt-data-api.cgi"}
	h.prompt.Confirms = []bool{false}

	out := h.mustRun("profile", "add")

	if !strings.Contains(out, "prod") {
		t.Errorf("いまあるプロファイルが出ていません:\n%s", out)
	}
	f, err := h.store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if f.Profiles["stg"] == nil {
		t.Fatalf("聞いた名前で登録されていません: %+v", f.Profiles)
	}
	if f.Profiles["stg"].BaseURL != "https://stg.example.com/mt-data-api.cgi" {
		t.Errorf("URL が違います: %+v", f.Profiles["stg"])
	}
	// 既定は付け替えない。追加しただけで向き先が変わるのは危ない。
	if f.Current != "prod" {
		t.Errorf("既定が %q に変わっています", f.Current)
	}
}

// 対話できない場面（エージェント・CI）では、黙って止まらずに名前を要求する。
func TestProfileAddWithoutNameNeedsNameWhenNotInteractive(t *testing.T) {
	h := newHarness(t)
	err := h.run("profile", "add")
	if err == nil {
		t.Fatal("エラーになるはずです")
	}
	if !strings.Contains(err.Error(), "プロファイル名") {
		t.Errorf("何が足りないのか分かりません: %v", err)
	}
}

// 例は「見せるだけ」で、既定値ではない。Line の def に入れてしまうと、
// 何も打たずに Enter を押した人のプロファイルに example.com が登録され、
// しかも成功したように見える。
func TestProfileAddExampleIsNotADefault(t *testing.T) {
	h := newHarness(t, withoutProfile())
	h.prompt.NotInteractive = false
	h.prompt.Lines = []string{""} // Enter だけ

	err := h.run("profile", "add", "prod")
	if err == nil {
		t.Fatal("空入力はエラーになるはずです")
	}
	f, loadErr := h.store.Load()
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	if p, ok := f.Profiles["prod"]; ok {
		t.Errorf("空入力でプロファイルができています: %+v", p)
	}
}

func TestProfileUseSwitchesCurrent(t *testing.T) {
	h := newHarness(t)
	h.mustRun("profile", "add", "stg", "--base-url", "https://stg.example.com/api.cgi")
	h.mustRun("profile", "use", "stg")

	f, err := h.store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if f.Current != "stg" {
		t.Errorf("current = %q", f.Current)
	}
}

func TestProfileUseUnknownName(t *testing.T) {
	h := newHarness(t)
	err := h.run("profile", "use", "nope")
	if err == nil || !strings.Contains(err.Error(), "登録済み: prod") {
		t.Fatalf("err = %v", err)
	}
}

func TestProfileList(t *testing.T) {
	h := newHarness(t)
	h.mustRun("profile", "add", "stg", "--base-url", "https://stg.example.com/api.cgi")
	out := h.mustRun("profile", "list")

	if !strings.HasPrefix(lineWith(t, out, "prod"), "*") {
		t.Errorf("既定に印が付いていません:\n%s", out)
	}
	if strings.HasPrefix(lineWith(t, out, "stg"), "*") {
		t.Errorf("既定でないプロファイルに印が付いています:\n%s", out)
	}
	if !strings.Contains(out, "サインイン済み") {
		t.Errorf("認証状態が出ていません:\n%s", out)
	}
	if !strings.Contains(out, "未サインイン") {
		t.Errorf("未サインインのプロファイルが分かりません:\n%s", out)
	}
}

func TestProfileListWhenEmpty(t *testing.T) {
	h := newHarness(t, withoutProfile())
	out := h.mustRun("profile", "list")
	if !strings.Contains(out, "mtb profile add") {
		t.Errorf("出力 = %s", out)
	}
}

func TestProfileSetSiteID(t *testing.T) {
	h := newHarness(t)
	h.mustRun("profile", "set", "--site-id", "12")

	f, err := h.store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if f.Profiles["prod"].DefaultSiteID != "12" {
		t.Errorf("既定サイトIDが保存されていません: %+v", f.Profiles["prod"])
	}
}

func TestProfileSetWithoutFlags(t *testing.T) {
	h := newHarness(t)
	err := h.run("profile", "set")
	if err == nil || !strings.Contains(err.Error(), "--site-id") {
		t.Fatalf("err = %v", err)
	}
}

// 共通フラグを profile set が拾えること（ローカルに再定義していないため）。
func TestProfileSetUsesGlobalBaseURL(t *testing.T) {
	h := newHarness(t)
	h.mustRun("profile", "set", "--base-url", "https://moved.example.com/mt-data-api.cgi")

	f, err := h.store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if f.Profiles["prod"].BaseURL != "https://moved.example.com/mt-data-api.cgi" {
		t.Errorf("BaseURL = %q", f.Profiles["prod"].BaseURL)
	}
}

func TestProfileShow(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("profile", "show")
	for _, want := range []string{"prod", "https://example.com/mt/mt-data-api.cgi", "v7", "mem:config.json"} {
		if !strings.Contains(out, want) {
			t.Errorf("出力に %q がありません:\n%s", want, out)
		}
	}
}

func TestProfileRemoveNeedsConfirmation(t *testing.T) {
	h := newHarness(t)
	err := h.run("profile", "remove", "prod")
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("err = %v", err)
	}
	f, err := h.store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := f.Profiles["prod"]; !ok {
		t.Error("確認前に削除されています")
	}
}

func TestProfileRemoveAlsoRemovesCredential(t *testing.T) {
	h := newHarness(t)
	h.mustRun("profile", "add", "stg", "--base-url", "https://stg.example.com/api.cgi")
	h.mustRun("profile", "remove", "prod", "--yes")

	f, err := h.store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := f.Profiles["prod"]; ok {
		t.Error("プロファイルが残っています")
	}
	if f.Current != "stg" {
		t.Errorf("既定が引き継がれていません: %q", f.Current)
	}
	creds, err := h.store.LoadCredentials()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := creds["prod"]; ok {
		t.Error("認証情報が残っています")
	}
}

// --profile で接続先を切り替えられること。
func TestProfileFlagSwitchesEndpoint(t *testing.T) {
	h := newHarness(t)
	h.mustRun("profile", "add", "stg", "--base-url", "https://stg.example.com/mt-data-api.cgi")
	if err := h.store.PutCredential("stg", &config.Credential{AccessToken: "stg-token", ExpiresAt: testNow + 3600}); err != nil {
		t.Fatal(err)
	}
	h.reply(okJSON(`{}`))
	h.mustRun("sites", "list", "--profile", "stg")

	if !strings.HasPrefix(h.last().URL, "https://stg.example.com/mt-data-api.cgi/v7/") {
		t.Errorf("URL = %s", h.last().URL)
	}
	if got := h.last().Header.Get("X-MT-Authorization"); got != "MTAuth accessToken=stg-token" {
		t.Errorf("プロファイルごとの認証情報が使われていません: %q", got)
	}
}

// lineWith は substr を含む最初の行を返します。
func lineWith(t *testing.T, out, substr string) string {
	t.Helper()
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, substr) {
			return line
		}
	}
	t.Fatalf("%q を含む行がありません:\n%s", substr, out)
	return ""
}

// 登録の次はサインイン、というのが普通の流れ。ここで突き放すと打ち直しが
// 1回増えるだけで誰も得をしない。
func TestProfileAddOffersSignIn(t *testing.T) {
	h := newHarness(t, withoutProfile())
	h.prompt.NotInteractive = false
	h.prompt.Lines = []string{"alice"}
	h.prompt.Secrets = []string{"web-service-password"}
	h.prompt.Confirms = []bool{true}
	h.reply(okJSON(`{"accessToken":"token-1","sessionId":"session-1","expiresIn":3600}`))

	out := h.mustRun("profile", "add", "prod", "--base-url", "https://example.com/mt-data-api.cgi")

	if !contains(h.prompt.Asked, "このままサインインしますか？") {
		t.Errorf("サインインするか聞いていません: %v", h.prompt.Asked)
	}
	if len(h.doer.sent) == 0 {
		t.Fatal("サインインしていません")
	}
	if !strings.Contains(out, "サインインしました") {
		t.Errorf("出力 = %s", out)
	}
	// 追加したプロファイルに保存されること（--profile の指定に引きずられない）
	cred, err := h.store.Credential("prod")
	if err != nil {
		t.Fatal(err)
	}
	if cred.SessionID != "session-1" {
		t.Errorf("認証情報が prod に保存されていません: %+v", cred)
	}
}

// 断ったら、次にやることだけ伝えて終わる。
func TestProfileAddSignInDeclined(t *testing.T) {
	h := newHarness(t, withoutProfile())
	h.prompt.NotInteractive = false
	h.prompt.Confirms = []bool{false}

	out := h.mustRun("profile", "add", "prod", "--base-url", "https://example.com/mt-data-api.cgi")

	if len(h.doer.sent) != 0 {
		t.Error("断ったのにサインインしています")
	}
	if !strings.Contains(out, "mtb auth login --profile prod") {
		t.Errorf("次にやることを案内していません: %s", out)
	}
}

// 対話できない場面（エージェント・CI）では聞かない。黙って止まるのが一番まずい。
func TestProfileAddDoesNotAskWhenNotInteractive(t *testing.T) {
	h := newHarness(t, withoutProfile())

	out := h.mustRun("profile", "add", "prod", "--base-url", "https://example.com/mt-data-api.cgi")

	if len(h.prompt.Asked) != 0 {
		t.Errorf("聞いています: %v", h.prompt.Asked)
	}
	if !strings.Contains(out, "mtb auth login --profile prod") {
		t.Errorf("次にやることを案内していません: %s", out)
	}
}
