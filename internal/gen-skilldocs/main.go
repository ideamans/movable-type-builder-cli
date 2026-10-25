// Command gen-skilldocs は、配布スキルの参照ファイルを原本から作り直します。
//
// 原本は internal/mtdocs/source/*.json（事実だけ）と summaries.json（自前の要約）。
// **マニュアルの文章は書き出さない。** 各タグは要約・属性の書き方・マニュアルの
// ページへのリンクだけにする（internal/mtdocs のパッケージ説明を参照）。出力先は
// plugins/movable-type-builder-cli/skills/*/references/*.md です。
// go generate から呼ばれ、CI が再生成して差分が出たら失敗します。
//
// -fetch を付けると、原本そのものを movabletype.jp から取り直します
// （ネットワークに出るので、MT 側のドキュメントが更新されたときだけ）。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ideamans/movable-type-builder-cli/internal/mtdocs"
)

var (
	fetch   = flag.Bool("fetch", false, "movabletype.jp から原本を取り直す")
	out     = flag.String("out", "../../plugins/movable-type-builder-cli/skills", "スキルの置き場")
	srcDir  = flag.String("source", "../mtdocs/source", "原本の置き場（-fetch のとき書き出す先）")
	cache   = flag.String("cache", "", "取得した HTML の置き場（既定は一時ディレクトリ）")
	verbose = flag.Bool("v", false, "経過を出す")
)

// group は機能別のまとまり。ファイル名と見出しを決める。
type group struct {
	key, title, lead string
}

var groups = []group{
	{"entry", "記事 (Entry)", "記事の一覧・取得・属性。ブログ記事の中心。"},
	{"page", "ウェブページ (Page)", "日付を持たない固定ページ。フォルダで階層化する。"},
	{"content", "コンテンツタイプ (Content Data)", "MT7 以降の構造化コンテンツ。任意のフィールドを定義できる。"},
	{"category", "カテゴリ", "記事の分類。階層を持つ。"},
	{"folder", "フォルダ", "ウェブページの分類。階層を持つ。"},
	{"asset", "アセット（画像・ファイル）", "アップロードした画像やファイル。"},
	{"tag", "タグ（ラベル）", "記事・ページ・アセットに付ける自由なラベル。"},
	{"site", "サイト・ブログ", "サイトの属性と、サイトをまたぐ出力。"},
	{"archive", "アーカイブ・日付", "アーカイブページと日付の出力。"},
	{"control", "制御構文・変数・インクルード", "条件分岐、繰り返し、変数、モジュールの読み込み。"},
	{"pagination", "ページ送り", "一覧の分割表示。"},
	{"user", "ユーザー", "投稿者とサインイン。"},
	{"comment", "コメント・トラックバック", "フィードバック機能。"},
	{"search", "検索", "サイト内検索の結果出力。"},
	{"customfield", "カスタムフィールド", "記事・ページなどに足した独自項目。"},
	{"format", "テキストフォーマット", "本文の整形方式。"},
	{"system", "システム・パス", "設置パスやバージョンなど。"},
}

var spaces = regexp.MustCompile(`\s+`)

func flat(s string) string { return strings.TrimSpace(spaces.ReplaceAllString(s, " ")) }

func main() {
	flag.Parse()
	if *fetch {
		if err := runFetch(); err != nil {
			fmt.Fprintln(os.Stderr, "gen-skilldocs -fetch:", err)
			os.Exit(1)
		}
		return
	}
	if err := build(); err != nil {
		fmt.Fprintln(os.Stderr, "gen-skilldocs:", err)
		os.Exit(1)
	}
}

func build() error {
	tags, err := mtdocs.Tags()
	if err != nil {
		return err
	}
	tagsDir := filepath.Join(*out, "mt-template-tags", "references")
	if err := os.MkdirAll(tagsDir, 0o755); err != nil {
		return err
	}

	byGroup := map[string][]mtdocs.Tag{}
	var deprecated []mtdocs.Tag
	for _, t := range tags {
		if t.Deprecated {
			deprecated = append(deprecated, t)
			continue
		}
		byGroup[t.Group] = append(byGroup[t.Group], t)
	}

	var index []string
	live := 0
	for _, g := range groups {
		rows := byGroup[g.key]
		if len(rows) == 0 {
			continue
		}
		sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
		live += len(rows)

		b := &strings.Builder{}
		fmt.Fprintf(b, "# %s\n\n%s\n\n現行のタグ %d 件。非推奨のものは `tags-deprecated.md` にまとめてあります。\n"+
			"説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**\n\n",
			g.title, g.lead, len(rows))
		for _, t := range rows {
			writeTag(b, t)
		}
		if err := write(filepath.Join(tagsDir, "tags-"+g.key+".md"), b.String()); err != nil {
			return err
		}

		index = append(index, "", "## `tags-"+g.key+".md`", "")
		for _, t := range rows {
			index = append(index, fmt.Sprintf("- %s (%s) — %s", t.Name, kindLetter(t.Type), t.Summary))
		}
	}

	sort.Slice(deprecated, func(i, j int) bool { return deprecated[i].Name < deprecated[j].Name })
	b := &strings.Builder{}
	fmt.Fprintf(b, "# 非推奨のテンプレートタグ\n\n"+
		"**新しく書くテンプレートでは使わないこと。** 既存テンプレートを読むときの参照用です。\n\n%d 件。\n\n", len(deprecated))
	for _, t := range deprecated {
		fmt.Fprintf(b, "- **%s** (%s) — %s [マニュアル](%s)\n", t.Name, t.Type, t.Summary, t.URL())
	}
	if err := write(filepath.Join(tagsDir, "tags-deprecated.md"), b.String()); err != nil {
		return err
	}

	head := []string{"# テンプレートタグ索引", "",
		fmt.Sprintf("現行 %d 件 / 非推奨 %d 件。", live, len(deprecated)), "",
		"**タグ名から、属性とマニュアルへのリンクが載っているファイルを引く表です。**",
		"説明は自前の1行の要約です。詳しい使い方はそのファイルからマニュアルへ。",
		"", "B=ブロックタグ / F=ファンクションタグ"}
	if err := write(filepath.Join(tagsDir, "tags-index.md"), strings.Join(append(head, index...), "\n")+"\n"); err != nil {
		return err
	}

	if err := buildModifiers(tagsDir); err != nil {
		return err
	}
	if err := buildPermissions(); err != nil {
		return err
	}
	if err := buildArchives(); err != nil {
		return err
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "現行 %d / 非推奨 %d のタグを書き出しました\n", live, len(deprecated))
	}
	return nil
}

func kindLetter(t string) string {
	if t == "" {
		return "?"
	}
	return t[:1]
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// writeTag はタグ1件です。要約・属性の書き方・マニュアルへのリンクだけ。
func writeTag(b *strings.Builder, t mtdocs.Tag) {
	fmt.Fprintf(b, "### %s — %s\n", t.Name, orQuestion(t.Type))
	fmt.Fprintf(b, "%s\n", t.Summary)
	if len(t.Attributes) > 0 {
		attrs := make([]string, len(t.Attributes))
		for i, a := range t.Attributes {
			attrs[i] = "`" + strings.ReplaceAll(a, "`", "") + "`"
		}
		fmt.Fprintf(b, "\n属性: %s\n", strings.Join(attrs, " "))
	}
	fmt.Fprintf(b, "\n[マニュアル](%s)\n\n", t.URL())
}

func orQuestion(s string) string {
	if s == "" {
		return "?"
	}
	return s
}

func write(path, content string) error {
	return os.WriteFile(path, []byte(strings.TrimRight(content, "\n")+"\n"), 0o644)
}
