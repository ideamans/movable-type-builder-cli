# 制御構文・変数・インクルード

条件分岐、繰り返し、変数、モジュールの読み込み。

現行のタグ 78 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTApp:ActionBar — FUNCTION
管理画面用。一覧・編集画面の操作バーを組む

属性: `bar_position="foo"` `hide_pager="0 | 1"` `form_id="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/app:actionbar.html)

### MTApp:Form — BLOCK
管理画面用。フォームを組む

属性: `action="url"` `object_id="id"` `blog_id="id"` `object_type="foo"` `name="foo"` `enctype="enctype_foo"` `id="id"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/app:form.html)

### MTApp:Link — FUNCTION
管理画面用。管理画面内へのリンクを組む

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/app:link.html)

### MTApp:ListFilters — FUNCTION
管理画面用。一覧の絞り込みを組む

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/app:listfilters.html)

### MTApp:Listing — BLOCK
管理画面用。一覧の表を組む

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/app:listing.html)

### MTApp:PageActions — FUNCTION
管理画面用。ページの操作欄を組む

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/app:pageactions.html)

### MTApp:Setting — BLOCK
管理画面用。設定項目の1行を組む

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/app:setting.html)

### MTApp:SettingGroup — BLOCK
管理画面用。設定項目のまとまりを組む

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/app:settinggroup.html)

### MTApp:StatusMsg — BLOCK
管理画面用。状態メッセージを出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/app:statusmsg.html)

### MTApp:Widget — BLOCK
管理画面用。ウィジェットの枠を組む

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/app:widget.html)

### MTBlockEditorBlockAsset — FUNCTION
ブロックエディタの画像・ファイルのブロックをアセットとして出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/mtblockeditorblockasset.html)

### MTBlockEditorBlocks — FUNCTION
ブロックエディタのブロックを並べる（マルチカラムは入れ子で）

属性: `tag="name_foo"` `var="name_foo"` `name="name_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/mtblockeditorblocks.html)

### MTElse — BLOCK
条件に合わなかったときの中身（終了タグは省略可）

属性: `name="foo"` `tag="TagName"` `like="foo"` `eq="foo"` `ne="foo"` `lt="foo"` `gt="foo"` `le="foo"` `ge="foo"` `op` `index` `key` `test`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/else.html)

### MTElseIf — BLOCK
MTIf の続きの条件

属性: `name="foo"` `tag="TagName"` `like="foo"` `eq="foo"` `ne="foo"` `lt="foo"` `gt="foo"` `le="foo"` `ge="foo"` `op` `index` `key` `test`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/elseif.html)

### MTFor — BLOCK
for ループ（from・to・step）

属性: `from` `start` `to` `end` `increment` `step` `glue="foo"` `var` `__first__` `__last__` `__odd__` `__even__` `__index__` `__counter__`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/for.html)

### MTIf — BLOCK
条件で出し分ける（変数・タグの値を比べる）

属性: `name="foo"` `tag="TagName"` `like="foo"` `eq="foo"` `ne="foo"` `lt="foo"` `gt="foo"` `le="foo"` `ge="foo"` `op` `index` `key` `test`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/if.html)

### MTIfAllowCommentHTML — BLOCK
コメントで一部の HTML を許していれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifallowcommenthtml.html)

### MTIfArchiveType — BLOCK
指定の種類のアーカイブを出力中なら中身を出す

属性: `archive_type="archive_type_foo"` `type="archive_type_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifarchivetype.html)

### MTIfArchiveTypeEnabled — BLOCK
指定の種類のアーカイブを使っていれば中身を出す

属性: `archive_type="archive_type_foo"` `type="archive_type_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifarchivetypeenabled.html)

### MTIfBlog — FUNCTION
ブログの文脈なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifblog.html)

### MTIfCategory — BLOCK
文脈か記事が指定のカテゴリに属していれば中身を出す

属性: `name="category_foo"` `label="category_foo"` `type="primary | secondary"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcategory.html)

### MTIfCommentParent — BLOCK
コメントに返信元があれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcommentparent.html)

### MTIfCommentReplies — BLOCK
コメントに返信があれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcommentreplies.html)

### MTIfCommenterIsAuthor — BLOCK
コメントした人が MT の登録ユーザーなら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcommenterisauthor.html)

### MTIfCommenterIsEntryAuthor — BLOCK
コメントした人が記事の作成者なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcommenterisentryauthor.html)

