package cmd

import (
	"testing"

	"github.com/ideamans/go-llm-cli-kit/skillcheck"
)

// TestPluginSkills は配布するプラグインを検証します。マニフェストの
// バージョンが PluginVersion に追随していること、SKILL.md の frontmatter が
// Agent Skills 標準の範囲に収まっていること（同じファイルが gh skill 経由で
// Copilot / Cursor / Gemini CLI にも入るため）、そして「利用者がこの道具を
// 必要とするときに口にする語」が description に残っていることを見ます。
func TestPluginSkills(t *testing.T) {
	report := skillcheck.CheckDir("../../plugins/movable-type-builder-cli", skillcheck.Options{
		Version:             PluginVersion,
		Keywords:            []string{"movable type", "data api", "entry", "install", "template", "content type"},
		RequireInstallSkill: true,
	})
	for _, problem := range report.Problems {
		t.Error(problem)
	}
}
