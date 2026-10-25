package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/ideamans/go-llm-cli-kit/llmcmd"

	"github.com/ideamans/movable-type-builder-cli/internal/llmdocs"
)

func TestEmbeddedReference(t *testing.T) {
	g, err := llmdocs.Docs().Markdown()
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}
	for _, want := range []string{
		"Webサービスパスワード",     // 認証で最も多い詰まり方
		"MTB_ACCESS_TOKEN", // 対話できない環境からの経路
		"mt-data-api.cgi",  // ベースURLの落とし穴
		"--yes",            // 破壊的操作
		"--dry-run",
		"totalResults", // 一覧が10件で切れる件
		"コマンド名とエンドポイントの対応", // 2章目
		"コマンドカタログ",         // 生成物
		"mtb sites entries list",
	} {
		if !strings.Contains(g, want) {
			t.Errorf("埋め込みリファレンスに %q がありません", want)
		}
	}
}

func TestChapterOrder(t *testing.T) {
	sections, err := llmdocs.Docs().Sections()
	if err != nil {
		t.Fatalf("Sections: %v", err)
	}
	var files []string
	for _, s := range sections {
		files = append(files, s.File)
	}
	want := "00-guide.md,10-endpoints.md,90-commands.md"
	if got := strings.Join(files, ","); got != want {
		t.Errorf("chapters = %s, want %s", got, want)
	}
}

// --llm は従来どおりコマンドラインのどの位置でも効く。
func TestLegacyLLMFlag(t *testing.T) {
	for _, args := range [][]string{{"--llm"}, {"sites", "entries", "list", "--llm"}} {
		var out bytes.Buffer
		handled, err := llmcmd.HandleLegacy(args, LLMConfig(), &out)
		if err != nil {
			t.Fatalf("HandleLegacy(%v): %v", args, err)
		}
		if !handled {
			t.Errorf("HandleLegacy(%v) が --llm を処理しませんでした", args)
		}
		if !strings.Contains(out.String(), "Webサービスパスワード") {
			t.Errorf("HandleLegacy(%v) の出力が想定と違います", args)
		}
	}
}

// カタログは実際のコマンドツリーを反映していること。
func TestCatalogCoversGeneratedCommands(t *testing.T) {
	g, err := llmdocs.Docs().Markdown()
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"mtb sites entries create",
		"mtb sites assets upload",
		"mtb profile add",
		"mtb auth login",
		"mtb api",
	} {
		if !strings.Contains(g, want) {
			t.Errorf("カタログに %q がありません。go generate ./... を実行してください", want)
		}
	}
}
