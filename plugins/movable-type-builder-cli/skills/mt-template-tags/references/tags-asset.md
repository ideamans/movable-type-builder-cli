# アセット（画像・ファイル）

アップロードした画像やファイル。

現行のタグ 52 件。非推奨のものは `tags-deprecated.md` にまとめてあります。
説明は1行の要約です。**属性の意味・使用例・注意はリンク先のマニュアルで確かめること。**

### MTAsset — BLOCK
id で指定した1つのアセットを文脈にする

属性: `id="ID"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/asset.html)

### MTAssetAddedBy — FUNCTION
アセットを追加したユーザーの表示名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetaddedby.html)

### MTAssetBlogID — FUNCTION
アセットが属するサイトの ID（システムのものは 0）

属性: `pad="0|1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetblogid.html)

### MTAssetCount — FUNCTION
サイトのアセットの数（サムネイルは数えない）

属性: `type="image | audio | video | file"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetcount.html)

### MTAssetDateAdded — FUNCTION
アセットを追加した日時（format で書式指定）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetdateadded.html)

### MTAssetDescription — FUNCTION
アセットの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetdescription.html)

### MTAssetFileExt — FUNCTION
アセットの拡張子（ドット無し）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetfileext.html)

### MTAssetFileName — FUNCTION
アセットのファイル名

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetfilename.html)

### MTAssetFilePath — FUNCTION
アセットのサーバ上のパス

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetfilepath.html)

### MTAssetID — FUNCTION
アセットの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetid.html)

### MTAssetIfTagged — BLOCK
アセットに指定のタグが付いていれば中身を出す条件タグ

属性: `tag="tag_foo"` `include_private="0"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetiftagged.html)

### MTAssetIsFirstInRow — BLOCK
MTAssets の assets_per_row で区切った行の先頭なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetisfirstinrow.html)

### MTAssetIsLastInRow — BLOCK
MTAssets の assets_per_row で区切った行の末尾なら中身を出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetislastinrow.html)

### MTAssetLabel — FUNCTION
アセットの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetlabel.html)

### MTAssetLink — FUNCTION
アセットへのリンク（a 要素）

属性: `new_window="1 | 0"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetlink.html)

### MTAssetMimeType — FUNCTION
アセットの MIME タイプ

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetmimetype.html)

### MTAssetProperty — FUNCTION
アセットの詳細データ（property で項目を選ぶ。寸法など）

属性: `property="file_size | image_width | image_height | label | description"` `image_width` `image_height` `label` `description`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetproperty.html)

### MTAssetRank — FUNCTION
アセットのスコアから出した順位（要スコア用プラグイン）

属性: `max="N"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetrank.html)

### MTAssetScore — FUNCTION
アセットのスコアの合計（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetscore.html)

### MTAssetScoreAvg — FUNCTION
アセットのスコアの平均（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetscoreavg.html)

### MTAssetScoreCount — FUNCTION
アセットにスコアが付いた回数（要スコア用プラグイン）

属性: `namespace="foo"` `singular="text"` `plural="text"` `none="text"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetscorecount.html)

### MTAssetScoreHigh — FUNCTION
アセットのスコアの最高値（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetscorehigh.html)

### MTAssetScoreLow — FUNCTION
アセットのスコアの最低値（要スコア用プラグイン）

属性: `namespace="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetscorelow.html)

### MTAssetSiteID — FUNCTION
アセットが属するサイトの ID

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetsiteid.html)

### MTAssetTags — BLOCK
アセットに付いたタグを並べるループ（glue で区切り）

属性: `glue="foo"` `include_private="0"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assettags.html)

### MTAssetThumbnailLink — FUNCTION
画像アセットのサムネイルを、元画像へのリンク付きで出す

属性: `width="value_foo"` `height="value_foo"` `scale="scale_foo"` `square="1 | 0"` `force="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetthumbnaillink.html)

