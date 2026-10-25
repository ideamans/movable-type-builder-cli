---
name: mtb-theme
description: Develop a Movable Type theme as files in Git — theme.yaml, templates and static files — deploy it to the server and apply it with the mtb CLI to build or rebuild a site. Use when someone wants MT templates under version control, asks to build a site from a theme, to package a design as a reusable theme, or to iterate on an MT design with a normal edit-deploy-verify loop.
license: MIT
compatibility: Requires the `mtb` binary on PATH (run the mtb-install skill if missing), a configured profile with `manage_themes`, **and a way to put files on the server** (rsync/scp, or the admin screen) — the Data API cannot install a theme. Applying a theme replaces every template on the target site.
---

# テーマをファイルとして開発する

テンプレートを Git で持ち、**ファイルからサイトを起こす**進め方。
サイトを組んでから `export-theme` で吸い出すのとは向きが逆で、
**テーマが正、サイトはその結果**になる。Git の履歴がそのまま設計の履歴になる。

## 0. この進め方が使えるか確かめる

**Data API はテーマを入れられない。** `/themes` にあるのは取得・削除・適用と
書き出しだけで、install も upload も無い（実機の 266 エンドポイントで確認）。
**ファイルをサーバに置く手段が別に要る。**

| 置く手段 | |
| --- | --- |
| rsync / scp | これがあるならこの進め方が一番きれい |
| 管理画面のテーマインストール | zip を人が上げる。CI には向かない |
| mtb-pro 拡張の `mtb themes files upload <id> --dir <dir>` | MT 側に MtbPro と `MtbProAllowFileUpload 1` が要る。rsync の代わりになる |
| **どちらも無い** | **この進め方は採れない。** `mtb-usage` でテンプレートを直接作るか、`sites templates pull/push` で往復させる |

相手に「サーバのファイルシステムに置けますか」と聞くこと。答えが No なら
別の道を勧める。

## 1. 置き場所

```
theme-v1/
  theme.yaml                    何を持つテーマかの定義
  templates/*.mtml              テンプレートの本文
  blog_static/<フォルダ>/...     静的ファイル（CSS・画像・フォント）
```

**静的ファイルは `blog_static/` の下に置く。** `static/` のような素の名前で
置いても運ばれず、**エラーも警告も出ないまま 404 になる**。

## 2. theme.yaml

実機で動いた最小の形。

```yaml
id: mtb_v1                        # ディレクトリ名と合わせる
name: mtb v1
version: '1.0'
class: website                    # blog なら blog
elements:
  template_set:
    importer: template_set
    data:
      base_path: templates        # .mtml の置き場所
      label: mtb v1
      templates:
        index:
          main_index:             # ← templates/main_index.mtml
            label: メインページ
            outfile: index.html
            rebuild_me: 1
        archive:
          entry_archive:
            label: 記事アーカイブ
            mappings:             # ここに書けばアーカイブマッピングも入る
              individual:
                archive_type: Individual
                preferred: 1
        module:
          header:
            label: ヘッダー
  blog_static_files:
    importer: blog_static_files
    data:
      - static                    # blog_static/static/ を運ぶ
```

**`mappings` を書けばアーカイブマッピングも一緒に入る。** これを書かないと
記事の個別ページが出力されず、後から手で作ることになる。

テンプレートの中身（MTタグ）を書くときは **`mt-template-tags` スキル**を使う。

## 3. 回す

```bash
rsync -a theme-v1/ prod:/path/to/themes/mtb_v1/        # 置く（mtb の外）
mtb themes list --profile prod --fields id,label       # 見えるか
mtb sites themes apply <site> mtb_v1 --profile prod --yes --confirm "<サイト名>"
mtb sites templates publish <site> <index-tid> --profile prod
curl -s -o /dev/null -w '%{http_code}\n' http://<host>/
```

**適用はテンプレートの総入れ替え。** `--yes` に加えて `--confirm "<サイト名>"`
が要る（取り違えたサイトに当てると、そのサイトの作業が消えるため）。

## 4. 確かめる

**API が success を返すことと、ページが正しいことは別。**

```bash
curl -s http://<host>/ | grep -c '</mt:'      # タグの閉じ漏れ（0 であること）
mtb dump --profile prod --dir ./after         # 入ったものの一覧
```

`</mt:` が本文に出ていたら、ブロックタグでないものをブロックとして書いている。

## テーマが運べるもの

`sites export-theme --include` の選択肢がそのまま「テーマが持てるもの」。

| | |
| --- | --- |
| `template_set` | テンプレートとアーカイブマッピング |
| `blog_static_files` | 静的ファイル |
| `custom_fields` | カスタムフィールドの定義 |
| `default_categories` / `default_folders` / `default_category_sets` | 初期の分類 |
| `default_content_types` | コンテンツタイプの定義 |
| **`be_block` / `be_config` / `be_pref`** | **ブロックエディタのカスタムブロック** |

**カスタムブロックはテーマでしか運べない。** Data API にエンドポイントが無く、
ダンプにも出ないので、**差分にすら現れない**。使っているなら必ず含める。

記事・ページ・アセットは運べない（テーマは器であって中身ではない）。

## 詰まったら

- **静的ファイルが 404** → `blog_static/` の下に置いているか
- **記事の個別ページが出ない** → `theme.yaml` に `mappings` を書いたか
- **適用が `--confirm` を要求する** → 正しい挙動。サイト名を確かめてから打つ
- **テーマが `themes list` に出ない** → ディレクトリ名と `id` が一致しているか、
  所有者がウェブサーバから読めるか
- **テーマを直したのにサイトが変わらない** → 適用済みのサイトはテーマのファイルを
  参照していない。`sites themes apply` をやり直す（テンプレートは作り直されて ID が変わる）
- **テンプレートが名前で見つからない** → 名前は MT の言語で訳される
  （Main Index → メインページ）。`templateType` か ID で引く

**繰り返し適用すると backup テンプレートが溜まる。** 適用のたびに旧
テンプレートが `backup` 型にリネームされて残る。**一覧には出ず、API からは
削除できない**（`403 Cannot delete backup template.`）。実機で2回適用したら
53 件中 38 件が backup だった。開発中は使い捨てのサイトで回し、
本番では管理画面から掃除する。
