---
name: mtb-migrate
description: Copy a Movable Type site's design from one server to another — staging to production, or between sites — by dumping both with mtb, diffing them, agreeing a plan with the user, and applying it. Use when someone asks to sync, migrate, deploy or promote MT templates, categories, custom fields or content types between environments, or asks what differs between two MT servers.
license: MIT
compatibility: Requires the `mtb` binary on PATH (run the mtb-install skill if missing) and **two configured profiles, both signed in** — `mtb auth status --profile <name>` must succeed for each. Writes to a live CMS on the receiving side.
---

# MT のサイトを別の MT へ同期する

ステージングで組んだデザインを本番へ持っていく、という作業を、
**どこが違うのかを機械で出してから**行う。

管理画面を2つ開いて見比べるのが一番時間を食う。`mtb dump` は接続先の中身を
種類ごとの YAML に書き出すので、差分は普通の diff で読める。

## 0. 何を運ぶかを先に決める（飛ばさない）

**「全部」と言われても、そのまま全部やらない。** MT のオブジェクトは
性質が4つに分かれていて、運んでよいものと運んではいけないものがある。

| | 中身 | 既定 |
| --- | --- | --- |
| **器** | テンプレート／アーカイブマッピング／ウィジェット／ウィジェットセット／カテゴリ／フォルダ／カテゴリセット／コンテンツタイプとそのフィールド／カスタムフィールド／整形済みテキスト | **運ぶ** |
| **判断が要る** | ウェブページ／コンテンツデータ／アセット | **必ず聞く** |
| **システム** | 利用者／グループ／ロール／権限／プラグイン／サイトそのもの | 依頼があれば |
| **コンテンツ** | 記事／コメント／タグ | 依頼があれば |
| 運ばない | ログ／アクセス統計 | — |

この4分類で API の 220 コマンドが漏れなく重なりなく収まる。**「これはどれ？」
と迷うものが出たら、それは相手に聞く合図**——たいてい真ん中の行に入る。

**テーマはこの表に入らない。** 器を1つずつ運ぶのではなく、器を丸ごと
入れ替える別の手段だから。同期の途中で当ててはいけない（後述）。

**真ん中の3つは、そのプロジェクトでどちらなのかを聞かないと決められない。**

- **ウェブページ** — 会社概要・料金・特商法のような固定ページは、体裁の一部
  なので器として運ぶことが多い。一方、更新の速いお知らせページなら
  コンテンツ側。**「そのページは誰が書き換えるか」で決まる**
- **コンテンツデータ** — 製品マスタや事例のように構成の一部なら器。
  日々増えるものならコンテンツ
- **アセット** — `theme-images/` のような体裁の画像は器。記事に貼った写真は
  コンテンツ。**同じサイトに両方あるのが普通**なので、パスで分けられるか聞く

聞き方は、分類を見せて「このプロジェクトではどれを運びますか」でよい。
決まったら**確認のために復唱してから**先へ進む。

## 1. 両方をダンプして差分を見る

```bash
mtb dump --profile stg  --dir ./dump/stg
mtb dump --profile prod --dir ./dump/prod
diff -ru ./dump/prod ./dump/stg
```

読むのは、0 で決めた種類のファイルだけでよい。器だけなら
`templates.yaml` `categories.yaml` `custom-fields.yaml` `content-types.yaml`
`widgets.yaml` `widgetsets.yaml` `formatted-texts.yaml`。

**サイト設定は2つのファイルに分かれている。**

- `site.yaml` — 運ぶべき設定（`smartReplace` `convertParas` `language` など）
- `site-environment.yaml` — **運んではいけない**（`host` `sitePath` `archivePath`
  `relativeUrl` `extraPath` `uploadDestination`）。運ぶと移行先が行方不明になる

**後者に差分があるのは正常。** 見るのは前者だけでよい。

**アセットの中身は環境をまたいで比べられない。** MT は取り込むときに JPEG を
再エンコードするので、同じ画像を送っても保存されるバイト列が変わる
（実測: 95362 → 95339）。ダンプは `fileSize` を出さず、寸法（`width` /
`height`）で見る——再エンコードでは変わらないため。`--asset-hash` は
**同じ環境で「いつ変わったか」を見る用途**で、移行の突き合わせには使えない。

### 運ぶ前に、移行先の前提を確かめる

**サイトを完全に再現しても、サーバの設定が違えば動きは違う。** ここは運べない
ので、人が揃える。`system/` の差分を先に見ること。

```bash
diff -ru ./dump/prod/system ./dump/stg/system
```

- `plugins.yaml` — 受け側に無いプラグインに依存したテンプレートは公開で落ちる
- `text-filters.yaml` — 受け側に無い書式（`markdown` `block_editor` など）の
  記事を運んでも読めない
- `endpoints.yaml` — 受け側に無い機能は叩いた時点で 404

**`mt-config.cgi` は API から読めない。** 差分にも出ないので、相手に聞くか
SSH で見る。移行で効いてくるのは主にこの3つ。

