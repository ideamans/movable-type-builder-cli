# テンプレートの種類とアーカイブ

## テンプレートの種類

| 種類 | API の `type` | 役割 |
| --- | --- | --- |
| インデックステンプレート | `index` | サイト全体に関わる単独ファイル。メインページ、CSS、RSS、JavaScript など。`outputFile` を持つ |
| アーカイブテンプレート | `individual` `page` `archive` | 記事・ページ・一覧の出力。**アーカイブマッピングが要る** |
| コンテンツタイプテンプレート | `ct` `ct_archive` | コンテンツデータの出力。作成時に `contentType` の指定が要る |
| テンプレートモジュール | `custom` | 他から `mt:Include` で読み込む断片。ヘッダ・フッタの共通化に |
| ウィジェットテンプレート | `widget` | サイドバーなどの部品 |
| ウィジェットセット | `widgetset` | ウィジェットをまとめたもの |
| ダッシュボードウィジェット | — | 管理画面のダッシュボードに出す |
| システムテンプレート | — | 検索・コメントなどMTが使うもの。**削除と名前変更はできない**（編集と初期化は可能） |

テンプレートの作成には**「テンプレートの管理」権限**（サイトまたはシステム）が要る。

## アーカイブマッピング

**どのアーカイブテンプレートで、どの種類を、どのパスに出すか**の対応。
インデックステンプレート以外は、これが無いと何も出力されない。

```bash
mtb sites templates templatemaps list 4 <template-id>
mtb sites templates templatemaps create 4 <template-id> --templatemap '{"archiveType":"Individual"}'
```

**選べるアーカイブ種別は、テンプレートの種類で決まっている。**
噛み合っていないと `409 Invalid archive type` になる。

| テンプレートの種類 | 選べるアーカイブ種別 |
| --- | --- |
| 記事アーカイブ (`individual`) | `Individual` のみ |
| 記事リストアーカイブ (`archive`) | `Daily` `Weekly` `Monthly` `Yearly` `Author` `Author-Daily` `Author-Weekly` `Author-Monthly` `Author-Yearly` `Category` `Category-Daily` `Category-Weekly` `Category-Monthly` `Category-Yearly` |
| ウェブページアーカイブ (`page`) | `Page` のみ |
| コンテンツタイプアーカイブ (`ct`) | `ContentType` |
| コンテンツタイプリストアーカイブ (`ct_archive`) | `ContentType-Daily` `ContentType-Weekly` `ContentType-Monthly` `ContentType-Yearly` `ContentType-Author` `ContentType-Author-Daily` `ContentType-Author-Weekly` `ContentType-Author-Monthly` `ContentType-Author-Yearly` `ContentType-Category` `ContentType-Category-Daily` `ContentType-Category-Weekly` `ContentType-Category-Monthly` `ContentType-Category-Yearly` |

<!-- archives:begin -->
全 31 種: `Author-Daily` `Author-Monthly` `Author-Weekly` `Author-Yearly` `Author` `Category-Daily` `Category-Monthly` `Category-Weekly` `Category-Yearly` `Category` `ContentType-Author-Daily` `ContentType-Author-Monthly` `ContentType-Author-Weekly` `ContentType-Author-Yearly` `ContentType-Author` `ContentType-Category-Daily` `ContentType-Category-Monthly` `ContentType-Category-Weekly` `ContentType-Category-Yearly` `ContentType-Category` `ContentType-Daily` `ContentType-Monthly` `ContentType-Weekly` `ContentType-Yearly` `ContentType` `Daily` `Individual` `Monthly` `Page` `Weekly` `Yearly`
<!-- archives:end -->

個別記事のテンプレートに `Yearly` を付けようとして 409 になるのは、この表に無いから。

## パス

マッピングの [パス] が、出力先とファイル名の形を決める。
MT が一般的な例をあらかじめ登録しているが、自由に変えられる。**テンプレートタグも書ける。**

## 優先するアーカイブマッピング

同じ種類のマッピングが複数あるとき、**どれを「正」の URL とするか**を選べる。

`MTEntryPermalink` / `MTEntryLink` / `MTArchiveLink` が返す URL は、
**優先指定されたマッピングのもの**になる。

優先でないマッピングでもページ自体は出力される。その URL を出したいときは、
マッピングの [パス] の内容をテンプレートタグで自分で組み立てる。

**同じ種類のマッピングを増やすときは、どれを優先にするか必ず決めること。**
決めないと、リンク先が意図しない方に向く。

## 出力されないときの確かめ方

1. そのテンプレートに**マッピングがあるか**（`templatemaps list`）
2. マッピングの種類が**テンプレートの種類と噛み合っているか**（上の表）
3. **再構築したか**（テンプレートを保存しただけでは出ない）
4. インデックステンプレートなら `outputFile` が入っているか
