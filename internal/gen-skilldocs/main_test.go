package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/mtdocs"
)

// 分類の順番には意味がある。folder を category より先に見ないと、
// HasSubFolders が category に落ちる。
func TestGroupOf(t *testing.T) {
	tests := map[string]string{
		"MTEntries":           "entry",
		"MTEntryTitle":        "entry",
		"MTPages":             "page",
		"MTPageTitle":         "page",
		"MTContentField":      "content",
		"MTCategories":        "category",
		"MTHasSubCategories":  "category",
		"MTFolderLabel":       "folder",
		"MTHasSubFolders":     "folder",
		"MTTopLevelParent":    "folder",
		"MTAssetURL":          "asset",
		"MTTagName":           "tag",
		"MTAuthorDisplayName": "user",
		"MTSigninLink":        "user",
		"MTCommentBody":       "comment",
		"MTArchiveTitle":      "archive",
		"MTSearchString":      "search",
		"MTBlogName":          "site",
		"MTMultiBlog":         "site",
		"MTNextLink":          "pagination",
		"MTIf":                "control",
		"MTSetHashVar":        "control",
		"MTIndexList":         "control",
		"MTCustomFieldValue":  "customfield",
		"MTTextFormat":        "format",
		"MTCGIPath":           "system",
		"MTDataAPIScript":     "system",
	}
	for name, want := range tests {
		if got := groupOf(name); got != want {
			t.Errorf("groupOf(%q) = %q, want %q", name, got, want)
		}
	}
}

// 未分類が出たら、規則に足すべきものが増えたということ。
// 分類漏れはリファレンスからそのタグが消えることを意味するので落とす。
func TestNoUngroupedTags(t *testing.T) {
	tags := mustTags(t)
	var orphans []string
	for _, tag := range tags {
		if tag.Deprecated {
			continue
		}
		if groupOf(tag.Name) == "other" {
			orphans = append(orphans, tag.Name)
		}
	}
	if len(orphans) > 0 {
		t.Errorf("群に割り当てられないタグが %d 件あります: %s",
			len(orphans), strings.Join(orphans, ", "))
	}
}

// 原本に入っている群は、すべて出力先のファイルを持っていること。
func TestEveryGroupHasAFile(t *testing.T) {
	known := map[string]bool{}
	for _, g := range groups {
		known[g.key] = true
	}
	seen := map[string]bool{}
	for _, tag := range mustTags(t) {
		if !tag.Deprecated {
			seen[groupOf(tag.Name)] = true
		}
	}
	for g := range seen {
		if !known[g] {
			t.Errorf("群 %q に対応する出力ファイルの定義がありません（groups に足してください）", g)
		}
	}
}

func TestTruncKeepsRunes(t *testing.T) {
	// 日本語をバイトで切ると壊れる。ルーンで切ること。
	got := trunc("あいうえおかきくけこ", 5)
	if got != "あいうえお…" {
		t.Errorf("trunc = %q", got)
	}
	if got := trunc("あいう", 5); got != "あいう" {
		t.Errorf("短い文字列を切っています: %q", got)
	}
}

func TestFlatCollapsesWhitespace(t *testing.T) {
	if got := flat(" a \n b\t c "); got != "a b c" {
		t.Errorf("flat = %q", got)
	}
}

func TestReplaceBlock(t *testing.T) {
	body := "前\n<!-- x:begin -->\n古い\n<!-- x:end -->\n後"
	got := replaceBlock(body, "<!-- x:begin -->", "<!-- x:end -->", "新しい")
	want := "前\n<!-- x:begin -->\n新しい\n<!-- x:end -->\n後"
	if got != want {
		t.Errorf("replaceBlock = %q", got)
	}
	// 目印が無ければ触らない
	if got := replaceBlock("なし", "<!-- x:begin -->", "<!-- x:end -->", "新"); got != "なし" {
		t.Errorf("目印が無いのに書き換えています: %q", got)
	}
}

func mustTags(t *testing.T) []mtdocs.Tag {
	t.Helper()
	tags, err := mtdocs.Tags()
	if err != nil {
		t.Fatal(err)
	}
	return tags
}

