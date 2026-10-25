package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

// extraQueryFlag と extraFormFlag は、定義に載っていないパラメータを送るための
// 逃げ道です。プラグインが足したパラメータや、MT のバージョン差で
// リファレンスに未反映のものがここを通ります。
const (
	extraQueryFlag = "query"
	extraFormFlag  = "form"
	// fieldsFlag は応答の属性を絞るフラグです。MT は GET にしか fields を
	// 用意していませんが（226 中 67）、使えたり使えなかったりする方が
	// 混乱するので、対応していないエンドポイントでは mtb が手元で絞ります。
	fieldsFlag = "fields"
)

// addGeneratedCommands は埋め込んだ API 定義からコマンドツリーを作ります。
func (a *App) addGeneratedCommands() {
	groups := map[string]*cobra.Command{}

	// Groups は名前順に並んでいるので、親が必ず子より先に来ます。
	for _, g := range a.env.Spec.Groups {
		c := &cobra.Command{
			Use:     g.Command[len(g.Command)-1],
			Short:   groupShort(g),
			Aliases: aliasesFor(g.Command[len(g.Command)-1]),
		}
		groups[strings.Join(g.Command, " ")] = c
		a.attach(groups, g.Command, c)
	}

	for _, sc := range a.env.Spec.Commands {
		a.attach(groups, sc.Command, a.newOperationCmd(sc))
	}
}

// attach は親グループ（無ければルート）にぶら下げます。
func (a *App) attach(groups map[string]*cobra.Command, path []string, c *cobra.Command) {
	if len(path) == 1 {
		a.root.AddCommand(c)
		return
	}
	parent, ok := groups[strings.Join(path[:len(path)-1], " ")]
	if !ok {
		// 生成器がグループを取りこぼしたときだけ起きます。黙って落とすと
		// コマンドが1つ消えるだけで気づけないので、ルート直下に出します。
		a.root.AddCommand(c)
		return
	}
	parent.AddCommand(c)
}

func groupShort(g *spec.Group) string {
	if g.Short == "" {
		return strings.Join(g.Command, " ")
	}
	return g.Short
}

// aliasesFor は複数形のグループ名に単数形の別名を足します。
// `mtb sites entry list` のように書けると打ちやすいためです。
func aliasesFor(name string) []string {
	switch {
	case strings.HasSuffix(name, "ies"):
		return []string{strings.TrimSuffix(name, "ies") + "y"}
	case strings.HasSuffix(name, "ss"), strings.HasSuffix(name, "us"):
		return nil
	case strings.HasSuffix(name, "s") && len(name) > 2:
		return []string{strings.TrimSuffix(name, "s")}
	}
	return nil
}

// newOperationCmd は末端コマンド1つを作ります。
func (a *App) newOperationCmd(sc *spec.Command) *cobra.Command {
	canonical := canonicalPathParams(sc)

	cmd := &cobra.Command{
		Use:     usageLine(sc.Name(), canonical),
		Short:   sc.Short,
		Long:    longDescription(sc, canonical),
		Args:    cobra.MaximumNArgs(len(canonical)),
		RunE:    func(cmd *cobra.Command, args []string) error { return a.runOperation(cmd, sc, args) },
		Aliases: nil,
	}

	f := cmd.Flags()
	for _, p := range allParams(sc) {
		f.String(p.Flag, "", flagUsage(p))
	}
	if !hasServerFields(sc) {
		f.String(fieldsFlag, "", i18n.T(
			"応答から取り出す属性をカンマ区切りで指定（このエンドポイントはサーバが未対応のため手元で絞ります）",
			"comma-separated fields to keep from the response (this endpoint has no server-side support, so the filtering happens locally)"))
	}
	if hasPagination(sc) {
		f.Bool(allFlag, false, i18n.T(
			"全ページを取得して1つの応答にまとめる（limit と offset は自動で動かします）",
			"fetch every page and merge them into one response (limit and offset are driven automatically)"))
	}
	if hasObjectBody(sc) {
		f.StringArray(setFlag, nil, i18n.T(
			"リソース本体を項目ごとに組み立てる（<フラグ>.<項目>=<値>、値の @ファイル名 でファイルから読む、繰り返し可）",
			"build the resource body field by field (<flag>.<field>=<value>; a value of @filename is read from that file; repeatable)"))
		f.StringArray(setJSONFlag, nil, i18n.T(
			"--set と同じだが値を JSON として解釈する（数値・真偽値・入れ子）",
			"like --set but the value is parsed as JSON (numbers, booleans, nested objects)"))
	}
	if hasUploadPath(sc) {
		f.Bool(replaceFlag, false, i18n.T(
			"送り先に同名のアセットがあれば消してから送る（MT は上書きできません）",
			"delete any asset with the same name at the destination first (MT cannot overwrite)"))
	}
	if hasFileParam(sc) {
		f.String(dirFlag, "", i18n.T(
			"ディレクトリを丸ごと送る（--file の代わり。下の階層も送り先に写します）",
			"send a whole directory instead of --file (subdirectories are mirrored at the destination)"))
	}
	if isRebuild(sc.Ops[0]) {
		f.Bool(waitFlag, false, i18n.T(
			"restIds が空になるまで呼び直す（進まなくなったら、その旨を出して終了コード 8 で止まります）",
			"keep calling until restIds is empty (if it stops progressing, say so and exit with code 8)"))
	}
	if _, ok := isSweeping(sc.Ops[0]); ok {
		f.String(confirmFlag, "", i18n.T(
			"対象の名前。入れ物ごと壊す操作では --yes に加えて必要です",
			"name of the target; operations that destroy a whole container need this on top of --yes"))
	}
	f.StringArray(extraQueryFlag, nil, i18n.T(
		"定義に無いクエリを足す（key=value、値の @ファイル名 でファイルから読む、繰り返し可）",
		"add a query parameter the definition does not list (key=value; a value of @filename is read from that file; repeatable)"))
	f.StringArray(extraFormFlag, nil, i18n.T(
		"定義に無い本文パラメータを足す（key=value、値の @ファイル名 でファイルから読む、繰り返し可）",
		"add a body parameter the definition does not list (key=value; a value of @filename is read from that file; repeatable)"))
	return cmd
}

