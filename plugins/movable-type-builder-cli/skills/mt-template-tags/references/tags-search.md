# 検索

サイト内検索の結果出力。

現行のタグ 12 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTNoSearch — BLOCK
キーワード無しで検索したときに中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/nosearch.html)

### MTNoSearchResults — BLOCK
検索結果が0件のときに中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/nosearchresults.html)

### MTSearchIncludeBlogs — FUNCTION
検索対象のサイトの ID（カンマ区切り）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/searchincludeblogs.html)

### MTSearchMaxResults — FUNCTION
mt-config.cgi の SearchMaxResults の値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/searchmaxresults.html)

### MTSearchResultCount — FUNCTION
検索結果の件数（複数サイトなら合計）

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/searchresultcount.html)

### MTSearchResults — BLOCK
検索結果を並べるループ（中では記事のタグが使える）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/searchresults.html)

### MTSearchResultsFooter — BLOCK
検索結果の最後の1件で出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/searchresultsfooter.html)

### MTSearchResultsHeader — BLOCK
検索結果の最初の1件で出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/searchresultsheader.html)

### MTSearchScript — FUNCTION
検索の CGI のファイル名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/searchscript.html)

### MTSearchString — FUNCTION
検索したキーワード

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/searchstring.html)

### MTSearchTemplateBlogID — FUNCTION
検索結果に使ったテンプレートのサイトの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/searchtemplateblogid.html)

### MTSearchTemplateID — FUNCTION
検索結果に使ったテンプレートの識別子

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/searchtemplateid.html)
