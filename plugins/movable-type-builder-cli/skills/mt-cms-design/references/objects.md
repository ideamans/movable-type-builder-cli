# Movable Type のオブジェクトモデル

サイトを設計するときは、**何をどの入れ物に入れるか**を最初に決める。
MT には似た入れ物が複数あり、選び違えると後から作り直しになる。

## 全体の形

```
システム
└── ウェブサイト (website)          … 独立したサイト。URL とディレクトリを持つ
    ├── ブログ (blog)               … ウェブサイトの下にぶら下がる子サイト
    ├── 記事 (entry)                … 日付を持つ。カテゴリで分類
    ├── ウェブページ (page)          … 日付を持たない固定ページ。フォルダで分類
    ├── コンテンツデータ (content data) … 任意の項目を定義した構造化コンテンツ
    ├── アセット (asset)            … 画像・ファイル
    ├── テンプレート (template)      … 出力の定義
    └── カスタムフィールド (field)    … 上記に足す独自項目
```

**サイトは2階層まで。** ブログの下にブログは作れない
（`400 Cannot create a blog under blog` になる）。

## 入れ物の選び分け

| 作りたいもの | 使う入れ物 | 理由 |
| --- | --- | --- |
| 時系列のお知らせ・ブログ記事 | **記事** | 日付アーカイブ、カテゴリ、タグが最初から使える |
| 会社概要・サービス紹介などの固定ページ | **ウェブページ** | 日付を持たず、フォルダで階層 URL を作れる |
| 製品・事例・スタッフなど、**項目が決まっているもの** | **コンテンツタイプ** | 項目ごとに入力欄と検証が付く。一覧の並べ替えも項目単位 |
| 上のどれかに1〜2項目足したいだけ | **カスタムフィールド** | コンテンツタイプを作るより軽い |
| 複数の場所で同じ文言を使い回したい（住所・注意書き） | **整形済みテキスト** | サイト単位の名前付きテキスト。`sites formatted-texts` |

**迷ったら「項目が3つ以上あって、入力を揃えたいか」で決める。**
揃えたいならコンテンツタイプ、そうでなければ記事＋カスタムフィールド。

## 記事 (entry)

日付 (`date`) を持ち、`status` で公開状態を持つ。

- `status` — `Publish` / `Draft` / `Review` / `Future` / `Spam` / `Unpublish`
- **`status` を省くと下書きになる。** 公開したいなら明示する
- `categories` は配列。最初の1件が主カテゴリ
- `tags` は自由なラベル。カテゴリと違って階層を持たない
- `basename` が出力ファイル名になる。日本語タイトルからは自動生成されない

```bash
mtb sites entries create 4 --entry '{"title":"…","body":"…","status":"Publish","categories":[{"id":13}],"tags":["お知らせ"]}'
```

## ウェブページ (page)

記事から日付とカテゴリを外し、代わりに**フォルダ**を持つもの。

- フォルダは URL の階層になる（`about` フォルダ → `/about/～`）
- **作成時に `folder` を指定しても効かない。作ってから更新で設定する**
- `basename` を `index` にするとフォルダの既定ページになる

```bash
mtb sites pages create 1 --page @page.json --fields id
mtb sites pages update 1 <id> --page '{"folder":{"id":11}}'
```

## コンテンツタイプ / コンテンツデータ (MT7 以降)

「項目を自分で定義した記事」。項目は**コンテンツフィールド**。

```bash
mtb sites content-types create 4 --content-type '{"name":"製品"}'
mtb sites content-types fields create 4 <ct-id> --content-field '{"label":"本文","type":"multi_line_text","required":true,"options":{"label":"本文","input_format":"richtext"}}'
mtb sites content-types data create 4 <ct-id> --content-data '{"label":"製品A","data":[{"id":<field-id>,"data":"…"}]}'
```

### 使えるコンテンツフィールド

テキスト / テキスト(複数行) / 数値 / URL / 日付と時刻 / 日付 / 時刻 /
セレクトボックス / ラジオボタン / チェックボックス / 埋め込みテキスト /
**コンテンツタイプ**（他のコンテンツデータを参照）/ アセット / オーディオ /
ビデオ / 画像 / **カテゴリ**（カテゴリセットから選ぶ）/ タグ / リスト / テーブル

「コンテンツタイプフィールド」で**コンテンツデータ同士を関連付けられる**のが、
記事＋カスタムフィールドでは作れないところ。製品と事例を結ぶ、といった構造に使う。

