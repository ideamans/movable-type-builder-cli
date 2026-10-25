package cmd

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

const twoTemplates = `{"totalResults":2,"items":[
 {"id":218,"name":"インデックスページ","type":"index","text":"<html></html>","outputFile":"index.html","buildType":"Static","linkToFile":""},
 {"id":217,"name":"HTML/ヘッダー","type":"custom","text":"<head></head>","outputFile":"","buildType":"","linkToFile":""}
]}`

// 1テンプレート1ファイル＋マニフェストで書き出せること。
func TestTemplatePullWritesFilesAndManifest(t *testing.T) {
	h := newHarness(t).reply(okJSON(twoTemplates), okJSON(`{"totalResults":2,"items":[]}`))

	h.mustRun("sites", "templates", "pull", "4", "--dir", "tpl")

	// 名前に / が入っていても、ファイル名としては落とす。
	if got := string(h.written["tpl/HTML_ヘッダー.mtml"]); got != "<head></head>" {
		t.Errorf("本文が違います: %q", got)
	}
	if got := string(h.written["tpl/インデックスページ.mtml"]); got != "<html></html>" {
		t.Errorf("本文が違います: %q", got)
	}

	var m manifest
	if err := json.Unmarshal(h.written["tpl/mtb-templates.json"], &m); err != nil {
		t.Fatal(err)
	}
	if m.SiteID != "4" || len(m.Templates) != 2 {
		t.Fatalf("マニフェストが違います: %+v", m)
	}
	// 種別と出力ファイル名はマニフェストが持つ（本文のファイルには入れない）。
	if m.Templates[0].Type != "index" || m.Templates[0].OutputFile != "index.html" {
		t.Errorf("項目が落ちています: %+v", m.Templates[0])
	}
}

// 変わっていないものは送らない。1つ直しただけで全件にリビジョンが積まれると、
// MT 側の更新履歴が「いつ何を変えたか」の役に立たなくなる。
func TestTemplatePushSendsOnlyChanged(t *testing.T) {
	h := newHarness(t)
	h.files["tpl/mtb-templates.json"] = `{"siteId":"4","templates":[
	 {"id":"218","file":"a.mtml","name":"インデックスページ","type":"index","outputFile":"index.html","buildType":"Static"},
	 {"id":"217","file":"b.mtml","name":"HTML/ヘッダー","type":"custom"}]}`
	h.files["tpl/a.mtml"] = "<html>変えました</html>"
	h.files["tpl/b.mtml"] = "<head></head>"
	h.reply(okJSON(twoTemplates), okJSON(`{"totalResults":2,"items":[]}`), okJSON(`{"id":"218"}`), okJSON(twoTemplates), okJSON(`{"totalResults":2,"items":[]}`))

	h.mustRun("sites", "templates", "push", "4", "--dir", "tpl", "--yes")

	var puts []recorded
	for _, sent := range h.doer.sent {
		if sent.Method == "PUT" {
			puts = append(puts, sent)
		}
	}
	if len(puts) != 1 {
		t.Fatalf("PUT = %d 件（変わった1件だけのはず）", len(puts))
	}
	if !strings.HasSuffix(puts[0].URL, "/templates/218") {
		t.Errorf("送り先が違います: %s", puts[0].URL)
	}
	if !strings.Contains(h.errOut.String(), "変わっていない 1 件") {
		t.Errorf("送らなかった件数を報告していません: %s", h.errOut.String())
	}
}

// id が無い項目は新規に作り、採番された ID を書き戻す。
func TestTemplatePushCreatesAndWritesBackID(t *testing.T) {
	h := newHarness(t)
	h.files["tpl/mtb-templates.json"] = `{"siteId":"4","templates":[{"file":"n.mtml","name":"新着","type":"custom"}]}`
	h.files["tpl/n.mtml"] = "<ul></ul>"
	h.reply(okJSON(`{"totalResults":0,"items":[]}`), okJSON(`{"id":"339"}`), okJSON(`{"totalResults":0,"items":[]}`))

	h.mustRun("sites", "templates", "push", "4", "--dir", "tpl", "--yes")

	sent := h.doer.sent[1]
	if sent.Method != "POST" {
		t.Fatalf("新規作成していません: %s %s", sent.Method, sent.URL)
	}
	// 種別は作るときだけ送る（既存の種別は変えられない）。
	if !strings.Contains(sent.Body, "custom") {
		t.Errorf("種別を送っていません: %s", sent.Body)
	}
	var m manifest
	if err := json.Unmarshal(h.written["tpl/mtb-templates.json"], &m); err != nil {
		t.Fatal(err)
	}
	if m.Templates[0].ID != "339" {
		t.Errorf("ID を書き戻していません: %+v", m.Templates[0])
	}
}

