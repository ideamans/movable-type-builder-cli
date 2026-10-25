# ウェブページ (Page)

日付を持たない固定ページ。フォルダで階層化する。

現行のタグ 35 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTBlogPageCount — FUNCTION
サイトの公開済みウェブページの数

属性: `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogpagecount.html)

### MTPageAssets — BLOCK
ウェブページが使っているアセットを並べるループ

属性: `lastn="N"` `offset="N"` `limit="N"` `author="author_foo"` `type="image | audio | video | file"` `sort_order="ascend | descend"` `sort_by="file_name | created_by | created_on"` `assets_per_row="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pageassets.html)

### MTPageAuthorDisplayName — FUNCTION
ウェブページの作成者の表示名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pageauthordisplayname.html)

### MTPageAuthorEmail — FUNCTION
ウェブページの作成者のメールアドレス

属性: `spam_protect="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pageauthoremail.html)

### MTPageAuthorLink — FUNCTION
ウェブページの作成者の表示名をリンク付きで

属性: `show_url="0 | 1"` `show_email="0 | 1"` `new_window="0 | 1"` `spam_protect="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pageauthorlink.html)

### MTPageAuthorURL — FUNCTION
ウェブページの作成者のウェブサイト URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pageauthorurl.html)

### MTPageBasename — FUNCTION
ウェブページのベースネーム（出力ファイル名）

属性: `separator="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagebasename.html)

### MTPageBody — FUNCTION
ウェブページの本文

属性: `convert_breaks="0 | 1"` `words="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagebody.html)

### MTPageCustomFieldDescription — FUNCTION
ウェブページのカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagecustomfielddescription.html)

### MTPageCustomFieldName — FUNCTION
ウェブページのカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagecustomfieldname.html)

### MTPageCustomFieldValue — FUNCTION
ウェブページのカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagecustomfieldvalue.html)

### MTPageCustomFields — BLOCK
ウェブページのカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagecustomfields.html)

### MTPageDate — FUNCTION
ウェブページの公開日時（format で書式指定）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagedate.html)

### MTPageExcerpt — FUNCTION
ウェブページの概要（空なら本文の先頭）

属性: `convert_breaks="0 | 1"` `no_generate="0 | 1"` `words="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pageexcerpt.html)

### MTPageFolder — BLOCK
ウェブページのフォルダを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagefolder.html)

### MTPageID — FUNCTION
ウェブページの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pageid.html)

### MTPageIfTagged — BLOCK
ウェブページに（指定の）タグが付いていれば中身を出す

属性: `tag="tag_foo"` `name="tag_foo"` `include_private="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pageiftagged.html)

### MTPageKeywords — FUNCTION
ウェブページのキーワード

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagekeywords.html)

### MTPageModifiedAuthorDisplayName — FUNCTION
ウェブページを最後に更新したユーザーの表示名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagemodifiedauthordisplayname.html)

### MTPageModifiedAuthorEmail — FUNCTION
ウェブページを最後に更新したユーザーのメールアドレス

属性: `spam_protect="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagemodifiedauthoremail.html)

### MTPageModifiedAuthorLink — FUNCTION
ウェブページを最後に更新したユーザーの表示名をリンク付きで

属性: `show_url="0 | 1"` `show_email="0 | 1"` `new_window="0 | 1"` `spam_protect="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagemodifiedauthorlink.html)

### MTPageModifiedAuthorURL — FUNCTION
ウェブページを最後に更新したユーザーのウェブサイト URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagemodifiedauthorurl.html)

### MTPageModifiedDate — FUNCTION
ウェブページを最後に更新した日時

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagemodifieddate.html)

### MTPageMore — FUNCTION
ウェブページの続き

属性: `convert_breaks="0 | 1"` `words="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagemore.html)

### MTPageNext — BLOCK
次のウェブページを文脈にする

属性: `by_author="0 | 1"` `by_folder="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagenext.html)

### MTPagePermalink — FUNCTION
ウェブページの絶対 URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagepermalink.html)

### MTPagePrevious — BLOCK
前のウェブページを文脈にする

属性: `by_author="0 | 1"` `by_folder="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pageprevious.html)

### MTPageTags — BLOCK
ウェブページのタグを並べるループ

属性: `glue="foo"` `include_private="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagetags.html)

### MTPageTitle — FUNCTION
ウェブページのタイトル

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagetitle.html)

### MTPageUnpublishedDate — FUNCTION
ウェブページの公開終了日時

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/mtpageunpublisheddate.html)

### MTPagerBlock — BLOCK
検索結果のページ数だけ回すループ（ページ番号の並び）

属性: `__first__` `__last__` `__odd__` `__even__` `__counter__` `__value__`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagerblock.html)

### MTPagerLink — FUNCTION
MTPagerBlock の中で、各ページの URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagerlink.html)

### MTPages — BLOCK
ウェブページを並べるループ

属性: `lastn="N"` `limit="N"` `sort_by="foo"` `sort_order="ascend | descend"` `field:customfieldbasename="foo"` `namespace="foo"` `offset="N"` `folder="folder_foo"` `folders="folder_foo"` `no_folder="0 | 1"` `include_subfolders="1 | 0"` `tag="tag_foo"` `tags="tag_foo AND tag_bar"` `author="author_user_name"` `id="page_id"` `min_score="score"` `max_score="score"` `min_rate="rate"` `max_rate="rate"` `min_count="count"` `max_count="count"` `scored_by="author_user_name"` `days="N"` `recently_commented_on="N"` `unique="1 | 0"` `glue="foo"` `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pages.html)

### MTPagesFooter — BLOCK
MTPages の最後の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagesfooter.html)

### MTPagesHeader — BLOCK
MTPages の最初の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pagesheader.html)