**踏みやすい罠**

- フィールドの名前は `name` ではなく **`label`**
- コンテンツデータには **`label` が必須**（データ自体の名前。フィールドではない）
- `data` は**フィールドIDの配列**。IDは `fields list` で調べる
- コンテンツデータのアーカイブテンプレート（`type: ct`）が無いとプレビューできない。
  `ct` テンプレートを作るには `contentType` の指定が要る

## カスタムフィールド

既存のオブジェクトに項目を足す。**作成後に「システムオブジェクト」と「種類」は変更できない。**

```bash
mtb sites fields create 4 --field '{"name":"想定読者","basename":"target_reader","systemObject":"entry","type":"select","options":"開発者,運用者","tag":"TargetReader"}'
```

### 対象にできるオブジェクト（`systemObject`）

サイト配下の API (`/sites/{id}/fields`) で作れるもの（実測）:

`entry` `page` `category` `folder` `comment` `template` `audio` `video` `image` `file`
＋ ウェブサイトでは `website`、ブログでは `blog`

**`author`（ユーザー）はシステム全体のカスタムフィールドで、サイト配下の API では作れない。**
管理画面のシステム側から作る。`asset` というキーは無く、**アセット全般は `file`**。

### 種類（`type`）

`text` `textarea` `checkbox` `url` `datetime` `select` `radio` `embed` `post_type`
`audio` `video` `image` `file`

- **`select` `radio` `datetime` は `options` が必須**（`datetime` は `date_and_time` など）
- 選択肢は `赤,青,緑` のようにカンマ区切り
- **表示と保存を分けたいときは `保存値=表示値`** — `red=赤,blue=青`
- 既定値は後から変えられる。変えても入力済みの値は残る

### ベースネームとテンプレートタグ

- `basename` は**サイト内で一意**。重複すると値が正しく保存されないことがある
- `tag` はテンプレートで出力するときのタグ名。**MT 標準のタグと衝突させないこと**。
  日付なら `CFEntryDate` のように接頭辞を付けて避ける

## カテゴリとフォルダ

どちらも階層を持ち、`basename` が URL になる。

- **カテゴリ**は記事用、**フォルダ**はウェブページ用。混ぜられない
- **カテゴリセット**（MT7 以降）はコンテンツタイプ用のカテゴリ。上の2つとは別物
- 並び替えは `permutate`。**`{"id":N}` だけの配列では通らない**。
  `list` の出力をそのまま並べ替えて渡す

```bash
mtb sites categories list 4 --compact > cats.json   # 並べ替えてから
mtb sites categories permutate 4 --categories @cats.json
```

## アセット (asset)

アップロードした画像・ファイル。記事やページに `assets` で紐づける。

- **MT は画像を再保存する。** JPEG は quality 85 で入れ直され、元のバイト列は残らない
- 画像以外（PDF・テキストなど）はバイト単位でそのまま
- サムネイルは `assets thumbnail` で任意サイズを生成できる
- `--normalize-orientation 1` で EXIF の向きを物理回転して正す（既定は 1）

## テンプレート (template)

出力の定義。`type` で役割が決まる。

| type | 役割 |
| --- | --- |
| `index` | 単独のファイルを出す（トップページ、CSS、フィードなど）。`outputFile` を持つ |
| `individual` | 記事1件のページ |
| `page` | ウェブページ1件 |
| `archive` | カテゴリ別・日付別などの一覧 |
| `ct` / `ct_archive` | コンテンツデータ用。作成時に `contentType` が要る |
| `custom` | 他から読み込むモジュール（`mt:Include`） |
| `widget` | ウィジェット |

**`index` 以外は、アーカイブマッピングでアーカイブ種別と結びつける**（`archives.md`）。

`mtb sites refresh-templates` は**再構築ではなく初期化**。書いたテンプレートが消える。

## ユーザー・グループ・ロール

`permissions.md` を参照。要点だけ:

- 権限は**ロールにまとめ、サイト単位で**ユーザーかグループに与える
- **Data API で作ったユーザーは、そのままでは Data API にサインインできない。**
  `sign_in_data_api` 権限・Webサービスパスワード・サイト側の Data API 許可の
  3つが要る。詳細は `permissions.md`
- 記事の作成者は**認証しているユーザー**になる。別名義で投稿するならその人のプロファイルでサインインする
