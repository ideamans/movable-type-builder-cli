---
name: mt-template-tags
description: Write and debug Movable Type templates — look up the right MT template tag and its modifiers, apply global modifiers, and get the block/function distinction and context nesting right. Covers all 614 current tags and 48 global modifiers as bundled reference files. Use when writing or fixing MT template code, when a template outputs nothing or leaks raw tags, or when someone asks what an MT tag or modifier does.
license: MIT
compatibility: Reference only — needs no binary. Pair it with the mtb CLI (mtb-usage skill) to read and write the templates on a live Movable Type.
allowed-tools: Bash(mtb:*) Read Write Grep
---

# mt-template-tags

Movable Type のテンプレート言語を書く。**タグの仕様は暗記せず、同梱の
リファレンスを引く。** 614 のタグと 48 のグローバルモディファイアが入っている。

## 引き方

1. **`references/tags-index.md`** — タグ名と1行の要約から、載っているファイルを引く
2. **`references/tags-<群>.md`** — 要約・属性の書き方・マニュアルのページへのリンク
3. **`references/modifiers.md`** — グローバルモディファイア48件の要約とリンク

**参照ファイルの説明は1行の要約だけ。** 属性の意味、値の取り方、使用例、注意点は
リンク先のマニュアル（movabletype.jp）にある。**タグを使う前に、迷ったらそのページを
読むこと**（WebFetch など）。要約だけから属性の挙動を推測しない。
4. **`references/tags-deprecated.md`** — 非推奨82件。**新しく書くときは使わない**

群は用途で分かれている。

| ファイル | 中身 |
| --- | --- |
| `tags-entry.md` | 記事 |
| `tags-page.md` | ウェブページ |
| `tags-content.md` | コンテンツタイプ（MT7以降） |
| `tags-category.md` / `tags-folder.md` | 分類 |
| `tags-asset.md` | 画像・ファイル |
| `tags-tag.md` | タグ（ラベル） |
| `tags-site.md` | サイト属性・サイト横断 |
| `tags-archive.md` | アーカイブと日付 |
| `tags-control.md` | 条件分岐・変数・インクルード |
| `tags-pagination.md` | ページ送り |
| `tags-user.md` / `tags-comment.md` / `tags-search.md` | 投稿者・フィードバック・検索 |
| `tags-customfield.md` / `tags-format.md` / `tags-system.md` | 独自項目・整形・パス |

**目的の群が分かっているなら、索引を飛ばして直接その群を開く方が速い。**

## 書き方の基本

**ファンクションタグ**は値を出す。`<$mt:～$>`。

```html
<$mt:EntryTitle$>
<$mt:EntryDate format="%Y年%m月%d日"$>
```

**ブロックタグ**は囲んで繰り返す・分岐する。`<mt:～>…</mt:～>`。

```html
<mt:Entries lastn="10">
  <h2><$mt:EntryTitle$></h2>
</mt:Entries>
```

**この2つを取り違えると、閉じタグが本文にそのまま出る。**
例えば `MTEntryCategory` はファンクションタグなので、
`<mt:EntryCategory>…</mt:EntryCategory>` と書くと `</mt:EntryCategory>` が
文字列として出力される。**種別は必ずリファレンスで確かめること**（各項目に
`— BLOCK` / `— FUNCTION` と書いてある）。

## コンテキスト

**タグは「今どのオブジェクトの中にいるか」で使えるものが変わる。**

```html
<mt:Entries>          ← ここから中は記事コンテキスト
  <$mt:EntryTitle$>   ← 使える
</mt:Entries>
<$mt:EntryTitle$>     ← 記事アーカイブ以外では空になる
```

出力が空になったら、まず**そのタグが要求するコンテキストの中にいるか**を疑う。
記事一覧を出したいなら `MTEntries` で囲む、カテゴリの中の記事なら
`MTCategories` の中で `MTEntries` を使う、という具合。

## モディファイア

**タグ固有のモディファイア**は、そのタグの項目に列挙してある。

```html
<mt:Entries lastn="10" category="お知らせ" sort_by="authored_on" sort_order="descend">
```

**グローバルモディファイア**はどのファンクションタグにも付く。`modifiers.md` 参照。

```html
<$mt:EntryTitle trim_to="20+…" encode_html="1"$>
```

**複数指定すると書いた順に左から適用される。** よく使うもの:

| 用途 | モディファイア |
| --- | --- |
| HTML として安全に出す | `encode_html="1"` |
| 文字数で切る | `trim_to="20+…"` |
| タグを剥がす | `remove_html="1"` |
| 空のときの既定値 | `_default="未設定"` |
| URL に埋める | `encode_url="1"` |
| 変数に入れる | `setvar="foo"` |

## 条件分岐と変数

```html
<mt:If tag="EntryTitle">…</mt:If>              タグの出力が真なら
<mt:If name="foo" eq="bar">…</mt:If>           変数の比較
<mt:Unless tag="EntryTags">…</mt:Unless>       否定
<mt:SetVar name="foo" value="1">
<$mt:Var name="foo"$>
<$mt:Include module="ヘッダー"$>
```

カスタムフィールドは、作成時に決めた**テンプレートタグ名**で引ける
（`tag` に `PageLead` と付けたなら `<$mt:PageLead$>`）。

## 詰まったときの見方

| 症状 | 原因 |
| --- | --- |
| 出力に `</mt:～>` が混ざる | ファンクションタグをブロックとして書いている |
| 何も出ない | コンテキストの外にいる／条件が偽 |
| 一覧が10件で切れる | `lastn` / `limit` を指定していない |
| 日付の書式が英語 | `format` を指定していない（既定はサイトの言語設定） |
| 変更が反映されない | テンプレートを保存しただけ。**再構築が要る** |

再構築は `mtb sites templates publish`（index テンプレート）か、
記事・ページの保存し直し。詳しくは `mt-cms-design` スキル。

## 実物を読む

テーマのテンプレートを読むのが一番速い。

```bash
mtb sites templates list <site> --fields id,name,type,outputFile --limit 200
mtb sites templates get <site> <id> --fields text --format raw
```
