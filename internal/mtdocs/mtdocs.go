// Package mtdocs は Movable Type のテンプレートタグ・権限・ロール・アーカイブの
// 資料を保持します。ここから配布スキルの参照ファイル
// （plugins/movable-type-builder-cli/skills/*/references/*.md）を生成します。
// 生成器は internal/gen-skilldocs です。**参照ファイルを手編集しないこと。**
//
// **マニュアルの文章は持ちません。** movabletype.jp のマニュアルは Six Apart の
// 著作物で、再配布の許諾がありません。持つのは事実（タグ名・スラッグ・種類・
// 非推奨か・属性の書き方）だけで、説明は自前で書いた1行の要約（summaries.json）と、
// マニュアルの該当ページへのリンクにします。
//
// 原本の出どころ:
//
//	source/tags.json        movabletype.jp のテンプレートタグ索引（事実だけ）
//	source/modifiers.json   同 グローバルモディファイア索引（事実だけ）
//	source/permissions.json 実機の Movable Type（MT::Permission->perms）
//	source/roles.json       実機の Movable Type（mt_role テーブル）
//	source/archives.json    実機の Movable Type（archive_types レジストリ）
//	summaries.json          自前の要約（タグ・モディファイア・ロール）。手で書く
//
// MT 側のドキュメントが更新されたら、取り直してから生成し直します。新しい
// タグが増えたら summaries.json に要約を足すこと（無いとテストが落ちる）。
//
//	go run ./internal/gen-skilldocs -fetch   # movabletype.jp から取り直す（時間がかかる）
//	go generate ./...                        # 原本から参照ファイルを作り直す
package mtdocs

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed source/*.json summaries.json
var files embed.FS

// ManualBase はマニュアルのテンプレートタグ・モディファイアの置き場です。
const ManualBase = "https://www.movabletype.jp/documentation/appendices"

// Tag はテンプレートタグ1件です。
type Tag struct {
	Name       string   `json:"name"`
	Slug       string   `json:"slug"`
	Type       string   `json:"type"`
	Deprecated bool     `json:"deprecated"`
	Group      string   `json:"group"`
	Labels     []string `json:"labels"`
	// Attributes はタグ固有のモディファイアの書き方です（lastn="N" など）。
	Attributes []string `json:"attributes"`
	// Summary は自前の要約です（summaries.json）。
	Summary string `json:"-"`
}

// URL はマニュアルのそのタグのページです。
func (t Tag) URL() string { return ManualBase + "/tags/" + t.Slug + ".html" }

// Modifier はグローバルモディファイア1件です。
type Modifier struct {
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Summary string `json:"-"`
}

// URL はマニュアルのそのモディファイアのページです。
func (m Modifier) URL() string { return ManualBase + "/modifiers/" + m.Slug + ".html" }

// Permission は権限1件です。
type Permission struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Group string `json:"group"`
}

// Role は組み込みロール1件です。
type Role struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
	Summary     string   `json:"-"`
}

// ArchiveType はアーカイブ種別1件です。
type ArchiveType struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// Summaries は自前の要約です。
type Summaries struct {
	Tags      map[string]string `json:"tags"`
	Modifiers map[string]string `json:"modifiers"`
	Roles     map[string]string `json:"roles"`
}

var (
	summariesOnce sync.Once
	summaries     Summaries
	summariesErr  error
)

// LoadSummaries は summaries.json を読みます。
func LoadSummaries() (Summaries, error) {
	summariesOnce.Do(func() {
		buf, err := files.ReadFile("summaries.json")
		if err != nil {
			summariesErr = err
			return
		}
		summariesErr = json.Unmarshal(buf, &summaries)
	})
	return summaries, summariesErr
}

func load[T any](name string) ([]T, error) {
	buf, err := files.ReadFile("source/" + name)
	if err != nil {
		return nil, err
	}
	var v []T
	if err := json.Unmarshal(buf, &v); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return v, nil
}

// Tags はテンプレートタグの一覧です（要約つき）。
func Tags() ([]Tag, error) {
	tags, err := load[Tag]("tags.json")
	if err != nil {
		return nil, err
	}
	sum, err := LoadSummaries()
	if err != nil {
		return nil, err
	}
	for i := range tags {
		tags[i].Summary = sum.Tags[tags[i].Name]
	}
	return tags, nil
}

// Modifiers はグローバルモディファイアの一覧です（要約つき）。
func Modifiers() ([]Modifier, error) {
	mods, err := load[Modifier]("modifiers.json")
	if err != nil {
		return nil, err
	}
	sum, err := LoadSummaries()
	if err != nil {
		return nil, err
	}
	for i := range mods {
		mods[i].Summary = sum.Modifiers[mods[i].Name]
	}
	return mods, nil
}

// Roles は組み込みロールの一覧です（要約つき）。
func Roles() ([]Role, error) {
	roles, err := load[Role]("roles.json")
	if err != nil {
		return nil, err
	}
	sum, err := LoadSummaries()
	if err != nil {
		return nil, err
	}
	for i := range roles {
		roles[i].Summary = sum.Roles[roles[i].Name]
	}
	return roles, nil
}

// ArchiveTypes はアーカイブ種別の一覧です。
func ArchiveTypes() ([]ArchiveType, error) { return load[ArchiveType]("archives.json") }

// Permissions はスコープ（system / blog）ごとの権限一覧です。
func Permissions() (map[string][]Permission, error) {
	buf, err := files.ReadFile("source/permissions.json")
	if err != nil {
		return nil, err
	}
	var v map[string][]Permission
	if err := json.Unmarshal(buf, &v); err != nil {
		return nil, fmt.Errorf("permissions.json: %w", err)
	}
	return v, nil
}
