# グローバルモディファイア

**どのファンクションタグにも付けられる**共通の修飾子です。
ブロックタグには付きません（ブロックタグの属性は、そのタグ固有のモディファイアです）。

```html
<$mt:EntryTitle trim_to="20+…" encode_html="1"$>
```

**複数指定したときは、書いた順に左から適用されます。**

全 48 件。説明は1行の要約で、値の取り方や例はリンク先のマニュアルにあります。

## 出力の加工

- `_default` — 値が空のときに代わりに出す値 [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/default.html)
- `capitalize` — 単語の頭を大文字、残りを小文字にする [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/capitalize.html)
- `cat` — 値の後ろに指定の文字列をつなぐ [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/cat.html)
- `indent` — 各行の頭に指定の数の空白を入れる [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/indent.html)
- `lower_case` — 小文字にする [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/lower_case.html)
- `ltrim` — 先頭の空白を取る [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/ltrim.html)
- `nl2br` — 改行を br 要素にする [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/nl2br.html)
- `numify` — 数字を3桁ごとにカンマで区切る（区切り文字は指定可） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/numify.html)
- `regex_replace` — 正規表現で置き換える（"パターン","置換後"） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/regex_replace.html)
- `remove_html` — HTML のタグを取り除く [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/remove_html.html)
- `replace` — 文字列を置き換える（"探す","置換後"） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/replace.html)
- `rtrim` — 末尾の空白を取る [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/rtrim.html)
- `space_pad` — 指定の文字数に満たなければ先頭を空白で埋める [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/space_pad.html)
- `spacify` — 1文字ずつの間に指定の文字を入れる [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/spacify.html)
- `sprintf` — printf の書式で整える [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/sprintf.html)
- `string_format` — printf の書式で整える（sprintf と同じ） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/string_format.html)
- `strip` — 連続する空白を指定の文字に置き換える [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/strip.html)
- `strip_linefeeds` — 改行を取り除く [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/strip_linefeeds.html)
- `strip_tags` — HTML のタグを取り除く [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/strip_tags.html)
- `trim` — 前後の空白を取る [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/trim.html)
- `trim_to` — 先頭から指定の文字数だけ残す（"20+…" で末尾に付ける文字） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/trim_to.html)
- `upper_case` — 大文字にする [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/upper_case.html)
- `wrap_text` — 指定の長さで折り返す [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/wrap_text.html)
- `zero_pad` — 指定の文字数に満たなければ先頭を 0 で埋める [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/zero_pad.html)

## エスケープ・エンコード

- `decode_html` — HTML の実体参照を元の文字に戻す [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/decode_html.html)
- `decode_xml` — XML の実体参照を元の文字に戻す [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/decode_xml.html)
- `dirify` — ファイル名・URL に使える形にする（記号を除き、空白を _ など） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/dirify.html)
- `encode_html` — HTML の特殊文字をエスケープする [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/encode_html.html)
- `encode_js` — JavaScript の文字列に入れられるようにエスケープする [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/encode_js.html)
- `encode_json` — JSON の文字列に入れられるようにエスケープする [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/encode_json.html)
- `encode_php` — PHP のコードに入れられるようにエスケープする（q・qq・here） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/encode_php.html)
- `encode_sha1` — SHA-1 の16進にする [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/encode_sha1.html)
- `encode_url` — URL エンコードする [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/encode_url.html)
- `encode_xml` — XML に入れられるようにエンコードする [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/encode_xml.html)
- `escape` — 指定の方式でエスケープする（html・url・javascript など） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/escape.html)
- `nofollowfy` — 値の中の a 要素に rel="nofollow" を付ける [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/nofollowfy.html)
- `sanitize` — 許可したタグ以外の HTML を取り除く [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/sanitize.html)
- `spam_protect` — メールアドレスの記号を数値文字参照にする（収集ロボット対策） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/spam_protect.html)

## 数える

- `count_characters` — 値の代わりに文字数（空白を含む）を出す [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/count_characters.html)
- `count_paragraphs` — 値の代わりに行数（空行を除く）を出す [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/count_paragraphs.html)
- `count_words` — 値の代わりに単語数を出す [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/count_words.html)

## テンプレートの制御

- `filters` — テキストフォーマットを当てる（markdown など） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/filters.html)
- `mteval` — 値の中のテンプレートタグを処理する [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/mteval.html)
- `setvar` — 出力せず、結果を指定の名前の変数に入れる [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/setvar.html)

## 記号の整形（要プラグイン）

- `smart_dashes` — ダッシュを整形する（要 SmartyPants） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/smart_dashes.html)
- `smart_ellipses` — 「...」を省略記号にする（要 SmartyPants） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/smart_ellipses.html)
- `smart_quotes` — 引用符を整形する（要 SmartyPants） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/smart_quotes.html)
- `smarty_pants` — ダッシュ・引用符・省略記号をまとめて整形する（要 SmartyPants） [マニュアル](https://www.movabletype.jp/documentation/appendices/modifiers/smarty_pants.html)
