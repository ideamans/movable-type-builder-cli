# サイト・ブログ

サイトの属性と、サイトをまたぐ出力。

現行のタグ 80 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTBlogArchiveURL — FUNCTION
サイトのアーカイブ URL（未設定ならサイト URL）

属性: `id="blog_id"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogarchiveurl.html)

### MTBlogCategoryCount — FUNCTION
サイトのカテゴリの数

属性: `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogcategorycount.html)

### MTBlogCommentCount — FUNCTION
サイトのコメントの数

属性: `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogcommentcount.html)

### MTBlogCustomFieldDescription — FUNCTION
ブログのカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogcustomfielddescription.html)

### MTBlogCustomFieldName — FUNCTION
ブログのカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogcustomfieldname.html)

### MTBlogCustomFieldValue — FUNCTION
ブログのカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogcustomfieldvalue.html)

### MTBlogCustomFields — BLOCK
ブログのカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogcustomfields.html)

### MTBlogDateLanguage — FUNCTION
サイトの「日付の言語」の設定

属性: `locale="0 | 1"` `ietf="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogdatelanguage.html)

### MTBlogDescription — FUNCTION
サイトの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogdescription.html)

### MTBlogFileExtension — FUNCTION
アーカイブの拡張子（ドット付き）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogfileextension.html)

### MTBlogHost — FUNCTION
サイト URL のホスト名

属性: `exclude_port="0 | 1"` `signature="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/bloghost.html)

### MTBlogID — FUNCTION
サイトの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogid.html)

### MTBlogIfCommentsOpen — BLOCK
サイトがコメント（匿名を含む）を受け付けていれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogifcommentsopen.html)

### MTBlogLanguage — FUNCTION
サイトの使用言語

属性: `locale="1"` `ietf="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/bloglanguage.html)

### MTBlogName — FUNCTION
サイトの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogname.html)

### MTBlogParentWebsite — BLOCK
ブログの親ウェブサイトを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogparentwebsite.html)

### MTBlogPingCount — FUNCTION
サイトの公開済みトラックバックの数

属性: `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogpingcount.html)

### MTBlogRelativeURL — FUNCTION
サイト URL のホストを除いた相対 URL

属性: `id="blog_id"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogrelativeurl.html)

### MTBlogResultFooter — BLOCK
検索結果でサイトが切り替わる直前に出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogresultfooter.html)

### MTBlogResultHeader — BLOCK
検索結果でサイトが切り替わった直後に出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogresultheader.html)

### MTBlogSitePath — FUNCTION
サイトの公開パス（サーバ上のディレクトリ）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogsitepath.html)

### MTBlogTemplateSetID — FUNCTION
テーマの識別子。MT5 以降は MTBlogThemeID を使う

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogtemplatesetid.html)

### MTBlogThemeID — FUNCTION
サイトに適用したテーマの識別子

属性: `raw="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogthemeid.html)

### MTBlogTimezone — FUNCTION
サイトのタイムゾーン（UTC との時差）

属性: `no_colon="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogtimezone.html)

### MTBlogURL — FUNCTION
サイトの絶対 URL

属性: `id="blog_id"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogurl.html)

### MTBlogs — BLOCK
サイト（ブログ）を並べるループ

属性: `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/blogs.html)

### MTChildSites — BLOCK
子サイトを並べるループ

属性: `mode` `include_sites` `exclude_sites`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/childsites.html)

### MTMultiBlog — BLOCK
複数のサイトをまたいで中身を出す（include_blogs などで選ぶ）

属性: `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"` `mode="loop | context"` `context` `ignore_archive_context="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/multiblog.html)

### MTMultiBlogIfLocalBlog — BLOCK
MTMultiBlog の中で、いまのサイトが元のサイトなら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/multiblogiflocalblog.html)

### MTMultiBlogLocalBlog — BLOCK
MTMultiBlog の中で、一時的に元のサイトの文脈に戻す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/multibloglocalblog.html)

### MTOtherBlog — BLOCK
blog_id で指定したサイトの文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/otherblog.html)

### MTParentSite — BLOCK
子サイトから親サイトを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/parentsite.html)

### MTSiteArchiveURL — FUNCTION
サイトのアーカイブ URL（未設定ならサイト URL）

属性: `id="site_id"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitearchiveurl.html)

### MTSiteCategoryCount — FUNCTION
サイトのカテゴリの数（カテゴリセットのものは数えない）

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitecategorycount.html)

### MTSiteCommentCount — FUNCTION
サイトのコメントの数（MTBlogCommentCount と同じ）

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitecommentcount.html)

### MTSiteContentCount — FUNCTION
サイトの公開済みコンテンツデータの数（content_type で絞る）

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitecontentcount.html)

### MTSiteDateLanguage — FUNCTION
サイトの「日付の言語」（MTBlogDateLanguage と同じ）

属性: `locale="1"` `ietf="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitedatelanguage.html)

### MTSiteDescription — FUNCTION
サイトの説明（MTBlogDescription と同じ）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitedescription.html)

### MTSiteEntryCount — FUNCTION
サイトの公開済み記事の数（MTBlogEntryCount と同じ）

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/siteentrycount.html)

### MTSiteFileExtension — FUNCTION
アーカイブの拡張子（ドット付き）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitefileextension.html)

### MTSiteHasChildSite — BLOCK
サイトに子サイトがあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitehaschildsite.html)

### MTSiteHost — FUNCTION
サイト URL のホスト名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitehost.html)

### MTSiteID — FUNCTION
サイトの ID（MTBlogID と同じ）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/siteid.html)

### MTSiteIfCommentsOpen — BLOCK
サイトがコメント（匿名を含む）を受け付けていれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/siteifcommentsopen.html)

### MTSiteLanguage — FUNCTION
サイトの使用言語

属性: `locale="1"` `ietf="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitelanguage.html)

