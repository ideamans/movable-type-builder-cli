package cmd

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// 差分を取るための出力なので、**同じ入力なら必ず同じバイト列**が出ること。
// API が返す順や Go のマップの順に引きずられると、中身が同じでも差分が出る。
func TestDumpYAMLIsDeterministic(t *testing.T) {
	obj := map[string]any{
		"zebra": 1, "alpha": 2, "middle": map[string]any{"z": "1", "a": "2"},
		"list": []any{"b", "a"},
	}
	first, err := yamlBytes(obj)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		again, err := yamlBytes(obj)
		if err != nil {
			t.Fatal(err)
		}
		if string(again) != string(first) {
			t.Fatalf("%d 回目で揺れました:\n%s\n---\n%s", i, first, again)
		}
	}
	// キーは名前順。配列の順は入力のまま（意味のある順序なので並べ替えない）。
	if got := string(first); !strings.HasPrefix(got, "alpha: 2\n") {
		t.Errorf("キーが名前順ではありません:\n%s", got)
	}
	if !strings.Contains(string(first), "- b\n    - a\n") && !strings.Contains(string(first), "- b\n  - a\n") {
		t.Errorf("配列の順を変えています:\n%s", first)
	}
}

// 複数行はリテラルブロックで書く。テンプレート本文は 1500 行を超えることが
// あり、1行の引用符付き文字列にすると差分が「1行変わった」としか出ない。
func TestDumpYAMLUsesLiteralBlocks(t *testing.T) {
	out, err := yamlBytes(map[string]any{"text": "<html>\n  <body></body>\n</html>\n"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "text: |\n") {
		t.Errorf("リテラルブロックになっていません:\n%s", out)
	}
}

// **ダンプで中身が変わってはいけない。** ブロックスカラーは行末の空白と
// タブを復元できないので、そういう文字列は引用に落とす必要がある。
// ここを間違えると「差分は出ないのに実物が違う」という最悪の壊れ方をする。
func TestDumpYAMLRoundTripsExactly(t *testing.T) {
	for _, tc := range []struct{ name, text string }{
		{"ふつうの複数行", "<html>\n<body>\n</html>\n"},
		{"行末に空白", "line one   \nline two\n"},
		{"タブで字下げ", "if (x) {\n\treturn 1;\n}\n"},
		{"空白だけの行", "a\n   \nb\n"},
		{"末尾に改行なし", "no trailing newline"},
		{"改行が連続", "a\n\n\nb\n"},
		{"MTタグ", "<$mt:BlogName$>\n<mt:Entries lastn=\"5\">\n</mt:Entries>\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := yamlBytes(map[string]any{"text": tc.text})
			if err != nil {
				t.Fatal(err)
			}
			var back map[string]string
			if err := yaml.Unmarshal(out, &back); err != nil {
				t.Fatalf("読み戻せません: %v\n%s", err, out)
			}
			if back["text"] != tc.text {
				t.Errorf("中身が変わりました\n元 : %q\n後 : %q\nYAML:\n%s", tc.text, back["text"], out)
			}
		})
	}
}

// ID・時刻・作成者は落とす。残すと中身が同じでも全行が差分になる。
func TestDumpDropsVolatileKeys(t *testing.T) {
	in := map[string]any{
		"id": 42, "name": "メインページ", "text": "x",
		"createdDate": "2026-08-30T00:00:00+09:00", "modifiedDate": "2026-08-30T01:00:00+09:00",
		"createdBy": map[string]any{"id": 1, "displayName": "miyanaga"},
		"updatable": true, "permalink": "http://stg.example.com/a.html",
		"blog": map[string]any{"id": "6"},
	}
	out, _ := clean(in).(map[string]any)
	for _, k := range []string{"id", "createdDate", "modifiedDate", "createdBy", "updatable", "permalink", "blog"} {
		if _, ok := out[k]; ok {
			t.Errorf("%s が残っています", k)
		}
	}
	if out["name"] != "メインページ" || out["text"] != "x" {
		t.Errorf("残すべきものが消えました: %+v", out)
	}
}

