package cmd

import (
	"sort"
	"strings"
)

// ダンプの中身の決め方。
//
// **環境をまたぐと ID は一致しません。** ステージングの記事 42 と本番の記事 42
// は別物です。だから ID で並べても突き合わせても意味がなく、**自然キー**
// （basename / name / label）で並べ、参照も自然キーに書き換えます。
//
// コンテンツタイプとコンテンツフィールドだけは `uniqueID`（40桁のhex）を
// 持っていますが、**これも環境をまたいで一致するとは限りません**——テーマ経由で
// 配ったものは同じ値を持ちますが、両方で別々に作れば別の値になります。
// だから鍵には使わず、データとして残します（同じ血統かどうかが差分で分かる）。

// volatileKeys は、どのオブジェクトからも落とす項目です。
//
// **これを落とさないとダンプは差分検出に使えません。** ID も時刻も作成者も
// 環境ごとに違うので、中身が同じでも全行が差分になります。
var volatileKeys = map[string]bool{
	"id":           true,
	"createdDate":  true,
	"modifiedDate": true,
	"createdBy":    true,
	"modifiedBy":   true,
	"updatable":    true,
	"permalink":    true, // ホスト名を含む
	"archiveLink":  true, // 同上
	"blog":         true, // 親サイトへの ID 参照
	"site":         true,
	"filePath":     true, // サーバの絶対パス
	"url":          true, // ホスト名を含む（アセットは path に畳んでから落とす）
	// ここから下は「受け側の MT が作る値」。書いたものではないので、
	// 環境をまたいで比べても意味が無い。
	//
	// uniqueID はコンテンツタイプとフィールドが持つ40桁のhex。両方の環境で
	// 別々に作れば必ず違う値になる。参照している側（dataLabel）は、指している
	// **フィールドの名前**に畳んでから落とす。
	//
	// fileSize は保存されたファイルの大きさ。**MT は JPEG を取り込むときに
	// 再エンコードする**ので、同じ画像を送っても送り側と受け側で違う値になる。
	// 残すと画像のあるサイトでは永久に差分が出続け、道具として使えなくなる。
	// 中身が同じかを見るなら寸法（width / height）を使う——再エンコードでは変わらない。
	"uniqueID": true,
	"fileSize": true,
}

// urlKeys は「絶対URLが入っていたら落とす」項目です。
//
// permalink や archiveLink のように必ず URL のものは volatileKeys で落とせますが、
// path は**アセットではサイト相対、フォルダでは絶対URL**という二枚舌になっています。
// 名前ではなく中身で判断しないと、どちらかを取りこぼします。
var urlKeys = map[string]bool{"path": true, "url": true, "archiveUrl": true, "siteUrl": true}

// environmentKeys はサイト設定のうち「そのサーバがどこにあるか」です。
//
// **これを移行先へ運ぶとサイトが行方不明になります。** 一方で smartReplace や
// convertParas のような挙動の設定は運ぶべきもので、同じファイルに混ぜると
// 「差分は出るが運んではいけない」行と「運ぶべき」行が並んで判断できません。
// だから site.yaml と site-environment.yaml に分けます。
var environmentKeys = map[string]bool{
	"host":              true,
	"sitePath":          true,
	"archivePath":       true,
	"relativeUrl":       true,
	"extraPath":         true,
	"uploadDestination": true,
}

