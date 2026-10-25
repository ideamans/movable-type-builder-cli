package spec_test

import (
	"strings"
	"testing"

	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

func TestLoad(t *testing.T) {
	s, err := spec.Load()
	if err != nil {
		t.Fatal(err)
	}
	if s.APIVersion != "v7" {
		t.Errorf("APIVersion = %q, want v7", s.APIVersion)
	}
	if s.SpecVersion == "" {
		t.Error("SpecVersion が空です")
	}
	if len(s.Commands) < 200 {
		t.Errorf("コマンド数 = %d。取り込みが欠けている可能性があります", len(s.Commands))
	}
}

// コマンド名は一意で、必ず1つ以上のオペレーションを持つこと。
func TestCommandInvariants(t *testing.T) {
	s := spec.MustLoad()
	seen := map[string]bool{}
	for _, c := range s.Commands {
		key := strings.Join(c.Command, " ")
		if seen[key] {
			t.Errorf("コマンド名が重複しています: %s", key)
		}
		seen[key] = true

		if len(c.Command) == 0 {
			t.Error("コマンドパスが空です")
			continue
		}
		if len(c.Ops) == 0 {
			t.Errorf("%s にオペレーションがありません", key)
			continue
		}
		if c.Short == "" {
			t.Errorf("%s に説明がありません", key)
		}
		for _, op := range c.Ops {
			if op.Method == "" || op.Path == "" {
				t.Errorf("%s のオペレーションが不完全です: %+v", key, op)
			}
			for _, p := range op.PathParams {
				if !strings.Contains(op.Path, "{"+p.Name+"}") {
					t.Errorf("%s: パスパラメータ %s が %s に現れません", key, p.Name, op.Path)
				}
			}
		}
	}
}

// 1コマンドに複数のオペレーションが載るのは、末尾のパスパラメータの有無だけが
// 違う場合に限る。ここが崩れると、意図しないエンドポイントが呼ばれる。
func TestMergedOperationsDifferOnlyByTrailingPathParam(t *testing.T) {
	for _, c := range spec.MustLoad().Commands {
		if len(c.Ops) < 2 {
			continue
		}
		key := strings.Join(c.Command, " ")
		for _, op := range c.Ops[1:] {
			if op.Method != c.Ops[0].Method {
				t.Errorf("%s: メソッドが違う操作が畳まれています（%s と %s）", key, c.Ops[0].Method, op.Method)
			}
			if len(op.PathParams) == len(c.Ops[0].PathParams) {
				t.Errorf("%s: 同形の操作が畳まれています（%s と %s）", key, c.Ops[0].Path, op.Path)
			}
		}
		// パスパラメータの多い順に並んでいること（実行時の選択がこれに依存する）
		for i := 1; i < len(c.Ops); i++ {
			if len(c.Ops[i-1].PathParams) < len(c.Ops[i].PathParams) {
				t.Errorf("%s: オペレーションの並び順が違います", key)
			}
		}
	}
}

// パラメータのフラグ名は、同じコマンド内で衝突しないこと。
func TestParamFlagsAreUniquePerCommand(t *testing.T) {
	for _, c := range spec.MustLoad().Commands {
		seen := map[string]string{}
		for _, op := range c.Ops {
			for _, group := range [][]*spec.Param{op.PathParams, op.Query, op.Body} {
				for _, p := range group {
					if p.Flag == "" {
						t.Errorf("%s: %s にフラグ名がありません", strings.Join(c.Command, " "), p.Name)
					}
					if prev, ok := seen[p.Flag]; ok && prev != p.Name {
						t.Errorf("%s: フラグ --%s が %s と %s に重複しています",
							strings.Join(c.Command, " "), p.Flag, prev, p.Name)
					}
					seen[p.Flag] = p.Name
				}
			}
		}
	}
}

// 認証エンドポイントは手書きの mtb auth が受け持つので、生成しない。
func TestAuthenticationEndpointsAreNotGenerated(t *testing.T) {
	for _, c := range spec.MustLoad().Commands {
		for _, op := range c.Ops {
			switch op.Path {
			case "/authentication", "/token", "/authorization":
				t.Errorf("%s が生成されています（mtb auth が受け持つはずです）", op.Path)
			}
		}
	}
}

// 代表的なエンドポイントが期待どおりの名前になっていること。
// 命名規則を変えたらここが落ちるのが正しい挙動です。
func TestRepresentativeCommandNames(t *testing.T) {
	want := map[string]string{
		"GET /sites/{site_id}/entries":                           "sites entries list",
		"POST /sites/{site_id}/entries":                          "sites entries create",
		"GET /sites/{site_id}/entries/{entry_id}":                "sites entries get",
		"PUT /sites/{site_id}/entries/{entry_id}":                "sites entries update",
		"DELETE /sites/{site_id}/entries/{entry_id}":             "sites entries delete",
		"GET /sites/{site_id}/entries/export":                    "sites entries export",
		"POST /sites/{site_id}/assets/upload":                    "sites assets upload",
		"GET /sites/{site_id}/categories/{category_id}/children": "sites categories children list",
		"POST /sites/{site_id}/themes/{theme_id}/apply":          "sites themes apply",
		"GET /search":                  "search",
		"GET /version":                 "server-version",
		"GET /":                        "openapi",
		"POST /sites/{site_id}":        "sites create-blog",
		"DELETE /sites/{site_id}/logs": "sites logs reset",
		"GET /publish/entries":         "publish entries",
	}
	got := map[string]string{}
	for _, c := range spec.MustLoad().Commands {
		for _, op := range c.Ops {
			got[op.Method+" "+op.Path] = strings.Join(c.Command, " ")
		}
	}
	for endpoint, wantCmd := range want {
		if got[endpoint] != wantCmd {
			t.Errorf("%s -> %q, want %q", endpoint, got[endpoint], wantCmd)
		}
	}
}

// 生成済みの定義にも補正が乗っていること（生成器のテストだけだと、
// 再生成を忘れたコミットを見逃す）。
func TestSiteUpdateAcceptsBlogResource(t *testing.T) {
	for _, c := range spec.MustLoad().Commands {
		if strings.Join(c.Command, " ") != "sites update" {
			continue
		}
		var names []string
		for _, p := range c.Ops[0].Body {
			names = append(names, p.Name)
		}
		got := strings.Join(names, ",")
		if got != "blog,website" {
			t.Errorf("sites update の本文パラメータ = %q, want \"blog,website\"", got)
		}
		return
	}
	t.Fatal("sites update が見つかりません")
}

// 注記が生成済みの定義に乗っていること（再生成漏れの検出）。
func TestCustomFieldTypesAreDocumented(t *testing.T) {
	for _, c := range spec.MustLoad().Commands {
		if strings.Join(c.Command, " ") != "sites fields create" {
			continue
		}
		for _, p := range c.Ops[0].Body {
			if p.Name != "field" {
				continue
			}
			if !strings.Contains(p.Description, "post_type") {
				t.Errorf("--field の説明に型の一覧がありません: %q", p.Description)
			}
			return
		}
		t.Fatal("field パラメータがありません")
	}
	t.Fatal("sites fields create が見つかりません")
}
