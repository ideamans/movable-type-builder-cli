package llmdocs

import (
	"fmt"
	"strings"

	kit "github.com/ideamans/go-llm-cli-kit/llmdocs"

	"github.com/ideamans/movable-type-builder-cli/internal/extension"
)

// WithExtensions は埋め込みリファレンスに拡張の章を足したものです。
//
// 章は日本語で固定します。手書きの章も生成されたコマンドカタログも日本語なので、
// 表示言語で拡張の章だけが入れ替わると、1つのリファレンスの中で言語が混ざります
// （gen-llmdocs が生成物の言語を固定しているのと同じ理由）。
func WithExtensions(exts []*extension.Extension) *kit.Docs {
	if len(exts) == 0 {
		return Docs()
	}
	extra := map[string]string{}
	for _, e := range exts {
		extra["80-extension-"+e.Name+".md"] = ExtensionChapter(e)
	}
	return DocsWith(extra)
}

// ExtensionChapter は拡張1つ分の章です。拡張の llm.md の前に、どこから来た
// コマンドで、MT 側に何が要るのかを機械的に書き足します。
func ExtensionChapter(e *extension.Extension) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# 拡張: %s\n\n", e.Name)
	desc := e.DescriptionJa
	if desc == "" {
		desc = e.Description
	}
	fmt.Fprintf(&b, "%s（%s %s）\n\n", desc, e.Name, e.Version)
	b.WriteString("以下のコマンドは MT 本体の Data API には無く、この拡張が足したもの。\n")
	if p := e.RequiredPlugin(); p != nil {
		fmt.Fprintf(&b, "**MT 側に %s プラグインが要る。** 入っていなければ 404 になる。"+
			"`mtb extensions check %s` で確かめられる。\n", p.Label(), e.Name)
	}
	b.WriteString("\n")
	for _, sc := range e.Spec.Commands {
		fmt.Fprintf(&b, "- `mtb %s` — %s\n", strings.Join(sc.Command, " "), sc.Short)
		for _, op := range sc.Ops {
			fmt.Fprintf(&b, "  - `%s %s`\n", op.Method, op.Path)
		}
	}
	if doc := strings.TrimSpace(e.LLMDoc); doc != "" {
		b.WriteString("\n" + demoteHeadings(doc) + "\n")
	}
	return b.String()
}

// demoteHeadings は拡張の章の見出しを1段下げます。章の題は mtb が付けるので、
// 拡張の書いた # が他の章と同じ階層に並ばないようにします。コードブロックの
// 中（シェルのコメント）は触りません。
func demoteHeadings(md string) string {
	lines := strings.Split(md, "\n")
	inFence := false
	for i, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "```") {
			inFence = !inFence
			continue
		}
		if !inFence && strings.HasPrefix(l, "#") {
			lines[i] = "#" + l
		}
	}
	return strings.Join(lines, "\n")
}