// canonicalPathParams は位置引数の並び順です。パスパラメータが最も多い
// オペレーションの順番をそのまま使います。
// hasServerFields はサーバ側に fields クエリがあるかを返します。
func hasServerFields(sc *spec.Command) bool {
	for _, op := range sc.Ops {
		for _, p := range op.Query {
			if p.Name == fieldsFlag {
				return true
			}
		}
	}
	return false
}

func canonicalPathParams(sc *spec.Command) []*spec.Param {
	var best []*spec.Param
	for _, op := range sc.Ops {
		if len(op.PathParams) > len(best) {
			best = op.PathParams
		}
	}
	return best
}

// allParams はコマンドが持つべきフラグを、重複を除いて返します。
func allParams(sc *spec.Command) []*spec.Param {
	seen := map[string]bool{}
	var out []*spec.Param
	add := func(ps []*spec.Param) {
		for _, p := range ps {
			if seen[p.Flag] {
				continue
			}
			seen[p.Flag] = true
			out = append(out, p)
		}
	}
	for _, op := range sc.Ops {
		add(op.PathParams)
	}
	for _, op := range sc.Ops {
		add(op.Query)
		add(op.Body)
	}
	return out
}

func usageLine(name string, path []*spec.Param) string {
	if len(path) == 0 {
		return name
	}
	parts := make([]string, 0, len(path)+1)
	parts = append(parts, name)
	for _, p := range path {
		parts = append(parts, "["+p.Flag+"]")
	}
	return strings.Join(parts, " ")
}

// flagUsage はフラグの説明です。API 側の名前がケバブケースと違うときは
// 併記します。リファレンスと突き合わせるときに要るためです。
func flagUsage(p *spec.Param) string {
	var b strings.Builder
	b.WriteString(oneLine(p.Describe()))
	var notes []string
	if p.Name != p.Flag {
		notes = append(notes, "API: "+p.Name)
	}
	if p.Type != "string" {
		note := p.Type
		// リソース本体は (object) としか出ないと、何を入れればよいか
		// 分からない。定義に書いてある項目名をそのまま並べる。
		if len(p.Fields) > 0 {
			note += ": " + strings.Join(p.Fields, ", ")
		}
		notes = append(notes, note)
	}
	if len(p.Enum) > 0 {
		notes = append(notes, strings.Join(p.Enum, "|"))
	}
	if p.Default != "" {
		notes = append(notes, i18n.T("既定 ", "default ")+p.Default)
	}
	if p.Required {
		notes = append(notes, i18n.T("必須", "required"))
	}
	if len(notes) > 0 {
		if b.Len() > 0 {
			b.WriteString(" ")
		}
		b.WriteString("(" + strings.Join(notes, ", ") + ")")
	}
	return b.String()
}

func longDescription(sc *spec.Command, canonical []*spec.Param) string {
	var b strings.Builder
	b.WriteString(sc.Short)
	b.WriteString("\n")
	for _, op := range sc.Ops {
		fmt.Fprintf(&b, "\n%s %s\n", op.Method, op.Path)
		if op.Deprecated {
			b.WriteString(i18n.T("（このエンドポイントは非推奨です）\n", "(this endpoint is deprecated)\n"))
		}
		if op.Description != "" {
			b.WriteString("\n" + strings.TrimSpace(op.Description) + "\n")
		}
		if op.DocURL != "" {
			fmt.Fprintf(&b, i18n.T("\n詳しい説明: %s\n", "\nFull description: %s\n"), op.DocURL)
		}
	}
	if len(sc.Ops) > 1 {
		b.WriteString(i18n.T(
			"\n末尾のIDを省くと、IDを取らない方のエンドポイントを呼びます。\n",
			"\nLeave off the trailing ID to call the endpoint that does not take one.\n"))
	}
	if len(canonical) > 0 {
		names := make([]string, len(canonical))
		for i, p := range canonical {
			names[i] = p.Flag
		}
		fmt.Fprintf(&b, i18n.T(
			"\n位置引数は --%s と同じ意味です（この順番）。\n",
			"\nThe positional arguments mean the same as --%s, in that order.\n"), strings.Join(names, " --"))
	}
	if hasObjectBody(sc) {
		b.WriteString(i18n.T(
			"\nリソース本体を取るフラグには JSON を渡します。@ファイル名 でファイルから、\n"+
				"- で標準入力から読み込めます。\n",
			"\nFlags that take a resource body expect JSON. Use @filename to read it from a\n"+
				"file, or - to read it from standard input.\n"))
	}
	return b.String()
}

func hasObjectBody(sc *spec.Command) bool {
	for _, op := range sc.Ops {
		for _, p := range op.Body {
			if p.Type == "object" || p.Type == "array" {
				return true
			}
		}
	}
	return false
}

func oneLine(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if i := strings.Index(s, "\n"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(strings.TrimPrefix(s, "- "))
}

// commandNames はツリー上の全コマンド名を返します。テストが重複を検出するために使います。
func commandNames(c *cobra.Command) []string {
	var out []string
	for _, sub := range c.Commands() {
		out = append(out, sub.Name())
		out = append(out, commandNames(sub)...)
	}
	sort.Strings(out)
	return out
}