### MTAssetThumbnailURL — FUNCTION
画像アセットのサムネイルの URL（大きさを指定できる）

属性: `width="value_foo"` `height="value_foo"` `scale="scale_foo"` `square="1 | 0"` `force="1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetthumbnailurl.html)

### MTAssetType — FUNCTION
アセットの種類（image・audio・video・file など）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assettype.html)

### MTAssetURL — FUNCTION
アセットの元ファイルの URL

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/asseturl.html)

### MTAssets — BLOCK
アセットを並べるループ（種類・タグ・並び順で絞れる）

属性: `lastn="N"` `offset="N"` `limit="N"` `author="author_foo"` `days="N"` `type="image | audio | video | file"` `file_ext="extension_foo"` `sort_order="ascend | descend"` `sort_by="file_name | created_by | created_on"` `blog_ids="all | ID"` `namespace="foo"` `blog_ids="all | children | siblings | ID"` `include_blogs="all | children | siblings | ID"` `include_with_website="1"` `include_websites="1"` `exclude_blogs="ID"` `exclude_websites="1"` `assets_per_row="N"` `ignore_archive_context="0 | 1"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assets.html)

### MTAssetsFooter — BLOCK
MTAssets の最後の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetsfooter.html)

### MTAssetsHeader — BLOCK
MTAssets の最初の1回だけ出す

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/assetsheader.html)

### MTAudioCustomFieldDescription — FUNCTION
オーディオのカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/audiocustomfielddescription.html)

### MTAudioCustomFieldName — FUNCTION
オーディオのカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/audiocustomfieldname.html)

### MTAudioCustomFieldValue — FUNCTION
オーディオのカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/audiocustomfieldvalue.html)

### MTAudioCustomFields — BLOCK
オーディオのカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/audiocustomfields.html)

### MTFileCustomFieldDescription — FUNCTION
ファイル（アイテム）のカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/filecustomfielddescription.html)

### MTFileCustomFieldName — FUNCTION
ファイル（アイテム）のカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/filecustomfieldname.html)

### MTFileCustomFieldValue — FUNCTION
ファイル（アイテム）のカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/filecustomfieldvalue.html)

### MTFileCustomFields — BLOCK
ファイル（アイテム）のカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/filecustomfields.html)

### MTFileTemplate — FUNCTION
アーカイブマッピングの規則に沿ったアーカイブのパス

属性: `format="format"` `separator="foo"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/filetemplate.html)

### MTImageCustomFieldDescription — FUNCTION
画像のカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/imagecustomfielddescription.html)

### MTImageCustomFieldName — FUNCTION
画像のカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/imagecustomfieldname.html)

### MTImageCustomFieldValue — FUNCTION
画像のカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/imagecustomfieldvalue.html)

### MTImageCustomFields — BLOCK
画像のカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/imagecustomfields.html)

### MTImageHeight — FUNCTION
アップロードした画像の高さ（ポップアップ画像のテンプレート向け）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/imageheight.html)

### MTImageURL — FUNCTION
アップロードした画像の URL（ポップアップ画像のテンプレート向け）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/imageurl.html)

### MTImageWidth — FUNCTION
アップロードした画像の幅（ポップアップ画像のテンプレート向け）

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/imagewidth.html)

### MTVideoCustomFieldDescription — FUNCTION
ビデオのカスタムフィールドの説明

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/videocustomfielddescription.html)

### MTVideoCustomFieldName — FUNCTION
ビデオのカスタムフィールドの名前

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/videocustomfieldname.html)

### MTVideoCustomFieldValue — FUNCTION
ビデオのカスタムフィールドの値

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/videocustomfieldvalue.html)

### MTVideoCustomFields — BLOCK
ビデオのカスタムフィールドを並べるループ

属性: `include="カスタムフィールドの名前"` `exclude="カスタムフィールドの名前"`

[マニュアル](https://www.movabletype.jp/documentation/appendices/tags/videocustomfields.html)
