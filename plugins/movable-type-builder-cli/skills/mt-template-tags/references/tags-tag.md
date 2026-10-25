# タグ（ラベル）

記事・ページ・アセットに付ける自由なラベル。

現行のタグ 7 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTTagCount — FUNCTION
いまのタグが付いた記事・ページの数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/tagcount.html)

### MTTagID — FUNCTION
タグの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/tagid.html)

### MTTagLabel — FUNCTION
タグの名前（MTTagName と同じ）

属性: `quote="0 | 1"` `normalize="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/taglabel.html)

### MTTagName — FUNCTION
タグの名前

属性: `quote="0 | 1"` `normalize="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/tagname.html)

### MTTagRank — FUNCTION
タグの使用頻度の順位（小さいほど多い。タグクラウド向け）

属性: `max="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/tagrank.html)

### MTTagSearchLink — FUNCTION
そのタグで検索する URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/tagsearchlink.html)

### MTTags — BLOCK
タグを並べるループ

属性: `type="type"` `content_type="content_type"` `glue="foo"` `sort_by="name | rank"` `sort_order="ascend | descend"` `limit="N"` `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"` `top="N"` `include_private="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/tags.html)
