// Command gen-spec は公式の OpenAPI 定義を、CLI が必要とする形だけに絞った
// spec.json に変換します。`make spec` から呼ばれます。
//
// **公式の説明文は持ち込みません。** 定義の原本（Six Apart の
// mt-docs-data-api-reference）にはライセンスの指定が無く、説明文をそのまま
// 再配布できるとは言えないためです。spec.json に残すのは、パス・メソッド・
// パラメータの名前・型・選択肢といった事実と、操作の短い題名（summary）だけで、
// 説明は公式リファレンスの該当箇所への URL（docUrl）にします。mtb が補った
// 注記（paramNotes ほか）は自前の文章なので残します。
//
// 原本もリポジトリに置きません。固定したコミットから取ってきて、
// 利用者のキャッシュディレクトリに置きます（-in で手元のファイルも渡せます）。
//
// ここがコマンド名の決定を一手に引き受けています。規則は README と
// internal/llmdocs/10-endpoints.md に書いてあるものと同じで、
// 例外は overrides の表に全部載っています（推測で足さないこと）。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ideamans/movable-type-builder-cli/internal/spec"
)

const apiVersion = "v7"

var (
	inPath  = flag.String("in", "", "公式の OpenAPI 定義（省くと固定したコミットから取ってくる）")
	outPath = flag.String("out", "spec.json", "書き出す先")
)

// sourceCommit は取り込む公式定義のコミットです。上げるときはここだけを
// 書き換えて make spec を流し、差分（コマンド名の動き）を確かめてから
// コミットします。2026-02-04 の src/openapi/v7.json（info.version 9.1.0）。
const sourceCommit = "8863d203b632f5947cfebbe55642bb9615b49d36"

const sourceURL = "https://raw.githubusercontent.com/movabletype/mt-docs-data-api-reference/" +
	sourceCommit + "/src/openapi/v7.json"

// referenceURL は公式リファレンスです。各操作へのアンカーは
// #tag/<最初のタグ>/paths/<パスの / を ~1 にしたもの>/<メソッド> の形をしている。
const referenceURL = "https://movabletype.github.io/mt-docs-data-api-reference/v7.html"

// ownShorts は、公式定義に題名（summary）が無い操作の題名です。説明文の
// 1行目を流用すると公式の文章を持ち込むことになるので、ここで書きます。
var ownShorts = map[string]string{
	"POST /groups/{group_id}/members":               "Add a user to a group",
	"DELETE /groups/{group_id}/members/{member_id}": "Remove a user from a group",
}

// actions は「末尾に来たら動詞として扱う」リテラルセグメントです。
// ここに無いリテラルはリソース名とみなし、動詞は HTTP メソッドから決めます。
var actions = map[string]bool{
	"upload":            true,
	"export":            true,
	"import":            true,
	"preview":           true,
	"permutate":         true,
	"grant":             true,
	"revoke":            true,
	"clone":             true,
	"publish":           true,
	"refresh":           true,
	"apply":             true,
	"unlock":            true,
	"enable":            true,
	"disable":           true,
	"synchronize":       true,
	"recover_password":  true,
	"backup":            true,
	"thumbnail":         true,
	"export_theme":      true,
	"refresh_templates": true,
	"search":            true,
}

// overrides は規則どおりだと意味を取り違える／衝突するエンドポイントの
// コマンド名を明示します。増やすときは理由をコメントに残すこと。
var overrides = map[string][]string{
	// ルートは自分自身の OpenAPI 定義を返す。"list" ではない。
	"GET /": {"openapi"},
	// サーバ側の Data API バージョン。CLI 自身の --version と紛らわしいので改名。
	"GET /version": {"server-version"},
	// site_id 配下に「ブログを作る」。/sites の POST（ウェブサイト作成）と衝突する。
	"POST /sites/{site_id}": {"sites", "create-blog"},
	// コレクションへの DELETE は「全ログ削除」。1件削除と同名になると事故る。
	"DELETE /sites/{site_id}/logs": {"sites", "logs", "reset"},
	// GET だが再構築の実行。"list" ではない。
	"GET /publish/entries":     {"publish", "entries"},
	"GET /publish/contentData": {"publish", "content-data"},
	// 現在有効なプロバイダを1つ返す。コレクションではない。
	"GET /sites/{site_id}/stats/provider": {"sites", "stats", "provider"},
}

// extraBody は公式定義に載っていない本文パラメータを補います。
// 実物の MT が受け付けるのに定義に無いものだけを、理由つきで足すこと。
// 「定義に無いものは存在しない」が原則なので、ここは最後の手段です。
var extraBody = map[string][]*spec.Param{
	// サイトの更新は、対象が website なら website、blog なら blog を要求する。
	// 定義は website しか書いていないため、そのままでは子サイト（ブログ）を
	// 更新できず、400 A resource "blog" is required. で初めて気づくことになる。
	"PUT /sites/{site_id}": {
		{
			Name: "blog", Flag: "blog", Type: "object",
			Description:   "更新内容。対象が blog のときはこちら（website のときは --website）",
			DescriptionEn: "What to update. Use this when the target is a blog (use --website for a website).",
		},
	},
}