| | 揃っていないと |
| --- | --- |
| `EmailAddressMain` | メール送信系が 500。**実際に片方だけ未設定だった** |
| `DeniedAssetFileExtensions` | `.js` `.html` がアセットにできない |
| `DefaultLanguage` | 取り込みの文字コードの有効値が変わる |

そのほか API では触れないもの: **Webサービスパスワード**（移行先で毎回手作業）、
**`theme_export_settings`**（静的ファイルのテーマ書き出し。管理画面のみ）。

**カスタムブロック（ブロックエディタ）は Data API で扱えない。** MT では
`be_block` というオブジェクトだが、エンドポイントが1つも無い（実機の 266 個を
確認）。ダンプにも出ないので、**差分にすら現れない**。

**運ぶならテーマ経由。** これだけが道。

```bash
mtb sites export-theme <site> --include be_block,be_config,template_set \
  --theme-id my-theme --theme-name "My Theme" --overwrite-yes 1
```

`block_editor_blocks/<識別子>.json` としてテーマに入る。**適用はテンプレートの
総入れ替えを伴う**ので、移行の最初にやること（後から当てると作業が消える）。

記事の側は普通に運べる。カスタムブロックを使った本文は
`<!-- mt-beb t="<識別子>" -->…<!-- /mt-beb -->` の形で、原稿のまま往復する。
**ただし定義が無い移行先では、その記事は生の HTML ブロックとして開く。**

## 2. 計画を立てて見せる

差分を「作る／直す／消す」に分けて、**実行前に相手に見せる**。

```
作る  : テンプレート「製品詳細」、カテゴリ「導入事例」
直す  : テンプレート「HTMLヘッダー」「スタイルシート」
消す  : （なし）
運ばない: site.yaml の url / host / sitePath、記事279件、利用者
```

**消すのは既定でやらない。** 受け側にしか無いものは、本番で誰かが足した
ものかもしれない。**一覧にして見せ、消すかどうかを個別に聞く。**

## 3. 順番を守って適用する

```
サイトの設定 → カテゴリ・フォルダ → カスタムフィールド
  → コンテンツタイプとフィールド → テンプレート → アーカイブマッピング
  → （必要なら）アセット・ページ・記事 → 関連付け → 公開
```

**サイトの設定を最初に運ぶ。** `smartReplace` や `convertParas` は
**あとで運ぶと、それまでに作った記事やページの見え方が変わる**
（`format: null` の記事はサイトの既定に従うため）。

```bash
# 運ぶのは site.yaml の項目だけ。site-environment.yaml には触れない
mtb sites update <site> --profile prod --website @site-settings.json
```

**分類と項目を先に作る。** テンプレートが参照するものが無いと、貼った時点では
通っても公開で落ちる。

テンプレートは**名前で突き合わせる**。

```bash
# 受け側に同じ名前があるか
mtb sites templates list <site> --profile prod --limit 300 --fields id,name --extract items

# 無ければ作る
mtb sites templates create <site> --profile prod \
  --set template.name="製品詳細" --set template.type=index \
  --set template.outputFile=products.html --set template.text=@products.mtml

# あれば直す（本文だけ差し替える）
mtb sites templates update <site> <tid> --profile prod --set template.text=@products.mtml
```

**`mtb sites templates push` を別の MT に向けないこと。** マニフェストが持って
いる ID は出し側のもので、受け側では別の何かを指すか存在しない。移行では
名前で突き合わせて create / update する。

**名前は MT の言語で訳される。** テーマから入ったテンプレートは、テーマの
`name: Main Index` が日本語の MT では「メインページ」になる。言語の違う MT どうしでは
名前が合わないので、インデックス以外は `type`、インデックスは出力ファイル名
（`outputFile`）で突き合わせる。

### 実機で踏んだ罠（どれもエラーが出ないか、出ても意味が読めない）

- **システムテンプレートは消せない**（403）。検索結果・コメント関連など。
  名前ではなく **`type` で突き合わせて更新**する。通常のテンプレートは消せる
- **`templateType` は作成時に送らないと欠ける**（`index_page` などが null になる）。
  一覧の差分では気づきにくい
- **`templatemaps list` はアーカイブ系以外で 400**。
  `archive` `individual` `page` `category` `ct_archive` に絞ってから引く
- **コンテンツタイプの `dataLabel` はフィールドの uniqueID を指す。**
  受け側では別の値になるので、①型を作る ②フィールドを作る
  ③受け側の uniqueID で貼り直す、の3手順。ダンプでは `dataLabelField`
  としてフィールド名で出る
- **記事の中の `categories` は `basename` を持たない**（`id` と `label` だけ）。
  送り側の一覧から `id → 自然キー` の対応表を作ってから変換する。
  これを怠ると空配列を送ることになり、**受け側のカテゴリを消す**
- **タイトルの記号が化ける。** em ダッシュ「—」が「--」になる。smartReplace は
  **作成のときだけ**効くので、**作った直後に同じ値で更新し直すと原文に戻る**
- **`description` の `null` と `""` は別物。** 空文字を送ると null に戻せる
  （`{"description":null}`）が、送らないと `""` のまま残る