### MTIfCommenterTrusted — BLOCK
コメントした人が登録済み（信頼済み）なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcommentertrusted.html)

### MTIfCommentsAccepted — BLOCK
サイトと記事の両方がコメントを受け付けていれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcommentsaccepted.html)

### MTIfCommentsActive — BLOCK
コメントを受け付けているか、すでにコメントがあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcommentsactive.html)

### MTIfCommentsAllowed — BLOCK
サイトがコメントを受け付けていれば中身を出す（記事の設定は見ない）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcommentsallowed.html)

### MTIfCommentsModerated — BLOCK
コメントをすぐ公開しない設定なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcommentsmoderated.html)

### MTIfCurrentPage — BLOCK
MTPagerBlock の中で、いまのページなら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifcurrentpage.html)

### MTIfDynamic — BLOCK
テンプレートがダイナミック公開なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifdynamic.html)

### MTIfExternalUserManagement — BLOCK
外部ユーザー管理（ExternalUserManagement）が有効なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/mtifexternalusermanagement.html)

### MTIfFolder — BLOCK
文脈が指定のフォルダを含めば中身を出す

属性: `name="folder_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/iffolder.html)

### MTIfImageSupport — BLOCK
サーバに画像処理ライブラリがあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifimagesupport.html)

### MTIfIsAncestor — BLOCK
child のカテゴリがいまのカテゴリの子孫なら中身を出す

属性: `child="category_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifisancestor.html)

### MTIfIsDescendant — BLOCK
parent のカテゴリがいまのカテゴリの祖先なら中身を出す

属性: `parent="category_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifisdescendant.html)

### MTIfMoreResults — BLOCK
検索結果が1ページの件数より多ければ中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifmoreresults.html)

### MTIfNeedEmail — BLOCK
コメントでメールアドレスが必須なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifneedemail.html)

### MTIfPingsAccepted — BLOCK
サイトと記事の両方がトラックバックを受け付けていれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifpingsaccepted.html)

### MTIfPingsActive — BLOCK
トラックバックを受け付けているか、すでにあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifpingsactive.html)

### MTIfPingsAllowed — BLOCK
サイトがトラックバックを受け付けていれば中身を出す（記事の設定は見ない）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifpingsallowed.html)

### MTIfPingsModerated — BLOCK
トラックバックをすぐ公開しない設定なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifpingsmoderated.html)

### MTIfPreviousResults — BLOCK
検索結果に前のページがあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifpreviousresults.html)

### MTIfRegistrationAllowed — BLOCK
サイトからのユーザー登録を許していれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifregistrationallowed.html)

### MTIfRegistrationNotRequired — BLOCK
匿名・未認証のコメントを許していれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifregistrationnotrequired.html)

### MTIfRegistrationRequired — BLOCK
コメントに認証が必要なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifregistrationrequired.html)

### MTIfRequireCommentEmails — BLOCK
コメントでメールアドレスが必須なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifrequirecommentemails.html)

### MTIfStatic — BLOCK
テンプレートがスタティック公開なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifstatic.html)

### MTIfStraightSearch — BLOCK
キーワード検索なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifstraightsearch.html)

### MTIfTagSearch — BLOCK
タグ検索なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/iftagsearch.html)

### MTIfWebsite — BLOCK
ウェブサイトの文脈なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ifwebsite.html)

### MTIgnore — BLOCK
囲んだ中身を処理も出力もしない（コメントアウト）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/ignore.html)

### MTInclude — FUNCTION
テンプレートモジュールやファイルを読み込む（module・file など）

属性: `file="file_foo"` `module="module_foo"` `widget="widget_foo"` `identifier="index_template_identifier"` `name="foo"` `blog_id="ID"` `site_id="ID"` `variable_foo="value"` `global="0 | 1"` `parent="0 | 1"` `local="0 | 1"` `cache="0 | 1"` `cache_key="cache_key"` `key="cache_key"` `ttl="0"` `ssi="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/include.html)

### MTIncludeBlock — BLOCK
モジュールを読み込み、囲んだ中身をモジュールの決まった位置に差し込む

属性: `var="variable_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/includeblock.html)

### MTIndexBasename — FUNCTION
インデックスのファイル名（IndexBasename。既定は index）

属性: `extension="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/indexbasename.html)