### MTSiteName — FUNCTION
サイトの名前（MTBlogName と同じ）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitename.html)

### MTSitePageCount — FUNCTION
サイトの公開済みウェブページの数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitepagecount.html)

### MTSiteParentSite — FUNCTION
子サイトから親サイトを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/siteparentsite.html)

### MTSitePath — FUNCTION
サイトの公開パス（サーバ上のディレクトリ）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitepath.html)

### MTSitePingCount — FUNCTION
サイトの公開済みトラックバックの数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitepingcount.html)

### MTSiteRelativeURL — FUNCTION
サイト URL のホストを除いた相対 URL

属性: `id="site_id"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/siterelativeurl.html)

### MTSiteThemeID — FUNCTION
サイトに適用したテーマの識別子

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitethemeid.html)

### MTSiteTimezone — FUNCTION
サイトのタイムゾーン（UTC との時差）

属性: `no_colon="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitetimezone.html)

### MTSiteURL — FUNCTION
サイトの絶対 URL（末尾は必ず /）

属性: `id="site_id"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/siteurl.html)

### MTSites — BLOCK
サイトを並べるループ（子サイトは含まない）

属性: `include_sites` `exclude_sites`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sites.html)

### MTSitesIfLocalSite — BLOCK
MTSites の中で、いまのサイトが再構築中のサイトなら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/sitesiflocalsite.html)

### MTSitesLocalSite — BLOCK
MTSites の中で、一時的に再構築中のサイトの文脈に戻す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/siteslocalsite.html)

### MTWebsiteCommentCount — FUNCTION
ウェブサイトのコメントの数（配下のブログは含まない）

属性: `site_ids="all | WebsiteID"` `include_websites="all | WebsiteID"` `exclude_websites="WebsiteID"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitecommentcount.html)

### MTWebsiteCustomFieldDescription — FUNCTION
ウェブサイトのカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitecustomfielddescription.html)

### MTWebsiteCustomFieldName — FUNCTION
ウェブサイトのカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitecustomfieldname.html)

### MTWebsiteCustomFieldValue — FUNCTION
ウェブサイトのカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitecustomfieldvalue.html)

### MTWebsiteCustomFields — BLOCK
ウェブサイトのカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitecustomfields.html)

### MTWebsiteDateLanguage — FUNCTION
ウェブサイトの「日付の言語」

属性: `locale="0 | 1"` `ietf="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitedatelanguage.html)

### MTWebsiteDescription — FUNCTION
ウェブサイトの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitedescription.html)

### MTWebsiteEntryCount — FUNCTION
ウェブサイトの公開済み記事の数（配下のブログは既定で含まない）

属性: `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"` `include_websites="all | ID"` `exclude_websites="ID"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websiteentrycount.html)

### MTWebsiteFileExtension — FUNCTION
ウェブサイトのアーカイブの拡張子（ドット付き）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitefileextension.html)

### MTWebsiteHasBlog — BLOCK
ウェブサイトの配下にブログがあれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitehasblog.html)

### MTWebsiteHost — FUNCTION
ウェブサイト URL のホスト名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitehost.html)

### MTWebsiteID — FUNCTION
ウェブサイトの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websiteid.html)

### MTWebsiteIfCommentsOpen — BLOCK
ウェブサイトがコメント（匿名を含む）を受け付けていれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websiteifcommentsopen.html)

### MTWebsiteLanguage — FUNCTION
ウェブサイトの使用言語

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitelanguage.html)

### MTWebsiteName — FUNCTION
ウェブサイトの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitename.html)

### MTWebsitePageCount — FUNCTION
ウェブサイトの公開済みウェブページの数（配下のブログは含まない）

属性: `site_ids="all | WebsiteID"` `include_websites="all | WebsiteID"` `exclude_websites="WebsiteID"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitepagecount.html)

### MTWebsitePath — FUNCTION
ウェブサイトの公開パス（サーバ上のディレクトリ）

属性: `id="website_id"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitepath.html)

### MTWebsitePingCount — FUNCTION
ウェブサイトの公開済みトラックバックの数

属性: `site_ids="all | WebsiteID"` `include_websites="all | WebsiteID"` `exclude_websites="WebsiteID"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitepingcount.html)

### MTWebsiteRelativeURL — FUNCTION
ウェブサイト URL のホストを除いた相対 URL

属性: `id="website_id"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websiterelativeurl.html)

### MTWebsiteThemeID — FUNCTION
ウェブサイトに適用したテーマの識別子

属性: `raw="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitethemeid.html)

### MTWebsiteTimezone — FUNCTION
ウェブサイトのタイムゾーン（UTC との時差）

属性: `no_colon="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websitetimezone.html)

### MTWebsiteURL — FUNCTION
ウェブサイトの絶対 URL

属性: `id="website_id"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websiteurl.html)

### MTWebsites — BLOCK
ウェブサイトを並べるループ

属性: `site_ids="all | WebsiteID"` `include_websites="all | WebsiteID"` `exclude_websites="WebsiteID"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/websites.html)
