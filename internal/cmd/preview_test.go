package cmd

import (
	"net/url"
	"strings"
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// ID を渡して本体を省いたら、保存済みの中身を取ってきて添える。
// MT は本体も必須で、そのままだと A resource "template" is required. になる。
func TestPreviewFetchesStoredBodyWhenOmitted(t *testing.T) {
	h := newHarness(t)
	h.reply(
		okJSON(`{"id":"96","name":"記事","text":"<mt:EntryTitle>"}`), // 保存済みの取得
		okJSON(`{"preview":"<h1>タイトル</h1>","status":"success"}`),
	)
	out := h.mustRun("sites", "templates", "preview", "6", "96", "--raw", "1")
	if !strings.Contains(out, "タイトル") {
		t.Errorf("プレビューが返っていません: %s", out)
	}
	if len(h.doer.sent) != 2 {
		t.Fatalf("送信回数 = %d, want 2", len(h.doer.sent))
	}
	// 取得先は /preview を落としたパス。
	u, err := url.Parse(h.doer.sent[0].URL)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(u.Path, "/sites/6/templates/96") {
		t.Errorf("取得先 = %q", u.Path)
	}
	// 取ってきた中身が本体として送られていること。
	body, err := url.ParseQuery(h.doer.sent[1].Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body.Get("template"), "mt:EntryTitle") {
		t.Errorf("保存済みの中身が送られていません: %q", body.Get("template"))
	}
}

// 本体を渡したときは、それが勝つ（差し替えて試すのが本来の使い方）。
func TestPreviewKeepsGivenBody(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"preview":"x","status":"success"}`))
	h.files["snippet.mtml"] = "<mt:SiteName>"
	if err := h.run("sites", "templates", "preview", "6", "96",
		"--set", "template.text=@snippet.mtml", "--raw", "1"); err != nil {
		t.Fatal(err)
	}
	if len(h.doer.sent) != 1 {
		t.Errorf("送信回数 = %d, want 1（取り直しは不要）", len(h.doer.sent))
	}
	body, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body.Get("template"), "mt:SiteName") {
		t.Errorf("渡した本体が使われていません: %q", body.Get("template"))
	}
}

// ID を省いたときは取りに行かない（保存せずに本文だけを描く呼び方）。
func TestPreviewWithoutIDDoesNotFetch(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"preview":"x","status":"success"}`))
	h.files["snippet.mtml"] = "<mt:SiteName>"
	if err := h.run("sites", "templates", "preview", "6",
		"--set", "template.text=@snippet.mtml", "--set", "template.type=index", "--raw", "1"); err != nil {
		t.Fatal(err)
	}
	if len(h.doer.sent) != 1 {
		t.Fatalf("送信回数 = %d, want 1", len(h.doer.sent))
	}
	u, err := url.Parse(h.last().URL)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(u.Path, "/sites/6/templates/preview") {
		t.Errorf("ID なしのエンドポイントを呼んでいません: %q", u.Path)
	}
}

// 記事とページは原稿のまま取る。変換後の HTML を渡すと二重に変換される。
func TestPreviewFetchesEntryWithoutTextFilter(t *testing.T) {
	h := newHarness(t)
	h.reply(
		okJSON(`{"id":"7","title":"記事","body":"# 見出し"}`),
		okJSON(`{"preview":"x","status":"success"}`),
	)
	if err := h.run("sites", "entries", "preview", "6", "7"); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(h.doer.sent[0].URL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("noTextFilter") != "1" {
		t.Errorf("noTextFilter が付いていません: %q", u.RawQuery)
	}
}

// テンプレートの GET は noTextFilter を持たないので、付けない。
func TestPreviewDoesNotSendNoTextFilterWhereUnsupported(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{"id":"96","text":"x"}`), okJSON(`{"preview":"x","status":"success"}`))
	if err := h.run("sites", "templates", "preview", "6", "96", "--raw", "1"); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(h.doer.sent[0].URL)
	if err != nil {
		t.Fatal(err)
	}
	if u.Query().Get("noTextFilter") != "" {
		t.Errorf("対応していないのに noTextFilter を送っています: %q", u.RawQuery)
	}
}

// previewOps のキーが定義に実在すること。綴りがズレると黙って無効になる。
func TestPreviewOpsExistInSpec(t *testing.T) {
	s := spec.MustLoad()
	have := map[string]bool{}
	for _, c := range s.Commands {
		for _, op := range c.Ops {
			have[op.Method+" "+op.Path] = true
		}
	}
	for key := range previewOps {
		if !have[key] {
			t.Errorf("previewOps の %q は API 定義にありません", key)
		}
	}
	// ID を取らない方（本体だけを描く呼び方）も定義にあること。
	// こちらが消えると「サンドボックス不要」の前提が崩れる。
	for _, key := range []string{
		"POST /sites/{site_id}/templates/preview",
		"POST /sites/{site_id}/entries/preview",
		"POST /sites/{site_id}/pages/preview",
		"POST /sites/{site_id}/contentTypes/{content_type_id}/data/preview",
	} {
		if !have[key] {
			t.Errorf("ID なしのプレビュー %q が定義にありません", key)
		}
	}
}

// samePathShape は ID の値によらず形で突き合わせる。
func TestSamePathShape(t *testing.T) {
	cases := []struct {
		pattern, path string
		want          bool
	}{
		{"/sites/{site_id}/entries/{entry_id}", "/sites/6/entries/309", true},
		{"/sites/{site_id}/entries", "/sites/6/entries/309", false},
		{"/sites/{site_id}/pages/{page_id}", "/sites/6/entries/309", false},
		{"/sites/{site_id}/templates/{template_id}", "/sites/6/templates/96", true},
	}
	for _, c := range cases {
		if got := samePathShape(c.pattern, c.path); got != c.want {
			t.Errorf("samePathShape(%q, %q) = %v, want %v", c.pattern, c.path, got, c.want)
		}
	}
}
