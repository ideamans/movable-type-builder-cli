package cmd

import (
	"net/url"
	"strings"
	"testing"
)

// mtb api は定義に無いエンドポイントを叩くためのもの。**プラグインが足した
// ものこそ**、ページ繰りや本体の組み立てが要る（実機の MT でも Core 以外が
// 20 個あった）。生成コマンドでできることが、ここでできない理由は無い。

func TestAPIAllFollowsPages(t *testing.T) {
	h := newHarness(t).reply(
		okJSON(`{"totalResults":3,"items":[{"id":1},{"id":2}]}`),
		okJSON(`{"totalResults":3,"items":[{"id":3}]}`),
		okJSON(`{"totalResults":3,"items":[]}`),
	)
	out := h.mustRun("api", "GET", "/sites/1/entries", "--all")

	if len(h.doer.sent) < 2 {
		t.Fatalf("ページを繰っていません: %d 回", len(h.doer.sent))
	}
	if !strings.Contains(out, `"totalResults": 3`) {
		t.Errorf("まとめた応答になっていません:\n%s", out)
	}
}

func TestAPIFieldsFiltersLocally(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":"42","title":"あ","body":"い"}`))
	out := h.mustRun("api", "GET", "/sites/1/entries/42", "--fields", "id,title")

	if strings.Contains(out, `"body"`) {
		t.Errorf("絞れていません:\n%s", out)
	}
	if !strings.Contains(out, `"title"`) {
		t.Errorf("残すべきものが消えました:\n%s", out)
	}
}

// --set は定義を持たないので、本体の名前を検査しない。プラグインが
// どんな名前の本体を取るかは事前に分からないため。
func TestAPISetBuildsAnyBody(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"id":"1"}`))
	h.files["note.txt"] = "本文です"
	h.mustRun("api", "POST", "/plugins/acme/things",
		"--set", "thing.label=ラベル", "--set", "thing.note=@note.txt",
		"--set-json", "thing.count=3")

	values, err := url.ParseQuery(h.last().Body)
	if err != nil {
		t.Fatal(err)
	}
	got := values.Get("thing")
	for _, want := range []string{`"label":"ラベル"`, `"note":"本文です"`, `"count":3`} {
		if !strings.Contains(got, want) {
			t.Errorf("%s がありません: %s", want, got)
		}
	}
}

// --form で渡した本体の上に --set を重ねられること（ファイルを土台にする使い方）。
func TestAPISetLayersOnForm(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.mustRun("api", "POST", "/x",
		"--form", `thing={"label":"もと","keep":"のこる"}`,
		"--set", "thing.label=あと")

	values, _ := url.ParseQuery(h.last().Body)
	got := values.Get("thing")
	if !strings.Contains(got, `"label":"あと"`) || !strings.Contains(got, `"keep":"のこる"`) {
		t.Errorf("重ね方が違います: %s", got)
	}
}
