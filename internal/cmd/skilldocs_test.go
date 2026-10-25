package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/mtdocs"
)

const skillsDir = "../../plugins/movable-type-builder-cli/skills"

// 配布する参照ファイルは 500KB を超える生成物で、壊れても誰も気づかない。
// 原本との対応をここで縛る。

// 現行のタグが1件残らずリファレンスに載っていること。
// 分類規則を変えたときに、どこにも出力されないタグができるのを防ぐ。
func TestEveryTagIsDocumented(t *testing.T) {
	tags, err := mtdocs.Tags()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(skillsDir, "mt-template-tags", "references")
	corpus := readAll(t, dir, "tags-")

	var missing []string
	live := 0
	for _, tag := range tags {
		if tag.Deprecated {
			continue
		}
		live++
		if !strings.Contains(corpus, "### "+tag.Name+" — ") {
			missing = append(missing, tag.Name)
		}
	}
	if live < 500 {
		t.Fatalf("現行タグが %d 件しかありません。原本が壊れている可能性があります", live)
	}
	if len(missing) > 0 {
		t.Errorf("リファレンスに載っていないタグが %d 件あります: %s",
			len(missing), strings.Join(missing[:min(10, len(missing))], ", "))
	}
}

// 索引に挙がっているタグは、指しているファイルに実在すること。
func TestIndexPointsAtRealFiles(t *testing.T) {
	dir := filepath.Join(skillsDir, "mt-template-tags", "references")
	index := read(t, filepath.Join(dir, "tags-index.md"))

	fileRe := regexp.MustCompile("^## `(tags-[a-z]+\\.md)`$")
	tagRe := regexp.MustCompile(`^- (MT\S+) \(([BF?])\) — \S`)

	var current string
	checked := 0
	for _, line := range strings.Split(index, "\n") {
		if m := fileRe.FindStringSubmatch(line); m != nil {
			current = m[1]
			if _, err := os.Stat(filepath.Join(dir, current)); err != nil {
				t.Fatalf("索引が存在しないファイル %s を指しています", current)
			}
			continue
		}
		m := tagRe.FindStringSubmatch(line)
		if m == nil || current == "" {
			continue
		}
		body := read(t, filepath.Join(dir, current))
		if !strings.Contains(body, "### "+m[1]+" — ") {
			t.Errorf("%s に %s がありません（索引はここにあると言っています）", current, m[1])
		}
		checked++
	}
	if checked < 500 {
		t.Errorf("索引から検査できたのは %d 件だけです", checked)
	}
}

// グローバルモディファイアが全件載っていること。
func TestEveryModifierIsDocumented(t *testing.T) {
	mods, err := mtdocs.Modifiers()
	if err != nil {
		t.Fatal(err)
	}
	body := read(t, filepath.Join(skillsDir, "mt-template-tags", "references", "modifiers.md"))
	if len(mods) < 40 {
		t.Fatalf("モディファイアが %d 件しかありません", len(mods))
	}
	for _, m := range mods {
		if !strings.Contains(body, "- `"+m.Name+"` — ") || !strings.Contains(body, m.URL()) {
			t.Errorf("modifiers.md に %s（要約とマニュアルへのリンク）がありません", m.Name)
		}
	}
}

// 非推奨タグが現行のファイルに混ざっていないこと。
// 混ざると、エージェントが新しいテンプレートで使ってしまう。
func TestDeprecatedTagsAreIsolated(t *testing.T) {
	tags, err := mtdocs.Tags()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(skillsDir, "mt-template-tags", "references")
	corpus := readAll(t, dir, "tags-")
	deprecatedFile := read(t, filepath.Join(dir, "tags-deprecated.md"))
	corpus = strings.Replace(corpus, deprecatedFile, "", 1)

	for _, tag := range tags {
		if !tag.Deprecated {
			continue
		}
		if strings.Contains(corpus, "### "+tag.Name+" — ") {
			t.Errorf("非推奨の %s が現行のファイルに載っています", tag.Name)
		}
		if !strings.Contains(deprecatedFile, "**"+tag.Name+"**") {
			t.Errorf("非推奨の %s が tags-deprecated.md にありません", tag.Name)
		}
	}
}

// 組み込みロールとアーカイブ種別が、実機から取った原本と一致していること。
func TestCMSDesignReferencesMatchSource(t *testing.T) {
	dir := filepath.Join(skillsDir, "mt-cms-design", "references")

	roles, err := mtdocs.Roles()
	if err != nil {
		t.Fatal(err)
	}
	perms := read(t, filepath.Join(dir, "permissions.md"))
	for _, r := range roles {
		if !strings.Contains(perms, "| "+r.Name+" |") {
			t.Errorf("permissions.md にロール %s がありません", r.Name)
		}
	}

	types, err := mtdocs.ArchiveTypes()
	if err != nil {
		t.Fatal(err)
	}
	arch := read(t, filepath.Join(dir, "archives.md"))
	for _, a := range types {
		if !strings.Contains(arch, "`"+a.Key+"`") {
			t.Errorf("archives.md にアーカイブ種別 %s がありません", a.Key)
		}
	}
}

// スキルが参照しろと書いているファイルが実在すること。
// SKILL.md だけ直してファイルを作り忘れる事故を防ぐ。
func TestSkillsReferenceExistingFiles(t *testing.T) {
	ref := regexp.MustCompile("`(references/[a-z0-9-]+\\.md|tags-[a-z]+\\.md|modifiers\\.md|objects\\.md|permissions\\.md|archives\\.md|recipes\\.md)`")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skill := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		body, err := os.ReadFile(skill)
		if err != nil {
			continue
		}
		for _, m := range ref.FindAllStringSubmatch(string(body), -1) {
			name := strings.TrimPrefix(m[1], "references/")
			if _, err := os.Stat(filepath.Join(skillsDir, e.Name(), "references", name)); err != nil {
				t.Errorf("%s が %s を参照していますが、そのファイルはありません", skill, m[1])
			}
		}
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return string(buf)
}

func readAll(t *testing.T, dir, prefix string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), prefix) && strings.HasSuffix(e.Name(), ".md") {
			b.WriteString(read(t, filepath.Join(dir, e.Name())))
		}
	}
	return b.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// 参照ファイルはマニュアルの文章を持たず、要約とリンクで誘導すること。
// 各タグの見出しの下に要約があり、マニュアルのそのページへのリンクがある。
func TestEveryTagHasSummaryAndManualLink(t *testing.T) {
	tags, err := mtdocs.Tags()
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(skillsDir, "mt-template-tags", "references")
	var all strings.Builder
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		all.WriteString(read(t, filepath.Join(dir, e.Name())))
	}
	body := all.String()
	for _, tag := range tags {
		if tag.Summary == "" {
			t.Errorf("%s の要約が summaries.json にありません", tag.Name)
			continue
		}
		if !strings.Contains(body, tag.URL()) {
			t.Errorf("%s のマニュアルへのリンク（%s）がありません", tag.Name, tag.URL())
		}
	}
}
