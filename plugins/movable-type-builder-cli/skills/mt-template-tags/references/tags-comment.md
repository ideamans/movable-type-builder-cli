# コメント・トラックバック

フィードバック機能。

現行のタグ 67 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTCaptchaFields — FUNCTION
CAPTCHA の画像と入力欄

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/captchafields.html)

### MTCommentAuthor — FUNCTION
コメント投稿者の名前（default で無記名時の名前）

属性: `default="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentauthor.html)

### MTCommentAuthorIdentity — FUNCTION
認証してコメントした人のプロフィール画像（登録先へのリンク付き）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentauthoridentity.html)

### MTCommentAuthorLink — FUNCTION
コメント投稿者の名前をリンク付きで（default_name で無記名時の名前）

属性: `default_name="Anonymous"` `show_url="0 | 1"` `show_email="0 | 1"` `new_window="0 | 1"` `spam_protect="0 | 1"` `no_redirect="0 | 1"` `nofollowfy="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentauthorlink.html)

### MTCommentBlogID — FUNCTION
コメントが属するサイトの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentblogid.html)

### MTCommentBody — FUNCTION
コメントの本文

属性: `autolink="0 | 1"` `convert_breaks="0 | 1"` `words="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentbody.html)

### MTCommentCustomFieldDescription — FUNCTION
コメントのカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentcustomfielddescription.html)

### MTCommentCustomFieldName — FUNCTION
コメントのカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentcustomfieldname.html)

### MTCommentCustomFieldValue — FUNCTION
コメントのカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentcustomfieldvalue.html)

### MTCommentCustomFields — BLOCK
コメントのカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentcustomfields.html)

### MTCommentDate — FUNCTION
コメントの投稿日時（format で書式指定）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentdate.html)

### MTCommentEmail — FUNCTION
コメント投稿者のメールアドレス（spam_protect で難読化）

属性: `spam_protect="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentemail.html)

### MTCommentEntry — BLOCK
コメント先の記事を文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commententry.html)

### MTCommentEntryID — FUNCTION
コメント先の記事の ID

属性: `pad="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commententryid.html)

### MTCommentID — FUNCTION
コメントの ID（アンカーに使う）

属性: `pad="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentid.html)

### MTCommentIP — FUNCTION
コメントを送ったホストの IP アドレス

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentip.html)

### MTCommentIfModerated — BLOCK
コメントが保留中なら中身を出す（管理画面・メール向け）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentifmoderated.html)

### MTCommentLink — FUNCTION
コメントのパーマリンク

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentlink.html)

### MTCommentOrderNumber — FUNCTION
一覧の中でのコメントの通し番号（1 から）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentordernumber.html)

### MTCommentParent — BLOCK
返信元のコメントを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentparent.html)

### MTCommentParentID — FUNCTION
返信元のコメントの ID（無ければ 0）

属性: `pad="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentparentid.html)

### MTCommentPreviewIsStatic — FUNCTION
システム用。コメント後に戻るページの情報を渡す（プレビュー・エラーのテンプレート）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentpreviewisstatic.html)

### MTCommentRank — FUNCTION
コメントのスコアから出した順位（要スコア用プラグイン）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentrank.html)

### MTCommentReplies — BLOCK
コメントへの返信を並べるループ

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentreplies.html)

### MTCommentRepliesRecurse — FUNCTION
MTCommentReplies の中身を返信の深さまで繰り返す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentrepliesrecurse.html)

### MTCommentReplyToLink — FUNCTION
コメントに返信するリンク（標準テーマの JavaScript と組で使う）

属性: `text="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentreplylink.html)

### MTCommentScore — FUNCTION
コメントのスコアの合計（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentscore.html)

### MTCommentScoreAvg — FUNCTION
コメントのスコアの平均（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentscoreavg.html)

### MTCommentScoreCount — FUNCTION
コメントにスコアが付いた回数（要スコア用プラグイン）