// MT にあって手元に無いものは消さない。取り違えると戻せない。
func TestTemplatePushLeavesUntrackedAlone(t *testing.T) {
	h := newHarness(t)
	h.files["tpl/mtb-templates.json"] = `{"siteId":"4","templates":[{"id":"218","file":"a.mtml","name":"インデックスページ","type":"index","outputFile":"index.html","buildType":"Static"}]}`
	h.files["tpl/a.mtml"] = "<html></html>"
	h.reply(okJSON(twoTemplates), okJSON(`{"totalResults":2,"items":[]}`))

	h.mustRun("sites", "templates", "push", "4", "--dir", "tpl", "--yes")

	for _, sent := range h.doer.sent {
		if sent.Method == "DELETE" {
			t.Errorf("消しています: %s", sent.URL)
		}
	}
	if !strings.Contains(h.errOut.String(), "HTML/ヘッダー") {
		t.Errorf("手元に無いものを報告していません: %s", h.errOut.String())
	}
}

// 別のサイトのマニフェストを取り違えて上書きしない。
func TestTemplatePushRefusesWrongSite(t *testing.T) {
	h := newHarness(t)
	h.files["tpl/mtb-templates.json"] = `{"siteId":"4","templates":[{"id":"218","file":"a.mtml","name":"x","type":"index"}]}`
	h.files["tpl/a.mtml"] = "x"

	err := h.run("sites", "templates", "push", "5", "--dir", "tpl", "--yes")
	if err == nil {
		t.Fatal("別サイトへ戻せてしまっています")
	}
	if code := ExitCode(err); code != ExitUsage {
		t.Errorf("exit = %d, want %d", code, ExitUsage)
	}
	if len(h.doer.sent) != 0 {
		t.Error("送信しています")
	}
}

