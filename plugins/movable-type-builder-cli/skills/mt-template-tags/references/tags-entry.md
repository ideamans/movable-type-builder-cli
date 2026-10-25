# 記事 (Entry)

記事の一覧・取得・属性。ブログ記事の中心。

現行のタグ 75 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTBlogEntryCount — FUNCTION
サイトの公開済み記事の数

属性: `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogentrycount.html)

### MTEntries — BLOCK
記事を並べるループ（件数・カテゴリ・タグ・並び順で絞れる）

属性: `lastn="N"` `limit="N"` `sort_by="foo"` `sort_order="ascend | descend"` `field:customfieldbasename="foo"` `namespace="foo"` `class_type="entry | page"` `offset="N"` `category="category_foo"` `categories="category_foo"` `include_subcategories="1 | 0"` `tag="tag_foo"` `tags="tag_foo AND tag_bar"` `author="author_user_name"` `id="entry_id"` `min_score="score"` `max_score="score"` `min_rate="rate"` `max_rate="rate"` `min_count="count"` `max_count="count"` `scored_by="author_user_name"` `days="N"` `recently_commented_on="N"` `unique="1 | 0"` `glue="foo"` `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entries.html)

### MTEntriesCount — FUNCTION
いまの文脈で出る記事の数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entriescount.html)

### MTEntriesFooter — BLOCK
MTEntries の最後の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entriesfooter.html)

### MTEntriesHeader — BLOCK
MTEntries の最初の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entriesheader.html)

### MTEntriesWithSubCategories — BLOCK
サブカテゴリの記事も含めて並べるループ

属性: `lastn="N"` `limit="N"` `sort_by="foo"` `sort_order="ascend | descend"` `field:customfieldbasename="foo"` `namespace="foo"` `offset="N"` `category="category_foo"` `categories="category_foo"` `tag="tag_foo"` `tags="tag_foo AND tag_bar"` `author="author_user_name"` `id="entry_id"` `min_score="score"` `max_score="score"` `min_rate="rate"` `max_rate="rate"` `min_count="count"` `max_count="count"` `scored_by="author_user_name"` `days="N"` `recently_commented_on="N"` `unique="1 | 0"` `glue="foo"` `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrieswithsubcategories.html)

### MTEntryAdditionalCategories — BLOCK
記事のメイン以外のカテゴリを並べるループ

属性: `glue="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryadditionalcategories.html)

### MTEntryAssets — BLOCK
記事が使っているアセットを並べるループ

属性: `lastn="N"` `offset="N"` `limit="N"` `author="author_foo"` `type="image | audio | video | file"` `sort_order="ascend | descend"` `sort_by="file_name | created_by | created_on"` `assets_per_row="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryassets.html)

### MTEntryAtomID — FUNCTION
記事の Atom ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryatomid.html)

### MTEntryAuthorDisplayName — FUNCTION
記事の作成者の表示名（未設定なら空）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryauthordisplayname.html)

### MTEntryAuthorEmail — FUNCTION
記事の作成者のメールアドレス

属性: `spam_protect="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryauthoremail.html)

### MTEntryAuthorID — FUNCTION
記事・ページの作成者の ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryauthorid.html)

### MTEntryAuthorLink — FUNCTION
記事の作成者の表示名をリンク付きで（ウェブサイト URL へ）

属性: `show_url="0 | 1"` `show_email="0 | 1"` `new_window="0 | 1"` `spam_protect="0 | 1"` `type="url | email | archive"` `show_hcard="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryauthorlink.html)

### MTEntryAuthorURL — FUNCTION
記事の作成者のウェブサイト URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryauthorurl.html)

### MTEntryAuthorUsername — FUNCTION
記事の作成者のユーザー名（公開ページでは表示名を推奨）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryauthorusername.html)

### MTEntryAuthorUserpic — FUNCTION
記事の作成者のプロフィール画像の img 要素

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryauthoruserpic.html)

### MTEntryAuthorUserpicAsset — BLOCK
記事の作成者のプロフィール画像をアセットとして文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryauthoruserpicasset.html)

### MTEntryAuthorUserpicURL — FUNCTION
記事の作成者のプロフィール画像の URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryauthoruserpicurl.html)

### MTEntryBasename — FUNCTION
記事のベースネーム（出力ファイル名）

属性: `separator="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrybasename.html)