// extraQuery は定義に無いクエリパラメータを足します。extraBody と同じ趣旨で、
// 「実物は受け付けるのに定義に書かれていない」ものだけが対象です。
//
// noTextFilter は一覧（9エンドポイント）には定義があるのに、個別取得には
// 無い。**無いと原稿が取り出せません。** get が返す body は変換後の HTML で、
// それを書き戻すと Markdown やブロックエディタの原稿が HTML に置き換わる。
// format は markdown のままなので、誰もエラーに気づきません。
// MT 側は MT::DataAPI::Resource::v2::Entry の _from_object_text が
// no_text_filter / noTextFilter のどちらでも見ています。
var extraQuery = map[string][]*spec.Param{
	"GET /sites/{site_id}/entries/{entry_id}":                                    {noTextFilterParam()},
	"GET /sites/{site_id}/pages/{page_id}":                                       {noTextFilterParam()},
	"GET /sites/{site_id}/contentTypes/{content_type_id}/data/{content_data_id}": {noTextFilterParam()},
}

func noTextFilterParam() *spec.Param {
	return &spec.Param{
		Name: "noTextFilter", Flag: "no-text-filter", Type: "integer", Enum: []string{"0", "1"},
		Description: "1 を指定すると本文をテキストフィルタに通さず、保存されている原稿のまま返す。" +
			"既定では変換後の HTML が返るので、取り出して書き戻すと Markdown や" +
			"ブロックエディタの原稿が HTML に置き換わる（format は変わらないので気づけない）。" +
			"編集のために取り出すなら必ず付ける",
		DescriptionEn: "Set to 1 to return the stored source instead of running the body through the " +
			"text filter. By default the converted HTML comes back, so reading a body and writing it " +
			"straight back replaces Markdown or Block Editor source with HTML (the format is unchanged, " +
			"so nothing looks wrong). Always pass it when you read a body in order to edit it.",
	}
}

// paramNotes は公式定義が説明しきれていないパラメータの説明を厚くします。
// 値の扱いは変えず、--help に出る文言だけを足すもの。「MT はエラーで
// 選択肢を教えてくれないが、実物には決まった集合がある」ものが対象です。
// note は補足の文言です。公式定義の説明は英語なので、注記だけを訳し分けます。
type note struct{ ja, en string }

var paramNotes = map[string]map[string]note{
	"POST /sites/{site_id}/entries/import":                {"encoding": encodingNote, "publish": entryPublishNote},
	"GET /sites/{site_id}/logs/export":                    {"encoding": encodingNote},
	"GET /search":                                         {"cdSearch": cdSearchNote},
	"POST /sites/{site_id}/entries":                       {"publish": entryPublishNote, "entry": entryTextNote},
	"PUT /sites/{site_id}/entries/{entry_id}":             {"publish": entryPublishNote, "entry": entryTextNote},
	"POST /sites/{site_id}/pages":                         {"publish": entryPublishNote},
	"PUT /sites/{site_id}/pages/{page_id}":                {"publish": entryPublishNote},
	"POST /sites/{site_id}/fields":                        {"field": customFieldNote},
	"PUT /sites/{site_id}/fields/{field_id}":              {"field": customFieldNote},
	"POST /sites/{site_id}/contentTypes":                  {"content_type": contentTypeNote},
	"PUT /sites/{site_id}/contentTypes/{content_type_id}": {"content_type": contentTypeNote},
	"POST /sites":                                {"website": smartReplaceNote},
	"POST /sites/{site_id}/blogs":                {"blog": smartReplaceNote},
	"PUT /sites/{site_id}":                       {"website": smartReplaceNote, "blog": smartReplaceNote},
	"POST /users":                                {"user": userPasswordNote},
	"GET /publish/entries":                       {"ids": rebuildNote, "startTime": rebuildStartTimeNote},
	"POST /sites/{site_id}/categories/permutate": {"categories": permutateNote},
	"POST /sites/{site_id}/folders/permutate":    {"folders": permutateNote},
	"POST /sites/{site_id}/categorySets/{category_set_id}/categories/permutate":  {"categories": permutateNote},
	"POST /sites/{site_id}/contentTypes/{content_type_id}/fields/permutate":      {"content_fields": permutateNote},
	"POST /sites/{site_id}/contentTypes/{content_type_id}/data":                  {"content_data": contentDataNote},
	"PUT /sites/{site_id}/contentTypes/{content_type_id}/data/{content_data_id}": {"content_data": contentDataNote},
	"GET /publish/contentData":           {"ids": rebuildNote, "startTime": rebuildStartTimeNote},
	"POST /sites/{site_id}/export_theme": {"include": themeIncludeNote},
	"PUT /users/{user_id}":               {"user": userPasswordNote},
}

// rebuildNote は再構築が1回で終わらないこと。応答の {"status":"Rebuilding"} は
// 受け付けたという意味でしかなく、残りは restIds に入る。空になるまで呼び直す。
var rebuildNote = note{
	ja: "1回では終わらない。応答の Rebuilding は受け付けたという意味で、" +
		"残りは restIds に入る。空になるまで呼び直すこと（--wait を付ければ mtb が回す）。" +
		"すぐ反映したいなら対象を更新する方が早い（保存すると同期で再構築される）",
	en: "One call does not finish it. Rebuilding in the response only means the request was accepted; " +
		"what is left comes back in restIds, and you call again until it is empty (--wait does that for you). " +
		"To see a change right away, updating the object is quicker (saving rebuilds it synchronously).",
}

