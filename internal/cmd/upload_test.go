package cmd

import (
	"net/url"
	"strings"
	"testing"
)

// dist を1つ用意します。下の階層と隠しファイルを混ぜてあります。
func withThemeFiles() harnessOption {
	return func(h *harness) {
		h.files["dist/theme.css"] = "body{color:#111}"
		h.files["dist/logo.png"] = "PNG"
		h.files["dist/icons/menu.svg"] = "<svg/>"
		h.files["dist/icons/close.svg"] = "<svg/>"
		h.files["dist/.DS_Store"] = "junk"
		h.files["dist/.git/config"] = "junk"
	}
}

// ディレクトリを丸ごと送れること。下の階層は送り先にも写す。
func TestUploadDirMirrorsSubdirectories(t *testing.T) {
	h := newHarness(t, withThemeFiles())
	h.reply(okJSON(`{"id":"1"}`), okJSON(`{"id":"2"}`), okJSON(`{"id":"3"}`), okJSON(`{"id":"4"}`))

	out := h.mustRun("assets", "upload", "--site-id", "1", "--dir", "dist", "--path", "/theme-images")

	if len(h.doer.sent) != 4 {
		t.Fatalf("送信 = %d 件（隠しファイルを除いた4件のはず）", len(h.doer.sent))
	}
	// 送り先は --path の下にローカルの構成をそのまま写す。
	want := map[string]string{
		"close.svg": "/theme-images/icons",
		"menu.svg":  "/theme-images/icons",
		"logo.png":  "/theme-images",
		"theme.css": "/theme-images",
	}
	for _, sent := range h.doer.sent {
		name, path := multipartFilename(t, sent.Body), multipartField(t, sent.Body, "path")
		if want[name] == "" {
			t.Errorf("送るはずのないファイル: %q", name)
			continue
		}
		if path != want[name] {
			t.Errorf("%s の送り先 = %q, want %q", name, path, want[name])
		}
	}
	// 応答は --all と同じ形にまとめる。
	if !strings.Contains(out, `"totalResults": 4`) {
		t.Errorf("まとめた応答になっていません:\n%s", out)
	}
}

// 隠しファイルは送らない。.DS_Store をサイトの公開領域へ配るのは、
// まず意図ではない。
func TestUploadDirSkipsHiddenFiles(t *testing.T) {
	h := newHarness(t, withThemeFiles())
	h.reply(okJSON(`{}`), okJSON(`{}`), okJSON(`{}`), okJSON(`{}`))

	h.mustRun("assets", "upload", "--site-id", "1", "--dir", "dist")

	for _, sent := range h.doer.sent {
		if name := multipartFilename(t, sent.Body); strings.HasPrefix(name, ".") {
			t.Errorf("隠しファイルを送っています: %q", name)
		}
	}
}

// 途中で失敗したら止める。どこまで通ったのかが分からない状態を広げない。
func TestUploadDirStopsAtFirstFailure(t *testing.T) {
	h := newHarness(t, withThemeFiles())
	h.reply(okJSON(`{"id":"1"}`), stub{status: 500, body: `{"error":{"message":"boom"}}`, contentType: "application/json"})

	err := h.run("assets", "upload", "--site-id", "1", "--dir", "dist")
	if err == nil {
		t.Fatal("エラーになるはずです")
	}
	if len(h.doer.sent) != 2 {
		t.Errorf("送信 = %d 件（2件目で止まるはず）", len(h.doer.sent))
	}
	// 何件通ったのかを言わないと、やり直しの判断ができない。
	if got := h.errOut.String(); !strings.Contains(got, "2/4") || !strings.Contains(got, "1 件") {
		t.Errorf("どこで止まったのか分かりません: %q", got)
	}
}

func TestUploadDirConflictsWithFile(t *testing.T) {
	h := newHarness(t, withThemeFiles())
	err := h.run("assets", "upload", "--site-id", "1", "--dir", "dist", "--file", "dist/logo.png")
	if err == nil {
		t.Fatal("両方指定して成功しています")
	}
	if code := ExitCode(err); code != ExitUsage {
		t.Errorf("exit = %d, want %d", code, ExitUsage)
	}
	if len(h.doer.sent) != 0 {
		t.Error("送信しています")
	}
}

// --dry-run は1件も送らずに、何をどこへ送るつもりかを並べる。
func TestUploadDirDryRun(t *testing.T) {
	h := newHarness(t, withThemeFiles())
	out := h.mustRun("assets", "upload", "--site-id", "1", "--dir", "dist", "--path", "/theme-images", "--dry-run")

	if len(h.doer.sent) != 0 {
		t.Error("--dry-run なのに送信しています")
	}
	for _, want := range []string{"/theme-images/icons", "dist/icons/menu.svg", "4"} {
		if !strings.Contains(out, want) {
			t.Errorf("%q が出ていません:\n%s", want, out)
		}
	}
}

func TestRemotePath(t *testing.T) {
	for _, tc := range []struct{ base, name, want string }{
		{"/theme-images", "logo.png", "/theme-images"},
		{"/theme-images", "icons/menu.svg", "/theme-images/icons"},
		{"/theme-images/", "icons/menu.svg", "/theme-images/icons"},
		{"", "icons/menu.svg", "/icons"},
		{"", "logo.png", ""},
	} {
		if got := remotePath(tc.base, tc.name); got != tc.want {
			t.Errorf("remotePath(%q, %q) = %q, want %q", tc.base, tc.name, got, tc.want)
		}
	}
}

// multipartFilename は送られた multipart 本文からファイル名を取り出します。
func multipartFilename(t *testing.T, body string) string {
	t.Helper()
	const marker = `filename="`
	i := strings.Index(body, marker)
	if i < 0 {
		t.Fatalf("multipart にファイル名がありません: %.200s", body)
	}
	rest := body[i+len(marker):]
	return rest[:strings.Index(rest, `"`)]
}

// multipartField は送られた multipart 本文から通常のフィールドを取り出します。
func multipartField(t *testing.T, body, name string) string {
	t.Helper()
	marker := `name="` + name + `"`
	i := strings.Index(body, marker)
	if i < 0 {
		return ""
	}
	rest := body[i+len(marker):]
	rest = strings.TrimLeft(rest, "\r\n")
	end := strings.Index(rest, "\r\n")
	if end < 0 {
		return ""
	}
	v, err := url.QueryUnescape(strings.TrimSpace(rest[:end]))
	if err != nil {
		return strings.TrimSpace(rest[:end])
	}
	return v
}