### MTIndexLink — FUNCTION
インデックステンプレートの出力の絶対 URL（MTIndexList の中）

属性: `with_index="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/indexlink.html)

### MTIndexList — BLOCK
インデックステンプレートを並べるループ

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/indexlist.html)

### MTIndexName — FUNCTION
インデックステンプレートの名前（MTIndexList の中）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/indexname.html)

### MTLoop — BLOCK
配列・ハッシュの変数を回すループ（name で変数を選ぶ）

属性: `name="foo"` `var="foo"` `sort_by="foo"` `glue="foo"` `__first__` `__last__` `__odd__` `__even__` `__key__` `__value__`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/loop.html)

### MTSection — BLOCK
囲んだ中身をキャッシュする（管理画面用）

属性: `cache_prefix="key_foo"` `period="Sec"` `by_blog="0 | 1"` `by_user="0 | 1"` `html_tag="element_foo"` `id="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/section.html)

### MTSetHashVar — BLOCK
キーと値の組をまとめてハッシュ変数にする

属性: `name="variable_foo"` `var="variable_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sethashvar.html)

### MTSetVar — FUNCTION
変数に値を入れる（name・value。呼ぶのは MTVar）

属性: `name="foo"` `var="foo"` `value="bar"` `append="0 | 1"` `prepend="0 | 1"` `op` `index` `key` `function`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/setvar.html)

### MTSetVarBlock — BLOCK
囲んだ中身（タグを処理した結果）を変数に入れる

属性: `name="name_foo"` `var="name_foo"` `append="0 | 1"` `prepend="0 | 1"` `op` `index` `key` `function`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/setvarblock.html)

### MTSetVarTemplate — BLOCK
囲んだ中身をテンプレートとして変数に入れる（呼ぶたびに処理される）

属性: `name="name_foo"` `var="name_foo"` `append="0 | 1"` `prepend="0 | 1"` `index` `key` `function`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/setvartemplate.html)

### MTSetVars — BLOCK
複数の変数を「名前=値」の行でまとめて入れる

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/setvars.html)

### MTSubCatIsFirst — BLOCK
カテゴリのループの最初の1回で中身を出す（ul の開始など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/subcatisfirst.html)

### MTSubCatIsLast — BLOCK
カテゴリのループの最後の1回で中身を出す（ul の終了など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/subcatislast.html)

### MTSubCatsRecurse — FUNCTION
MTSubCategories を子カテゴリへ繰り返す（max_depth で深さ）

属性: `max_depth="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/subcatsrecurse.html)

### MTTemplateCreatedOn — FUNCTION
テンプレートを作った日時

属性: `language="format"` `format="format"` `format_name="format_name"` `utc="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/templatecreatedon.html)

### MTTemplateCustomFieldDescription — FUNCTION
テンプレートのカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/templatecustomfielddescription.html)

### MTTemplateCustomFieldName — FUNCTION
テンプレートのカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/templatecustomfieldname.html)

### MTTemplateCustomFieldValue — FUNCTION
テンプレートのカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/templatecustomfieldvalue.html)

### MTTemplateCustomFields — BLOCK
テンプレートのカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/templatecustomfields.html)

### MTTemplateNote — FUNCTION
何も出さない（属性に書いた内容をテンプレート内のメモにする）

属性: `value="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/templatenote.html)

### MTUnless — BLOCK
条件に合わないときに中身を出す（MTIf の逆）

属性: `name="foo"` `tag="TagName"` `like="foo"` `eq="foo"` `ne="foo"` `lt="foo"` `gt="foo"` `le="foo"` `ge="foo"` `op` `index` `key` `test`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/unless.html)

### MTVar — FUNCTION
変数の値（MTGetVar と同じ。配列・ハッシュの要素も取れる）

属性: `name="foo"` `var="foo"` `value="bar"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/var.html)

### MTWidgetSet — FUNCTION
name で指定したウィジェットセットを出す

属性: `file="file_foo"` `module="module_foo"` `widget="widget_foo"` `identifier="index_template_identifier"` `name="foo"` `blog_id="ID"` `variable_foo="value"` `global="0 | 1"` `parent="0 | 1"` `local="0 | 1"` `cache="0 | 1"` `cache_key="cache_key"` `key="cache_key"` `ttl="0"` `ssi="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/widgetset.html)