// rebuildStartTimeNote は呼び直しの必須条件。**実測: startTime を渡さないと
// restIds はいつまでも減らない。** 毎回新しい再構築として受け付け直されるため、
// 12回呼んでも6件のまま返り続けた。渡した途端に2回で空になる。
// 応答が success を返し続けるので、これは自力では絶対に気づけない。
var rebuildStartTimeNote = note{
	ja: "**呼び直すときは応答の startTime をそのまま渡すこと。** " +
		"渡さないと毎回新しい再構築として受け付け直され、restIds がいつまでも減らない" +
		"（エラーにはならず、Rebuilding のまま終わらない）。--wait を使えば mtb が引き継ぐ",
	en: "**When calling again, pass back the startTime from the response.** " +
		"Without it MT treats each call as a fresh rebuild and restIds never shrinks — " +
		"no error, it just never finishes. --wait carries it over for you.",
}

// permutateNote は並べ替えの渡し方。id だけの配列では
// A parameter "..." is invalid. になる。list の結果をそのまま並べ替えて渡す。
var permutateNote = note{
	ja: "並べ替えた後の完全なリソース配列を渡す。" +
		"[{\"id\":1}] のような id だけの配列は通らない（list の出力をそのまま並べ替えるのが早い）。" +
		"**取り出すときは --sort-by user_custom --sort-order ascend を付けること。** " +
		"list の既定は descend なので、そのまま並べ替えて送ると順序が逆になる",
	en: "Pass the complete resource array in its new order. " +
		"An array of bare ids such as [{\"id\":1}] is rejected; reordering the output of the list command is the quickest way. " +
		"**Read the current order with --sort-by user_custom --sort-order ascend.** " +
		"The list default is descend, so reordering what list gives you and sending it back reverses the order.",
}

// contentDataNote はコンテンツデータの必須項目。label が無いと
// "Data Label" is required. になるが、定義には現れない。
var contentDataNote = note{
	ja: "label（データの名前）が必須。" +
		"data はフィールドIDの配列 [{\"id\":<フィールドID>,\"data\":\"値\"}]。" +
		"フィールドIDは sites content-types fields list で調べる",
	en: "label (the name of the data) is required. " +
		"data is an array keyed by field ID: [{\"id\":<field ID>,\"data\":\"value\"}]. " +
		"Look the field IDs up with sites content-types fields list.",
}

// encodingNote は取り込みの文字コード。MT は不正な値に Invalid encoding としか
// 返さず、しかも有効な値は MT の DefaultLanguage によって変わる。
var encodingNote = note{
	ja: "有効な値は MT の DefaultLanguage で変わる。" +
		"英語環境は guess / utf8 / ascii、日本語環境は sjis / euc-jp など。" +
		"既定は guess（自動判別）なので通常は指定しなくてよい",
	en: "The valid values depend on MT's DefaultLanguage: guess / utf8 / ascii in an English " +
		"installation, sjis / euc-jp and others in a Japanese one. " +
		"The default is guess (auto-detect), so you rarely need to set it.",
}

// cdSearchNote は MT 側の不具合の回避策。cdSearch=1 だけを渡すと
// 既定の並び順が壊れて 500 Unknown column 'cd_' in 'ORDER BY' になる。
var cdSearchNote = note{
	ja: "1 を指定するときは SearchSortBy も一緒に渡すこと。" +
		"単独で使うと MT 側が 500（Unknown column 'cd_' in 'ORDER BY'）を返す",
	en: "When you set this to 1, pass SearchSortBy as well. " +
		"On its own it makes MT return 500 (Unknown column 'cd_' in 'ORDER BY').",
}

// entryPublishNote は publish の意味。0 にしても記事のステータスは変わらず、
// 再構築をしないだけ。「下書きにする」と誤解しやすい。
var entryPublishNote = note{
	ja: "0 にしても記事のステータスは変わらない。再構築をしないだけ。" +
		"下書きにしたいなら entry の status を Draft にする",
	en: "Setting this to 0 does not change the status of the entry; it only skips the rebuild. " +
		"To make it a draft, set the entry's status to Draft.",
}

// userPasswordNote は Data API で作ったユーザーがそのままでは Data API に
// サインインできないこと。password はサインイン用で、API の認証に要る
// Webサービスパスワードは Data API からは設定できない。気づくのは
// 401 Invalid login を踏んだ後になるので、作る前に見えるようにしておく。
var userPasswordNote = note{
	ja: "password はサインイン用。" +
		"Data API のサインインには (1) sign_in_data_api システム権限 " +
		"(2) Webサービスパスワード（MT が自動生成。API からは設定も取得もできず管理画面で発行する） " +
		"(3) サイト側の Data API 許可 の3つが要る",
	en: "password is for signing in to MT itself. Signing in to the Data API needs three more things: " +
		"(1) the sign_in_data_api system permission, " +
		"(2) the Web Services Password (generated by MT; it can be neither set nor read through the API, " +
		"only issued from the admin screen), and (3) Data API access enabled on the site.",
}