// isAbsoluteURL は環境依存の値かを返します。
func isAbsoluteURL(v any) bool {
	s, ok := v.(string)
	return ok && (strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://"))
}

// dumpKind は1種類のオブジェクトの取り方です。
type dumpKind struct {
	// File は書き出すファイル名（拡張子なし）です。
	File string
	// Path は一覧のパスです。%s はサイトIDに置き換えます。
	Path string
	// Key は自然キーを作る項目です。前から順に、空でないものを使います。
	Key []string
	// Query は一覧に足すクエリです。
	Query map[string]string
	// Hierarchical を立てると parent を親の自然キーに書き換えます。
	Hierarchical bool
}

// siteKinds はサイト配下のオブジェクトです。
//
// 統計とログは入れません。環境で違って当たり前のもので、差分に出しても
// ノイズにしかならない。コメントも同じ理由で入れません。
var siteKinds = []dumpKind{
	{File: "templates", Path: "/sites/%s/templates", Key: []string{"name"}},
	{File: "categories", Path: "/sites/%s/categories", Key: []string{"basename", "label"}, Hierarchical: true},
	{File: "folders", Path: "/sites/%s/folders", Key: []string{"basename", "label"}, Hierarchical: true},
	{File: "category-sets", Path: "/sites/%s/categorySets", Key: []string{"name"}},
	{File: "content-types", Path: "/sites/%s/contentTypes", Key: []string{"name"}},
	{File: "custom-fields", Path: "/sites/%s/fields", Key: []string{"systemObject", "basename"}},
	{File: "widgets", Path: "/sites/%s/widgets", Key: []string{"name"}},
	{File: "widgetsets", Path: "/sites/%s/widgetsets", Key: []string{"name"}},
	{File: "formatted-texts", Path: "/sites/%s/formatted_texts", Key: []string{"label"}},
	{File: "entries", Path: "/sites/%s/entries", Key: []string{"basename", "title"},
		Query: map[string]string{"status": "", "noTextFilter": "1"}},
	{File: "pages", Path: "/sites/%s/pages", Key: []string{"basename", "title"},
		Query: map[string]string{"status": "", "noTextFilter": "1"}},
	{File: "assets", Path: "/sites/%s/assets", Key: []string{"filename"}},
}

// systemKinds はサイトに属さないオブジェクトです。
var systemKinds = []dumpKind{
	{File: "roles", Path: "/roles", Key: []string{"name"}},
	{File: "users", Path: "/users", Key: []string{"name"}},
	{File: "groups", Path: "/groups", Key: []string{"name"}},
}

// naturalKey は並べ替えと突き合わせに使う値を作ります。
func (k dumpKind) naturalKey(obj map[string]any) string {
	var parts []string
	for _, name := range k.Key {
		if v, ok := obj[name].(string); ok && v != "" {
			parts = append(parts, v)
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "/")
}

// clean は1オブジェクトから揮発する項目を落とし、ID の参照を自然キーに
// 書き換えます。
//
// 入れ子のオブジェクトが `id` と自然キーの両方を持っているとき（記事の
// categories や assets）、**自然キーだけを残します**。ID をそのままにすると、
// 中身が同じでも環境間で必ず差分になります。
func clean(v any) any {
	switch t := v.(type) {
	case map[string]any:
		// {"id": ...} だけ、あるいは id ＋ 自然キーの参照は、キーに畳みます。
		if key, ok := referenceKey(t); ok {
			return key
		}
		out := map[string]any{}
		for k, val := range t {
			if volatileKeys[k] || (urlKeys[k] && isAbsoluteURL(val)) {
				continue
			}
			out[k] = clean(val)
		}
		if len(out) == 0 {
			return nil
		}
		return out
	case []any:
		out := make([]any, 0, len(t))
		for _, item := range t {
			out = append(out, clean(item))
		}
		return sortIfKeyed(out)
	}
	return v
}

// sortIfKeyed は「並び順に意味が無い配列」だけを並べ替えます。
//
// MT はカスタムフィールドの値を毎回同じ順では返しません。そのまま出すと
// **中身が同じなのに差分が出ます**（実機で記事6件すべてがそうなった）。
// 一方、カテゴリの並びのように順序に意味があるものは触ってはいけないので、
// **自然キーを持つ要素だけ**を対象にします（それが「集合であって列ではない」印）。
func sortIfKeyed(items []any) []any {
	keys := make([]string, len(items))
	for i, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			return items
		}
		var k string
		for _, name := range []string{"basename", "uniqueID", "name", "label"} {
			if v, ok := m[name].(string); ok && v != "" {
				k = v
				break
			}
		}
		if k == "" {
			return items
		}
		keys[i] = k
	}
	idx := make([]int, len(items))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool { return keys[idx[a]] < keys[idx[b]] })
	out := make([]any, len(items))
	for i, j := range idx {
		out[i] = items[j]
	}
	return out
}

// referenceKey は「他のオブジェクトを指しているだけ」のマップを1つの値に
// 畳みます。畳めなければ false を返します。
func referenceKey(m map[string]any) (any, bool) {
	if _, hasID := m["id"]; !hasID {
		return nil, false
	}
	// アセットの参照は filename しか自然キーを持ちません（url は環境ごとに
	// 違うので使えない）。別フォルダの同名ファイルは見分けが付きませんが、
	// 参照の側にそれ以上の手がかりが無いので、ここは割り切ります。
	// displayName は記事の著者が持つ唯一の手がかりです（author は
	// id / displayName / userpicUrl しか返さない）。ここに入れないと
	// **著者が null に潰れて、誰のものか分からなくなります**。
	for _, name := range []string{"basename", "filename", "name", "label", "uniqueID", "displayName"} {
		if v, ok := m[name].(string); ok && v != "" {
			// 参照先を指す値が1つあれば、それだけで足ります。
			if isReferenceOnly(m) {
				return v, true
			}
			return nil, false
		}
	}
	if isReferenceOnly(m) {
		// 自然キーが無い純粋な ID 参照は、環境間で必ず食い違うので落とします。
		return nil, true
	}
	return nil, false
}

// isReferenceOnly は「参照のための項目しか持っていない」かを返します。
func isReferenceOnly(m map[string]any) bool {
	for k := range m {
		switch k {
		case "id", "basename", "name", "label", "uniqueID", "displayName", "userpicUrl", "url", "filename", "class", "type":
		default:
			return false
		}
	}
	return true
}

// sortObjects は自然キーで並べます。キーが同じものは、安定させるために
// YAML にしたときの中身で並べます。
func sortObjects(items []dumpItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Key != items[j].Key {
			return items[i].Key < items[j].Key
		}
		return items[i].tiebreak < items[j].tiebreak
	})
}

// dumpItem は書き出す1件です。
type dumpItem struct {
	Key      string
	Object   map[string]any
	tiebreak string
}
