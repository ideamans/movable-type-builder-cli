package cmd

import (
	"strings"
	"testing"
)

// TestEnglishOutput は英語環境での文言を確かめます。日本語側は他のテストが
// 見ているので、ここでは「切り替わること」と「訳し漏れが混ざらないこと」を見ます。
func TestEnglishOutput(t *testing.T) {
	h := newHarness(t, withLang("en"))
	out := h.mustRun("auth", "status")

	for _, want := range []string{"Profile: prod", "Endpoint: ", "Session: yes", "Access token: "} {
		if !strings.Contains(out, want) {
			t.Errorf("出力に %q がありません:\n%s", want, out)
		}
	}
	if hasJapanese(out) {
		t.Errorf("英語の出力に日本語が混ざっています:\n%s", out)
	}
}

// TestLangFlagBeatsEnvironment は --lang が環境変数より強いことを確かめます。
// ヘルプの文言はコマンドツリーを組む時点で決まるので、フラグ解析を待たずに
// 先読みできていないと日本語のままになります。
func TestLangFlagBeatsEnvironment(t *testing.T) {
	h := newHarness(t, withLang("ja"))
	if err := h.run("--lang", "en", "profile", "show"); err != nil {
		t.Fatal(err)
	}
	if out := h.out.String(); !strings.Contains(out, "Base URL") {
		t.Errorf("--lang en が効いていません:\n%s", out)
	}

	h2 := newHarness(t, withLang("en"))
	if err := h2.run("--lang", "ja", "profile", "show"); err != nil {
		t.Fatal(err)
	}
	if out := h2.out.String(); !strings.Contains(out, "ベースURL") {
		t.Errorf("--lang ja が効いていません:\n%s", out)
	}
}

// TestHelpFollowsLanguage はヘルプ本文とフラグの説明が切り替わることを見ます。
func TestHelpFollowsLanguage(t *testing.T) {
	for _, tc := range []struct{ lang, want string }{
		{"ja", "サインイン"},
		{"en", "Web Services Password"},
	} {
		h := newHarness(t, withLang(tc.lang))
		if err := h.run("auth", "login", "--help"); err != nil {
			t.Fatal(err)
		}
		if out := h.out.String(); !strings.Contains(out, tc.want) {
			t.Errorf("%s のヘルプに %q がありません:\n%s", tc.lang, tc.want, out)
		}
	}
}

// TestErrorsFollowLanguage は、パッケージ変数として先に作られるエラーでも
// 表示時の言語で出ることを確かめます（api.ErrNotAuthenticated など）。
func TestErrorsFollowLanguage(t *testing.T) {
	for _, tc := range []struct{ lang, want string }{
		{"ja", "認証されていません"},
		{"en", "not authenticated"},
	} {
		h := newHarness(t, withLang(tc.lang), withoutProfile())
		if err := h.store.Save(mustProfile("prod", "https://example.com/mt/mt-data-api.cgi", "")); err != nil {
			t.Fatal(err)
		}
		err := h.run("sites", "entries", "list", "--site-id", "1")
		if err == nil {
			t.Fatal("エラーになるはずです")
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s のエラーに %q がありません: %v", tc.lang, tc.want, err)
		}
	}
}

func hasJapanese(s string) bool {
	for _, r := range s {
		if (r >= 0x3040 && r <= 0x30ff) || (r >= 0x4e00 && r <= 0x9fff) {
			return true
		}
	}
	return false
}
