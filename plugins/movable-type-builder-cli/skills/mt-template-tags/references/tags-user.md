# ユーザー

投稿者とサインイン。

現行のタグ 34 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTAuthorAuthIconURL — FUNCTION
ユーザーの登録元サービスのアイコン URL

属性: `size="logo_small"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorauthiconurl.html)

### MTAuthorAuthType — FUNCTION
ユーザーの登録元（通常は MT）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorauthtype.html)

### MTAuthorBasename — FUNCTION
ユーザーのベースネーム（ユーザーアーカイブのフォルダ名など）

属性: `separator="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorbasename.html)

### MTAuthorCommentCount — FUNCTION
ユーザーが投稿したコメントの数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorcommentcount.html)

### MTAuthorContentCount — FUNCTION
ユーザーが公開しているコンテンツデータの数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorcontentcount.html)

### MTAuthorCustomFields — BLOCK
ユーザーのカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorcustomfields.html)

### MTAuthorDisplayName — FUNCTION
ユーザーの表示名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authordisplayname.html)

### MTAuthorEmail — FUNCTION
ユーザーのメールアドレス

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authoremail.html)

### MTAuthorEntriesCount — FUNCTION
ユーザーの記事の数（未公開を含む）

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorentriescount.html)

### MTAuthorEntryCount — FUNCTION
ユーザーの公開済み記事の数

属性: `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorentrycount.html)

### MTAuthorHasContent — BLOCK
ユーザーが公開済みのコンテンツデータを持っていれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorhascontent.html)

### MTAuthorHasEntry — BLOCK
ユーザーが公開済みの記事を持っていれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorhasentry.html)

### MTAuthorHasPage — BLOCK
ユーザーが公開済みのウェブページを持っていれば中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorhaspage.html)

### MTAuthorID — FUNCTION
ユーザーの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorid.html)

### MTAuthorName — FUNCTION
ユーザー名（サインイン名。公開ページでは表示名の方を推奨）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorname.html)

### MTAuthorNext — BLOCK
次のユーザーアーカイブを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authornext.html)

### MTAuthorPrevious — BLOCK
前のユーザーアーカイブを文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorprevious.html)

### MTAuthorRank — FUNCTION
ユーザーのスコアから出した順位（要スコア用プラグイン）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorrank.html)

### MTAuthorScore — FUNCTION
ユーザーのスコアの合計（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorscore.html)

### MTAuthorScoreAvg — FUNCTION
ユーザーのスコアの平均（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorscoreavg.html)

### MTAuthorScoreCount — FUNCTION
ユーザーにスコアが付いた回数（要スコア用プラグイン）

属性: `namespace="foo"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorscorecount.html)

### MTAuthorScoreHigh — FUNCTION
ユーザーのスコアの最高値（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorscorehigh.html)

### MTAuthorScoreLow — FUNCTION
ユーザーのスコアの最低値（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorscorelow.html)

### MTAuthorURL — FUNCTION
ユーザーのウェブサイト URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authorurl.html)

### MTAuthorUserpic — FUNCTION
ユーザーのプロフィール画像の img 要素

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authoruserpic.html)

### MTAuthorUserpicAsset — BLOCK
ユーザーのプロフィール画像をアセットとして文脈にする

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authoruserpicasset.html)

### MTAuthorUserpicURL — FUNCTION
ユーザーのプロフィール画像の URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authoruserpicurl.html)

### MTAuthors — BLOCK
ユーザーを並べるループ（並び順・件数・役割で絞れる）

属性: `display_name="foo"` `lastn="N"` `sort_order="ascend | descend"` `sort_by="created_on | name | display_name | email | url"` `role="ロール (役割)"` `need_entry="0 | 1"` `any_type="0 | 1"` `need_association="0 | 1"` `status="enabled | disabled"` `namespace="foo"` `scoring_to="foo"` `min_score="N"` `max_score="N"` `min_rate="N"` `max_rate="N"` `min_count="N"` `max_count="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/authors.html)

### MTOtherProfileVar — FUNCTION
外部サービスの値を取り出す

属性: `name`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/otherprofilevar.html)

### MTPasswordValidation — FUNCTION
パスワード検証用の JavaScript

属性: `form="foo"` `password="foo"` `username="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/passwordvalidation.html)

### MTPasswordValidationRule — FUNCTION
パスワード検証の規則（最低文字数など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/passwordvalidationrule.html)

### MTSignOutLink — FUNCTION
コメント投稿者のサインアウトのリンク

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/signoutlink.html)

### MTSigninLink — FUNCTION
コメント投稿者のサインインのリンク

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/signinlink.html)

### MTUserSessionState — FUNCTION
サインイン中のユーザーを表す JSON

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/usersessionstate.html)