// themeIncludeNote は静的ファイルが黙って落ちること。実機で確かめた。
// blog_static_files を指定しても、サイトに theme_export_settings が
// 保存されていなければ何も入らない。しかも status は success を返すので、
// 空のテーマができたことに気づけない。MT 側は
// MT::Theme::StaticFiles::export が $blog->theme_export_settings を読んで
// おり、リクエストのパラメータからは渡せない。
var themeIncludeNote = note{
	ja: "カンマ区切りで複数指定する。" +
		"blog_static_files は、対象フォルダが管理画面で保存されていないと" +
		"何も含まれない（status は success を返すので気づけない）。" +
		"サイトの theme_export_settings に入るもので、本体の API からは設定できない" +
		"（拡張 mtb-pro があれば sites theme-export-settings update。テーマの適用で消える）",
	en: "Pass several values separated by commas. " +
		"blog_static_files includes nothing unless the target folders have been saved " +
		"from the admin screen (the status still comes back as success, so it is easy to miss). " +
		"They live in the site's theme_export_settings, which MT's own API cannot set " +
		"(with the mtb-pro extension: sites theme-export-settings update; applying a theme clears it).",
}

// customFieldNote はカスタムフィールドの型と対象。MT は不正な値に対して
// The type "..." is invalid. としか返さず、有効な値を教えてくれない。
// contentTypeNote は dataLabel の渡し方。**フィールドの名前ではなくユニークID**。
// 名前を渡してもコンテンツタイプの作成は通ってしまい、データ作成のときに
// 409 "Data Label" is required. になります。文面から dataLabel が原因だとは
// 分からないので書いておく。
var contentTypeNote = note{
	ja: "**dataLabel にはコンテンツフィールドのユニークIDを渡す**（フィールド名ではない）。" +
		"名前を渡してもコンテンツタイプの作成は通ってしまい、" +
		"データ作成のときに 409 \"Data Label\" is required. になる。" +
		"ユニークIDは先にフィールドを作ってから content-types get か content-types fields list で調べる",
	en: "**dataLabel takes a content field's uniqueID**, not its label. " +
		"A label is accepted when the content type is created, and the failure only shows up later as " +
		"409 \"Data Label\" is required. when you create data. " +
		"Create the fields first, then read the uniqueID from content-types get or content-types fields list.",
}

// smartReplaceNote は文字の自動変換。**どちらの設定でも作成時に書き換わります**。
// 実測（MT 9）:
//
//	smartReplace=0（既定） 作成: — → --        更新: そのまま
//	smartReplace=1        作成: — → &#8212;   更新: そのまま
//
// 「1 にすれば原文どおり」ではないのが要点で、切っても数値文字参照になります。
// 表示は同じなので、格納値を突き合わせるまで気づけません。
var smartReplaceNote = note{
	ja: "**記号の書き換えは記事・ページの「作成」でだけ起きる**（更新では起きない）。" +
		"smartReplace=0（既定）だと em ダッシュ「—」が「--」に、" +
		"1（変換しない）にしても「&#8212;」（数値文字参照）になる。**1 にしても原文のままにはならない。**" +
		"バイトどおりに入れたいなら、作成したあとに同じ値で更新し直すのが確実。" +
		"対象は smartReplaceFields（既定で title / text / text_more / keywords / excerpt / tags）",
	en: "**Punctuation is rewritten only when an entry or page is created**, never on update. " +
		"With smartReplace=0 (the default) an em dash becomes two hyphens; with 1 it becomes &#8212; " +
		"(a numeric character reference). **Setting it to 1 does not store the text verbatim.** " +
		"To store the exact bytes, update the object with the same value after creating it. " +
		"smartReplaceFields lists what it touches (title, text, text_more, keywords, excerpt, tags by default).",
}

// entryTextNote は上の症状の側から見た注記。記事を入れるときに気づけるように、
// サイト設定の側だけでなく entry にも書いておく。
var entryTextNote = note{
	ja: "**作成のときだけ、タイトルと本文の記号が書き換わる**（更新では起きない）。" +
		"em ダッシュ「—」はサイトの smartReplace が 0（既定）なら「--」に、1 でも「&#8212;」になる。" +
		"原文のバイトどおりに入れたいなら、**作成したあとに同じ値で更新し直す**",
	en: "**The title and body are rewritten only on create**, never on update. " +
		"An em dash becomes two hyphens when the site's smartReplace is 0 (the default), " +
		"and &#8212; even when it is 1. " +
		"To store the exact bytes, update the entry with the same value right after creating it.",
}

