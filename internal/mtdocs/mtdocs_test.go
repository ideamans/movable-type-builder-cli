package mtdocs_test

import (
	"strings"
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/mtdocs"
)

// 原本が読めること、そして中身が痩せていないこと。
// 取り直しに失敗した原本をコミットする事故を防ぐ。
func TestSourcesLoad(t *testing.T) {
	tags, err := mtdocs.Tags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) < 600 {
		t.Errorf("タグが %d 件しかありません", len(tags))
	}
	var withMod, live int
	for _, tag := range tags {
		if tag.Name == "" || tag.Slug == "" {
			t.Errorf("名前かスラッグが空のタグがあります: %+v", tag)
		}
		if tag.Type != "BLOCK" && tag.Type != "FUNCTION" {
			t.Errorf("%s の種別が %q です", tag.Name, tag.Type)
		}
		if !tag.Deprecated {
			live++
		}
		if len(tag.Attributes) > 0 {
			withMod++
		}
	}
	if live < 500 {
		t.Errorf("現行タグが %d 件しかありません", live)
	}
	if withMod < 200 {
		t.Errorf("モディファイアを持つタグが %d 件しかありません（取得に失敗した原本かもしれません）", withMod)
	}

	mods, err := mtdocs.Modifiers()
	if err != nil {
		t.Fatal(err)
	}
	if len(mods) < 40 {
		t.Errorf("グローバルモディファイアが %d 件しかありません", len(mods))
	}

	roles, err := mtdocs.Roles()
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) < 9 {
		t.Errorf("組み込みロールが %d 件しかありません", len(roles))
	}
	for _, r := range roles {
		if r.Name == "" || len(r.Permissions) == 0 {
			t.Errorf("中身の無いロールがあります: %+v", r)
		}
	}

	types, err := mtdocs.ArchiveTypes()
	if err != nil {
		t.Fatal(err)
	}
	if len(types) < 30 {
		t.Errorf("アーカイブ種別が %d 件しかありません", len(types))
	}

	perms, err := mtdocs.Permissions()
	if err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{"system", "blog"} {
		if len(perms[scope]) == 0 {
			t.Errorf("%s スコープの権限がありません", scope)
		}
	}
	// Data API のサインインに要る権限は、資料の要になっている
	found := false
	for _, p := range perms["system"] {
		if p.Key == "sign_in_data_api" {
			found = true
		}
	}
	if !found {
		t.Error("system 権限に sign_in_data_api がありません")
	}
}

// 代表的なタグが、期待する群と種別で入っていること。
func TestKnownTags(t *testing.T) {
	tags, err := mtdocs.Tags()
	if err != nil {
		t.Fatal(err)
	}
	by := map[string]mtdocs.Tag{}
	for _, tag := range tags {
		by[tag.Name] = tag
	}
	for name, want := range map[string]struct{ group, kind string }{
		"MTEntries":       {"entry", "BLOCK"},
		"MTEntryTitle":    {"entry", "FUNCTION"},
		"MTEntryCategory": {"entry", "FUNCTION"}, // 記事の文脈で使うので entry
		"MTIf":            {"control", "BLOCK"},
		"MTContentField":  {"content", "BLOCK"},
	} {
		tag, ok := by[name]
		if !ok {
			t.Errorf("%s が原本にありません", name)
			continue
		}
		if tag.Group != want.group || tag.Type != want.kind {
			t.Errorf("%s = group %q / type %q, want %q / %q", name, tag.Group, tag.Type, want.group, want.kind)
		}
	}
	if e := by["MTEntries"]; !strings.Contains(strings.Join(modNames(e), " "), "lastn") {
		t.Error("MTEntries に lastn モディファイアがありません")
	}
}

func modNames(t mtdocs.Tag) []string {
	var out []string
	out = append(out, t.Attributes...)
	return out
}

// 要約は自前で書いたもので、全件にある。マニュアルの文章は原本に持たない。
func TestSummariesCoverEverything(t *testing.T) {
	sum, err := mtdocs.LoadSummaries()
	if err != nil {
		t.Fatal(err)
	}
	tags, _ := mtdocs.Tags()
	mods, _ := mtdocs.Modifiers()
	roles, _ := mtdocs.Roles()
	for _, x := range tags {
		if x.Summary == "" {
			t.Errorf("タグ %s の要約がありません（summaries.json の tags に足す）", x.Name)
		}
	}
	for _, x := range mods {
		if x.Summary == "" {
			t.Errorf("モディファイア %s の要約がありません", x.Name)
		}
	}
	for _, x := range roles {
		if x.Summary == "" {
			t.Errorf("ロール %s の要約がありません", x.Name)
		}
	}
	// 消えたタグの要約が残っていたら、取り直しのときの消し忘れ
	known := map[string]bool{}
	for _, x := range tags {
		known[x.Name] = true
	}
	for name := range sum.Tags {
		if !known[name] {
			t.Errorf("summaries.json に存在しないタグ %s の要約があります", name)
		}
	}
}

// 属性は書き方（name="value"）だけ。マニュアルの説明文が紛れ込んでいないこと。
func TestAttributesAreSyntaxOnly(t *testing.T) {
	tags, err := mtdocs.Tags()
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range tags {
		for _, a := range tag.Attributes {
			if strings.ContainsAny(a, "\n。") || len([]rune(a)) > 120 {
				t.Errorf("%s の属性に文章が混ざっています: %q", tag.Name, a)
			}
		}
	}
}
