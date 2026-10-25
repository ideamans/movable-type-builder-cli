package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

func TestKebab(t *testing.T) {
	tests := map[string]string{
		"contentTypes":    "content-types",
		"categorySets":    "category-sets",
		"formatted_texts": "formatted-texts",
		"textFilters":     "text-filters",
		"IncludeBlogs":    "include-blogs",
		"SearchSortBy":    "search-sort-by",
		"site_id":         "site-id",
		"screenPageViews": "screen-page-views",
		"yearMonth":       "year-month",
		"entries":         "entries",
		"widgetsets":      "widgetsets",
		"Category Sets":   "category-sets",
	}
	for in, want := range tests {
		if got := kebab(in); got != want {
			t.Errorf("kebab(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCommandFor(t *testing.T) {
	tests := []struct {
		method string
		path   string
		want   string
	}{
		// コレクションと個別
		{"GET", "/sites/{site_id}/entries", "sites entries list"},
		{"POST", "/sites/{site_id}/entries", "sites entries create"},
		{"GET", "/sites/{site_id}/entries/{entry_id}", "sites entries get"},
		{"PUT", "/sites/{site_id}/entries/{entry_id}", "sites entries update"},
		{"DELETE", "/sites/{site_id}/entries/{entry_id}", "sites entries delete"},
		// 末尾が動詞
		{"POST", "/assets/upload", "assets upload"},
		{"GET", "/sites/{site_id}/entries/export", "sites entries export"},
		{"POST", "/sites/{site_id}/categories/permutate", "sites categories permutate"},
		{"POST", "/groups/{group_id}/permissions/grant", "groups permissions grant"},
		{"POST", "/users/{user_id}/recover_password", "users recover-password"},
		{"POST", "/recover_password", "recover-password"},
		// 末尾が下位リソース（動詞ではない）
		{"GET", "/sites/{site_id}/categories/{category_id}/children", "sites categories children list"},
		{"POST", "/sites/{site_id}/entries/{entry_id}/comments/{comment_id}/replies", "sites entries comments replies create"},
		// 名前を明示したもの
		{"GET", "/", "openapi"},
		{"GET", "/version", "server-version"},
		{"POST", "/sites/{site_id}", "sites create-blog"},
		{"DELETE", "/sites/{site_id}/logs", "sites logs reset"},
		{"GET", "/publish/entries", "publish entries"},
		{"GET", "/sites/{site_id}/stats/provider", "sites stats provider"},
		// キャメルケースのセグメント
		{"GET", "/sites/{site_id}/contentTypes/{content_type_id}/data", "sites content-types data list"},
		{"GET", "/sites/{site_id}/stats/yearMonth/sessions", "sites stats year-month sessions list"},
	}
	for _, tt := range tests {
		got := strings.Join(commandFor(tt.method, tt.path), " ")
		if got != tt.want {
			t.Errorf("commandFor(%s %s) = %q, want %q", tt.method, tt.path, got, tt.want)
		}
	}
}

func TestVerbFor(t *testing.T) {
	tests := []struct {
		method string
		item   bool
		want   string
	}{
		{"GET", false, "list"},
		{"GET", true, "get"},
		{"POST", false, "create"},
		{"PUT", true, "update"},
		{"PATCH", true, "update"},
		{"DELETE", true, "delete"},
	}
	for _, tt := range tests {
		if got := verbFor(tt.method, tt.item); got != tt.want {
			t.Errorf("verbFor(%s, item=%v) = %q, want %q", tt.method, tt.item, got, tt.want)
		}
	}
}

// 名前が衝突したときは黙って畳まず、生成を失敗させる。
func TestCheckMergeRejectsRealCollisions(t *testing.T) {
	t.Run("メソッドが違う", func(t *testing.T) {
		c := &spec.Command{Command: []string{"x"}, Ops: []*spec.Operation{
			{Method: "GET", Path: "/a"},
			{Method: "POST", Path: "/b"},
		}}
		if err := checkMerge(c); err == nil {
			t.Error("エラーになるべきです")
		}
	})
	t.Run("同じ形", func(t *testing.T) {
		c := &spec.Command{Command: []string{"x"}, Ops: []*spec.Operation{
			{Method: "GET", Path: "/a/{id}", PathParams: []*spec.Param{{Name: "id"}}},
			{Method: "GET", Path: "/b/{id}", PathParams: []*spec.Param{{Name: "id"}}},
		}}
		if err := checkMerge(c); err == nil {
			t.Error("エラーになるべきです")
		}
	})
	t.Run("末尾のIDの有無だけが違うなら畳んでよい", func(t *testing.T) {
		c := &spec.Command{Command: []string{"x"}, Ops: []*spec.Operation{
			{Method: "POST", Path: "/p/{id}/enable", PathParams: []*spec.Param{{Name: "id"}}},
			{Method: "POST", Path: "/p/enable"},
		}}
		if err := checkMerge(c); err != nil {
			t.Errorf("err = %v", err)
		}
	})
}

func TestGroupNaming(t *testing.T) {
	commands := []*spec.Command{
		{Command: []string{"sites", "list"}, Ops: []*spec.Operation{{Tags: []string{"Sites"}}}},
		{Command: []string{"sites", "entries", "list"}, Ops: []*spec.Operation{{Tags: []string{"Entries"}}}},
		{Command: []string{"sites", "permissions", "list"}, Ops: []*spec.Operation{{Tags: []string{"Sites", "Permissions"}}}},
		{Command: []string{"publish", "entries"}, Ops: []*spec.Operation{{Tags: []string{"Entries", "Publish"}}}},
		{Command: []string{"publish", "content-data"}, Ops: []*spec.Operation{{Tags: []string{"Content Data", "Publish"}}}},
	}
	got := map[string]string{}
	for _, g := range groupsOf(commands) {
		got[strings.Join(g.Command, " ")] = g.Short
	}
	want := map[string]string{
		"sites":             "Sites",
		"sites entries":     "Entries",
		"sites permissions": "Permissions", // タグ名がグループ名と一致する方を選ぶ
		"publish":           "Publish",     // 配下すべてに共通するタグ
	}
	for key, wantShort := range want {
		if got[key] != wantShort {
			t.Errorf("グループ %q の説明 = %q, want %q", key, got[key], wantShort)
		}
	}
}

// 公式定義に足りない本文パラメータが補われていること。
// PUT /sites/{site_id} は対象が blog なら blog を要求するが、定義は website
// しか書いていない。ここが抜けると子サイトを更新できない。
func TestExtraBodyParamsAreApplied(t *testing.T) {
	extra, ok := extraBody["PUT /sites/{site_id}"]
	if !ok {
		t.Fatal("PUT /sites/{site_id} の補正がありません")
	}
	var names []string
	for _, p := range extra {
		names = append(names, p.Name)
		if p.Flag == "" || p.Type == "" || p.Description == "" {
			t.Errorf("補正パラメータ %q に flag/type/description のいずれかがありません", p.Name)
		}
	}
	if strings.Join(names, ",") != "blog" {
		t.Errorf("補正パラメータ = %v", names)
	}
}

// MT は不正なカスタムフィールド型に The type "..." is invalid. としか返さない。
// 有効な値が --help に出ていないと、当てずっぽうで叩くことになる。
func TestParamNotesAreApplied(t *testing.T) {
	notes, ok := paramNotes["POST /sites/{site_id}/fields"]
	if !ok {
		t.Fatal("カスタムフィールド作成の注記がありません")
	}
	n := notes["field"]
	for _, want := range []string{"text", "textarea", "select", "systemObject", "options"} {
		if !strings.Contains(n.ja, want) {
			t.Errorf("和文の注記に %q がありません: %s", want, n.ja)
		}
		if !strings.Contains(n.en, want) {
			t.Errorf("英文の注記に %q がありません: %s", want, n.en)
		}
	}
}

// 再構築が非同期であることは、応答が Rebuilding としか言わないので
// --help に出ていないと気づけない。
func TestRebuildNoteIsApplied(t *testing.T) {
	n := paramNotes["GET /publish/entries"]["ids"]
	for _, want := range []string{"restIds", "呼び直す", "--wait", "更新"} {
		if !strings.Contains(n.ja, want) {
			t.Errorf("注記に %q がありません: %s", want, n.ja)
		}
	}
	for _, want := range []string{"restIds", "--wait", "updating"} {
		if !strings.Contains(n.en, want) {
			t.Errorf("英文の注記に %q がありません: %s", want, n.en)
		}
	}
}

// startTime の注記は、再構築を呼び直す両方のエンドポイントに付いていること。
//
// **これを渡さないと restIds はいつまでも減りません**（実測: 12回呼んでも
// 6件のまま。渡した途端に2回で空になった）。エラーにならず Rebuilding が
// 返り続けるので、注記が無いと自力では気づけません。
func TestRebuildStartTimeNoteIsApplied(t *testing.T) {
	for _, key := range []string{"GET /publish/entries", "GET /publish/contentData"} {
		n, ok := paramNotes[key]["startTime"]
		if !ok {
			t.Errorf("%s の startTime に注記がありません", key)
			continue
		}
		for _, want := range []string{"startTime", "restIds", "--wait"} {
			if !strings.Contains(n.ja, want) {
				t.Errorf("%s の注記に %q がありません: %s", key, want, n.ja)
			}
		}
		if !strings.Contains(n.en, "startTime") || !strings.Contains(n.en, "restIds") {
			t.Errorf("%s の英文の注記が足りません: %s", key, n.en)
		}
	}
}

// dataLabel の注記。**フィールド名を渡しても作成は通ってしまい**、
// 失敗するのはデータ作成のとき（409 "Data Label" is required.）。
// 文面から原因が分からないので、注記が消えると誰も辿り着けなくなる。
func TestContentTypeNoteIsApplied(t *testing.T) {
	for _, key := range []string{
		"POST /sites/{site_id}/contentTypes",
		"PUT /sites/{site_id}/contentTypes/{content_type_id}",
	} {
		n, ok := paramNotes[key]["content_type"]
		if !ok {
			t.Errorf("%s の content_type に注記がありません", key)
			continue
		}
		for _, want := range []string{"dataLabel", "ユニークID", "Data Label"} {
			if !strings.Contains(n.ja, want) {
				t.Errorf("%s の注記に %q がありません: %s", key, want, n.ja)
			}
		}
		if !strings.Contains(n.en, "uniqueID") {
			t.Errorf("%s の英文の注記に uniqueID がありません: %s", key, n.en)
		}
	}
}

// smartReplace の注記。既定で有効なので、何もしないとタイトルの
// em ダッシュが「--」に化ける。サイト側と記事側の両方から気づけるようにする。
func TestSmartReplaceNoteIsApplied(t *testing.T) {
	siteKeys := map[string]string{
		"POST /sites":                 "website",
		"POST /sites/{site_id}/blogs": "blog",
		"PUT /sites/{site_id}":        "website",
	}
	for key, param := range siteKeys {
		n, ok := paramNotes[key][param]
		if !ok {
			t.Errorf("%s の %s に注記がありません", key, param)
			continue
		}
		if !strings.Contains(n.ja, "smartReplace") || !strings.Contains(n.en, "smartReplace") {
			t.Errorf("%s の注記に smartReplace がありません", key)
		}
		// **1 にしても原文のままにはならない**（&#8212; になる）。実測。
		if !strings.Contains(n.ja, "&#8212;") {
			t.Errorf("%s の注記に、1 にしたときの実際の値がありません: %s", key, n.ja)
		}
	}
	// 症状が出るのは記事を入れるときなので、そちらにも書いてある。
	for _, key := range []string{
		"POST /sites/{site_id}/entries",
		"PUT /sites/{site_id}/entries/{entry_id}",
	} {
		n, ok := paramNotes[key]["entry"]
		if !ok {
			t.Errorf("%s の entry に注記がありません", key)
			continue
		}
		// 「作成のときだけ」が要点。1 にすれば直る、と読めては困る。
		for _, want := range []string{"作成", "更新では起きない"} {
			if !strings.Contains(n.ja, want) {
				t.Errorf("%s の注記に %q がありません: %s", key, want, n.ja)
			}
		}
	}
}

// カスタムフィールドの tag はテンプレートタグとしては生えない（MT 9 で実測）。
// fields list が tag を返すので、使えると思い込みやすい。
func TestCustomFieldNoteWarnsAboutTag(t *testing.T) {
	n := paramNotes["POST /sites/{site_id}/fields"]["field"]
	for _, want := range []string{"tag", "EntryCustomFieldValue"} {
		if !strings.Contains(n.ja, want) {
			t.Errorf("注記に %q がありません: %s", want, n.ja)
		}
		if !strings.Contains(n.en, want) {
			t.Errorf("英文の注記に %q がありません: %s", want, n.en)
		}
	}
}

// permutate の型補正が効いていること。string のままだと @ファイル名 が
// 展開されず、MT 側で 500 malformed JSON string になる。
func TestPermutateParamsAreArrays(t *testing.T) {
	want := []string{
		"POST /sites/{site_id}/categories/permutate",
		"POST /sites/{site_id}/folders/permutate",
		"POST /sites/{site_id}/categorySets/{category_set_id}/categories/permutate",
		"POST /sites/{site_id}/contentTypes/{content_type_id}/fields/permutate",
	}
	for _, key := range want {
		types, ok := paramTypes[key]
		if !ok {
			t.Errorf("%s の型補正がありません", key)
			continue
		}
		for name, typ := range types {
			if typ != "array" {
				t.Errorf("%s の %s = %q, want array", key, name, typ)
			}
		}
	}
}

// 注記は定数を書いただけでは効かない。paramNotes の表に載っていることまで見る
// （gofmt の整形で挿入が空振りし、定数だけ増えていたことがある）。
func TestParamNotesAreWiredUp(t *testing.T) {
	for _, tc := range []struct{ endpoint, param, want string }{
		{"POST /sites/{site_id}/entries/import", "encoding", "DefaultLanguage"},
		{"GET /search", "cdSearch", "SearchSortBy"},
		{"POST /sites/{site_id}/entries", "publish", "ステータスは変わらない"},
		{"POST /sites/{site_id}/pages", "publish", "ステータスは変わらない"},
		{"GET /sites/{site_id}/logs/export", "encoding", "DefaultLanguage"},
	} {
		n := paramNotes[tc.endpoint][tc.param]
		if !strings.Contains(n.ja, tc.want) {
			t.Errorf("%s の %s に注記がありません（%q）", tc.endpoint, tc.param, n.ja)
		}
	}
}

// 注記は英語環境でも出る。片方だけ書いて済ませると、英語では説明が
// 丸ごと消えるのではなく日本語のまま出る（Describe が和文に落ちるため）。
func TestParamNotesAreBilingual(t *testing.T) {
	for endpoint, notes := range paramNotes {
		for param, n := range notes {
			if n.ja == "" || n.en == "" {
				t.Errorf("%s の %s の注記が片方だけです（ja=%q en=%q）", endpoint, param, n.ja, n.en)
			}
		}
	}
	for endpoint, params := range extraBody {
		for _, p := range params {
			if p.DescriptionEn == "" {
				t.Errorf("%s の %s に英語の説明がありません", endpoint, p.Name)
			}
		}
	}
}

// 小さな OpenAPI を実際に通し、パラメータの取り出しと補正を確かめる。
// 命名規則は commandFor で見ているが、$ref の解決・本文の型・enum や既定値の
// 取り出し・補正表の適用は、ここまで生成物経由でしか見ていなかった。
func TestRunOnFixture(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "openapi.json")
	out := filepath.Join(dir, "spec.json")
	if err := os.WriteFile(in, []byte(fixtureOpenAPI), 0o644); err != nil {
		t.Fatal(err)
	}

	oldIn, oldOut := *inPath, *outPath
	*inPath, *outPath = in, out
	defer func() { *inPath, *outPath = oldIn, oldOut }()

	if err := run(); err != nil {
		t.Fatal(err)
	}
	buf, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var s spec.Spec
	if err := json.Unmarshal(buf, &s); err != nil {
		t.Fatal(err)
	}

	byName := map[string]*spec.Command{}
	for _, c := range s.Commands {
		byName[strings.Join(c.Command, " ")] = c
	}

	list, ok := byName["sites entries list"]
	if !ok {
		t.Fatalf("コマンドが作られていません: %v", byName)
	}
	op := list.Ops[0]

	// パスパラメータ: 説明が空でも補われ、必須になる
	if len(op.PathParams) != 1 || op.PathParams[0].Flag != "site-id" {
		t.Fatalf("パスパラメータ = %+v", op.PathParams)
	}
	if !op.PathParams[0].Required || op.PathParams[0].Description == "" {
		t.Errorf("site_id = %+v", op.PathParams[0])
	}

	// $ref のクエリが解決され、ケバブ化・enum・既定値が取れていること
	q := map[string]*spec.Param{}
	for _, p := range op.Query {
		q[p.Name] = p
	}
	sortBy, ok := q["sortBy"]
	if !ok {
		t.Fatalf("$ref のクエリが解決されていません: %v", q)
	}
	if sortBy.Flag != "sort-by" {
		t.Errorf("flag = %q, want sort-by", sortBy.Flag)
	}
	if strings.Join(sortBy.Enum, ",") != "title,created_on" {
		t.Errorf("enum = %v", sortBy.Enum)
	}
	if sortBy.Default != "title" {
		t.Errorf("default = %q", sortBy.Default)
	}
	if q["limit"].Type != "integer" {
		t.Errorf("limit の型 = %q", q["limit"].Type)
	}
	// 認証ヘッダはフラグにしない
	if _, bad := q["X-MT-Authorization"]; bad {
		t.Error("ヘッダがクエリに混ざっています")
	}

	// 本文: $ref はオブジェクト、binary はファイル
	create := byName["sites entries create"]
	body := map[string]*spec.Param{}
	for _, p := range create.Ops[0].Body {
		body[p.Name] = p
	}
	if body["entry"].Type != "object" {
		t.Errorf("entry の型 = %q", body["entry"].Type)
	}
	if body["file"].Type != "file" {
		t.Errorf("file の型 = %q", body["file"].Type)
	}
	if create.Ops[0].BodyKind != "multipart" {
		t.Errorf("bodyKind = %q", create.Ops[0].BodyKind)
	}

	// 補正表が効いていること（型の直しと注記）
	perm := byName["sites folders permutate"]
	pb := perm.Ops[0].Body[0]
	if pb.Type != "array" {
		t.Errorf("folders の型 = %q, want array（paramTypes の補正）", pb.Type)
	}
	if !strings.Contains(pb.Description, "並べ替えた後") {
		t.Errorf("folders に注記がありません: %q", pb.Description)
	}
}

const fixtureOpenAPI = `{
 "info": {"version": "9.9.9"},
 "components": {
  "parameters": {
   "entry_sortBy": {"name": "sortBy", "in": "query", "description": "並び替え",
     "schema": {"type": "string", "default": "title", "enum": ["title", "created_on"]}},
   "entry_limit": {"name": "limit", "in": "query", "schema": {"type": "integer", "default": 10}}
  },
  "schemas": {"entry_updatable": {"type": "object", "properties": {"title": {"type": "string"}}}}
 },
 "paths": {
  "/sites/{site_id}/entries": {
   "get": {"tags": ["Entries"], "summary": "一覧", "parameters": [
     {"name": "site_id", "in": "path", "required": true, "schema": {"type": "integer"}},
     {"name": "X-MT-Authorization", "in": "header", "schema": {"type": "string"}},
     {"$ref": "#/components/parameters/entry_sortBy"},
     {"$ref": "#/components/parameters/entry_limit"}]},
   "post": {"tags": ["Entries"], "summary": "作成", "parameters": [
     {"name": "site_id", "in": "path", "required": true, "schema": {"type": "integer"}}],
    "requestBody": {"required": true, "content": {"multipart/form-data": {"schema": {"type": "object",
      "required": ["entry"],
      "properties": {"entry": {"$ref": "#/components/schemas/entry_updatable"},
                     "file": {"type": "string", "format": "binary"}}}}}}}
  },
  "/sites/{site_id}/folders/permutate": {
   "post": {"tags": ["Folders"], "summary": "並べ替え", "parameters": [
     {"name": "site_id", "in": "path", "required": true, "schema": {"type": "integer"}}],
    "requestBody": {"content": {"application/x-www-form-urlencoded": {"schema": {"type": "object",
      "properties": {"folders": {"type": "string", "description": "Array of folder resource."}}}}}}}
  }
 }
}`

// 配列の選択肢は items の側にある。拾い損ねると、渡せる値が --help の
// どこにも出ない（MT もエラーでは教えてくれない）。
func TestArrayItemEnumIsCarried(t *testing.T) {
	sp, err := spec.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range sp.Commands {
		for _, op := range c.Ops {
			if op.Path != "/sites/{site_id}/export_theme" {
				continue
			}
			for _, p := range op.Body {
				if p.Name != "include" {
					continue
				}
				if len(p.Enum) == 0 {
					t.Fatal("--include の選択肢が落ちています")
				}
				for _, want := range []string{"blog_static_files", "template_set"} {
					if !slices.Contains(p.Enum, want) {
						t.Errorf("--include の選択肢に %q がありません: %v", want, p.Enum)
					}
				}
				return
			}
		}
	}
	t.Fatal("export_theme の include が見つかりません")
}
