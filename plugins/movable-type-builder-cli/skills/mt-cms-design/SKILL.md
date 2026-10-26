---
name: mt-cms-design
description: Design and build a Movable Type site from a requirement — decide whether content belongs in entries, pages or content types, lay out categories, folders and custom fields, set up users and roles, then create it all through the mtb CLI. Use when someone asks for a site, blog, corporate site, product catalogue or membership structure to be built in Movable Type, or asks how MT's objects, archives or permissions work.
license: MIT
compatibility: Requires the `mtb` binary on PATH and a logged-in profile pointing at the target Movable Type (see the mtb-usage and mtb-install skills). Creates and modifies live CMS content.
allowed-tools: Bash(mtb:*) Bash(jq:*) Bash(command:*) Read Write
---

# mt-cms-design

Movable Type でサイトを組み立てる。**設計を決めてから作る**のがこのスキルの筋で、
作る手順そのものは `mtb` に任せる。

## 1. 何を作るのか決める

要件を聞いたら、まず**入れ物の割り当て**を決める。ここを間違えると後から作り直しになる。

| 要件の形 | 入れ物 |
| --- | --- |
| 時系列に増える。日付順に並べたい | **記事** |
| 増えない。URL を固定したい（会社概要・料金など） | **ウェブページ** |
| 項目が決まっていて、入力を揃えたい（製品・事例・スタッフ） | **コンテンツタイプ** |
| 既存のものに1〜2項目足すだけ | **カスタムフィールド** |

**判断に迷ったら「項目が3つ以上あって、毎回同じ形で入れたいか」で決める。**

決めたら、その根拠を1行で相手に伝えてから着手する。
「製品は項目が決まっているのでコンテンツタイプにします」で十分。

詳細は `references/objects.md`。

## 2. 接続先を確かめる

**作る前に必ず、どのサーバに書こうとしているかを声に出す。**
同じコマンドが本番にもステージングにも当たる。

```bash
mtb profile show
mtb auth status
mtb sites list --fields id,name,class,url
```

## 3. 作る順番を守る

```
権限の関門 → サイトとテーマ → 器の型 → 分類と項目 → ユーザー
           → コンテンツ → アーカイブ設計とテンプレート → 公開と確認
```

**テーマは最初に当てる。最後ではない。** テーマを適用すると、そのサイトの
テンプレートは全件 `backup` 型にリネームされ、**アーカイブマッピングは削除、
出力ファイル名も消える**（`MT::Theme::TemplateSet::apply`）。作り込んだ後に
当てると URL 設計ごとやり直しになる。マニュアルの適用ページに警告は無い。

**分類より先にコンテンツを作らない。** 後から付け替えると URL が変わる。

**権限は最後ではなく前提。** 編集者に入れさせるなら、コンテンツより先に
権限が要る（後から付けると、それまでのものが全部あなたの所有になる）。
そもそも API を叩く時点で「4. 権限を決める」の3つの関門が必要。

具体的なコマンドの並びは `references/recipes.md` にレシピとして置いてある。
コーポレート＋ブログ、製品カタログ、既存テーマの流用の3本。

## 4. 権限を決める

**まず組み込みロールで足りないか確かめる。** 9種類ある。

```bash
mtb roles list --fields id,name
mtb sites permissions grant --site-id 4 --user-id 2 --role-id 2
```

権限の一覧、ロールごとの中身、システム権限との違いは `references/permissions.md`。

**Data API で作ったユーザーは、そのままでは Data API にサインインできない。**
3つ揃って初めて通る。

1. `sign_in_data_api` システム権限 — これは API で付けられる
2. Webサービスパスワード — **MT が自動生成する**。API からは設定も取得もできず、
   管理画面の [新しいパスワードを発行] で出た値をコピーして渡す
3. サイト設定 → Web サービスで **Data API のアクセスが許可されていること**

ユーザーを作ったら、2 と 3 が管理画面の作業として残ることを必ず伝える。
**mtb-pro 拡張が入っていれば（`mtb extensions list`）2 と 3 も API で済む**:
`mtb users api-password issue <id>` と `mtb sites data-api update <site> --enabled true`。
**新しいサイトは Data API が無効で始まる**（インストーラが作った最初のサイトも、
API で作ったサイトも）。システム管理者は素通りするので気づかない（一般ユーザーだけ
中身の無い 403 になる）。サイトを足す設計なら、編集者を入れる前に許可する手順を含める。

**管理画面で更新する人には `sign_in_cms` も付ける。** 1 の `sign_in_data_api` だけでは、
Data API には入れても管理画面には入れない（ロールが正しくても「サインインできませんでした」）。