// 他のオブジェクトへの参照は自然キーに畳む。ID のままだと環境間で必ず違う。
func TestDumpFoldsReferencesToNaturalKeys(t *testing.T) {
	in := map[string]any{
		"title":      "記事",
		"categories": []any{map[string]any{"id": 23, "label": "導入事例", "basename": "case-study"}},
		"assets":     []any{map[string]any{"id": 70, "filename": "hero.jpg", "url": "http://stg/x/hero.jpg"}},
	}
	out, _ := clean(in).(map[string]any)
	cats, _ := out["categories"].([]any)
	if len(cats) != 1 || cats[0] != "case-study" {
		t.Errorf("カテゴリが自然キーに畳まれていません: %#v", out["categories"])
	}
	assets, _ := out["assets"].([]any)
	if len(assets) != 1 || assets[0] != "hero.jpg" {
		t.Errorf("アセットが自然キーに畳まれていません: %#v", out["assets"])
	}
}

// 階層のあるものは、親を ID ではなく根からの道で持つ。
func TestDumpResolvesParentPath(t *testing.T) {
	kind := dumpKind{Key: []string{"basename"}, Hierarchical: true}
	items := []map[string]any{
		{"id": 1, "basename": "products", "parent": 0},
		{"id": 2, "basename": "saas", "parent": 1},
		{"id": 3, "basename": "enterprise", "parent": 2},
	}
	keys := parentKeys(items, kind)
	if keys["3"] != "products/saas/enterprise" {
		t.Errorf("道 = %q, want products/saas/enterprise", keys["3"])
	}
	child := withParentKey(items[2], keys)
	if child["parent"] != "products/saas" {
		t.Errorf("親 = %#v, want products/saas", child["parent"])
	}
	if _, ok := withParentKey(items[0], keys)["parent"]; ok {
		t.Error("根の parent は落とすべきです")
	}
}

