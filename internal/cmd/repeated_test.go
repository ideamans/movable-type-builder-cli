package cmd

import (
	"net/url"
	"strings"
	"testing"
)

// 選択肢の配列は「同じ名前を繰り返して」送ること。
//
// JSON の配列を1つの値として送ると MT は黙って無視し、しかも success を
// 返す（テーマのエクスポートが空のテーマを作る、という形で出た）。
// エラーにならない間違いなので、テストで押さえる。
func TestRepeatedArrayIsSentAsRepeatedFields(t *testing.T) {
	for _, tc := range []struct{ name, value string }{
		{"カンマ区切り", "blog_static_files,template_set"},
		{"JSON の配列", `["blog_static_files","template_set"]`},
		{"空白まじり", "blog_static_files , template_set"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t).reply(okJSON(`{"status":"success"}`))
			h.mustRun("sites", "export-theme", "4", "--theme-id", "t", "--theme-name", "T", "--include", tc.value)

			values, err := url.ParseQuery(h.last().Body)
			if err != nil {
				t.Fatal(err)
			}
			got := values["include"]
			if len(got) != 2 || got[0] != "blog_static_files" || got[1] != "template_set" {
				t.Errorf("include = %v（include=a&include=b の形で送るべきです）\n本文: %s", got, h.last().Body)
			}
			if strings.Contains(h.last().Body, "%5B") {
				t.Errorf("JSON の配列のまま送っています: %s", h.last().Body)
			}
		})
	}
}

// 1つだけでも同じ形で送れること（エージェントが最初に試す形）。
func TestRepeatedArrayAcceptsSingleValue(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{"status":"success"}`))
	h.mustRun("sites", "export-theme", "4", "--theme-id", "t", "--theme-name", "T", "--include", "template_set")

	values, _ := url.ParseQuery(h.last().Body)
	if got := values["include"]; len(got) != 1 || got[0] != "template_set" {
		t.Errorf("include = %v", got)
	}
}

// オブジェクトの配列は JSON のまま送ること（並べ替えは JSON で通ると実測済み）。
func TestObjectArrayStaysJSON(t *testing.T) {
	h := newHarness(t).reply(okJSON(`{}`))
	h.files["cats.json"] = `[{"id":2},{"id":1}]`
	h.mustRun("sites", "categories", "permutate", "4", "--categories", "@cats.json")

	values, _ := url.ParseQuery(h.last().Body)
	if got := values.Get("categories"); got != `[{"id":2},{"id":1}]` {
		t.Errorf("categories = %q（JSON のまま送るべきです）", got)
	}
}
