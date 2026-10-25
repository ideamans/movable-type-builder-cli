package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ideamans/movable-type-builder-cli/internal/mtdocs"
)

// modifierCategories はグローバルモディファイアの並べ方です。
// 用途でまとめないと、48件がただの列挙になって引けません。
var modifierCategories = []struct {
	title string
	names []string
}{
	{"出力の加工", strings.Fields("trim trim_to ltrim rtrim strip strip_linefeeds strip_tags remove_html nl2br indent wrap_text space_pad zero_pad spacify cat replace regex_replace lower_case upper_case capitalize sprintf string_format numify _default")},
	{"エスケープ・エンコード", strings.Fields("encode_html encode_xml encode_js encode_json encode_php encode_url encode_sha1 decode_html decode_xml escape sanitize spam_protect nofollowfy dirify")},
	{"数える", strings.Fields("count_characters count_words count_paragraphs")},
	{"テンプレートの制御", strings.Fields("setvar mteval filters")},
	{"記号の整形（要プラグイン）", strings.Fields("smarty_pants smart_dashes smart_quotes smart_ellipses")},
}

func buildModifiers(dir string) error {
	mods, err := mtdocs.Modifiers()
	if err != nil {
		return err
	}
	by := map[string]mtdocs.Modifier{}
	for _, m := range mods {
		by[m.Name] = m
	}

	placed := map[string]bool{}
	cats := modifierCategories
	for _, c := range cats {
		for _, n := range c.names {
			placed[n] = true
		}
	}
	var rest []string
	for _, m := range mods {
		if !placed[m.Name] {
			rest = append(rest, m.Name)
		}
	}
	sort.Strings(rest)

	b := &strings.Builder{}
	fmt.Fprintf(b, "# グローバルモディファイア\n\n"+
		"**どのファンクションタグにも付けられる**共通の修飾子です。\n"+
		"ブロックタグには付きません（ブロックタグの属性は、そのタグ固有のモディファイアです）。\n\n"+
		"```html\n<$mt:EntryTitle trim_to=\"20+…\" encode_html=\"1\"$>\n```\n\n"+
		"**複数指定したときは、書いた順に左から適用されます。**\n\n全 %d 件。説明は1行の要約で、"+
		"値の取り方や例はリンク先のマニュアルにあります。\n\n", len(mods))

	emit := func(title string, names []string) {
		var present []string
		for _, n := range names {
			if _, ok := by[n]; ok {
				present = append(present, n)
			}
		}
		if len(present) == 0 {
			return
		}
		sort.Strings(present)
		fmt.Fprintf(b, "## %s\n\n", title)
		// 要約とマニュアルへのリンクだけ。マニュアルの文章と使用例は持ち込まない。
		for _, n := range present {
			m := by[n]
			fmt.Fprintf(b, "- `%s` — %s [マニュアル](%s)\n", m.Name, m.Summary, m.URL())
		}
		b.WriteString("\n")
	}
	for _, c := range cats {
		emit(c.title, c.names)
	}
	emit("その他", rest)
	return write(filepath.Join(dir, "modifiers.md"), b.String())
}

// buildPermissions は権限表のうち、機械から出せる部分だけを差し込みます。
// ロールの説明は自前の要約（summaries.json）。目印の中だけを置き換えます。
func buildPermissions() error {
	roles, err := mtdocs.Roles()
	if err != nil {
		return err
	}
	path := filepath.Join(*out, "mt-cms-design", "references", "permissions.md")
	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s は手書きの土台が要ります: %w", path, err)
	}
	body := string(buf)

	rows := &strings.Builder{}
	for _, r := range roles {
		perm := fmt.Sprintf("%d", len(r.Permissions))
		if len(r.Permissions) == 1 {
			perm = "`" + r.Permissions[0] + "`"
		}
		fmt.Fprintf(rows, "| %d | %s | %s | %s |\n", r.ID, r.Name, perm, r.Summary)
	}
	return write(path, replaceBlock(body, "<!-- roles:begin -->", "<!-- roles:end -->", strings.TrimRight(rows.String(), "\n")))
}

func buildArchives() error {
	types, err := mtdocs.ArchiveTypes()
	if err != nil {
		return err
	}
	path := filepath.Join(*out, "mt-cms-design", "references", "archives.md")
	buf, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("%s は手書きの土台が要ります: %w", path, err)
	}
	var names []string
	for _, t := range types {
		names = append(names, "`"+t.Key+"`")
	}
	sort.Strings(names)
	line := fmt.Sprintf("全 %d 種: %s", len(names), strings.Join(names, " "))
	return write(path, replaceBlock(string(buf), "<!-- archives:begin -->", "<!-- archives:end -->", line))
}

// replaceBlock は目印に挟まれた部分だけを差し替えます。目印が無ければそのまま返します。
func replaceBlock(body, begin, end, content string) string {
	i := strings.Index(body, begin)
	j := strings.Index(body, end)
	if i < 0 || j < 0 || j < i {
		return body
	}
	return body[:i+len(begin)] + "\n" + content + "\n" + body[j:]
}