- **ページを更新すると、指定しなかった関連が外れる。** タグだけ直すつもりで
  送ったら `folder` が null になった。**関連は毎回まとめて送り直す**
- **`format: null` は「サイトの既定に従う」という意味。** 送らないと受け側の
  既定が入る。原文どおりにしたいなら `null` を明示的に送る
- **記事の著者は変えられない。** `author` は Data API では読み取り専用で、
  `{"author":{"id":5}}` を送っても **200 が返って黙って無視される**。
  作成時の著者は**サインインしている利用者**になるので、運んだ記事は
  すべて「運んだ人」のものになる。元の著者を保ちたいなら、その人の
  プロファイルでサインインして運ぶしかない（人数分プロファイルが要る）。
  **mtb-pro 拡張があれば、運んでから `mtb sites entries author update <site> <id>
  --author-name <名前>` で直せる**（ID は環境で違うので名前で指定する）。
  ダンプは著者を出すので、**差分を見れば誰のものになったか分かる**

アセットを運ぶなら **`--replace` が要る**。MT はアセットを上書きできず、
同じ名前で送ると 409 になる。

```bash
mtb assets upload --profile prod --site-id <site> --dir ./assets --path /theme-images --replace --yes
```

## 4. 公開して、反映まで確かめる

```bash
mtb publish entries --profile prod --ids <id>,<id> --wait --verify-publish
```

`--wait` が無いと1回で終わり、成功したように見える。`--verify-publish` は
静的ファイルの書き出しに失敗していないかを MT のログで確かめる（失敗なら
終了コード 8）。**API が 200 を返すことと、サイトが変わったことは別。**

## 5. もう一度ダンプして、差分が消えたことを見る

```bash
mtb dump --profile prod --dir ./dump/prod-after
diff -ru ./dump/prod-after ./dump/stg
```

**これをやらないと、やったつもりで終わる。** 残る差分は 0 で決めた
「運ばないもの」と `site-environment.yaml` だけのはず。そうでなければ、
まだ終わっていない。

実際に器＋コンテンツを丸ごと運んだときは、最後にこうなった。

```diff
-  host: 13.114.133.144
+  host: 54.249.24.57
```

**これ1行だけ**（`site-environment.yaml`）。ここまで来て初めて終わり。

## テーマをファイルとして開発する（サーバに触れる場合）

**テーマを外から MT に入れる経路は Data API に無い。** `/themes` は GET と
DELETE と apply だけで、install も upload も無い（実機の 266 個を確認）。
入れるにはサーバのファイルシステムか管理画面しかない。
**mtb-pro 拡張があれば `mtb themes files upload <id> --dir <dir>` で送れる**
（MT 側で `MtbProAllowFileUpload 1` が有効なとき）。

**触れるなら、これが一番きれいな進め方。** テーマを Git のファイルとして持ち、
配置は rsync、適用と確認は mtb が受け持つ。

```
theme-v1/
  theme.yaml                     テンプレートの一覧・アーカイブマッピング・静的ファイル
  templates/*.mtml               本文
  blog_static/<フォルダ>/...      静的ファイル
```

```bash
rsync -a theme-v1/ prod:/path/to/themes/mtb_v1/     # 配置（mtb の外）
mtb themes list --profile prod                       # 見えることを確認
mtb sites themes apply <site> mtb_v1 --profile prod --yes --confirm "<サイト名>"
mtb sites templates publish <site> <index-tid> --profile prod
```

**静的ファイルは `blog_static/` の下に置く。** `static/` のような素の名前で
置いても運ばれず、**エラーも出ないまま 404 になる**（実際に踏んだ）。
`theme.yaml` の `blog_static_files` の `data` には、`blog_static/` の下の
フォルダ名を並べる。

`theme.yaml` に `mappings` を書けば**アーカイブマッピングも一緒に入る**ので、
移行で別途作り直す必要が無い。

```yaml
archive:
  entry_archive:
    label: 記事アーカイブ
    mappings:
      individual:
        archive_type: Individual
        preferred: 1
```

**この経路なら、テーマが唯一の正になる。** サイトから export するのではなく、
ファイルからサイトを起こすので、Git の履歴がそのまま設計の履歴になる。
手順の詳細は **`mtb-theme` スキル**。

## 壊さない

- **テーマを当てない。** `sites themes apply` はテンプレートを総入れ替えし、
  アーカイブマッピングを削除して出力ファイル名も消す。同期の途中で当てると
  それまでの作業ごと消える
- **`site.yaml` の url / sitePath を運ばない。** 受け側のサイトが行方不明になる
- **消す操作は個別に確認する。** 受け側にしか無いものは本番で足されたもので
  ありうる
- 破壊的な操作の前に、対象を `get` して相手に見せる
- 迷ったら `--dry-run` で送る内容を確かめる

## 詰まったら

```bash
mtb llm | head -80                    # 鉄則と失敗モード
mtb dump --help                       # 何が出て何が出ないか
mtb <コマンド> --help                  # パラメータと公式リファレンスへのリンク
```

CMS の設計そのものを相談されたら `mt-cms-design`、テンプレートの中身を
書くなら `mt-template-tags` を使う。
