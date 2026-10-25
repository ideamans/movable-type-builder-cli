# テキストフォーマット

本文の整形方式。

現行のタグ 6 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTMarkdownOptions — BLOCK
Markdown の変換結果の形式を変える

属性: `output="html4 | raw"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/markdownoptions.html)

### MTSmartyPantsVersion — FUNCTION
Markdown プラグイン（SmartyPants）の版

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/smartypantsversion.html)

### MTTextFormat — FUNCTION
複数行テキストのテキストフォーマット

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/text-format.html)

### MTTextile — BLOCK
囲んだ中身を Textile で処理する

属性: `head_offset="N"` `preserve_spaces="0 | 1"` `trim_spaces="0 | 1"` `flavor="html | html/css | xhtml1 | xhtml2"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/textile.html)

### MTTextileHeadOffset — FUNCTION
Textile の見出しの段を start の分だけ下げる

属性: `start="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/textileheadoffset.html)

### MTTextileOptions — FUNCTION
Textile の変換のオプション（テンプレートの最初に書く）

属性: `head_offset="N"` `preserve_spaces="0 | 1"` `trim_spaces="0 | 1"` `flavor="html | html/css | xhtml1 | xhtml2"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/textileoptions.html)
