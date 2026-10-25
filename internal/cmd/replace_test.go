package cmd

import (
	"strings"
	"testing"
)

// 同名のアセットを消してから送ること。
//
// MT はアセットを上書きできない（409）。デザインの差し替えは日常の操作
// なので、これが無いと利用者が毎回「消してから送る」を書くことになる。
func TestReplaceDeletesExistingAssetFirst(t *testing.T) {
	h := newHarness(t)
	h.files["dist/theme.css"] = "body{}"
	h.reply(
		okJSON(`{"url":"http://example.com/blog/"}`),                                           // サイトのURL
		okJSON(`{"items":[{"id":38,"url":"http://example.com/blog/theme-assets/theme.css"}]}`), // 検索
		okJSON(`{"id":"38"}`), // 削除
		okJSON(`{"id":"41"}`), // 送信
	)

	h.mustRun("assets", "upload", "--site-id", "4", "--file", "dist/theme.css",
		"--path", "/theme-assets", "--replace", "--yes")

	if len(h.doer.sent) != 4 {
		t.Fatalf("送信 = %d 件: %+v", len(h.doer.sent), h.doer.sent)
	}
	if got := h.doer.sent[2]; got.Method != "DELETE" || !strings.HasSuffix(got.URL, "/sites/4/assets/38") {
		t.Errorf("削除していません: %s %s", got.Method, got.URL)
	}
	if got := h.doer.sent[3]; got.Method != "POST" || !strings.Contains(got.URL, "/assets/upload") {
		t.Errorf("送信していません: %s %s", got.Method, got.URL)
	}
}

// 送り先が違う同名ファイルは消さない。ファイル名だけで一致を見ると、
// 別フォルダの同名ファイルを巻き込む。
func TestReplaceMatchesTheDestinationNotJustTheName(t *testing.T) {
	h := newHarness(t)
	h.files["dist/theme.css"] = "body{}"
	h.reply(
		okJSON(`{"url":"http://example.com/blog/"}`),
		okJSON(`{"items":[{"id":38,"url":"http://example.com/blog/other-folder/theme.css"}]}`),
		okJSON(`{"id":"41"}`),
	)

	h.mustRun("assets", "upload", "--site-id", "4", "--file", "dist/theme.css",
		"--path", "/theme-assets", "--replace", "--yes")

	for _, sent := range h.doer.sent {
		if sent.Method == "DELETE" {
			t.Errorf("別の場所にある同名ファイルを消しています: %s", sent.URL)
		}
	}
}

// 消すのは破壊的な操作なので、対話できない場面では --yes を要求する。
func TestReplaceNeedsConfirmation(t *testing.T) {
	h := newHarness(t)
	h.files["dist/theme.css"] = "body{}"

	err := h.run("assets", "upload", "--site-id", "4", "--file", "dist/theme.css",
		"--path", "/theme-assets", "--replace")
	if err == nil {
		t.Fatal("確認なしで進んでいます")
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Errorf("--yes を案内すべきです: %v", err)
	}
	if len(h.doer.sent) != 0 {
		t.Error("送信しています")
	}
}

// id は数値で返ることも文字列で返ることもある（MT はここが揃っていない）。
func TestReplaceAcceptsNumericAndStringIDs(t *testing.T) {
	for _, id := range []string{`38`, `"38"`} {
		h := newHarness(t)
		h.files["dist/theme.css"] = "body{}"
		h.reply(
			okJSON(`{"url":"http://example.com/blog/"}`),
			okJSON(`{"items":[{"id":`+id+`,"url":"http://example.com/blog/theme-assets/theme.css"}]}`),
			okJSON(`{}`),
			okJSON(`{}`),
		)
		h.mustRun("assets", "upload", "--site-id", "4", "--file", "dist/theme.css",
			"--path", "/theme-assets", "--replace", "--yes")

		if got := h.doer.sent[2]; !strings.HasSuffix(got.URL, "/assets/38") {
			t.Errorf("id=%s のとき削除先が違います: %s", id, got.URL)
		}
	}
}

// テーマの適用はテンプレートの総入れ替え。refresh-templates と同じ重さなのに
// 確認が入っていなかった。
func TestThemeApplyIsGuarded(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"name":"アイデアマンズ ブログ"}`))

	err := h.run("sites", "themes", "apply", "4", "some-theme", "--yes")
	if err == nil {
		t.Fatal("--yes だけで通っています")
	}
	if !strings.Contains(err.Error(), "--confirm") {
		t.Errorf("対象の名前を要求すべきです: %v", err)
	}
	for _, sent := range h.doer.sent {
		if sent.Method == "POST" {
			t.Errorf("確認前に送信しています: %s", sent.URL)
		}
	}
}
