# 構築レシピ

「こういうサイトを作って」に対して、どの順で何を作るか。
**順番に意味がある**（後のものが前のものの ID を要求する）。

## 基本の順番

```
1. サイト        … ウェブサイト（＋必要ならブログ）
2. 分類          … フォルダ／カテゴリ／カテゴリセット
3. 構造          … コンテンツタイプとフィールド、カスタムフィールド
4. ユーザー       … ロールの付与
5. コンテンツ      … ページ・記事・コンテンツデータ
6. 見た目         … テンプレートと CSS
7. 公開          … テンプレートの publish と、対象の保存
```

**分類より先にコンテンツを作らない。** 後から付け替えると URL が変わる。

## レシピ1: コーポレートサイト＋ブログ

会社サイト（固定ページ中心）の下に、ブログをぶら下げる形。

```bash
# 1. サイト
mtb sites update 1 --website '{"name":"◯◯株式会社","description":"…","language":"ja","serverOffset":"9"}'
mtb sites create-blog --site-id 1 --blog '{"name":"◯◯ブログ","url":"blog/","sitePath":"/path/to/static/blog","themeId":"eigerwand","language":"ja","serverOffset":"9"}' --fields id
```

**子サイトの `url` と `sitePath` は親からの相対で書く。**
絶対 URL を渡すと連結されて `http://例/http://例/blog/` になる（エラーにならない）。

```bash
# 2. 分類
mtb sites folders create 1 --folder '{"label":"会社情報","basename":"about"}' --fields id
mtb sites categories create <blog> --category '{"label":"お知らせ","basename":"news"}' --fields id

# 3. 構造（記事に1〜2項目足すだけならカスタムフィールド）
mtb sites fields create <blog> --field '{"name":"想定読者","basename":"target_reader","systemObject":"entry","type":"select","options":"開発者,運用者"}'
# アセット系のフィールド（type は image / file / audio / video。asset は無い）
mtb sites fields create <blog> --set field.name=見出し画像 --set field.basename=hero \
  --set field.systemObject=category --set field.type=image --extract id
URL=$(mtb assets upload --site-id <blog> --file ./hero.png --path /catimg --extract url)
mtb sites categories update <blog> <cid> \
  --set-json category.customFields="[{\"basename\":\"hero\",\"value\":\"$URL\"}]"   # 値は URL

# 4. ユーザー
mtb users create --user '{"name":"editor","displayName":"編集部","email":"…","password":"…"}' --fields id
mtb sites permissions grant --site-id <blog> --user-id <uid> --role-id 2   # Editor

# 5. コンテンツ
mtb sites pages create 1 --page @about.json --fields id
mtb sites pages update 1 <pid> --page '{"folder":{"id":<fid>}}'            # フォルダは後付け
mtb sites entries create <blog> --entry @entry.json --fields id

# 6. 見た目（HTML をそのまま渡せる。JSON を手で組む必要は無い）
mtb sites templates list <blog> --fields id,name,type,outputFile
mtb sites templates update <blog> <tid> --set template.text=@styles.css
mtb assets upload --site-id <blog> --dir ./dist/theme-images --path /theme-images --replace --yes

# 7. 公開
mtb sites templates publish 1 <index-template-id>
mtb sites entries update <blog> <eid> --entry '{"excerpt":"（いまと違う値）"}'  # 値が変わると同期再構築
mtb publish entries --ids <eid> --wait --verify-publish                      # まとめて＋反映確認
```

## レシピ2: 製品カタログ（コンテンツタイプ）

項目が決まっているものは記事ではなくコンテンツタイプ。

```bash
CT=$(mtb sites content-types create 4 --content-type '{"name":"製品"}' --extract id)
mtb sites content-types fields create 4 $CT --content-field '{"label":"製品名","type":"single_line_text","required":true,"options":{"label":"製品名"}}'
mtb sites content-types fields create 4 $CT --content-field '{"label":"説明","type":"multi_line_text","options":{"label":"説明","input_format":"richtext"}}'
mtb sites content-types fields list 4 $CT --fields id,label     # ID を控える
mtb sites content-types data create 4 $CT --content-data '{"label":"製品A","data":[{"id":1,"data":"製品A"},{"id":2,"data":"<p>説明</p>"}]}'
```

出力するには `type: ct` のテンプレートと `ContentType` のアーカイブマッピングが要る。

```bash
mtb sites templates create 4 --template '{"name":"製品詳細","type":"ct","text":"…","contentType":{"id":'$CT'}}' --fields id
mtb sites templates templatemaps create 4 <tid> --templatemap '{"archiveType":"ContentType"}'
```

## レシピ3: 既存テーマを土台に見た目だけ変える

テーマのテンプレートを読んで、必要なものだけ差し替えるのが速い。

```bash
mtb sites templates list 4 --fields id,name,type,outputFile --limit 200
mtb sites templates get 4 <id> --fields text --format raw      # 中身を読む
mtb sites templates update 4 <id> --template @new.json
```

- CSS は `スタイルシート`（`index` / `styles.css`）
- ページ全体の枠は `レイアウト`、ヘッダは `バナーヘッダー`、ナビは `ナビゲーション`
- 一覧の1件は `記事一覧アイテム`、記事ページは `記事`

**テーマのクラス名に合わせて CSS を書くと、テンプレートを書き換えずに済む。**

## 公開（再構築）の落とし穴

| やりたいこと | 方法 |
| --- | --- |
| index テンプレート（トップ・CSS）を反映 | `mtb sites templates publish <site> <tid>` — 同期 |
| 記事・ページを反映 | **中身が変わる更新を送る** — 同期。同じ値では再構築されない |
| まとめて再構築 | `mtb publish entries --ids … --wait` — 1回では終わらない |
| 反映まで確かめる | どれかに `--verify-publish` を足す — 失敗なら終了コード 8 |

**`publish entries` の `{"status":"Rebuilding"}` は「受け付けた」だけ。**
残りは `restIds` に入るので、`--wait` を付けないと1回で終わって
成功したように見える。

**`--entry '{}'` のような空の更新を再公開目的で送らないこと。**
フォルダやカテゴリの関連が外れて URL が変わることがある。

## 壊さないための約束

- `sites delete` / `sites logs reset` / `sites refresh-templates` は
  **`--confirm <サイト名>` が要る**。IDの取り違えでサイトが飛ぶのを防ぐため
- 一覧は既定10件。**全件は `--all`**、必要な属性だけ `--fields`
- 迷ったら `--dry-run`
