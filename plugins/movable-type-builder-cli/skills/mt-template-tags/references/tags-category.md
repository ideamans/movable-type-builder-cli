# カテゴリ

記事の分類。階層を持つ。

現行のタグ 28 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTCategories — BLOCK
カテゴリを並べるループ（階層を区別しない。階層なら MTTopLevelCategories）

属性: `glue="foo"` `lastn="N"` `show_empty="1"` `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categories.html)

### MTCategoryArchiveLink — FUNCTION
カテゴリアーカイブの URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categoryarchivelink.html)

### MTCategoryBasename — FUNCTION
カテゴリのベースネーム（出力フォルダ名）

属性: `default="foo"` `separator="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorybasename.html)

### MTCategoryCommentCount — FUNCTION
カテゴリの記事に付いたコメントの数

属性: `lastn="N"` `sort_order="ascend | descend"` `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorycommentcount.html)

### MTCategoryCount — FUNCTION
カテゴリの公開済み記事の数

属性: `singular="text"` `plural="text"` `none="text"` `content_type="content_type_name"` `content_field="content_field_name"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorycount.html)

### MTCategoryCustomFieldDescription — FUNCTION
カテゴリのカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorycustomfielddescription.html)

### MTCategoryCustomFieldName — FUNCTION
カテゴリのカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorycustomfieldname.html)

### MTCategoryCustomFieldValue — FUNCTION
カテゴリのカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorycustomfieldvalue.html)

### MTCategoryCustomFields — BLOCK
カテゴリのカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorycustomfields.html)

### MTCategoryDescription — FUNCTION
カテゴリの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorydescription.html)

### MTCategoryID — FUNCTION
カテゴリの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categoryid.html)

### MTCategoryIfAllowPings — BLOCK
カテゴリがトラックバックを受け付けていれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categoryifallowpings.html)

### MTCategoryLabel — FUNCTION
カテゴリ名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorylabel.html)

### MTCategoryNext — BLOCK
次のカテゴリを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorynext.html)

### MTCategoryPrevious — BLOCK
前のカテゴリを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categoryprevious.html)

### MTCategorySetName — FUNCTION
カテゴリセットの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorysetname.html)

### MTCategorySets — BLOCK
カテゴリセットを並べるループ

属性: `id="N"` `blog_id="N"` `name="foo"` `content_type="コンテンツタイプのユニークID | コンテンツタイプのID | コンテンツタイプの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorysets.html)

### MTCategoryTrackbackCount — FUNCTION
カテゴリが受けたトラックバックの数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorytrackbackcount.html)

### MTCategoryTrackbackLink — FUNCTION
カテゴリのトラックバック URL

属性: `category="category_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/categorytrackbacklink.html)

### MTHasNoParentCategory — BLOCK
親カテゴリが無ければ中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/hasnoparentcategory.html)

### MTHasNoSubCategories — BLOCK
サブカテゴリが無ければ中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/hasnosubcategories.html)

### MTHasParentCategory — BLOCK
親カテゴリがあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/hasparentcategory.html)

### MTHasSubCategories — BLOCK
サブカテゴリがあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/hassubcategories.html)

### MTParentCategories — BLOCK
最上位からいまのカテゴリまでを回すループ（パンくず向け。glue で区切り）

属性: `glue="foo"` `exclude_current="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/parentcategories.html)

### MTParentCategory — BLOCK
親カテゴリを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/parentcategory.html)

### MTSubCategories — BLOCK
いまのカテゴリの子カテゴリを階層で並べるループ（MTSubCatsRecurse と組で）

属性: `include_current="0 | 1"` `top="0 | 1"` `category="category_foo"` `sort_by="label | description | basename | created_on | modified_on | user_custom"` `sort_order="ascend | descend"` `sort_method="sub_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/subcategories.html)

### MTSubCategoryPath — FUNCTION
最上位からいまのカテゴリまでのパス（ベースネームを / でつなぐ）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/subcategorypath.html)

### MTTopLevelCategories — BLOCK
最上位のカテゴリを並べるループ（いまのカテゴリに依らない）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/toplevelcategories.html)
