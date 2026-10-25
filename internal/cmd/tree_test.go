package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

func testRoot(t *testing.T) *cobra.Command {
	t.Helper()
	return New(&Env{Spec: spec.MustLoad()}).Root()
}

// 定義にあるエンドポイントが、1つ残らずコマンドとして生えていること。
// 取りこぼしても --help には出ないので、数で押さえる。
func TestEveryCommandInSpecExists(t *testing.T) {
	root := testRoot(t)
	s := spec.MustLoad()
	for _, sc := range s.Commands {
		cmd, _, err := root.Find(sc.Command)
		if err != nil {
			t.Errorf("mtb %s が見つかりません: %v", strings.Join(sc.Command, " "), err)
			continue
		}
		if cmd.Name() != sc.Name() {
			t.Errorf("mtb %s が %s に解決されました", strings.Join(sc.Command, " "), cmd.CommandPath())
		}
		if !cmd.Runnable() {
			t.Errorf("mtb %s が実行できません", strings.Join(sc.Command, " "))
		}
	}
}

// 兄弟に同じ名前や別名があると、片方が黙って呼べなくなる。
func TestNoDuplicateNamesAmongSiblings(t *testing.T) {
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		seen := map[string]string{}
		for _, sub := range c.Commands() {
			for _, name := range append([]string{sub.Name()}, sub.Aliases...) {
				if prev, ok := seen[name]; ok {
					t.Errorf("%s の下で %q が %s と %s に重複しています", c.CommandPath(), name, prev, sub.Name())
				}
				seen[name] = sub.Name()
			}
			walk(sub)
		}
	}
	walk(testRoot(t))
}

// 生成コマンドのフラグがグローバルフラグを覆い隠していないこと。
// 覆い隠すと --format などが静かに効かなくなる。
func TestGeneratedFlagsDoNotShadowGlobals(t *testing.T) {
	root := testRoot(t)
	globalNames := map[string]bool{}
	root.PersistentFlags().VisitAll(func(f *pflag.Flag) { globalNames[f.Name] = true })

	// llm は go-llm-cli-kit が持ち込むコマンドで、--format は
	// markdown|json（章の出し方）を意味します。API 応答の整形とは別物なので
	// 覆っていて構いません。
	exempt := map[string]bool{"mtb llm": true}

	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			if !exempt[sub.CommandPath()] {
				sub.LocalNonPersistentFlags().VisitAll(func(f *pflag.Flag) {
					if globalNames[f.Name] {
						t.Errorf("%s のフラグ --%s がグローバルフラグを覆っています", sub.CommandPath(), f.Name)
					}
				})
			}
			walk(sub)
		}
	}
	walk(root)
}

// 単数形の別名が使えること（打ちやすさのため）。
func TestSingularAliases(t *testing.T) {
	root := testRoot(t)
	for _, path := range [][]string{
		{"sites", "entry", "list"},
		{"site", "entries", "list"},
		{"user", "list"},
		{"sites", "category", "list"},
	} {
		if _, _, err := root.Find(path); err != nil {
			t.Errorf("mtb %s が使えません: %v", strings.Join(path, " "), err)
		}
	}
}

func TestAliasesFor(t *testing.T) {
	tests := map[string][]string{
		"entries":    {"entry"},
		"categories": {"category"},
		"sites":      {"site"},
		"children":   nil,
		"data":       nil,
		"stats":      {"stat"},
		"list":       nil,
		"get":        nil,
	}
	for in, want := range tests {
		got := aliasesFor(in)
		if len(got) != len(want) || (len(want) == 1 && got[0] != want[0]) {
			t.Errorf("aliasesFor(%q) = %v, want %v", in, got, want)
		}
	}
}

// 手書きのコマンドが生成ツリーに埋もれていないこと。
func TestHandWrittenCommandsPresent(t *testing.T) {
	root := testRoot(t)
	for _, path := range [][]string{
		{"auth", "login"}, {"auth", "status"}, {"auth", "refresh"}, {"auth", "logout"},
		{"profile", "list"}, {"profile", "add"}, {"profile", "use"}, {"profile", "show"},
		{"profile", "set"}, {"profile", "remove"},
		{"api"}, {"llm"},
	} {
		if _, _, err := root.Find(path); err != nil {
			t.Errorf("mtb %s がありません: %v", strings.Join(path, " "), err)
		}
	}
}