// build() を一時ディレクトリに向けて回し、書き出された中身を見る。
// 索引が指すファイル名を間違えたことがあるので、実際に描かせて確かめる。
func TestBuildWritesReferences(t *testing.T) {
	dir := t.TempDir()
	// permissions.md / archives.md は手書きの土台に差し込む形なので、
	// 目印だけ置いた最小のものを用意する。
	base := filepath.Join(dir, "mt-cms-design", "references")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}
	write(filepath.Join(base, "permissions.md"),
		"# 権限\n\n| ID | ロール | 権限 | できること |\n| --- | --- | --- | --- |\n<!-- roles:begin -->\n古い\n<!-- roles:end -->\n")
	write(filepath.Join(base, "archives.md"), "# アーカイブ\n\n<!-- archives:begin -->\n古い\n<!-- archives:end -->\n")

	restore := swapFlags(dir)
	defer restore()
	if err := build(); err != nil {
		t.Fatal(err)
	}

	tagsDir := filepath.Join(dir, "mt-template-tags", "references")
	for _, name := range []string{"tags-index.md", "tags-entry.md", "tags-control.md",
		"tags-deprecated.md", "modifiers.md"} {
		st, err := os.Stat(filepath.Join(tagsDir, name))
		if err != nil {
			t.Errorf("%s が書き出されていません", name)
			continue
		}
		if st.Size() < 500 {
			t.Errorf("%s が %d バイトしかありません", name, st.Size())
		}
	}

	entry := read(t, filepath.Join(tagsDir, "tags-entry.md"))
	if !strings.Contains(entry, "### MTEntries — BLOCK") {
		t.Error("tags-entry.md に MTEntries がありません")
	}
	if !strings.Contains(entry, "`lastn=\"N\"`") {
		t.Error("tags-entry.md に属性が出ていません")
	}
	if !strings.Contains(entry, "https://www.movabletype.jp/documentation/appendices/tags/entries.html") {
		t.Error("tags-entry.md にマニュアルへのリンクがありません")
	}
	// 群のファイルが指す非推奨ファイルの名前が正しいこと（実際に間違えた）
	if !strings.Contains(entry, "`tags-deprecated.md`") {
		t.Error("非推奨ファイルへの案内が違います")
	}

	index := read(t, filepath.Join(tagsDir, "tags-index.md"))
	if !strings.Contains(index, "## `tags-entry.md`") || !strings.Contains(index, "- MTEntries (B) — ") {
		t.Error("索引の中身が違います")
	}

	// 目印の中だけが差し替わり、外は残ること
	perms := read(t, filepath.Join(base, "permissions.md"))
	if strings.Contains(perms, "古い") {
		t.Error("目印の中が差し替わっていません")
	}
	if !strings.Contains(perms, "# 権限") || !strings.Contains(perms, "| Editor |") {
		t.Errorf("permissions.md = %q", trunc(perms, 200))
	}
	arch := read(t, filepath.Join(base, "archives.md"))
	if !strings.Contains(arch, "`Individual`") || strings.Contains(arch, "古い") {
		t.Errorf("archives.md = %q", trunc(arch, 200))
	}
}

// 手書きの土台が無ければ、黙って作らずに失敗すること。
func TestBuildFailsWithoutHandwrittenBase(t *testing.T) {
	dir := t.TempDir()
	restore := swapFlags(dir)
	defer restore()
	if err := build(); err == nil {
		t.Error("土台が無いのに成功しています")
	}
}

func swapFlags(dir string) func() {
	oldOut := *out
	*out = dir
	return func() { *out = oldOut }
}

func read(t *testing.T, path string) string {
	t.Helper()
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return string(buf)
}

// 当たることのない規則が混ざっていないこと。
// 先に来た規則に食われる書き方をしていると、読む人を誤解させる。
func TestNoDeadGroupRules(t *testing.T) {
	tags := mustTags(t)
	hit := map[string]bool{}
	for _, tag := range tags {
		for _, r := range groupRules {
			if r.re.MatchString(tag.Name) {
				hit[r.group] = true
				break
			}
		}
	}
	for _, r := range groupRules {
		if !hit[r.group] {
			t.Errorf("群 %q の規則は1件も当たりません（先の規則に食われている可能性）", r.group)
		}
	}
}