var customFieldNote = note{
	ja: "type は13種: text / textarea / checkbox / url / datetime / select / radio / embed / post_type / " +
		"image / file / audio / video。**アセット系は file（ラベルが Asset なので asset と書きたくなるが 409）**。" +
		"select・radio・datetime は options が必須（datetime は date_and_time / date_only / time_only）。" +
		"アセット系の値は URL を入れる（IDではない。検証されないので何でも入ってしまう）。" +
		"**API 経由ではアセットとの関連（objectasset）は作られない**ので、テンプレートでの出力はできるが" +
		"「このアセットを使っているもの」には出ない。" +
		"systemObject: entry / page / category / folder / comment / template はどのサイトでも可、" +
		"blog はブログのみ、website はウェブサイトのみ。author / asset / contentdata は使えない。" +
		"**tag に入れた名前はテンプレートタグとしては生えない**（MT 9 で実測。<mt:CFFeatured> は " +
		"unrecognized になる）。テンプレートから読むのは " +
		"<$mt:EntryCustomFieldValue basename=\"...\"$> か mt:EntryCustomFields のループ",
	en: "There are 13 types: text / textarea / checkbox / url / datetime / select / radio / embed / " +
		"post_type / image / file / audio / video. **The generic asset type is file** — its label is " +
		"\"Asset\", so asset is the tempting spelling, but it returns 409. " +
		"select, radio and datetime require options (for datetime: date_and_time / date_only / time_only). " +
		"Asset-typed fields take the URL as their value, not the ID, and it is not validated. " +
		"**The API does not create the objectasset association**, so the value renders in templates but the " +
		"object will not show up under \"where this asset is used\". " +
		"systemObject: entry / page / category / folder / comment / template work on any site; " +
		"blog only on a blog and website only on a website. author / asset / contentdata are not available. " +
		"**The tag name you set does not become a template tag** (measured on MT 9: <mt:CFFeatured> is " +
		"unrecognized); read the value with <$mt:EntryCustomFieldValue basename=\"...\"$> " +
		"or the mt:EntryCustomFields loop.",
}

// paramTypes は公式定義の型が実態と合わないパラメータを直します。
//
// permutate の4つは「並べ替えた後のリソース配列」を受け取るのに string と
// 書かれている。string のままだと mtb は @ファイル名 を展開せず、
// リテラルがそのまま送られて MT 側で JSON デコードに失敗する
// （500 malformed JSON string）。array にすれば他のリソース項目と同じ扱いになる。
var paramTypes = map[string]map[string]string{
	"POST /sites/{site_id}/categories/permutate":                                {"categories": "array"},
	"POST /sites/{site_id}/folders/permutate":                                   {"folders": "array"},
	"POST /sites/{site_id}/categorySets/{category_set_id}/categories/permutate": {"categories": "array"},
	"POST /sites/{site_id}/contentTypes/{content_type_id}/fields/permutate":     {"content_fields": "array"},
}

// skipPaths は手書きの `mtb auth` が受け持つため生成しないエンドポイントです。
var skipPaths = map[string]bool{
	"/authentication": true,
	"/token":          true,
	"/authorization":  true,
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gen-spec:", err)
		os.Exit(1)
	}
}

type rawSpec struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
	Components struct {
		Parameters map[string]json.RawMessage `json:"parameters"`
		Schemas    map[string]*rawSchema      `json:"schemas"`
	} `json:"components"`
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

type rawParam struct {
	Ref         string     `json:"$ref"`
	Name        string     `json:"name"`
	In          string     `json:"in"`
	Description string     `json:"description"`
	Required    bool       `json:"required"`
	Schema      *rawSchema `json:"schema"`
}

type rawSchema struct {
	Ref        string                `json:"$ref"`
	Desc       string                `json:"description"`
	Type       string                `json:"type"`
	Format     string                `json:"format"`
	Default    any                   `json:"default"`
	Enum       []any                 `json:"enum"`
	Required   []string              `json:"required"`
	Properties map[string]*rawSchema `json:"properties"`
	OneOf      []*rawSchema          `json:"oneOf"`
	Items      *rawSchema            `json:"items"`
}

type rawOp struct {
	Tags        []string   `json:"tags"`
	Summary     string     `json:"summary"`
	Description string     `json:"description"`
	Deprecated  bool       `json:"deprecated"`
	Parameters  []rawParam `json:"parameters"`
	RequestBody *struct {
		Required bool `json:"required"`
		Content  map[string]struct {
			Schema *rawSchema `json:"schema"`
		} `json:"content"`
	} `json:"requestBody"`
}

var httpMethods = []string{"get", "post", "put", "delete", "patch"}

func run() error {
	buf, err := readSource(*inPath)
	if err != nil {
		return err
	}
	var in rawSpec
	if err := json.Unmarshal(buf, &in); err != nil {
		return err
	}

	byCommand := map[string]*spec.Command{}
	var order []string

	for _, path := range sortedKeys(in.Paths) {
		if skipPaths[path] {
			continue
		}
		for _, method := range httpMethods {
			rawOpJSON, ok := in.Paths[path][method]
			if !ok {
				continue
			}
			var ro rawOp
			if err := json.Unmarshal(rawOpJSON, &ro); err != nil {
				return fmt.Errorf("%s %s: %w", method, path, err)
			}
			op, cmdPath, err := convert(&in, path, method, &ro)
			if err != nil {
				return err
			}
			key := strings.Join(cmdPath, " ")
			c, ok := byCommand[key]
			if !ok {
				c = &spec.Command{Command: cmdPath, Short: shortOf(op)}
				byCommand[key] = c
				order = append(order, key)
			}
			c.Ops = append(c.Ops, op)
		}
	}

	out := &spec.Spec{APIVersion: apiVersion, SpecVersion: in.Info.Version}
	sort.Strings(order)
	for _, key := range order {
		c := byCommand[key]
		if err := checkMerge(c); err != nil {
			return err
		}
		// パスパラメータの多い順。実行時は「与えられた引数で成立する最初の
		// オペレーション」を選ぶので、より具体的な方を先に置く。
		sort.SliceStable(c.Ops, func(i, j int) bool {
			return len(c.Ops[i].PathParams) > len(c.Ops[j].PathParams)
		})
		out.Commands = append(out.Commands, c)
	}
	out.Groups = groupsOf(out.Commands)

	enc, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	enc = append(enc, '\n')
	if err := os.WriteFile(*outPath, enc, 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d commands, %d operations, %d bytes)\n",
		*outPath, len(out.Commands), countOps(out.Commands), len(enc))
	return nil
}