## 5. 出力の形を決める

静的ファイルをどう書き出すかは、テンプレートとアーカイブ種別の組み合わせで決まる。
種別の一覧と、テンプレートの `type` との噛み合わせは `references/archives.md`。

テンプレートの中身（MTタグ）を書くときは **`mt-template-tags` スキル**を使う。
テンプレートを Git で持ちたい、同じ設計を何サイトにも配りたいなら
**`mtb-theme` スキル**（テーマをファイルとして開発する）。

## 6. 公開する

| 対象 | 方法 | |
| --- | --- | --- |
| index テンプレート（トップ・CSS） | `mtb sites templates publish <site> <tid>` | 同期 |
| 記事・ページ | **中身が変わる更新を送る** | 同期 |
| まとめて | `mtb publish entries --ids … --wait` | 繰り返し呼ぶ |

**`{"status":"Rebuilding"}` は「受け付けた」だけ。** `publish entries` は
1回では終わらず、残りは `restIds` に入る。**`--wait` を付けること**——
付けないと1回で終わり、成功したように見える。

**保存し直しで再構築されるのは、値が実際に変わったときだけ。**
同じタイトルを送り直しても MT は何もしない。

```bash
mtb sites entries update 4 42 --entry '{"excerpt":"（いまと違う値）"}'
```

**`--entry '{}'` を再公開の目的で送らない。** 指定しなかったフォルダや
カテゴリが外れて URL が変わることがある。

**保存できたことと、サイトに出たことは別。** MT は同期の再構築に失敗しても
200 を返す（原因がその記事と無関係のテンプレートのこともある）。
反映まで確かめるなら **`--verify-publish`**（失敗なら終了コード 8）。

## 7. 確かめる

**API が 200 を返すことと、ページが正しいことは別。**

```bash
curl -s -o /dev/null -w "%{http_code}\n" http://<host>/            # 出ているか
curl -s http://<host>/ | grep -c '</mt:'                           # タグの閉じ漏れが無いか
```

`</mt:` が本文に出ていたら、ブロックタグでないものをブロックとして書いている。

**API で確かめたことは、編集者の画面で確かめたことにならない。** 編集者が管理画面で使うなら、
次の4つは API の応答からは見えない（実際に、構築が通ったあとマニュアルを撮る段で全部出た）。

| 確かめること | 外れていると | 直し方 |
| --- | --- | --- |
| 編集者が管理画面にサインインできる | 「サインインできませんでした」 | システム権限に `sign_in_cms` |
| 担当のコンテンツタイプがメニューに出る | 権限は正しいのに出ない | MT の再起動（メニューは起動時に作られる） |
| 新しい記事の書式が想定どおり | ブロックエディタのつもりが「なし」 | サイトの `convertParas` |
| 使う項目が編集画面に出ている | 概要・カスタムフィールドが隠れている | 利用者ごとの表示オプション（API では変えられない。手順書に書く） |

## 壊さない

- `sites delete` / `sites logs reset` / `sites refresh-templates` は
  **`--confirm <サイト名>` が要る**。IDの取り違えでサイトごと消えるのを防ぐ仕掛け
- **`mtb sites refresh-templates` は再構築ではなく初期化。** テンプレートが消える
- 削除の前に対象を `get` して、何を消すのか相手に見せる
- 空の更新（`--entry '{}'`）を再公開目的で送らない。関連が外れて URL が変わる

## この流れが扱わないもの

構築の話ではないので、必要になったら別に考える。API の 220 コマンドのうち
70 ほどがここに入る。

| | 中身 |
| --- | --- |
| **公開後に外から来るもの** | コメント（`sites comments`）と検索（`mtb search`）。**コメントを使うなら、サイトの `allowCommentsDefault` を立ててからページを作る**——page リソースに `allowComments` が無いため、後から個別には付けられない |
| **運用・観測** | アクセス統計（`sites stats`・15コマンド）とログ（`sites logs`） |
| **移行・バックアップ** | `sites backup`、`sites entries export/import`、`users export/import` |
| **環境の前提** | `plugins list`（何が使えるか）、`endpoints list`（このMTが持つ機能）、`text-filters list`（`format` の語彙） |
| **利用者の日常** | パスワード回復、ロック解除、グループ（`groups`・14コマンド） |
| **タグ** | 先に作るものではなく、コンテンツから生まれる。API では改名と削除だけ |

## 詰まったら

```bash
mtb llm | head -80          # 鉄則と失敗モード
mtb endpoints list          # そのMTが実際に持つ機能（版によって無いものがある）
mtb <コマンド> --help        # パラメータと公式リファレンスへのリンク
```
