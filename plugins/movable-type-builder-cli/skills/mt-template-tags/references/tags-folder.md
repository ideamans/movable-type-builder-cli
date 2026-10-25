# フォルダ

ウェブページの分類。階層を持つ。

現行のタグ 26 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTFolderBasename — FUNCTION
フォルダのベースネーム（パス）

属性: `default="foo"` `separator="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/folderbasename.html)

### MTFolderCount — FUNCTION
フォルダのウェブページの数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/foldercount.html)

### MTFolderCustomFieldDescription — FUNCTION
フォルダのカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/foldercustomfielddescription.html)

### MTFolderCustomFieldName — FUNCTION
フォルダのカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/foldercustomfieldname.html)

### MTFolderCustomFieldValue — FUNCTION
フォルダのカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/foldercustomfieldvalue.html)

### MTFolderCustomFields — BLOCK
フォルダのカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/foldercustomfields.html)

### MTFolderDescription — FUNCTION
フォルダの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/folderdescription.html)

### MTFolderFooter — BLOCK
MTFolders の最後の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/folderfooter.html)

### MTFolderHeader — BLOCK
MTFolders の最初の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/folderheader.html)

### MTFolderID — FUNCTION
フォルダの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/folderid.html)

### MTFolderLabel — FUNCTION
フォルダ名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/folderlabel.html)

### MTFolderNext — BLOCK
次のフォルダを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/foldernext.html)

### MTFolderPath — FUNCTION
最上位からいまのフォルダまでのパス（foo/bar）

属性: `separator="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/folderpath.html)

### MTFolderPrevious — BLOCK
前のフォルダを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/folderprevious.html)

### MTFolders — BLOCK
フォルダを並べるループ（階層を区別しない。階層なら MTTopLevelFolders）

属性: `show_empty="0 | 1"` `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/folders.html)

### MTHasNoParentFolder — BLOCK
親フォルダが無ければ中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/hasnoparentfolder.html)

### MTHasNoSubFolders — BLOCK
サブフォルダが無ければ中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/hasnosubfolders.html)

### MTHasParentFolder — BLOCK
親フォルダがあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/hasparentfolder.html)

### MTHasSubFolders — BLOCK
サブフォルダがあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/hassubfolders.html)

### MTParentFolder — BLOCK
親フォルダを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/parentfolder.html)

### MTParentFolders — BLOCK
最上位からいまのフォルダまでを回すループ（glue で区切り）

属性: `glue="foo"` `exclude_current="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/parentfolders.html)

### MTSubFolderRecurse — FUNCTION
MTSubFolders を子フォルダへ繰り返す（max_depth で深さ）

属性: `max_depth="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/subfolderrecurse.html)

### MTSubFolders — BLOCK
いまのフォルダの子フォルダを階層で並べるループ

属性: `include_current="1"` `top="0 | 1"` `sort_by="user_custom"` `sort_order="ascend | descend"` `sort_method="sub_foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/subfolders.html)

### MTTopLevelFolder — BLOCK
いまのフォルダの最上位のフォルダを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/toplevelfolder.html)

### MTTopLevelFolders — BLOCK
最上位のフォルダを並べるループ（いまのフォルダに依らない）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/toplevelfolders.html)

### MTTopLevelParent — BLOCK
いまのカテゴリの最上位のカテゴリを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/toplevelparent.html)