// アセットはサイトルートからの相対パスで持つ。filePath も url も
// サーバの絶対パスとホスト名を含むので、そのままでは全件が差分になる。
func TestDumpAssetPathIsSiteRelative(t *testing.T) {
	site := dumpSite{URL: "http://stg.example.com/blog/", Path: "/var/www/blog"}
	for _, tc := range []struct {
		name string
		obj  map[string]any
		want string
	}{
		{"url から", map[string]any{"url": "http://stg.example.com/blog/assets/img/a.png"}, "assets/img/a.png"},
		{"filePath から", map[string]any{"filePath": "/var/www/blog/assets/b.png"}, "assets/b.png"},
		{"どちらも合わない", map[string]any{"filename": "c.png"}, "c.png"},
	} {
		if got := siteRelativePath(tc.obj, site); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}

// 並びは自然キー。API の返す順に引きずられない。
func TestDumpSortsByNaturalKey(t *testing.T) {
	items := []dumpItem{{Key: "zebra"}, {Key: "alpha"}, {Key: "middle"}}
	sortObjects(items)
	if items[0].Key != "alpha" || items[2].Key != "zebra" {
		t.Errorf("並び = %v", []string{items[0].Key, items[1].Key, items[2].Key})
	}
}

// サイト設定は「運ぶもの」と「その環境のもの」に分ける。
//
// **host や sitePath を移行先へ運ぶとサイトが行方不明になる。** 一方で
// smartReplace のような挙動の設定は運ぶべきもので、同じファイルに混ぜると
// 「差分だが運んではいけない」行と「運ぶべき」行が並んで判断できない。
func TestDumpSplitsSiteSettings(t *testing.T) {
	portable, environment := splitSiteSettings(map[string]any{
		"name": "SaasHub", "smartReplace": 1, "convertParas": "richtext",
		"host": "stg.example.com", "sitePath": "/var/www/stg", "archivePath": "/var/www/stg",
		"relativeUrl": "/", "extraPath": "", "uploadDestination": map[string]any{"path": "/var/www/stg"},
	})
	for _, k := range []string{"host", "sitePath", "archivePath", "relativeUrl", "extraPath", "uploadDestination"} {
		if _, ok := environment[k]; !ok {
			t.Errorf("%s が環境側に入っていません", k)
		}
		if _, ok := portable[k]; ok {
			t.Errorf("%s が運ぶ側に紛れています", k)
		}
	}
	for _, k := range []string{"name", "smartReplace", "convertParas"} {
		if _, ok := portable[k]; !ok {
			t.Errorf("%s は運ぶ側にあるべきです", k)
		}
	}
}

// 絶対URLの項目は環境依存。path はアセットではサイト相対、フォルダでは
// 絶対URLという二枚舌なので、名前ではなく中身で判断する。
func TestDumpDropsAbsoluteURLValues(t *testing.T) {
	out, _ := clean(map[string]any{
		"basename": "about",
		"path":     "http://stg.example.com/about", // フォルダ: 絶対URL → 落とす
		"label":    "会社情報",
	}).(map[string]any)
	if _, ok := out["path"]; ok {
		t.Errorf("絶対URLの path が残っています: %v", out["path"])
	}
	out2, _ := clean(map[string]any{
		"filename": "a.png",
		"path":     "assets/img/a.png", // アセット: サイト相対 → 残す
	}).(map[string]any)
	if out2["path"] != "assets/img/a.png" {
		t.Errorf("サイト相対の path まで落としています: %v", out2["path"])
	}
}

// 受け側の MT が作る値は差分に出さない。残すと永久に消えない差分になる。
func TestDumpDropsServerGeneratedValues(t *testing.T) {
	out, _ := clean(map[string]any{
		"name": "著者", "uniqueID": "fd12635ba6e7", // 環境ごとに生成される
		"meta": map[string]any{"fileSize": 95362, "width": 1200, "height": 630},
	}).(map[string]any)
	if _, ok := out["uniqueID"]; ok {
		t.Error("uniqueID が残っています")
	}
	meta, _ := out["meta"].(map[string]any)
	if _, ok := meta["fileSize"]; ok {
		t.Error("fileSize が残っています（MT は JPEG を再エンコードするので必ず違う値になる）")
	}
	if meta["width"] != 1200 {
		t.Error("寸法は残すべきです（再エンコードでも変わらないので、中身の手がかりになる）")
	}
}

// 並び順に意味の無い配列は並べ替える。MT は customFields を毎回同じ順では
// 返さないので、そのままだと中身が同じなのに差分が出る。
func TestDumpSortsUnorderedArrays(t *testing.T) {
	out, _ := clean(map[string]any{"customFields": []any{
		map[string]any{"basename": "zebra", "value": "1"},
		map[string]any{"basename": "alpha", "value": "2"},
	}}).(map[string]any)
	got, _ := out["customFields"].([]any)
	first, _ := got[0].(map[string]any)
	if first["basename"] != "alpha" {
		t.Errorf("並べ替えていません: %v", got)
	}
	// 順序に意味のあるもの（自然キーを持たない要素）は触らない。
	out2, _ := clean(map[string]any{"steps": []any{"z", "a"}}).(map[string]any)
	steps, _ := out2["steps"].([]any)
	if steps[0] != "z" {
		t.Errorf("順序に意味のある配列を並べ替えています: %v", steps)
	}
}

// 環境の記録は「運べないが、違うと分かるだけで価値がある」もの。
// プラグインの応答には SVG のアイコンが丸ごと入っていて、残すと差分が読めない。
func TestDumpStripsPluginNoise(t *testing.T) {
	obj := map[string]any{
		"name": "BlockEditor", "version": "1.0.4", "status": "Enabled",
		"icon": "<svg role=\"img\">…長大…</svg>", "description": "",
		"authorLink": "http://www.movabletype.org/", "authorName": "Six Apart Ltd.",
	}
	out := withoutKeys(obj, pluginNoise...)
	for _, k := range pluginNoise {
		if _, ok := out[k]; ok {
			t.Errorf("%s が残っています", k)
		}
	}
	for _, k := range []string{"name", "version", "status"} {
		if _, ok := out[k]; !ok {
			t.Errorf("%s は残すべきです（環境の比較に要る）", k)
		}
	}
}

// 著者は displayName しか手がかりが無い（author は id / displayName /
// userpicUrl しか返さない）。自然キーとして認めないと **null に潰れて、
// 誰のものか分からなくなる**。
//
// 著者は Data API では変更できないので（読み取り専用の項目）、移行では
// **必ず「運んだ人」のものになる**。差分に出ないと、それに気づけない。
func TestDumpKeepsEntryAuthor(t *testing.T) {
	out, _ := clean(map[string]any{
		"basename": "post-1",
		"author":   map[string]any{"id": 1, "displayName": "miyanaga", "userpicUrl": nil},
	}).(map[string]any)
	if out["author"] != "miyanaga" {
		t.Errorf("著者 = %#v, want miyanaga", out["author"])
	}
}