func TestUniqueFileName(t *testing.T) {
	used := map[string]bool{}
	for _, tc := range []struct{ name, id, want string }{
		{"インデックスページ", "1", "インデックスページ.mtml"},
		{"HTML/ヘッダー", "2", "HTML_ヘッダー.mtml"},
		{"インデックスページ", "3", "インデックスページ-2.mtml"},
		{"", "4", "template-4.mtml"},
	} {
		if got := uniqueFileName(tc.name, tc.id, used); got != tc.want {
			t.Errorf("uniqueFileName(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// tmplStub は fetchTemplates の応答を作ります。
func tmplListStub(items string) stub {
	return okJSON(`{"totalResults":1,"items":[` + items + `]}`)
}

// pull はアーカイブマッピングも書き出す。
//
// **マッピングは URL の設計そのもの**なので、これが無いとテンプレートを
// 往復させても記事の出力先（blog/post-1.html など）は戻りません。
func TestPullWritesArchiveMaps(t *testing.T) {
	h := newHarness(t)
	h.reply(
		tmplListStub(`{"id":"355","name":"記事","type":"individual","text":"<mt:EntryTitle>"}`),
		okJSON(`{"totalResults":1,"items":[{"id":32,"archiveType":"Individual","fileTemplate":"blog/%-b.html","buildType":"Static","isPreferred":true}]}`),
	)
	if err := h.run("sites", "templates", "pull", "6", "--dir", "t"); err != nil {
		t.Fatal(err)
	}
	var m manifest
	if err := json.Unmarshal(h.written["t/"+manifestFile], &m); err != nil {
		t.Fatal(err)
	}
	if len(m.Templates) != 1 || len(m.Templates[0].ArchiveMaps) != 1 {
		t.Fatalf("マッピングが書き出されていません: %s", h.written["t/"+manifestFile])
	}
	got := m.Templates[0].ArchiveMaps[0]
	if got.ID != "32" || got.ArchiveType != "Individual" || got.FileTemplate != "blog/%-b.html" || !got.IsPreferred {
		t.Errorf("マッピングの中身 = %+v", got)
	}
}

// マッピングを持てない種別には問い合わせない。30件超のテンプレートで
// 毎回その数だけ往復するのは無駄なので。
func TestPullSkipsTemplatemapsForNonArchiveTypes(t *testing.T) {
	h := newHarness(t)
	h.reply(tmplListStub(`{"id":"390","name":"ページ開始","type":"custom","text":"x"}`))
	if err := h.run("sites", "templates", "pull", "6", "--dir", "t"); err != nil {
		t.Fatal(err)
	}
	if len(h.doer.sent) != 1 {
		t.Errorf("送信回数 = %d, want 1（custom はマッピングを持てない）", len(h.doer.sent))
	}
}

// push はマッピングの変更も戻す。本文が変わっていなくても送る。
func TestPushUpdatesChangedArchiveMap(t *testing.T) {
	h := newHarness(t)
	h.files["t/"+manifestFile] = `{"siteId":"6","templates":[{"id":"355","file":"記事.mtml","name":"記事","type":"individual","archiveMaps":[{"id":"32","archiveType":"Individual","fileTemplate":"blog/%-b.html","buildType":"Static","isPreferred":true}]}]}`
	h.files["t/記事.mtml"] = "<mt:EntryTitle>"
	h.reply(
		// 本文は同じ（送らない）
		tmplListStub(`{"id":"355","name":"記事","type":"individual","text":"<mt:EntryTitle>"}`),
		// いまのマッピングは古い出力先
		okJSON(`{"totalResults":1,"items":[{"id":32,"archiveType":"Individual","fileTemplate":"%y/%-f","buildType":"Static","isPreferred":true}]}`),
		okJSON(`{"id":32,"archiveType":"Individual","fileTemplate":"blog/%-b.html"}`),
	)
	if err := h.run("sites", "templates", "push", "6", "--dir", "t", "--yes"); err != nil {
		t.Fatal(err)
	}
	last := h.last()
	if last.Method != "PUT" || !strings.Contains(last.URL, "/templates/355/templatemaps/32") {
		t.Fatalf("マッピングを更新していません: %s %s", last.Method, last.URL)
	}
	body, err := url.ParseQuery(last.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body.Get("templatemap"), "blog/%-b.html") {
		t.Errorf("新しい出力先が送られていません: %q", body.Get("templatemap"))
	}
}

// 変わっていないマッピングは送らない。
func TestPushSkipsUnchangedArchiveMap(t *testing.T) {
	h := newHarness(t)
	h.files["t/"+manifestFile] = `{"siteId":"6","templates":[{"id":"355","file":"記事.mtml","name":"記事","type":"individual","archiveMaps":[{"id":"32","archiveType":"Individual","fileTemplate":"blog/%-b.html","buildType":"Static","isPreferred":true}]}]}`
	h.files["t/記事.mtml"] = "<mt:EntryTitle>"
	h.reply(
		tmplListStub(`{"id":"355","name":"記事","type":"individual","text":"<mt:EntryTitle>"}`),
		okJSON(`{"totalResults":1,"items":[{"id":32,"archiveType":"Individual","fileTemplate":"blog/%-b.html","buildType":"Static","isPreferred":true}]}`),
	)
	if err := h.run("sites", "templates", "push", "6", "--dir", "t", "--yes"); err != nil {
		t.Fatal(err)
	}
	if len(h.doer.sent) != 2 {
		t.Errorf("送信回数 = %d, want 2（一覧と確認だけ）", len(h.doer.sent))
	}
}

// id の無いマッピングは新規作成し、採番された id をマニフェストに書き戻す。
func TestPushCreatesArchiveMapAndWritesBackID(t *testing.T) {
	h := newHarness(t)
	h.files["t/"+manifestFile] = `{"siteId":"6","templates":[{"id":"355","file":"記事.mtml","name":"記事","type":"individual","archiveMaps":[{"archiveType":"Individual","fileTemplate":"blog/%-b.html"}]}]}`
	h.files["t/記事.mtml"] = "<mt:EntryTitle>"
	h.reply(
		tmplListStub(`{"id":"355","name":"記事","type":"individual","text":"<mt:EntryTitle>"}`),
		okJSON(`{"totalResults":0,"items":[]}`),
		okJSON(`{"id":41,"archiveType":"Individual","fileTemplate":"blog/%-b.html"}`),
	)
	if err := h.run("sites", "templates", "push", "6", "--dir", "t", "--yes"); err != nil {
		t.Fatal(err)
	}
	if h.last().Method != "POST" {
		t.Errorf("新規作成になっていません: %s", h.last().Method)
	}
	var m manifest
	if err := json.Unmarshal(h.written["t/"+manifestFile], &m); err != nil {
		t.Fatal(err)
	}
	if got := m.Templates[0].ArchiveMaps[0].ID; got != "41" {
		t.Errorf("採番された id が書き戻されていません: %q", got)
	}
}