// checkMerge は1コマンドに複数のオペレーションが集まったとき、それが
// 「末尾のパスパラメータの有無だけが違う同じ操作」かを確かめます。
// そうでない衝突は名前の付け方の誤りなので、黙って畳まずに失敗させます。
func checkMerge(c *spec.Command) error {
	if len(c.Ops) < 2 {
		return nil
	}
	for _, op := range c.Ops[1:] {
		if op.Method != c.Ops[0].Method {
			return fmt.Errorf("コマンド %q が別メソッドの操作に衝突しています: %s %s と %s %s。overrides に名前を足してください",
				strings.Join(c.Command, " "),
				c.Ops[0].Method, c.Ops[0].Path, op.Method, op.Path)
		}
		if len(op.PathParams) == len(c.Ops[0].PathParams) {
			return fmt.Errorf("コマンド %q が同形の操作に衝突しています: %s と %s。overrides に名前を足してください",
				strings.Join(c.Command, " "), c.Ops[0].Path, op.Path)
		}
	}
	return nil
}

func convert(in *rawSpec, path, method string, ro *rawOp) (*spec.Operation, []string, error) {
	// 公式の説明文（ro.Description）は持ち込まない。リファレンスへ誘導する。
	op := &spec.Operation{
		Method:     strings.ToUpper(method),
		Path:       path,
		Tags:       ro.Tags,
		Summary:    strings.TrimSpace(ro.Summary),
		DocURL:     docURL(ro.Tags, path, method),
		Deprecated: ro.Deprecated,
	}

	for _, rp := range ro.Parameters {
		p := rp
		if p.Ref != "" {
			resolved, err := resolveParam(in, p.Ref)
			if err != nil {
				return nil, nil, fmt.Errorf("%s %s: %w", method, path, err)
			}
			p = *resolved
		}
		// 認証ヘッダはクライアントが付けるので、フラグには出さない。
		if p.In == "header" {
			continue
		}
		param := toParam(p.Name, p.Required, p.Schema)
		switch p.In {
		case "path":
			param.Required = true
			if param.Description == "" {
				param.Description, param.DescriptionEn = pathParamDesc(param.Name)
			}
			op.PathParams = append(op.PathParams, param)
		case "query":
			op.Query = append(op.Query, param)
		}
	}

	if ro.RequestBody != nil {
		for _, mime := range []string{"application/x-www-form-urlencoded", "multipart/form-data"} {
			content, ok := ro.RequestBody.Content[mime]
			if !ok || content.Schema == nil {
				continue
			}
			op.BodyKind = "form"
			if mime == "multipart/form-data" {
				op.BodyKind = "multipart"
			}
			required := map[string]bool{}
			for _, name := range content.Schema.Required {
				required[name] = true
			}
			for _, name := range sortedKeys(content.Schema.Properties) {
				sub := content.Schema.Properties[name]
				bp := toParam(name, required[name], sub)
				bp.Fields = schemaFields(in, sub)
				op.Body = append(op.Body, bp)
			}
			break
		}
	}

	if types, ok := paramTypes[op.Method+" "+path]; ok {
		for _, group := range [][]*spec.Param{op.Body, op.Query} {
			for _, p := range group {
				if t, ok := types[p.Name]; ok {
					p.Type = t
				}
			}
		}
	}

	if notes, ok := paramNotes[op.Method+" "+path]; ok {
		for _, group := range [][]*spec.Param{op.Body, op.Query, op.PathParams} {
			for _, p := range group {
				n, ok := notes[p.Name]
				if !ok {
					continue
				}
				base := strings.TrimSpace(p.Description)
				if base == "" {
					p.Description, p.DescriptionEn = n.ja, n.en
					continue
				}
				p.Description = strings.TrimSuffix(base, ".") + "。" + n.ja
				p.DescriptionEn = strings.TrimSuffix(base, ".") + ". " + n.en
			}
		}
	}

	if extra, ok := extraQuery[op.Method+" "+path]; ok {
		have := map[string]bool{}
		for _, p := range op.Query {
			have[p.Name] = true
		}
		for _, p := range extra {
			if !have[p.Name] {
				op.Query = append(op.Query, p)
			}
		}
	}

	if extra, ok := extraBody[op.Method+" "+path]; ok {
		if op.BodyKind == "" {
			op.BodyKind = "form"
		}
		have := map[string]bool{}
		for _, p := range op.Body {
			have[p.Name] = true
		}
		for _, p := range extra {
			if !have[p.Name] {
				op.Body = append(op.Body, p)
			}
		}
		sort.SliceStable(op.Body, func(i, j int) bool { return op.Body[i].Name < op.Body[j].Name })
	}

	sort.SliceStable(op.Query, func(i, j int) bool { return op.Query[i].Name < op.Query[j].Name })
	return op, commandFor(method, path), nil
}

