// Command mtb は Movable Type Data API のコマンドラインクライアントです。
package main

// 生成物の作り直しはこの2行が受け持ちます（ネットワークは使いません）。
// spec.json はここでは作りません。公式定義をネットワークから取ってくるので、
// `make spec` で作り直してから `go generate ./...` を流します（spec.json が
// コマンドツリーの素なので、この順でないとカタログが古いツリーから描かれる）。
//go:generate go run ../../internal/gen-llmdocs -out ../../internal/llmdocs/90-commands.md
//go:generate go run ../../internal/gen-skilldocs -out ../../plugins/movable-type-builder-cli/skills -source ../../internal/mtdocs/source

import (
	"fmt"
	"os"

	"github.com/ideamans/go-llm-cli-kit/llmcmd"

	"github.com/ideamans/movable-type-builder-cli/internal/cmd"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

func main() {
	cmd.SetLang(os.Args[1:], os.Getenv)
	env, err := cmd.DefaultEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.T("エラー:", "error:"), err)
		os.Exit(cmd.ExitError)
	}
	os.Exit(run(env, os.Args[1:]))
}

// run はコマンドライン1回分の処理です。終了コードを返すだけで、
// プロセスは終わらせません。配線（--llm の先回り、エラーの出し方、
// 終了コードの決め方）をテストから確かめられるように分けてあります。
func run(env *cmd.Env, args []string) int {
	// 非推奨の --llm はコマンドラインのどの位置でも効かせます。
	// cobra に入る前に処理しないとサブコマンド解決が先に走ってしまいます。
	if handled, err := llmcmd.HandleLegacy(args, cmd.LLMConfig(), env.Out); handled {
		if err != nil {
			fmt.Fprintln(env.Err, i18n.T("エラー:", "error:"), err)
			return cmd.ExitError
		}
		return cmd.ExitOK
	}

	if err := cmd.New(env, args...).Execute(args); err != nil {
		fmt.Fprintln(env.Err, i18n.T("エラー:", "error:"), err)
		return cmd.ExitCode(err)
	}
	return cmd.ExitOK
}
