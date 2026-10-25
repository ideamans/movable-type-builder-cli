# コンテンツタイプ (Content Data)

MT7 以降の構造化コンテンツ。任意のフィールドを定義できる。

現行のタグ 50 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTContentAuthorDisplayName — FUNCTION
コンテンツデータを作ったユーザーの表示名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentauthordisplayname.html)

### MTContentAuthorEmail — FUNCTION
コンテンツデータを作ったユーザーのメールアドレス

属性: `spam_protect="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentauthoremail.html)

### MTContentAuthorID — FUNCTION
コンテンツデータを作ったユーザーの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentauthorid.html)

### MTContentAuthorLink — FUNCTION
コンテンツデータを作ったユーザーの表示名をリンク付きで

属性: `show_url="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentauthorlink.html)

### MTContentAuthorURL — FUNCTION
コンテンツデータを作ったユーザーのウェブサイト URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentauthorurl.html)

### MTContentAuthorUsername — FUNCTION
コンテンツデータを作ったユーザーのユーザー名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentauthorusername.html)

### MTContentAuthorUserpic — FUNCTION
コンテンツデータを作ったユーザーのプロフィール画像の img 要素

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentauthoruserpic.html)

### MTContentAuthorUserpicAsset — BLOCK
コンテンツデータを作ったユーザーのプロフィール画像をアセットとして文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentauthoruserpicasset.html)

### MTContentAuthorUserpicURL — FUNCTION
コンテンツデータを作ったユーザーのプロフィール画像の URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentauthoruserpicurl.html)

### MTContentCalendar — BLOCK
コンテンツデータ用の1か月分のカレンダーを組むループ

属性: `content_type="コンテンツタイプのユニークID | コンテンツタイプのID | コンテンツタイプの名前"` `date_field="foo"` `month="this | YYYYMM | last | next"` `category_set="foo"` `category="foo"` `weeks_start_with="0 | 6 | 5 | 4 | 3 | 2 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentcalendar.html)

### MTContentCreatedDate — FUNCTION
コンテンツデータを作った日時（公開日ではない）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentcreateddate.html)

### MTContentDataSearchScript — FUNCTION
コンテンツデータ検索の CGI のファイル名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentdatasearchscript.html)

### MTContentDate — FUNCTION
コンテンツデータの公開日時（format で書式指定）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentdate.html)

### MTContentField — BLOCK
コンテンツデータの1つのフィールドを文脈にする（content_field で選ぶ）

属性: `content_field="foo"` `glue="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentfield.html)

### MTContentFieldFooter — BLOCK
MTContentField の値の最後の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentfieldfooter.html)

### MTContentFieldHeader — BLOCK
MTContentField の値の最初の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentfieldheader.html)

### MTContentFieldLabel — FUNCTION
コンテンツフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentfieldlabel.html)

### MTContentFieldValue — FUNCTION
いまのコンテンツフィールドの値

属性: `convert_breaks="0 | 1"` `words="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentfieldvalue.html)

### MTContentFields — BLOCK
コンテンツタイプの全フィールドを並べるループ

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentfields.html)

### MTContentFieldsFooter — BLOCK
MTContentFields の最後の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentfieldsfooter.html)

### MTContentFieldsHeader — BLOCK
MTContentFields の最初の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentfieldsheader.html)

### MTContentID — FUNCTION
コンテンツデータの ID

属性: `pad="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentid.html)

### MTContentIdentifier — FUNCTION
コンテンツデータのベースネーム（出力ファイル名）

属性: `separator="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentidentifier.html)

### MTContentLabel — FUNCTION
コンテンツデータのデータ識別ラベル

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentlabel.html)

### MTContentModifiedAuthorDisplayName — FUNCTION
コンテンツデータを最後に更新したユーザーの表示名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentmodifiedauthordisplay.html)

### MTContentModifiedAuthorEmail — FUNCTION
コンテンツデータを最後に更新したユーザーのメールアドレス

属性: `spam_protect="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentmodifiedauthoremail.html)

### MTContentModifiedAuthorID — FUNCTION
コンテンツデータを最後に更新したユーザーの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentmodifiedauthorid.html)

### MTContentModifiedAuthorLink — FUNCTION
コンテンツデータを最後に更新したユーザーの表示名をリンク付きで

属性: `show_url="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentmodifiedauthorlink.html)

### MTContentModifiedAuthorURL — FUNCTION
コンテンツデータを最後に更新したユーザーのウェブサイト URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentmodifiedauthorurl.html)

### MTContentModifiedAuthorUsername — FUNCTION
コンテンツデータを最後に更新したユーザーのユーザー名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentmodifiedauthorusername.html)

### MTContentModifiedAuthorUserpicURL — FUNCTION
コンテンツデータを最後に更新したユーザーのプロフィール画像の URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentmodifiedauthoruserpic.html)

### MTContentModifiedDate — FUNCTION
コンテンツデータを最後に保存した日時（再構築の日時ではない）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentmodifieddate.html)

### MTContentNext — BLOCK
次のコンテンツデータを文脈にする

属性: `by_author="0 | 1"` `category_field="foo"` `date_field="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentnext.html)

### MTContentPermalink — FUNCTION
コンテンツデータの公開 URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentpermalink.html)

### MTContentPrevious — BLOCK
前のコンテンツデータを文脈にする

属性: `by_author="0 | 1"` `category_field="foo"` `date_field="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentprevious.html)

### MTContentSiteDescription — FUNCTION
コンテンツデータが属するサイトの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentsitedescription.html)

### MTContentSiteID — FUNCTION
コンテンツデータが属するサイトの ID

属性: `pad="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentsiteid.html)

### MTContentSiteName — FUNCTION
コンテンツデータが属するサイトの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentsitename.html)

### MTContentSiteURL — FUNCTION
コンテンツデータが属するサイトの URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentsiteurl.html)

### MTContentStatus — FUNCTION
コンテンツデータの状態（Publish・Draft など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentstatus.html)

### MTContentTypeDescription — FUNCTION
コンテンツタイプの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contenttypedescription.html)

### MTContentTypeID — FUNCTION
コンテンツタイプの ID

属性: `pad="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contenttypeid.html)

### MTContentTypeName — FUNCTION
コンテンツタイプの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contenttypename.html)

### MTContentTypeUniqueId — FUNCTION
コンテンツタイプのユニーク ID（環境をまたいでも変わらない）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contenttypeuniqueid.html)

### MTContentUniqueID — FUNCTION
コンテンツデータのユニーク ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentuniqueid.html)

### MTContentUnpublishedDate — FUNCTION
コンテンツデータの公開終了日時

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentunpublisheddate.html)

### MTContents — BLOCK
コンテンツデータを並べるループ（content_type で種類を選ぶ）

属性: `content_type` `include_sites` `exclude_sites` `limit` `offset` `author` `sort_by` `sort_order="ascend | descend"` `field:コンテンツフィールドの名前またはコンテンツフィールドのユニークID="foo"` `id="N"` `unique_id="XX"` `days="N"` `date_field="foo"` `unique="0 | 1"` `glue="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contents.html)

### MTContentsCount — FUNCTION
いまの文脈で出るコンテンツデータの数

属性: `content_type` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentscount.html)

### MTContentsFooter — BLOCK
MTContents の最後の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentsfooter.html)

### MTContentsHeader — BLOCK
MTContents の最初の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/contentsheader.html)