### MTEntryBlogDescription — FUNCTION
記事が属するサイトの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryblogdescription.html)

### MTEntryBlogID — FUNCTION
記事が属するサイトの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryblogid.html)

### MTEntryBlogName — FUNCTION
記事が属するサイトの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryblogname.html)

### MTEntryBlogURL — FUNCTION
記事が属するサイトの URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryblogurl.html)

### MTEntryBody — FUNCTION
記事の本文

属性: `convert_breaks="0 | 1"` `words="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrybody.html)

### MTEntryCategories — BLOCK
記事のカテゴリを並べるループ（メインを含む）

属性: `glue="foo"` `type="primary"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrycategories.html)

### MTEntryCategory — FUNCTION
記事のメインカテゴリの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrycategory.html)

### MTEntryClass — FUNCTION
記事かウェブページか（entry / page。言語に依らない）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryclass.html)

### MTEntryClassLabel — FUNCTION
記事かウェブページかの表示名（言語に合わせる）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryclasslabel.html)

### MTEntryCommentCount — FUNCTION
記事のコメントの数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrycommentcount.html)

### MTEntryCreatedDate — FUNCTION
記事を作った日時（公開日ではない）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrycreateddate.html)

### MTEntryCustomFieldDescription — FUNCTION
記事のカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrycustomfielddescription.html)

### MTEntryCustomFieldName — FUNCTION
記事のカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrycustomfieldname.html)

### MTEntryCustomFieldValue — FUNCTION
記事のカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrycustomfieldvalue.html)

### MTEntryCustomFields — BLOCK
記事のカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrycustomfields.html)

### MTEntryDate — FUNCTION
記事の公開日時（format で書式指定）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrydate.html)

### MTEntryExcerpt — FUNCTION
記事の概要（空なら本文の先頭）

属性: `convert_breaks="0 | 1"` `no_generate="0 | 1"` `words="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryexcerpt.html)

### MTEntryFlag — FUNCTION
記事の項目の状態を 0/1 で（flag で項目を選ぶ）

属性: `flag="allow_comments | convert_breaks | allow_pings"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryflag.html)

### MTEntryID — FUNCTION
記事の ID（pad で桁を揃える）

属性: `pad="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryid.html)

### MTEntryIfExtended — BLOCK
記事に続きがあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryifextended.html)

### MTEntryIfTagged — BLOCK
記事に（指定の）タグが付いていれば中身を出す

属性: `tag="tag_foo"` `name="tag_foo"` `include_private="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryiftagged.html)

### MTEntryKeywords — FUNCTION
記事のキーワード

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrykeywords.html)

### MTEntryLink — FUNCTION
記事のアーカイブの絶対 URL（既定は個別アーカイブ）

属性: `archive_type="archive_type_foo"` `type="archive_type_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrylink.html)

### MTEntryModifiedAuthorDisplayName — FUNCTION
記事を最後に更新したユーザーの表示名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymodifiedauthordisplayname.html)

### MTEntryModifiedAuthorEmail — FUNCTION
記事を最後に更新したユーザーのメールアドレス

属性: `spam_protect="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymodifiedauthoremail.html)

### MTEntryModifiedAuthorID — FUNCTION
記事・ページを最後に更新したユーザーの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymodifiedauthorid.html)

### MTEntryModifiedAuthorLink — FUNCTION
記事を最後に更新したユーザーの表示名をリンク付きで

