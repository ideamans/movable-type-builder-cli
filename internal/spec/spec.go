// Package spec は Movable Type Data API v7 の OpenAPI 定義を、CLI が必要と
// する形だけに絞り込んだものを埋め込み、コマンドツリーの素として返します。
//
// 原本は openapi-v7.json（Six Apart 公式の
// movabletype/mt-docs-data-api-reference に置かれているもの）で、
// spec.json はそこから `go generate ./...` で作る派生物です。手編集しないこと。
// go:generate の指定は cmd/mtb/main.go にあります（spec.json を作ってから
// コマンドカタログを描く順序を保つため、生成の指定を1か所に集めています）。
// 埋め込みは spec.json のみで、原本はビルドに入りません。
package spec

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/ideamans/movable-type-builder-cli/internal/i18n"
)

//go:embed spec.json
var raw []byte

// Spec は絞り込み済みの API 定義です。
type Spec struct {
	// APIVersion はエンドポイントのバージョン識別子です（"v7"）。
	APIVersion string `json:"apiVersion"`
	// SpecVersion は原本の info.version（MT 本体のバージョン）です。
	SpecVersion string `json:"specVersion"`
	// Groups は中間ノード（コマンドの入れ物）です。
	Groups []*Group `json:"groups"`
	// Commands は末端のコマンド1つ分です。
	Commands []*Command `json:"commands"`
}

// Group はサブコマンドをぶら下げるだけの中間ノードです。
type Group struct {
	Command []string `json:"command"`
	Short   string   `json:"short"`
}

// Command は CLI のコマンド1つと、それが呼ぶ API オペレーションです。
//
// Ops が2件になるのは「末尾のパスパラメータがあってもなくても同じ操作」
// という API（POST /plugins/enable と POST /plugins/{plugin_id}/enable など）
// を1コマンドに畳んだときだけで、パスパラメータの多い順に並びます。
type Command struct {
	Command []string     `json:"command"`
	Short   string       `json:"short"`
	Ops     []*Operation `json:"ops"`
}

// Name はコマンド名（コマンドパスの末尾）です。
func (c *Command) Name() string { return c.Command[len(c.Command)-1] }

// Operation は API オペレーション1件です。
type Operation struct {
	Method  string   `json:"method"`
	Path    string   `json:"path"`
	Tags    []string `json:"tags,omitempty"`
	Summary string   `json:"summary,omitempty"`
	// Description は説明文です。公式定義の説明文は持ち込まない（gen-spec の
	// パッケージ説明を参照）ので、埋め込みの定義では空です。拡張が自前で
	// 書いた説明だけがここに入ります。
	Description string `json:"description,omitempty"`
	// DocURL は公式リファレンスのその操作の箇所です。--help に出します。
	DocURL     string   `json:"docUrl,omitempty"`
	Deprecated bool     `json:"deprecated,omitempty"`
	PathParams []*Param `json:"pathParams,omitempty"`
	Query      []*Param `json:"query,omitempty"`
	Body       []*Param `json:"body,omitempty"`
	// BodyKind は "form"（application/x-www-form-urlencoded）か
	// "multipart"（multipart/form-data）か、本文を取らないなら空です。
	BodyKind string `json:"bodyKind,omitempty"`
}

// Param はパス・クエリ・本文のパラメータ1つです。
type Param struct {
	// Name は API 側の名前です（sortBy など）。
	Name string `json:"name"`
	// Flag は CLI のフラグ名です（sort-by など）。Name をケバブケースにしたもの。
	Flag string `json:"flag"`
	// Type は string / integer / number / boolean / object / array / file のいずれか。
	Type string `json:"type"`
	// Description は mtb が補った注記です（公式定義の説明文は持ち込まない）。
	// 注記の無いパラメータでは空で、--help には名前・型・選択肢だけが出ます。
	Description string `json:"description,omitempty"`
	// DescriptionEn は Description の英語版です。補った注記が無い
	// パラメータでは空で、そのときは Description をそのまま使います。
	DescriptionEn string   `json:"descriptionEn,omitempty"`
	Required      bool     `json:"required,omitempty"`
	Default       string   `json:"default,omitempty"`
	Enum          []string `json:"enum,omitempty"`
	// Repeated は「同じ名前で繰り返して送る」配列です。MT は選択肢の
	// 並びを include=a&include=b の形で受け取ります。JSON の配列を1つの
	// 値として送ると、**エラーにならずに黙って無視されます**（テーマの
	// エクスポートが success を返して空のテーマを作る、という形で出ます）。
	Repeated bool `json:"repeated,omitempty"`
	// Fields はリソース本体（object）が取る項目の名前です。定義には
	// 書いてあるのに --help では (object) としか出ず、何を入れればよいか
	// 分からなかったので持たせています。--set の指定先でもあります。
	Fields []string `json:"fields,omitempty"`
}

// Describe は表示言語に合わせた説明です。
func (p *Param) Describe() string {
	if p.DescriptionEn == "" {
		return p.Description
	}
	return i18n.T(p.Description, p.DescriptionEn)
}

var (
	once   sync.Once
	loaded *Spec
	lofErr error
)

// Load は埋め込み済みの定義を返します。2回目以降はキャッシュを返します。
func Load() (*Spec, error) {
	once.Do(func() {
		s := &Spec{}
		if err := json.Unmarshal(raw, s); err != nil {
			lofErr = fmt.Errorf(i18n.T("埋め込みAPI定義を読めません: %w", "cannot read the embedded API definition: %w"), err)
			return
		}
		loaded = s
	})
	return loaded, lofErr
}

// MustLoad は Load に失敗したら panic します。埋め込みが壊れているのは
// ビルド時の事故であり、実行時に回復する余地がないためです。
func MustLoad() *Spec {
	s, err := Load()
	if err != nil {
		panic(err)
	}
	return s
}