func resolveParam(in *rawSpec, ref string) (*rawParam, error) {
	const prefix = "#/components/parameters/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, fmt.Errorf("解決できない $ref: %s", ref)
	}
	buf, ok := in.Components.Parameters[strings.TrimPrefix(ref, prefix)]
	if !ok {
		return nil, fmt.Errorf("components.parameters に %s がありません", ref)
	}
	var p rawParam
	if err := json.Unmarshal(buf, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// schemaFields はリソース本体が取る項目の名前を返します。$ref なら
// components.schemas を引きます。1段だけで、入れ子の中までは辿りません
// （--help に並べるための一覧なので、深さより見通しを優先します）。
func schemaFields(in *rawSpec, s *rawSchema) []string {
	if s == nil {
		return nil
	}
	if s.Ref != "" {
		const prefix = "#/components/schemas/"
		if !strings.HasPrefix(s.Ref, prefix) {
			return nil
		}
		s = in.Components.Schemas[strings.TrimPrefix(s.Ref, prefix)]
		if s == nil {
			return nil
		}
	}
	if len(s.Properties) == 0 {
		return nil
	}
	return sortedKeys(s.Properties)
}

// toParam はパラメータの事実（名前・型・既定値・選択肢）だけを取ります。
// 公式の説明文は持ち込みません（パッケージの説明を参照）。
func toParam(name string, required bool, s *rawSchema) *spec.Param {
	p := &spec.Param{
		Name:     name,
		Flag:     kebab(name),
		Type:     "string",
		Required: required,
	}
	if s == nil {
		return p
	}
	switch {
	case s.Ref != "":
		// entry / blog / template などのリソース本体。JSON をそのまま渡す。
		p.Type = "object"
	case s.Type == "object", len(s.Properties) > 0:
		p.Type = "object"
	case len(s.OneOf) > 0:
		p.Type = "object"
	case s.Type == "string" && s.Format == "binary":
		p.Type = "file"
	case s.Type != "":
		p.Type = s.Type
	}
	if s.Default != nil {
		p.Default = scalar(s.Default)
	}
	for _, v := range s.Enum {
		p.Enum = append(p.Enum, scalar(v))
	}
	// スカラーの配列は同じ名前を繰り返して送る。オブジェクトの配列
	// （permutate の並べ替えなど）は JSON のまま送るので区別する。
	if s.Type == "array" && s.Items != nil && s.Items.Ref == "" && len(s.Items.Properties) == 0 {
		p.Repeated = true
	}
	// 配列の選択肢は items の側にある。ここを拾わないと、渡せる値が
	// --help のどこにも出ない（MT もエラーで教えてくれない）。
	if len(p.Enum) == 0 && s.Items != nil {
		for _, v := range s.Items.Enum {
			p.Enum = append(p.Enum, scalar(v))
		}
	}
	return p
}

// pathParamDesc は説明が空のパスパラメータを補います。OpenAPI 側は
// path パラメータに description を持たないため、--help が名前だけになるのを防ぎます。
func pathParamDesc(name string) (ja, en string) {
	if name == "site_id" {
		return "対象サイトのID（プロファイルに既定サイトがあれば省略できます）",
			"ID of the target site (can be omitted when the profile has a default site)"
	}
	return "パスパラメータ " + name, "path parameter " + name
}

func scalar(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	case float64:
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'g', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

// commandFor はパスとメソッドからコマンドパスを決めます。
//
//	リテラルの並び      → コマンドのグループ
//	末尾がパラメータ    → GET=get / PUT=update / DELETE=delete / POST=create
//	末尾が動詞リテラル  → その動詞そのもの
//	末尾がリテラル      → GET=list / POST=create / PUT=update / DELETE=delete
func commandFor(method, path string) []string {
	if o, ok := overrides[strings.ToUpper(method)+" "+path]; ok {
		return append([]string(nil), o...)
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	var literals []string
	lastIsParam := false
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		if strings.HasPrefix(seg, "{") {
			lastIsParam = true
			continue
		}
		lastIsParam = false
		literals = append(literals, seg)
	}

	var verb string
	if !lastIsParam && len(literals) > 0 && actions[literals[len(literals)-1]] {
		verb = kebab(literals[len(literals)-1])
		literals = literals[:len(literals)-1]
	} else {
		verb = verbFor(method, lastIsParam)
	}

	cmd := make([]string, 0, len(literals)+1)
	for _, l := range literals {
		cmd = append(cmd, kebab(l))
	}
	return append(cmd, verb)
}

func verbFor(method string, item bool) string {
	switch strings.ToUpper(method) {
	case "GET":
		if item {
			return "get"
		}
		return "list"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	}
	return strings.ToLower(method)
}

// kebab は camelCase / PascalCase / snake_case をケバブケースにします。
// contentTypes→content-types, IncludeBlogs→include-blogs, site_id→site-id。
func kebab(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		switch {
		case r == '_' || r == ' ':
			b.WriteRune('-')
			continue
		case r >= 'A' && r <= 'Z':
			prevLower := i > 0 && (runes[i-1] >= 'a' && runes[i-1] <= 'z' || runes[i-1] >= '0' && runes[i-1] <= '9')
			nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
			prevUpper := i > 0 && runes[i-1] >= 'A' && runes[i-1] <= 'Z'
			if prevLower || (prevUpper && nextLower) {
				b.WriteRune('-')
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), "-")
}

func shortOf(op *spec.Operation) string {
	if op.Summary != "" {
		return firstLine(op.Summary)
	}
	if s, ok := ownShorts[op.Method+" "+op.Path]; ok {
		return s
	}
	return op.Method + " " + op.Path
}

// docURL は公式リファレンスの、その操作の箇所です。
func docURL(tags []string, path, method string) string {
	if len(tags) == 0 {
		return referenceURL
	}
	return referenceURL + "#tag/" + tags[0] + "/paths/" + strings.ReplaceAll(path, "/", "~1") + "/" + strings.ToLower(method)
}

// readSource は原本を読みます。path が空なら固定したコミットから取り、
// 利用者のキャッシュに置きます（2回目からは取りに行きません）。
func readSource(path string) ([]byte, error) {
	if path != "" {
		return os.ReadFile(path)
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	cache := filepath.Join(dir, "mtb-gen-spec", "openapi-v7-"+sourceCommit[:12]+".json")
	if buf, err := os.ReadFile(cache); err == nil {
		return buf, nil
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("公式定義を取得できません: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("公式定義を取得できません: %s %s", resp.Status, sourceURL)
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(cache), 0o755); err == nil {
		_ = os.WriteFile(cache, buf, 0o644)
	}
	return buf, nil
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\n"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(strings.TrimPrefix(s, "- "))
}

// groupsOf は中間ノードを洗い出し、説明文（OpenAPI のタグ名）を決めます。
//
// 単純に「配下で最も多いタグ」にすると、sites のように何でもぶら下がる
// グループが Categories を名乗ってしまいます。そこで、そのグループ自身の
// 一覧・取得・作成コマンドのタグ → 配下すべてに共通するタグ → 最頻タグ、
// の順で選びます。
func groupsOf(commands []*spec.Command) []*spec.Group {
	direct := map[string]string{}          // グループ自身の CRUD コマンドのタグ
	common := map[string]map[string]bool{} // 配下すべてに共通するタグ
	seen := map[string]bool{}              // common を初期化済みか
	counts := map[string]map[string]int{}

	for _, c := range commands {
		for i := 1; i < len(c.Command); i++ {
			key := strings.Join(c.Command[:i], " ")
			if counts[key] == nil {
				counts[key] = map[string]int{}
			}
			for _, op := range c.Ops {
				if len(op.Tags) == 0 {
					continue
				}
				for _, t := range op.Tags {
					counts[key][t]++
				}
				if i == len(c.Command)-1 && isCRUDVerb(c.Name()) && direct[key] == "" {
					direct[key] = op.Tags[0]
				}
				common[key] = intersectTags(common[key], seen[key], op.Tags)
				seen[key] = true
			}
		}
	}

	var groups []*spec.Group
	for _, key := range sortedKeys(counts) {
		segments := strings.Split(key, " ")
		short := matchingTag(counts[key], segments[len(segments)-1])
		if short == "" {
			short = direct[key]
		}
		if short == "" {
			short = firstTag(common[key])
		}
		if short == "" {
			short = topTag(counts[key])
		}
		groups = append(groups, &spec.Group{Command: strings.Split(key, " "), Short: short})
	}
	return groups
}

// isCRUDVerb はグループを代表するコマンドかどうかです。
func isCRUDVerb(name string) bool {
	return name == "list" || name == "get" || name == "create"
}

// intersectTags はこれまでのタグ集合と次の候補の積を取ります。
func intersectTags(current map[string]bool, initialized bool, tags []string) map[string]bool {
	next := map[string]bool{}
	if !initialized {
		for _, t := range tags {
			next[t] = true
		}
		return next
	}
	for _, t := range tags {
		if current[t] {
			next[t] = true
		}
	}
	return next
}

// matchingTag はグループ名そのものを指すタグを探します。
// Categories / Content Types のように、タグ名とパスのセグメントが
// 綴りだけ違うものを正しく拾うためです。
func matchingTag(counts map[string]int, segment string) string {
	for _, name := range sortedKeys(counts) {
		if kebab(name) == segment {
			return name
		}
	}
	return ""
}

func firstTag(set map[string]bool) string {
	if len(set) == 0 {
		return ""
	}
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names[0]
}

// topTag は最も多く現れたタグを返します。同数なら名前順で先のものです。
func topTag(counts map[string]int) string {
	best, bestN := "", -1
	for _, name := range sortedKeys(counts) {
		if counts[name] > bestN {
			best, bestN = name, counts[name]
		}
	}
	return best
}

func countOps(commands []*spec.Command) int {
	n := 0
	for _, c := range commands {
		n += len(c.Ops)
	}
	return n
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
