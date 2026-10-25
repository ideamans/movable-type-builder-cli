package cmd

import (
	"net/url"
	"strings"
	"testing"
)

func TestRawAPIGet(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"totalResults":0,"items":[]}`))
	h.mustRun("api", "GET", "/sites/1/entries", "--query", "limit=5")

	sent := h.last()
	if sent.Method != "GET" {
		t.Errorf("method = %s", sent.Method)
	}
	u, err := url.Parse(sent.URL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Path != "/mt/mt-data-api.cgi/v7/sites/1/entries" || u.Query().Get("limit") != "5" {
		t.Errorf("URL = %s", sent.URL)
	}
}

// 先頭のスラッシュを忘れても通す。
func TestRawAPIAcceptsPathWithoutSlash(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.mustRun("api", "get", "sites")
	if pathOf(t, h.last().URL) != "/mt/mt-data-api.cgi/v7/sites" {
		t.Errorf("URL = %s", h.last().URL)
	}
	if h.last().Method != "GET" {
		t.Errorf("メソッドが大文字化されていません: %s", h.last().Method)
	}
}

func TestRawAPIFormFromFile(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.files["entry.json"] = `{"title":"ファイル"}`
	h.mustRun("api", "POST", "/sites/1/entries", "--form", "entry=@entry.json")

	values, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	if values.Get("entry") != `{"title":"ファイル"}` {
		t.Errorf("entry = %q", values.Get("entry"))
	}
}

func TestRawAPIUpload(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.files["photo.jpg"] = "binary"
	h.mustRun("api", "POST", "/sites/1/assets/upload", "--file", "photo.jpg")

	if !strings.HasPrefix(h.last().Header.Get("Content-Type"), "multipart/form-data") {
		t.Errorf("Content-Type = %q", h.last().Header.Get("Content-Type"))
	}
	if !strings.Contains(h.last().Body, `filename="photo.jpg"`) {
		t.Errorf("body = %s", h.last().Body)
	}
}

func TestRawAPIDeleteNeedsConfirmation(t *testing.T) {
	h := newHarness(t)
	err := h.run("api", "DELETE", "/sites/1/entries/2")
	if err == nil || !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("err = %v", err)
	}
	if len(h.doer.sent) != 0 {
		t.Error("確認前に送信しています")
	}
}

func TestRawAPINoAuth(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"version":"9.1.0"}`))
	h.mustRun("api", "GET", "/version", "--no-auth")
	if got := h.last().Header.Get("X-MT-Authorization"); got != "" {
		t.Errorf("認証ヘッダが付いています: %q", got)
	}
}

func TestRawAPIDryRun(t *testing.T) {
	h := newHarness(t)
	out := h.mustRun("api", "POST", "/sites/1/entries", "--form", `entry={"title":"x"}`, "--dry-run")
	if len(h.doer.sent) != 0 {
		t.Error("--dry-run なのに送信しています")
	}
	if !strings.Contains(out, "/v7/sites/1/entries") {
		t.Errorf("出力 = %s", out)
	}
}

func TestRawAPIRejectsMalformedPair(t *testing.T) {
	h := newHarness(t)
	if err := h.run("api", "GET", "/sites", "--query", "broken"); err == nil {
		t.Error("key=value でない指定が通っています")
	}
}