属性: `namespace="foo"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentscorecount.html)

### MTCommentScoreHigh — FUNCTION
コメントのスコアの最高値（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentscorehigh.html)

### MTCommentScoreLow — FUNCTION
コメントのスコアの最低値（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentscorelow.html)

### MTCommentScript — FUNCTION
コメント処理の CGI のファイル名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentscript.html)

### MTCommentSiteID — FUNCTION
コメントが属するサイトの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentsiteid.html)

### MTCommentURL — FUNCTION
コメント投稿者が入れた URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commenturl.html)

### MTCommenterAuthIconURL — FUNCTION
コメントした人のサインイン元サービスのアイコン URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commenterauthiconurl.html)

### MTCommenterAuthType — FUNCTION
コメントした人のアカウントの登録元（MT など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commenterauthtype.html)

### MTCommenterEmail — FUNCTION
コメントした人のメールアドレス（spam_protect で難読化）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commenteremail.html)

### MTCommenterID — FUNCTION
コメントした人のユーザー ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commenterid.html)

### MTCommenterName — FUNCTION
コメントした人の名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentername.html)

### MTCommenterURL — FUNCTION
コメントした人のウェブサイト URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commenterurl.html)

### MTCommenterUsername — FUNCTION
コメントした人のユーザー名（無ければ空）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commenterusername.html)

### MTCommenterUserpic — FUNCTION
コメントした登録ユーザーのプロフィール画像の img 要素

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commenteruserpic.html)

### MTCommenterUserpicAsset — BLOCK
コメントした登録ユーザーのプロフィール画像をアセットとして文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commenteruserpicasset.html)

### MTCommenterUserpicURL — FUNCTION
コメントした登録ユーザーのプロフィール画像の URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commenteruserpicurl.html)

### MTComments — BLOCK
コメントを並べるループ（記事・ページの中ならそのコメント）

属性: `lastn="N"` `sort_order="ascend | descend"` `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `exclude_blogs="ID"` `top="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/comments.html)

### MTCommentsFooter — BLOCK
MTComments の最後の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentsfooter.html)

### MTCommentsHeader — BLOCK
MTComments の最初の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/commentsheader.html)

### MTPingBlogName — FUNCTION
トラックバック送信元のサイト名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingblogname.html)

### MTPingDate — FUNCTION
トラックバックを受けた日時（format で書式指定）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingdate.html)

### MTPingEntry — BLOCK
トラックバックを受けた記事を文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingentry.html)

### MTPingExcerpt — FUNCTION
トラックバックの概要

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingexcerpt.html)

### MTPingID — FUNCTION
トラックバックの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingid.html)

### MTPingIP — FUNCTION
トラックバック送信元の IP アドレス

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingip.html)

### MTPingRank — FUNCTION
トラックバックのスコアから出した順位（要スコア用プラグイン）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingrank.html)

### MTPingScore — FUNCTION
トラックバックのスコアの合計（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingscore.html)

### MTPingScoreAvg — FUNCTION
トラックバックのスコアの平均（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingscoreavg.html)

### MTPingScoreCount — FUNCTION
トラックバックにスコアが付いた回数（要スコア用プラグイン）

属性: `namespace="foo"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingscorecount.html)

### MTPingScoreHigh — FUNCTION
トラックバックのスコアの最高値（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingscorehigh.html)

### MTPingScoreLow — FUNCTION
トラックバックのスコアの最低値（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingscorelow.html)

### MTPingTitle — FUNCTION
トラックバック送信元のページのタイトル

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingtitle.html)

### MTPingURL — FUNCTION
トラックバック送信元のページの URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingurl.html)

### MTPings — BLOCK
トラックバックを並べるループ（記事の中ならその記事のもの）

属性: `category="category_foo"` `lastn="N"` `sort_order="ascend | descend"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pings.html)

### MTPingsFooter — BLOCK
MTPings の最後の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingsfooter.html)

### MTPingsHeader — BLOCK
MTPings の最初の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingsheader.html)

### MTPingsSent — BLOCK
記事が送ったトラックバックを並べるループ

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingssent.html)

### MTPingsSentURL — FUNCTION
送ったトラックバックの送り先 URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/pingssenturl.html)

### MTTrackbackScript — FUNCTION
トラックバック処理の CGI のファイル名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/trackbackscript.html)
