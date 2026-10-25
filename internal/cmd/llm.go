package cmd

import (
	"fmt"

	"github.com/ideamans/go-llm-cli-kit/llmcmd"

	"github.com/ideamans/movable-type-builder-cli/internal/extension"
	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
	"github.com/ideamans/movable-type-builder-cli/internal/llmdocs"
)

// LLMConfig は `mtb llm` サブコマンドの設定です。main からも参照し、
// 非推奨の --llm フラグをコマンドラインのどの位置でも受け付けるために使います。
//
// 読み込めた拡張があれば、その章を足します。拡張のコマンドは生成された
// コマンドカタログ（90-commands.md）に載らないので、拡張が足したことを
// エージェントが知る手段はここしかありません。
func LLMConfig(exts ...*extension.Extension) llmcmd.Config {
	return llmcmd.Config{
		Docs:  llmdocs.WithExtensions(exts),
		Short: i18n.T("AIエージェント向けの詳細リファレンスを表示する", "Print the detailed reference written for AI agents"),
	}
}

// warnExtensionErrors は読み込めなかった拡張を標準エラーに知らせます。
func (a *App) warnExtensionErrors() {
	if a.env.Err == nil {
		return
	}
	for _, err := range append(append([]error(nil), a.env.ExtensionErrors...), a.extensionErrors...) {
		fmt.Fprintln(a.env.Err, i18n.T("警告:", "warning:"), err)
	}
}
