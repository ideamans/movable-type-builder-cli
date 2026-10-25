# システム・パス

設置パスやバージョンなど。

現行のタグ 23 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTAdminCGIPath — FUNCTION
管理画面の CGI を置いた URL（mt-config.cgi の AdminCGIPath）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/admincgipath.html)

### MTAdminScript — FUNCTION
管理画面のスクリプト名（AdminScript の値。既定は mt.cgi）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/adminscript.html)

### MTBuildTemplateID — FUNCTION
いま出力しているファイルのテンプレートの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/buildtemplateid.html)

### MTCGIHost — FUNCTION
MT を置いたホスト名

属性: `exclude_port="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/cgihost.html)

### MTCGIPath — FUNCTION
MT の CGI の URL（CGIPath の値。末尾 /）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/cgipath.html)

### MTCGIRelativeURL — FUNCTION
MT の CGI の URL のパス部分

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/cgirelativeurl.html)

### MTCGIServerPath — FUNCTION
MT を置いたサーバ上のパス

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/cgiserverpath.html)

### MTCanonicalLink — FUNCTION
いまのページの正規 URL を link 要素で出す

属性: `with_index="1"` `current_mapping="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/canonicallink.html)

### MTCanonicalURL — FUNCTION
いまのページの正規 URL

属性: `with_index="1"` `current_mapping="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/canonicalurl.html)

### MTConfigFile — FUNCTION
mt-config.cgi のサーバ上のパス

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/configfile.html)

### MTDataAPIScript — FUNCTION
Data API の CGI の名前（DataAPIScript の値）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/dataapiscript.html)

### MTDataAPIVersion — FUNCTION
Data API の版

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/dataapiversion.html)

### MTDefaultLanguage — FUNCTION
MT の既定の言語

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/defaultlanguage.html)

### MTErrorMessage — FUNCTION
コメント投稿のエラーメッセージ

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/errormessage.html)

### MTHTTPContentType — FUNCTION
ダイナミック公開・検索結果の Content-Type ヘッダを決める

属性: `type="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/httpcontenttype.html)

### MTLink — FUNCTION
インデックステンプレートか記事の URL（template・entry_id で選ぶ）

属性: `template="template_foo"` `entry_id="ID"` `blog_id="ID"` `with_index="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/link.html)

### MTProductName — FUNCTION
製品名（Movable Type）

属性: `version="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/productname.html)

### MTPublishCharset — FUNCTION
出力の文字コード（PublishCharset の値）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/publishcharset.html)

### MTStaticFilePath — FUNCTION
mt-static のサーバ上の絶対パス

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/staticfilepath.html)

### MTStaticWebPath — FUNCTION
mt-static の URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/staticwebpath.html)

### MTStatsSnippet — FUNCTION
管理画面で設定したアクセス解析のトラッキングコード

属性: `gtag="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/statssnippet.html)

### MTSupportDirectoryURL — FUNCTION
サポートディレクトリの URL（SupportDirectoryURL の値）

属性: `with_domain="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/supportdirectoryurl.html)

### MTVersion — FUNCTION
Movable Type の版

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/version.html)
