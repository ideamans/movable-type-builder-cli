package cmd

import (
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

// テンプレートを部品から組めること。これが無いと、HTML をテンプレートに
// 入れるだけで jq が要る。
func TestSetBuildsResourceBody(t *testing.T) {
	h := newHarness(t)
	h.files["index.html"] = "<!doctype html>\n<html><$mt:BlogName$></html>\n"
	h.reply(okJSON(`{"id":"42"}`))

	h.mustRun("sites", "templates", "create", "--site-id", "1",
		"--set", "template.name=メインページ",
		"--set", "template.type=index",
		"--set", "template.outputFile=index.html",
		"--set", "template.text=@index.html")

	var doc struct {
		Name, Type, OutputFile, Text string
	}
	if err := json.Unmarshal([]byte(formField(t, h.last().Body, "template")), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Name != "メインページ" || doc.Type != "index" || doc.OutputFile != "index.html" {
		t.Errorf("組み立てが違います: %+v", doc)
	}
	if doc.Text != h.files["index.html"] {
		t.Errorf("本文がファイルの中身と違います: %q", doc.Text)
	}
}

// MT のテンプレートは < だらけ。< に逃がすと --dry-run で目視できず、
// エージェントには同じ内容が数倍のトークンになる。
func TestSetDoesNotEscapeHTML(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{}`))
	h.mustRun("sites", "templates", "create", "--site-id", "1", "--set", "template.text=<$mt:BlogName$>")

	// form エンコードを解いた「中の JSON」を見ること。生の本文は < が
	// %3C になるので、そこを見ても何も検査したことにならない。
	got := formField(t, h.last().Body, "template")
	if strings.Contains(got, `\u003c`) {
		t.Errorf("HTML が逃がされています: %s", got)
	}
	if !strings.Contains(got, `<$mt:BlogName$>`) {
		t.Errorf("そのままの形で入っていません: %s", got)
	}
}

// 明示の JSON を土台にして、その上に重ねられること。
func TestSetLayersOnExplicitJSON(t *testing.T) {
	h := newHarness(t)
	h.files["base.json"] = `{"name":"もとの名前","type":"index","outputFile":"index.html"}`
	h.reply(okJSON(`{}`))

	h.mustRun("sites", "templates", "create", "--site-id", "1",
		"--template", "@base.json", "--set", "template.name=あとの名前")

	var doc map[string]any
	if err := json.Unmarshal([]byte(formField(t, h.last().Body, "template")), &doc); err != nil {
		t.Fatal(err)
	}
	if doc["name"] != "あとの名前" {
		t.Errorf("--set が重なっていません: %v", doc["name"])
	}
	if doc["outputFile"] != "index.html" {
		t.Errorf("土台の項目が消えています: %v", doc)
	}
}

// --set-json は数値・真偽値・入れ子を運べる。
func TestSetJSONCarriesTypes(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{}`))
	h.mustRun("sites", "templates", "create", "--site-id", "1",
		"--set", "template.name=x", "--set-json", `template.contentType={"id":5}`)

	var doc struct {
		ContentType struct{ ID float64 } `json:"contentType"`
	}
	if err := json.Unmarshal([]byte(formField(t, h.last().Body, "template")), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.ContentType.ID != 5 {
		t.Errorf("数値として運べていません: %+v", doc)
	}
}

// 入れ子は途中を作りながら辿る。
func TestSetNestsPaths(t *testing.T) {
	h := newHarness(t)
	h.reply(okJSON(`{}`))
	h.mustRun("sites", "templates", "create", "--site-id", "1", "--set", "template.contentType.id=5")

	if got := formField(t, h.last().Body, "template"); !strings.Contains(got, `"contentType":{"id":"5"}`) {
		t.Errorf("入れ子になっていません: %s", got)
	}
}

// 打ち間違いを黙って飲み込まない。
func TestSetRejectsBadInput(t *testing.T) {
	for _, tc := range []struct{ name, arg string }{
		{"= が無い", "template.name"},
		{". が無い", "template=x"},
		{"項目名が空", "template.=x"},
		{"そんなフラグは無い", "entry.name=x"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			err := h.run("sites", "templates", "create", "--site-id", "1", "--set", tc.arg)
			if err == nil {
				t.Fatal("エラーになるはずです")
			}
			if code := ExitCode(err); code != ExitUsage {
				t.Errorf("exit = %d, want %d", code, ExitUsage)
			}
			if len(h.doer.sent) != 0 {
				t.Error("送信しています")
			}
		})
	}
}

// formField は送られた form 本文から1つの値を取り出します。
func formField(t *testing.T, body, name string) string {
	t.Helper()
	values, err := url.ParseQuery(body)
	if err != nil {
		t.Fatal(err)
	}
	return values.Get(name)
}