属性: `show_url="0 | 1"` `show_email="0 | 1"` `new_window="0 | 1"` `spam_protect="0 | 1"` `type="url | email | archive"` `show_hcard="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymodifiedauthorlink.html)

### MTEntryModifiedAuthorURL — FUNCTION
記事を最後に更新したユーザーのウェブサイト URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymodifiedauthorurl.html)

### MTEntryModifiedAuthorUsername — FUNCTION
記事を最後に更新したユーザーのユーザー名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymodifiedauthorusername.html)

### MTEntryModifiedAuthorUserpic — FUNCTION
記事を最後に更新したユーザーのプロフィール画像の img 要素

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymodifiedauthoruserpic.html)

### MTEntryModifiedAuthorUserpicAsset — FUNCTION
記事を最後に更新したユーザーのプロフィール画像をアセットとして文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymodifiedauthoruserpicas.html)

### MTEntryModifiedAuthorUserpicURL — FUNCTION
記事を最後に更新したユーザーのプロフィール画像の URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymodifiedauthoruserpicurl.html)

### MTEntryModifiedDate — FUNCTION
記事を最後に保存した日時（再構築の日時ではない）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymodifieddate.html)

### MTEntryMore — FUNCTION
記事の続き

属性: `convert_breaks="0 | 1"` `words="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrymore.html)

### MTEntryNext — BLOCK
次の記事を文脈にする

属性: `by_author="0 | 1"` `by_category="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrynext.html)

### MTEntryPermalink — FUNCTION
記事のパーマリンク（アーカイブの中の記事の位置まで）

属性: `archive_type="archive_type_foo"` `type="archive_type_foo"` `valid_html="0 | 1"` `with_index="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrypermalink.html)

### MTEntryPrevious — BLOCK
前の記事を文脈にする

属性: `by_author="0 | 1"` `by_category="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryprevious.html)

### MTEntryPrimaryCategory — BLOCK
記事のメインカテゴリを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryprimarycategory.html)

### MTEntryRank — FUNCTION
記事のスコアから出した順位（要スコア用プラグイン）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryrank.html)

### MTEntryScore — FUNCTION
記事のスコアの合計（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryscore.html)

### MTEntryScoreAvg — FUNCTION
記事のスコアの平均（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryscoreavg.html)

### MTEntryScoreCount — FUNCTION
記事にスコアが付いた回数（要スコア用プラグイン）

属性: `namespace="foo"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryscorecount.html)

### MTEntryScoreHigh — FUNCTION
記事のスコアの最高値（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryscorehigh.html)

### MTEntryScoreLow — FUNCTION
記事のスコアの最低値（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entryscorelow.html)

### MTEntrySiteDescription — FUNCTION
記事が属するサイトの説明（MTEntryBlogDescription と同じ）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrysitedescription.html)

### MTEntrySiteID — FUNCTION
記事が属するサイトの ID（MTEntryBlogID と同じ）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrysiteid.html)

### MTEntrySiteName — FUNCTION
記事が属するサイトの名前（MTEntryBlogName と同じ）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrysitename.html)

### MTEntrySiteURL — FUNCTION
記事が属するサイトの URL（MTEntryBlogURL と同じ）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrysiteurl.html)

### MTEntryStatus — FUNCTION
記事の状態（Publish・Draft・Future など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrystatus.html)

### MTEntryTags — BLOCK
記事のタグを並べるループ（glue で区切り）

属性: `glue="foo"` `include_private="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrytags.html)

### MTEntryTitle — FUNCTION
記事のタイトル

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrytitle.html)

### MTEntryTrackbackCount — FUNCTION
記事が受けたトラックバックの数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrytrackbackcount.html)

### MTEntryTrackbackData — FUNCTION
トラックバック自動検知用の RDF（画面には出ない）

属性: `comment_wrap="0 | 1"` `with_index="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrytrackbackdata.html)

### MTEntryTrackbackID — FUNCTION
記事のトラックバック ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrytrackbackid.html)

### MTEntryTrackbackLink — FUNCTION
記事のトラックバック URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/entrytrackbacklink.html)

### MTEntryUnpublishedDate — FUNCTION
記事の公開終了日時

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/mtentryunpublisheddate.html)
