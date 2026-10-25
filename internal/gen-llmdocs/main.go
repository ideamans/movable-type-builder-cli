// Command gen-llmdocs は埋め込みリファレンスの生成章を作り直します。
// go generate から呼ばれ、CI でも再生成してコミット済みの成果物がコマンド
// ツリーと一致していることを確かめます。
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ideamans/go-llm-cli-kit/catalog"

	"github.com/ideamans/movable-type-builder-cli/internal/cmd"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

var outPath = flag.String("out", "90-commands.md", "書き出す先")

func main() {
	flag.Parse()

	// 埋め込みリファレンスは手書きの章も日本語なので、生成する章も日本語に
	// 揃えます（生成器を動かす環境の言語で中身が入れ替わらないように）。
	i18n.SetLang(i18n.JA)

	md := catalog.Markdown(cmd.Root(), catalog.Options{
		Title: "コマンドカタログ",
		Intro: "cobra のコマンド定義から `go generate ./...` で生成しています。\n" +
			"リソース操作のコマンドは internal/spec の OpenAPI 定義から作られるため、\n" +
			"直すのは定義か生成器であって、このファイルではありません。",
		Skip:    []string{"llm"},
		Compact: true,
	})

	if err := os.WriteFile(*outPath, []byte(md), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "gen-llmdocs:", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d bytes)\n", *outPath, len(md))
}
