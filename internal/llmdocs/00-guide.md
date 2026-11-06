# mtb — AIエージェント向けリファレンス

Movable Type Data API (v7) のコマンドラインクライアント。バイナリ名は `mtb`。

## 鉄則

**1. パスワードは「Webサービスパスワード」。サインインパスワードではない。**
MT のユーザー設定画面にある別項目で、既定では未設定のことがある。ここを
取り違えると 401 になる。`mtb auth login` の失敗のほとんどがこれ。
MT 9 では**発行したときにしか表示されない**（ハッシュで保存される）。利用者が
分からないと言ったら、管理画面で削除して発行し直してもらう（人の操作。
拡張 `mtb-pro` があれば `users api-password issue`）。

**2. ベースURLは `mt-data-api.cgi` まで。`/v7` は付けても付けなくてもよい。**
サイトのトップURLを入れると 404 になる。付いていた `/vN` は取り除いて
バージョンとして解釈するので、リファレンスからコピーしたURLをそのまま貼れる。

**3. 対話できない環境では、破壊的操作に `--yes` が要る。**
DELETE は確認を挟む。端末でなければ確認できないので、`--yes` が無ければ
実行せずにエラーになる。**エージェントは意図した削除にだけ `--yes` を付けること。**

**4. 一覧の既定件数は 10 件。`--limit` を指定しないと「見つからない」ことがある。**
`--limit` の上限はエンドポイントによるが、多くは指定しないと 10 件で切れる。
全件を数えたいときは応答の `totalResults` を見る。

**5. 記事の既定の絞り込みは公開済みだけではない。`--status` で明示する。**
`Publish` / `Draft` / `Review` / `Future` / `Spam` / `Unpublish`。

**6. 送っていないパラメータは送られない。**
フラグを指定しなかったものは、既定値であっても API に送らない。MT 側の既定に
任せる設計なので、「既定値を明示的に送る」には自分でその値を書く。

**7. 迷ったら `--dry-run`。**
送るはずだったメソッド・URL・本文を表示して終わる。破壊的操作の下見に使う。

**8. 「保存できた」と「サイトに出た」は別。反映まで見るなら `--verify-publish`。**
MT は記事を保存すると同期で再構築するが、**そこで失敗しても API は 200 を返す**。
しかも原因はその記事と関係のないテンプレートのことがある（アーカイブテンプレートが
1つ壊れていると、記事の個別アーカイブごと書き出されない）。手がかりは MT の
ログにしか残らないので、探しに行かないかぎり気づけない。

```bash
mtb sites entries update 6 309 --entry '{"title":"…"}' --verify-publish
# 保存はできましたが、静的ファイルの書き出しに失敗しています（MT のログに 1 件のエラー）。
#   Error during publishing: ... Cannot find template 'main_index'
```

送信の前後でサイトのエラーログを見比べ、増えていればその内容を出して
**終了コード 8** で終わる。時刻ではなくログの ID で区切るので、サーバとこちらの
時計がずれていても取りこぼさない。リクエストが2回増えるので既定では off。
**ログが読めないとき（権限不足など）は、確かめられなかったと伝えるだけで、
成功した保存をエラーにはしない。**

**9. 全件ほしいときは `--all`。**
MT の一覧は既定10件。`--all` を付けると limit と offset を自動で動かして
全ページを取り、1つの応答に畳む。`totalResults` には**実際に集めた件数**が入る。

```bash
mtb sites entries list 4 --all --fields id,title    # 261件を3リクエストで
mtb sites entries list 4 --all --limit 50           # --limit は1ページの件数になる
```

`--all` が生えるのは limit と offset を両方持つコマンドだけ。

**10. 一覧は `--fields` で絞る。**
MT はリソースの全属性を返す。サイト2件で 5KB、記事50件なら数十KB になり、
読むだけで文脈を食い潰す。**必要な属性だけ挙げること。**

```bash
mtb sites list --fields id,name,url          # 200バイト弱で済む
mtb sites entries list --fields id,title,status --limit 50
```

**`--fields` はどのコマンドでも使える。** MT は GET にしか fields を用意して
いない（226 中 67）が、対応していないエンドポイントでは mtb が手元で絞る。
作成したもののIDを次のコマンドに渡すときはこれが効く。

```bash
mtb sites entries create --site-id 1 --entry @entry.json --fields id --compact
# => {"id":"42"}
```

**11. MT に書き込む作業の前に、`MTB_AUDIT_NOTE` に作業の説明を入れる。**
全リクエストに `X-MTB-Audit-Note` ヘッダで付き、サーバが監査ログを取っていれば
（MtbPro など）一緒に残る。何十回と叩く作業でも、あとで「どの依頼の操作か」を引ける。
mtb は `User-Agent: mtb/<版>` も名乗るので、Web サーバのログでも見分けられる。

```bash
export MTB_AUDIT_NOTE="記事の移行（stg→本番、依頼: 2026-10-25）"
```

## 認証

```bash
mtb profile add prod --base-url https://example.com/cgi-bin/mt/mt-data-api.cgi
mtb auth login --profile prod          # ユーザー名とパスワードを聞かれる
mtb auth status                        # 接続先と有効期限の確認
```

保存されるのは**セッションIDとアクセストークンだけ**で、パスワードは保存しない。
アクセストークンは1時間程度で切れるが、**セッションIDから自動で取り直す**ので
`mtb auth refresh` を明示的に叩く必要はない。

対話できない環境（CI・エージェント）では2通り。

```bash
# 1) パスワードを標準入力から渡してサインインする
pass show mt/prod | mtb auth login --username alice --password-stdin

# 2) すでに持っているアクセストークンだけで叩く（設定ファイルを一切使わない）
MTB_BASE_URL=https://example.com/cgi-bin/mt/mt-data-api.cgi \
MTB_ACCESS_TOKEN=... mtb sites entries list --site-id 1
```

`MTB_ACCESS_TOKEN` を使うときは、失効しても自動更新できない（セッションIDが
無いため）。長時間動かすなら 1) を使うこと。

**パスワードの渡し方は3つあり、残るものが違う。**

| | 残るところ |
| --- | --- |
| `--password-stdin` | どこにも残らない |
| `MTB_PASSWORD` | 同じホストの他のプロセスから見える |
| `--password` | プロセス一覧に残る。他に手が無いときだけ |

`--password-stdin` は末尾の改行だけを落とす（`echo pw \|` がそのまま使える）。
標準入力が空ならエラーで止まる——`MTB_PASSWORD` に落ちて別のパスワードで
通ってしまう方が危ないため。`--password` との併用は呼び出し方の誤り（終了コード5）。

## プロファイル（接続先の切り替え）

```bash
mtb profile list                       # 一覧。* が既定
mtb profile add stg --base-url https://stg.example.com/cgi-bin/mt/mt-data-api.cgi
mtb profile use stg                    # 既定を切り替える
mtb profile set --site-id 3            # 既定サイトIDを決める（--site-id を省ける）
mtb profile show                       # フラグ・環境変数・設定を重ねた結果
mtb sites entries list --profile stg   # その場だけ切り替える
```

`mtb profile add` は名前を省ける。省くと、いまあるプロファイルを一覧して
から名前と URL を聞く。**対話できない場面では聞かずにエラーにする**ので、
エージェントは `mtb profile add <名前> --base-url <URL>` の形で渡すこと。

優先順位は **フラグ → 環境変数 → プロファイル → 既定値**。
どのサーバを叩こうとしているか分からなくなったら `mtb profile show`。

設定は `~/.config/mtb/config.json`、認証情報は同じディレクトリの
`credentials.json`（0600）。`MTB_CONFIG_DIR` で移せる。

**カレントディレクトリの `.env` を読む。** `MTB_*` をプロジェクトごとに
置いておける。プロセスの環境変数が勝ち、`.env` は「打っていないときの既定」。
`export` 付きの行と `#` 行のコメント、値を囲む引用符に対応。行末コメントは
扱わない（値の一部として渡す）。

**作業ディレクトリによって接続先が変わりうる**ということでもある。叩く前に
`mtb profile show` を見ること——読み込んだ `.env` の場所もそこに出る。
確実に固定したいなら `--profile` か `--base-url` を明示する（どちらも .env より強い）。

人に向けた文言（メッセージ・ヘルプ・エラー）は日本語と英語を切り替える。
`--lang ja` / `--lang en`、または `MTB_LANG`。指定が無ければ `LC_ALL` →
`LC_MESSAGES` → `LANG` の順に見て、いずれも無ければ英語。エンドポイントの
説明は公式の OpenAPI 定義そのままなので、どちらの言語でも英語で出る
（mtb が補った注記だけが切り替わる）。**この埋め込みリファレンスは
言語によらず日本語**。出力を機械で判定するなら `--lang` を明示すること。

## リソースの組み立てと取り出し

**JSON を手で組み立てなくてよい。jq も要らない。**

```bash
mtb sites templates create --site-id 1 \
  --set template.name=メインページ \
  --set template.type=index \
  --set template.text=@index.html

mtb sites templates get 1 42 --extract text --out index.html
```

- `--set <フラグ>.<項目>=<値>` — 値は**文字列**。`@ファイル名` と `-` で読み込む
- `--set-json` — 値を JSON として解釈する（数値・真偽値・入れ子）
- `--template @base.json` と併用すると、ファイルを土台にして上から重ねる
- 指定できる項目名は `--help` に出る（`(object: name, outputFile, text, type, ...)`）
- **HTML を \uXXXX に逃がさない。** `<$mt:BlogName$>` はそのままの形で送られる

`--extract <項目>` は応答から1つを生の値で取り出す（`jq -r` と同じ）。

- 文字列はそのまま、それ以外は JSON
- `items.0.name` のように配列の添字も辿れる
- **`--out` と組み合わせるとバイトそのまま**。端末に出すときだけ末尾に改行が付く
- 項目名を間違えると、その応答にある項目名を並べて教える
- `--format raw` との併用は呼び出し方の誤り（raw は無加工という意味なので）

## 環境どうしの差分を取る

```bash
mtb dump --profile stg  --dir ./dump/stg
mtb dump --profile prod --dir ./dump/prod
diff -ru ./dump/prod ./dump/stg
```

`<dir>/system/*.yaml`（ロール・利用者・グループ）と
`<dir>/sites/<サイト名>/*.yaml`（テンプレート・カテゴリ・フォルダ・
カテゴリセット・コンテンツタイプ・カスタムフィールド・ウィジェット・
整形済みテキスト・記事・ページ・アセット）に分けて書き出す。

- **ID・時刻・作成者・permalink は書き出さない。** 環境ごとに違うので、
  残すと中身が同じでも全行が差分になる
- 並べ替えと突き合わせは `basename` / `name` / `label` の**自然キー**。
  **ID では突き合わせない**——ステージングの記事 42 と本番の記事 42 は別物
- `uniqueID`（コンテンツタイプとコンテンツフィールドだけが持つ40桁のhex）は
  鍵に使わずデータとして残す。テーマ経由で配ったものは一致するが、
  両方で別々に作れば別の値になるため
- 他のオブジェクトへの参照は自然キーに畳む（記事の categories は basename の配列）
- 階層のあるカテゴリ・フォルダは親を根からの道で持つ
- **複数行はリテラルブロック。** 行末の空白やタブを含むものは引用に落とすので、
  書き出しても中身は変わらない
- アセットはサイトルートからの相対パス。中身も見るなら `--asset-hash`（MD5）
- 統計・ログ・コメントは対象外
- サイトの `url` `host` `sitePath` は残る（環境で違って当然のもの。
  どちらを比べているかの確認になる）

**同じ入力なら必ず同じバイト列が出る**ので、diff に出たものは本当の差分。

`system/` にはサーバの素性も出る（`plugins` `text-filters` `endpoints` `server`）。
**運べないが、違うと分かるだけで価値がある**——受け側に無いプラグインや書式に
依存したものを運んでも動かない。

**`mt-config.cgi` は API から読めない。** `EmailAddressMain`（未設定だとメール
送信が 500）、`DeniedAssetFileExtensions`（`.js` `.html` が上げられない理由）は
差分に出ないので、移行の前に人が揃える。

## テンプレートの一括出し入れ

```bash
mtb sites templates pull 4 --dir ./templates
mtb sites templates push 4 --dir ./templates
```

**これはエンドポイントではない。** Data API に一括の取り出し／書き戻しは
無いので、mtb が list と update を回している（`--all` や `--dir` と同じ立場）。

- 本文は1テンプレート1ファイル（`.mtml`）、種別・出力ファイル名・buildType・
  linkToFile は `mtb-templates.json` が持つ
- **対応は ID で取る。** 名前やファイル名を変えても切れない
- **変わっていないものは送らない。** 本文だけでなく名前・出力ファイル名・
  buildType・linkToFile の変更も拾う
- **MT にあって手元に無いものには触れない。** 名前を報告するだけ。消すなら
  `sites templates delete` を明示的に使う
- マニフェストに `id` の無い項目を足すと新規作成し、ID を書き戻す
- **種別は新規作成のときだけ送る。** 既存テンプレートの種別は変えられない
  （システムテンプレートで通らない）
- マニフェストの `siteId` と対象が違えば止まる（取り違えて上書きしないため）
- **アーカイブマッピングも一緒に往復する。** マニフェストの `archiveMaps` に入る。
  マッピングは URL の設計そのものなので、これが無いと pull したディレクトリから
  出力先（`blog/%-b.html` など）を戻せない。**本文が変わっていなくても
  マッピングだけ変わっていれば送る**。持てるのは individual / page / archive /
  ct / ct_archive の種別だけで、それ以外には問い合わせない
- `pull` は手元を上書きする。`push` は書き換える件数を出して確認を求める（`--yes`）

サーバのファイルシステムに触れるなら `linkToFile` の方が MT 公式の仕組みで、
pull/push は要らない（`--set template.linkToFile=/path/to/file.mtml`）。

## テーマをファイルとして持つ

**テーマを外から MT に入れる経路は Data API に無い**（`/themes` は GET と
DELETE と apply だけ）。サーバのファイルシステムに触れるなら、テーマを
Git のファイルとして持ち、配置は rsync、適用と確認を mtb が受け持つのが
一番きれい。**サイトから export するのではなく、ファイルからサイトを起こす。**

**拡張 `mtb-pro` があれば rsync も要らない**（`mtb themes files upload <id> --dir <dir>`）。
MT 側で `MtbProAllowFileUpload 1` が有効になっていることが条件（→「拡張」）。

```
theme.yaml / templates/*.mtml / blog_static/<フォルダ>/...
```

- **静的ファイルは `blog_static/` の下**。素の `static/` に置いても運ばれず、
  **エラーも出ないまま 404 になる**
- `theme.yaml` の `mappings` でアーカイブマッピングも一緒に入る
- 適用は `sites themes apply`。**テンプレートの総入れ替え**なので `--confirm` が要る

## テーマは最初に当てる

**テーマを適用すると、そのサイトのテンプレートは全件 `backup` 型にリネームされ、
アーカイブマッピングは削除、出力ファイル名も消える**（`MT::Theme::TemplateSet::apply`）。
テンプレートを作り込んだ後に当てると URL 設計ごとやり直しになる。
マニュアルの適用ページに警告は無い。

- **当て直すとテンプレートは作り直され、ID が変わる。** ID を控えていたら引き直す
- **テーマのファイルを直しただけでは、適用済みのサイトは変わらない。** 当て直して初めて効く
- **テンプレートを名前で探さない。** テーマの `name: Main Index` は MT の言語で訳されて
  「メインページ」になる。`sites templates list` の `templateType`（index・archive など）か
  ID で引く（Data API のテンプレートには `identifier` が出ない）

`mtb sites themes apply` は `--yes` に加えて `--confirm <サイト名>` を要求する。

## 本文の書式（ここが一番危ない）

**`get` が返す `body` は変換後の HTML で、そのまま書き戻してはいけない。**

```bash
mtb sites entries get 4 42 --extract body                     # 変換後（表示用）
mtb sites entries get 4 42 --no-text-filter 1 --extract body  # 原稿（編集用）
```

`format` が `markdown` や `block_editor` の記事で、変換後の HTML を書き戻すと
**原稿が HTML に置き換わる**。`format` は変わらないので応答も見た目も正常で、
次に編集した人が壊れた本文に出会うまで誰も気づかない。**本文を編集するために
取り出すなら必ず `--no-text-filter 1` を付ける。**

本文に触れない部分更新（タイトルだけ、ステータスだけ）は安全。原稿はそのまま残る。

`format` に入れる値は **`mtb text-filters list`** が教える。

| key | 中身 |
| --- | --- |
| `0` | 変換しない（HTML をそのまま） |
| `__default__` | 改行を `<br>` に |
| `richtext` | リッチテキスト |
| `markdown` / `gfm` / `commonmark` / `markdown_with_smartypants` | 各種 Markdown |
| `block_editor` | ブロックエディタ |

**ブロックエディタも API から書ける。** 原稿は `<!-- mt-beb ... -->` で
ブロックを区切った HTML。

```html
<!-- mt-beb t="core-text" --><p>段落</p><!-- /mt-beb -->
<!-- mt-beb t="core-html" --><figure><img src="..."></figure><!-- /mt-beb -->
```

区切りが無い素の HTML を `block_editor` で送っても通り、`core-html` の
単一ブロックとして扱われる。公開の出力は同じだが、管理画面で開くと
1つの生HTMLブロックになる。**構造を持たせたいなら区切りを自分で書く。**

**カスタムブロックも同じ形で書ける。** `t=` に識別子を入れるだけ
（`<!-- mt-beb t="mtbnotice" -->…<!-- /mt-beb -->`）。原稿はそのまま保存され、
公開時には区切りが外れて中身だけが出る。**実機で確認済み。**

**ただしカスタムブロックの定義そのものは Data API で扱えない。**
MT では `be_block` というオブジェクトだが、**エンドポイントが1つも無い**
（実機の 266 個を確認）。作るのも運ぶのも API の外。

**運ぶ手段はテーマだけ。** `sites export-theme --include be_block,be_config`
でテーマに入り（`block_editor_blocks/<識別子>.json` として書き出される）、
適用先に持っていける。カスタムブロックを使うサイトの移行は、この経路が要る。

## カスタムフィールドと複合編集

```bash
# ① カテゴリに画像フィールドを作る
mtb sites fields create 4 --set field.name=見出し画像 --set field.basename=hero \
  --set field.systemObject=category --set field.type=image --extract id

# ② 画像を上げて URL を得る
URL=$(mtb assets upload --site-id 4 --file ./hero.png --path /catimg --extract url)

# ③ カテゴリに設定する
mtb sites categories update 4 18 \
  --set-json category.customFields="[{\"basename\":\"hero\",\"value\":\"$URL\"}]"
```

- **アセット系の値は URL。ID ではない。** 検証されないので、間違えても 200 が返る
- **`customFields` は指定したものだけでなく、そのオブジェクトの全フィールドが
  配列で返る**（未設定は `value: null`）。`--extract customFields.0.value` は
  別のフィールドを見ることになるので、basename で探すこと
- **API 経由ではアセットとの関連（objectasset）が作られない。** テンプレートでの
  出力はできるが、「このアセットを使っているもの」には出ない
- 部分更新でも `customFields` は配列ごと送る（送った要素だけが更新される）

## 並べ替えと親子構造

```bash
# 並べ替え：取り出すときの並び順に注意
mtb sites categories list 4 --sort-by user_custom --sort-order ascend --extract items --out c.json
# c.json を並べ替えて
mtb sites categories permutate 4 --categories @c.json
```

**`list` の既定は `--sort-order descend`。** そのまま並べ替えて送ると逆順になる。
`user_custom` の昇順が「管理画面で見える順」。

親子は `parent` に親の ID を入れるだけ。`--set`（文字列）でも通る。
`0` で根に戻す。

```bash
mtb sites categories update 4 19 --set category.parent=20   # 20 の子にする
mtb sites categories update 4 19 --set category.parent=0    # 根に戻す
```

## ファイルのアップロード

```bash
mtb assets upload --site-id 1 --file ./logo.png --path /theme-images
mtb assets upload --site-id 1 --dir ./dist/theme-images --path /theme-images
```

**`mtb sites assets upload` は非推奨。** 同じことをする `mtb assets upload`
（`POST /assets/upload`、site_id は本文で渡す）を使う。

`--dir` はディレクトリを丸ごと送る。Data API に一括アップロードは無いので
mtb が1ファイルずつ送っている（`--all` と同じ立場）。

- 下の階層は `--path` の下にそのまま写す（`icons/menu.svg` → `<path>/icons`）
- 隠しファイル（`.DS_Store` `.git/*` など）は送らない
- **途中で失敗したらそこで止まる。** 何件目で止まったかは標準エラーに出る
- **同名のファイルは上書きできない。** MT は 409 を返す
  （`A file named '...' already exists.`）。**Data API に上書きの手段は無い**。
  入れ替えるなら `--replace` を使う（送り先にある同名のアセットを消してから
  送る）。消すのは破壊的なので `--yes` が要る。**送り先まで一致したものだけ**を
  消すので、別フォルダの同名ファイルは巻き込まない。**消して上げ直すので ID が
  変わる。** ID を保ちたいなら拡張 `mtb-pro` の `mtb sites assets replace`
- **JS と HTML はアセットにできない。** MT の `DeniedAssetFileExtensions` が
  既定で拒否する（`js` `html` `htm` `shtml` ほか）。`'.js' is not allowed to
  upload by system settings.` が出たらこれ。**MT の作法ではスタイルシートと
  JavaScript はインデックステンプレート**なので、`--set template.type=index
  --set template.outputFile=theme.js` で置く。画像・CSS・SVG・フォント・webp は
  アセットで上げられる
- 応答は `--all` と同じ `{"items":[...],"totalResults":N}` にまとめる
- `--file` との併用は呼び出し方の誤り（終了コード5）

テーマの静的ファイルは**サイトルート直下のトップレベルフォルダ**に置く決まりで
（`theme-images` `theme-css` `theme-js` が慣例）、拡張子は
jpg/jpeg/gif/png/js/css/ico/swf/flv に限られる。

**ただし `--include blog_static_files` だけでは何も入らない。** どのフォルダを
含めるかはサイトの `theme_export_settings` に保存されている必要があり、
**Data API からは設定できない**（管理画面のテーマエクスポート画面で
フォルダ名を1行ずつ入れて保存する）。保存が無くても `status` は `success` を
返すので、空のテーマができたことに気づけない。**実機で確認済み。**
拡張 `mtb-pro` があれば `mtb sites theme-export-settings update <site> --static-directories '["css"]'`。
**テーマを適用するとこの設定は消える**ので、書き出す直前に入れる。

    mtb sites export-theme 4 --theme-id my-theme --theme-name "My Theme" \
      --include blog_static_files,template_set --overwrite-yes 1

`--include` はカンマ区切り。JSON の配列でも受け取る。

## コマンドの見つけ方

コマンド名は公式リファレンスのパスと1対1で対応する（詳細は次章）。

```
GET  /v7/sites/{site_id}/entries              → mtb sites entries list
POST /v7/sites/{site_id}/entries              → mtb sites entries create
GET  /v7/sites/{site_id}/entries/{entry_id}   → mtb sites entries get
```

**パスが分かっているなら、コマンド名を探すより `--help` を見る方が速い。**

```bash
mtb sites entries --help          # ぶら下がっている操作の一覧
mtb sites entries list --help     # パラメータと公式リファレンスへのリンク
```

複数形のグループには単数形の別名がある（`mtb sites entry list` も動く）。

## 終了コード

エージェントやスクリプトが「何が起きたか」で分岐できるように分けてある。

| コード | 意味 | 次の手 |
| --- | --- | --- |
| 0 | 成功 | |
| 1 | 分類できないエラー | 標準エラーを読む |
| 2 | 未認証・認証切れ（401） | `mtb auth login` をやり直す |
| 3 | 権限不足（403） | **リトライしても直らない。** MT のロールを見る |
| 4 | 対象が無い（404） | IDとベースURLを確認する |
| 5 | 呼び出し方の誤り | フラグ・JSON・key=value を直す。**送信していない** |
| 6 | 接続できない・時間切れ | リトライの余地がある |
| 7 | サーバ側のエラー（5xx） | MT のログを見る |
| 8 | 保存はできたが書き出せなかった | テンプレートの誤りを直して保存し直す。`--verify-publish` / `--wait` のときだけ返る |

**5 は送信前に弾いている**ので、副作用は起きていない。安心して直して再実行できる。

## 出力

既定は整形した JSON。日本語はそのまま読める形で出す
（MT は `\uXXXX` にエスケープして返すが、mtb が戻している。
バイト列そのものが要るときだけ `--format raw`）。JSON を返さないエンドポイント（エクスポート、バックアップ、
プレビューHTML）は本文をそのまま出す。

```bash
mtb sites entries list --site-id 1 --compact          # 1行 JSON（jq に流す）
mtb sites entries list --site-id 1 --format raw       # 無加工
mtb sites backup 1 --out backup.tar.gz                # ファイルに書き出す
```

`--out` を付けたときだけ、本文は標準出力に出ない（案内は標準エラー）。

## パラメータの渡し方

| 種類 | 書き方 |
| --- | --- |
| クエリ・本文の単純な値 | `--limit 50` `--status Publish` |
| リソース本体（JSON） | `--entry '{"title":"..."}'` |
| JSON をファイルから | `--entry @entry.json` |
| JSON を標準入力から | `--entry -` |
| ファイルの添付 | `--file ./photo.jpg` |
| 定義に無いクエリ | `--query key=value` |
| 定義に無い本文 | `--form key=value` |
| 定義に無い本文をファイルから | `--form key=@file.json` |

JSON は送る前に構文検査する。壊れていれば MT に届く前にエラーになる。

**`@ファイル名` が効くのは JSON を取る項目とファイル添付だけ。** 文字列の項目に
渡すと値そのものとして送られる（MT に `@SITE_LOGO` のようなタグ記法があるため、
一律には展開できない）。取り違えると mtb が警告する。

パスパラメータはフラグでも位置引数でも渡せる。**位置引数はパスの並び順**。

```bash
mtb sites entries get --site-id 1 --entry-id 42
mtb sites entries get 1 42                    # 同じ
```

フラグ名は API のパラメータ名をケバブケースにしたもの（`sortBy` → `--sort-by`）。
`--help` には `(API: sortBy)` と併記されるので、リファレンスと突き合わせられる。

## 代表的なワークフロー

### サイトを作る（ウェブサイトとブログ）

MT のサイトは**ウェブサイト（website）とその下のブログ（blog）**の2階層。
コーポレートサイトの下にブログをぶら下げる、という構成がこれにあたる。

```bash
mtb sites create-blog --site-id 1 --blog @blog.json    # 1 の下にブログを作る
```

ここに2つ罠がある。

**子サイトの `url` と `sitePath` は親からの相対で書く。**
絶対URLを渡すと親のURLに連結され、`http://example.com/http://example.com/blog/`
になる。**エラーにならず、後で気づく**ので注意。正しくは `"url": "blog/"`。

**更新の本文キーはサイトの種別で変わる。**
`website` を更新するなら `--website`、`blog` なら `--blog`。取り違えると
`400 A resource "blog" is required.` になる。

```bash
mtb sites update 1 --website @corp.json    # class: website
mtb sites update 2 --blog @blog.json       # class: blog
```

種別は `mtb sites get 2 --fields id,name,class` で確認できる。

### 大量データを扱う

`limit` に上限は無い（261件のところに `--limit 1000` を投げても全部返る）。
`offset` も素直に効き、`--sort-by` を指定しておけばページをまたいでも
重複や取りこぼしは出ない（261件を50件ずつ6ページで確認済み）。

**ただし `/search`（`mtb search`）の `totalResults` は当てにならない。**
limit で打ち切られると **0** を返す。limit が実件数以上のときだけ正しい。
件数を知りたいなら `--all` を使う（mtb が集めた件数を入れ直す）。

| | limit で打ち切られたとき |
| --- | --- |
| `sites entries list` などの一覧 | `totalResults` は全体の件数（正しい） |
| `mtb search` | **`totalResults` が 0 になる** |

### 検索

**一覧の `--search` は Data API 側の検索**で、下書きも対象になる（認証しているため）。

```bash
mtb sites entries list 4 --search "キーワード" --limit 50
mtb sites entries list 4 --search "キーワード" --search-fields body   # 対象を絞る
```

`--search-fields` の既定は `title,body,more,keywords,excerpt,basename`。

**`mtb search` は MT のサイト内検索**で、こちらは**公開済みのものだけ**が対象。

```bash
mtb search --search "キーワード" --blog-id 4
mtb search --search "キーワード" --include-blogs 1,4      # 複数サイト横断
mtb search --tag-search 1 --tag "タグ名" --blog-id 4      # タグ検索
mtb search --search "キーワード" --blog-id 4 --all        # 件数を正しく知るならこれ
```

### 同時に書き換えたとき

**MT に楽観ロックは無い**（ETag も If-Match も返さない）。同じ項目を同時に
更新すると**後勝ちで、両方とも成功として返る**。ぶつかったことは分からない。

一方で更新は項目単位なので、**別々の項目を同時に書く分には消し合わない**
（実測: `body` と `excerpt` を同時更新して両方残った）。

複数のプロセスから同じ記事を触る運用なら、順序を自分で決めること。

### そのMTに無いエンドポイントを見分ける

**仕様に載っていてもサーバに無い機能がある。** v7 の定義と実物を突き合わせると、
LDAP 連携まわりの4本が無いことがある（`users export` / `users import` /
`users synchronize` / `groups synchronize`）。これらは MT の Enterprise パック
（外部ディレクトリ連携）が入っていると生えるもので、AWS Marketplace の AMI には
`AWS.pack` と `Commercial.pack` しか入っていないため存在しない。

`/users/export` のように「無いパス」は `/users/{user_id}` に流れて
**`404 User not found`** になる。パスは合っているのに対象が無いと言われたら、
まずサーバの実際の一覧を見る。

**MTA（Movable Type Advanced）には Enterprise.pack が入っていて、4本とも生える**（MTA 9.3.0 で実測）。

| コマンド | 罠 |
| --- | --- |
| `mtb users export --out users.csv` | **CSV は Shift_JIS**（日本語の環境）。行は `update,名前,新しい名前,表示名,メール,言語` で、そのまま取り込みに戻せる |
| `mtb users import --file users.csv` | UTF-8 でよい。**`register` の行は13列ちょうど**（`register,名前,表示名,メール,言語,初期パスワード` ＋空の列7つ）。足りないと「カラムの数が不正」。`delete,名前` で削除。取り込んだ利用者には `sign_in_cms` と `sign_in_data_api` が付く |
| `mtb users synchronize --yes` / `mtb groups synchronize --yes` | LDAP 認証で外部ユーザー（グループ）管理のときだけ動く。**そうでなければシステム管理者でも 403**（ロールの問題ではない） |

LDAP で同期するときの罠: 初期化で作った管理者に LDAP の `external_id` が無いと、ユーザーの同期は
「すべてのシステム管理者が無効にされる」で止まる。`LDAPUserSyncAttributes fullname,email` が無いと success を返して何も変わらない。
`LDAPGroupFilter` が無いとグループの検索が OU 自身を拾って作成に失敗するが、**API は success を返す**（MT のログにだけ残る）。
**同期の応答の success は「同期が走った」で「変わった」ではない。** 結果は `users list` / `groups list` と MT のログで確かめる。
手元で試す環境は `~/dev/mtb/mtb-harness/mt-launcher`（配布 zip を Docker で起動。LDAP も付く）。
**拡張 `mtb-pro` があれば、こうした罠を埋めたコマンドを使う**: `directory get`（設定の診断）・`directory synchronize`
（変わったものと MT のログの失敗を返す）・`users bulk export`（UTF-8）・`users bulk import`（列を補い、行ごとの結果）。

```bash
mtb endpoints list --fields id,verb,route
```

### カスタムフィールドを足す

```bash
mtb sites fields create 4 --field '{"name":"想定読者","basename":"target_reader","systemObject":"entry","type":"select","options":"開発者,運用者"}'
```

**MT は不正な値に「invalid」としか返さない**ので、使える値を先に知っておく。

| type | 備考 |
| --- | --- |
| `text` `textarea` `checkbox` `url` `embed` `post_type` | そのまま作れる |
| `select` `radio` | `options` にカンマ区切りの選択肢が必須 |
| `datetime` | **`options` が必須**（`date_and_time` / `date_only` / `time_only`） |

`systemObject` は**サイトの種別で変わる**（実測）。

| 値 | ウェブサイト | ブログ |
| --- | --- | --- |
| `entry` `page` `category` `folder` `comment` `template` | ○ | ○ |
| `blog` | × | ○ |
| `website` | ○ | × |
| `author` `asset` `contentdata` | × | × |

### ユーザーを追加する

```bash
mtb users create --user @editor.json --fields id,name,displayName
mtb sites permissions grant --site-id 2 --user-id 2 --role-id 2   # role 2 = Editor
mtb roles list --fields id,name                                   # ロールIDはここで調べる
```

**作ったユーザーはそのままでは Data API にサインインできない。** 3つ要る。

1. **`sign_in_data_api` システム権限** — 認証エンドポイントが明示的に確認する。
   `mtb users update <id> --user '{"systemPermissions":["sign_in_data_api"]}'`
2. **Webサービスパスワード** — `password`（サインイン用）とは別物。
   **MT が自動生成するもので、Data API からは設定も取得もできない。**
   管理画面のユーザー編集で [新しいパスワードを発行] を押し、出た値を控える
   （拡張 `mtb-pro` があれば `mtb users api-password issue <id>`）
3. **サイト側で Data API が許可されていること** — サイト設定 → Web サービス
   （拡張 `mtb-pro` があれば `mtb sites data-api update <site> --enabled true`）。
   **新しいサイトは無効で始まる**（インストーラが作った最初のサイトも、
   `mtb sites create` で作ったサイトも）。システム管理者は素通りするので気づかず、
   一般ユーザーだけが中身の無い `403 Forbidden` を受ける（mtb はこの 403 に
   ロールではなく Data API の設定を疑うよう添える）

どれが欠けても `401 Invalid login` になり、区別は付かない。

**管理画面で更新する人なら `sign_in_cms` も付ける。** API で作った利用者は既定で
システム権限を何も持たず、`sign_in_data_api` だけを付けると、**Data API には入れるが
管理画面には入れない**（「サインインできませんでした。」とだけ出る）。サイトのロールが
正しくても入れない。編集者を作るなら `"systemPermissions":["sign_in_cms","sign_in_data_api"]`。
API から確かめたことは管理画面で確かめたことにならない（実際にこれで、構築は通り、
マニュアルを撮る段で編集者がサインインできないと分かった）。

**記事の作成者は、認証しているユーザーになる。** `entry` に作成者を指定する
項目は無い（`author` を送っても黙って無視される）。別の人の名前で投稿するには
2つの道がある。

- **その人のプロファイルでサインインする**（下の例）。その人の Webサービスパスワードが要る
- **拡張 `mtb-pro` で作ってから作成者を変える**（`mtb sites entries author update`）。
  MT 側に MtbPro プラグインが要る。入っているかは `mtb extensions list` と
  `mtb extensions check mtb-pro` で分かる（→「拡張」）

```bash
mtb profile add mt-editor --base-url http://example.com/cgi-bin/mt/mt-data-api.cgi
mtb auth login --profile mt-editor --username editor
mtb sites entries create --profile mt-editor --site-id 2 --entry @entry.json
```

### サイトを消す（入れ物ごと壊す操作）

サイトの削除・ログの全消し・テンプレートの初期化は、`--yes` だけでは実行できない。
**対象の名前を `--confirm` に打つ**必要がある。

```bash
mtb sites delete 2 --yes                       # 何が失われるかと、打つべき名前が出る
mtb sites delete 2 --yes --confirm "アイデアマンズ ブログ"
```

1件の削除と同じ重さで扱うと、IDの取り違えでサイトが丸ごと消える。
実際に総当たりテストでそれが起きたのでこの形にしてある。

### 記事を書いて公開する

```bash
mtb sites entries create --site-id 1 --entry '{
  "title": "タイトル",
  "body": "本文",
  "status": "Publish"
}'
```

`status` を省くと下書きになる。作成したら応答の `id` を控える。

### 記事を探して直す

```bash
mtb sites entries list --site-id 1 --search "キーワード" --limit 50
mtb sites entries update 1 42 --entry '{"title":"直したタイトル"}'
```

更新は**渡したフィールドだけ**を書き換える。全体を送り直す必要はない。

### 画像をアップロードして記事に添付する

```bash
mtb assets upload --site-id 1 --file ./photo.jpg --path /2026 --fields id,url,mimeType
mtb sites entries update 1 42 --entry '{"assets":[{"id":7}]}'
mtb sites entries assets list 1 42        # 添付されているもの
mtb sites assets entries list 1 7         # 逆引き（このアセットを使う記事）
mtb sites assets thumbnail 1 7 --width 200        # サムネイルを作る
```

**MT は画像を再保存する。** アップロードしたバイト列そのままでは保存されず、
JPEG は quality 85 で入れ直される（4000×3000 の 3.3MB が 1.8MB、画素差 0.4% を実測）。
PNG もメタデータが落ちて数バイト変わる。**画像以外（PDF・テキストなど）は
バイト単位でそのまま**保存される。

元のファイルをそのまま置きたいなら Data API 経由のアップロードは向かない。

| ファイル | 保存のされ方 |
| --- | --- |
| JPEG / PNG などの画像 | MT が再保存する（バイト列は変わる。寸法は保たれる） |
| PDF・テキスト・その他 | バイト単位でそのまま |

`mimeType` は mtb が拡張子から判定して送っている。日本語のファイル名も
そのまま通る（`会社ロゴ.png` → URL は percent-encoded になる）。

### コンテンツタイプ（MT7 のカスタムオブジェクト）

```bash
mtb sites content-types list --site-id 1              # まず content_type_id を調べる
mtb sites content-types fields list 1 3               # フィールドの構成を見る
mtb sites content-types data list 1 3
mtb sites content-types data create 1 3 --data '[{"id":10,"data":"値"}]'
```

**`data` はフィールドID付きの配列**で、記事のような名前付きオブジェクトではない。
フィールドIDは `fields list` で調べる。ここが最も間違えやすい。

**API で作ったコンテンツタイプは、すぐには管理画面のメニューに出ない。** MT はコンテンツ
データのメニューをプロセスの起動時に作る（`MT::CMS::ContentData::make_menus`）。PSGI（Starman）で
動いている MT は、ワーカーが作り直されるまで（`--max-requests` か再起動）古いメニューのまま。
権限は正しいのに編集者のメニューに出ないときはこれ。API からは直せないので、サーバの管理者に
MT の再起動を頼む。

**コンテンツタイプのカテゴリ別アーカイブ（`ContentType-Category`）は Data API では作れない。**
アーカイブマッピングに「どのカテゴリのフィールドで分けるか」を渡す項目が無い。
テーマに入れて適用する（`theme.yaml` の `mappings`。拡張 `mtb-pro` の `themes files upload` →
`sites themes apply`）。

### テンプレートとデザインを直す

```bash
mtb sites templates list --site-id 2 --fields id,name,type,outputFile
mtb sites templates get 2 96 --fields text          # 中身を読む
mtb sites templates update 2 96 --template @css.json
mtb sites templates publish 2 76                   # index テンプレートを書き出す
```

`--template` に渡すのは `{"text": "..."}`。CSS や HTML はファイルに書いて
`@ファイル名` で渡すのが楽。

**`mtb sites refresh-templates` は「初期状態に戻す」。** 名前から再構築だと
思って叩くと、書いたテンプレートが全部消える。再構築は `publish`。

**テンプレートを試すのに、捨て用のテンプレートもサンドボックスのサイトも要らない。**
`preview` は **テンプレートIDを省くと、保存せずに本文だけを評価する**
（`POST /sites/{site_id}/templates/preview`）。しかも実データが相手なので、
タグが本当に動くかをその場で確かめられる。

```bash
mtb sites templates preview 6 --set template.text=@snippet.mtml \
  --set template.type=index --raw 1 --extract preview
```

- **IDを付けると既存テンプレートのプレビュー**になる。省くと ID なしの方を呼ぶ
- `--raw 1` を付けないと HTML がエスケープされて返る
- 同じ形が `sites entries preview` / `sites pages preview` /
  `sites content-types data preview` にもある（記事やページを保存せずに描ける）

**タグの綴り違いは丸ごと 500 になる**が、エラー本文に「認識できないタグ」が
全部並ぶので、候補をまとめて1回で試すと速い。

### 再構築（静的ファイルの書き出し）

```bash
mtb publish entries --ids 42,43            # 記事に関係するアーカイブ
mtb sites templates publish 1 5            # index テンプレート（同期。すぐ反映される）
```

**`publish entries` は1回では終わらない。** 応答の `{"status":"Rebuilding"}` は
受け付けたという意味でしかなく、残りは `restIds` に入る。空になるまで呼び直すのが
本来の使い方で、**1回だけ呼んで終わると成功したように見える**。

```bash
mtb publish entries --ids 42,43 --wait     # restIds が空になるまで mtb が呼び直す
```

**手で呼び直すなら `--start-time` を必ず渡すこと。** 応答の `startTime` を
返さないと、MT は毎回「新しい再構築」として受け付け直し、**`restIds` が
いつまでも減らない**。エラーにはならず `Rebuilding` が返り続けるので、
これは自力では気づけない（実測: 渡さないと12回呼んでも6件のまま、
渡すと2回で空になった）。`--wait` はこれを引き継いでいる。

```bash
mtb publish entries --ids 42,43 --start-time 2026-08-29T08:29:57+00:00
```

`--wait` は進み具合も見ている。`restIds` が減らないまま3回続いたら、
待っても終わらないと判断して終了コード 8 で止まる。

すぐ反映したいなら対象を保存し直す。保存は同期で再構築される。

```bash
mtb sites entries update 2 7 --entry '{"excerpt":"（いまと違う値）"}'
```

**ただし値が変わらない更新では再構築されない。** 同じタイトルを送り直しても
MT は何もしないので、ファイルは古いままになる。実際に変わる値を送ること。

**このとき `--entry '{}'` を送らないこと。** 空の更新は「何もしない」ではなく、
指定しなかった関連（フォルダ・カテゴリ）が外れることがある。実際に
ページのフォルダが外れて URL が変わった例がある。mtb は警告を出すが、
送信は止めない。

**ページは、空でなくても `folder` を書かない更新でフォルダから外れる**
（MTA 9.3.0 で確かめた。`--set page.status=Publish` だけ、`page.title` だけでも
`folder` が `null` になり、URL がサイト直下に移る）。basename が `index` のページだと、
**公開した瞬間にサイトのトップページ（`/index.html`）を上書きする**（実際に起きた）。
ページを更新するときは、毎回 `folder` を一緒に送る。

```bash
mtb sites pages update 1 7 --page '{"status":"Publish","folder":{"id":4}}' --publish 1 --verify-publish
```

## 失敗モード

| 症状 | 原因 | 対処 |
| --- | --- | --- |
| 401 Invalid login | サインインパスワードを使っている | Webサービスパスワードを確認する |
| 401 Invalid login（Webサービスパスワードを DB の `mt_author.author_api_password` から読んで渡した） | MT 9 は Webサービスパスワードもハッシュ（`$6$…`）で持つ。読めるのはハッシュで、平文は DB から取り出せない | ハッシュは平文の代わりにならない。DB ごと別の MT に移したなら元の平文がそのまま通る。平文が分からなければ発行し直す（拡張 `mtb-pro` の `users api-password issue`） |
| 401 が実行のたびに出る | セッションが切れている | `mtb auth login` をやり直す |
| 403（文面あり） | ユーザーの権限不足 | MT の管理画面でロールを見る。`Only a system administrator` はロールでは足りない |
| 403 `Forbidden`（文面なし） | そのサイトの Data API が無効 | 下の「中身の無い 403」 |
| 404 `Unknown endpoint` | ベースURLが `mt-data-api.cgi` まで届いていない／その MT に無いエンドポイント | `mtb profile show` で確認 |
| 404（`Site not found` など文面あり） | エンドポイントはあり、対象が無い | ID・サイトの取り違えを見る |
| 405 Method Not Allowed | 前段が PUT / DELETE を通していない | `--method-override` を付ける |
| JSON でなく HTML が返る | ベースURLが管理画面を指している | 同上 |
| 一覧が10件で切れる | `--limit` を指定していない | `--limit` と `totalResults` |
| 変更したのにサイトに出ない | 再構築していない | `mtb publish entries --ids ...` |
| 対話入力で止まる/失敗する | 端末でない | `--yes` / `--username` / `--password-stdin` |
| `A resource "blog" is required.` | 子サイトを `--website` で更新した | `--blog` を使う |
| 子サイトのURLが二重になる | 絶対URLを渡した | `"url": "blog/"` のように相対で |
| 応答が大きすぎて読み切れない | 全属性が返っている | `--fields` で絞る |
| `Rebuilding` が返るのにファイルが変わらない | 1回では終わらない | `--wait` を付ける（`restIds` が空になるまで呼び直す） |
| 呼び直しても `restIds` が減らない | **`startTime` を渡していない**（毎回やり直しになる） | 応答の `startTime` を `--start-time` で返す。`--wait` なら自動 |
| 保存は 200 なのに静的ファイルが増えない | **関係のないテンプレートが壊れていて再構築ごと失敗している** | `--verify-publish` を付けるか `mtb sites logs list <site> --level error` |
| 保存し直しても再構築されない | 値が変わらない更新は MT が何もしない | 実際に変わる値を送る |
| `--fields` を付けたら `{}` が返った | 属性名の綴り違い | mtb が知らせる。サーバ側で絞る endpoint では `--fields` を外して名前を確かめる |
| 記事のタイトルの `—` が `--` や `&#8212;` になった | **作成のときだけ記号が書き換わる**（更新では起きない） | 作成したあとに同じ値で更新し直す。`smartReplace=1` にしても原文のままにはならない |
| 空の更新をしたら URL が変わった | `{}` でフォルダ・カテゴリが外れた | 変更する項目だけを書いて送る |
| 数値のはずのIDで 404 | パスに変な値が入った | mtb が手元で弾くので、その文言を読む |
| API で作った編集者が**管理画面に**サインインできない（Data API は通る） | システム権限に `sign_in_cms` が無い | `mtb users update <id> --user '{"systemPermissions":["sign_in_cms","sign_in_data_api"]}'` |
| 作ったコンテンツタイプが管理画面のメニューに出ない | MT がメニューを起動時に作る（権限は関係ない） | MT を再起動する（PSGI ならワーカーの入れ替わりでも直る） |
| 新しい記事を開くと書式が「なし」 | サイトの既定の書式（`convertParas`）。API で作ったサイトは `0` | ブロックエディタを使うサイトなら `mtb sites update <id> --set website.convertParas=block_editor` |
| 記事の概要・カスタムフィールドが編集画面に無い | 表示オプションで隠れている（利用者ごとの設定。API では変えられない） | 編集者に「表示オプション」で出してもらう。マニュアルに書く |
| API で作ったユーザーがサインインできない | `sign_in_data_api` 権限・Webサービスパスワード・サイトの Data API 許可のどれかが欠けている | 権限は API で付く。あとの2つは管理画面か、拡張 `mtb-pro` |
| 一般ユーザーだけ中身の無い `403 Forbidden` | そのサイトの Data API が無効（システム管理者は素通りする） | サイト設定 → Web サービス。拡張 `mtb-pro` なら `mtb sites data-api get` で確かめられる |
| `permutate` が 400 | id だけの配列を渡した | list の出力をそのまま並べ替えて渡す |
| `500 malformed JSON string` | `@ファイル名` が展開されず届いた | 文字列の項目では展開されない。警告が出ていないか見る |
| コンテンツデータが `"Data Label" is required.` | label を書いていない | `label` はフィールドではなくデータ自体の名前 |
| ページ作成時に folder を指定したのに root に出る | 作成時は無視される | 作ってから更新でフォルダを設定する |
| ページの状態やタイトルを変えたら URL がサイト直下に移った・トップページが別のページになった | 更新で `folder` を送らないと、ページがフォルダから外れる（basename `index` ならトップの index.html を上書きする） | 更新のたびに `folder` を一緒に送る。上書きしたトップはインデックスのテンプレートを `mtb sites templates publish` で書き出し直す |
| ページをフォルダの basename で選り分けたら、全部が「該当なし」になる | ページの一覧の `folder` は `id`・`label`・`parent` だけで、`basename` を持たない | `mtb sites folders list` で basename → id を引き、`folder.id` で比べる。**消す処理で間違えると自前のページまで消える** |
| 403 `turned off in the MtbPro plugin settings`（システム管理者なのに） | mtb-pro の MtbPro が運用時の設定でそのエンドポイントを閉じている | 権限の問題ではないので、プロファイルやロールを変えても通らない。Data API から開ける口は無い（わざと）。利用者に伝える。開いているものは `mtb endpoint-policy get` |
| `users synchronize` / `groups synchronize` が 403（システム管理者なのに） | LDAP の外部ユーザー管理ではない | ロールでは直らない。`mt-config.cgi` の `AuthenticationModule LDAP` と `ExternalUserManagement 1` |
| `users import` が「カラムの数が不正」 | `register` の行が13列でない | 後ろに空の列を足して13列にする |
| `Unknown endpoint` / 見覚えのない `404` | そのMTに無い機能 | `mtb endpoints list` でサーバの実際の一覧を見る |
| ページにコメントできない | page リソースに `allowComments` が無い | サイト側の `allowCommentsDefault` を立ててから**ページを作る** |
| メール送信系が 500 | `EmailAddressMain` 未設定 | MT の全般設定でシステムのメールアドレスを入れる |
| 統計が `400 startDate is required` | 集計期間は必須 | `--start-date` と `--end-date` を渡す |
| 取り込みで `Invalid encoding` | 有効な値は MT の言語設定次第 | 英語環境は `utf8`、日本語環境は `sjis` など。既定の `guess` に任せるのが早い |
| `--cd-search 1` で 500 | MT の不具合（並び順の列名が壊れる） | `--search-sort-by` を一緒に渡す |
| `--publish 0` にしたのに公開されている | これは再構築の抑止で、ステータスではない | 下書きにするなら `status` を `Draft` に |
| 記事の著者を変えたのに変わらない | `author` は読み取り専用。**200 が返って黙って無視される** | 本体では変えられない。拡張 `mtb-pro` の `mtb sites entries author update` |
| `409 A file named '...' already exists.` | MT はアセットを上書きできない | `--replace` で消してから送る（ID が変わる）。ID を保つなら拡張 `mtb-pro` の `sites assets replace`。`--auto-rename-if-exists` は名前がハッシュになる |
| `'.js' is not allowed to upload by system settings.` | `DeniedAssetFileExtensions` の既定が `js` `html` などを拒否 | MT の作法どおり**インデックステンプレート**で置く。ビルド済みの JS をそのまま置くなら拡張 `mtb-pro` の `sites files upload` |
| テーマのエクスポートが `success` なのに中身が空 | 静的ファイルの対象フォルダがサイトに保存されていない | 管理画面のテーマエクスポート画面でフォルダ名を保存する（本体の API からは設定できない）。拡張 `mtb-pro` の `sites theme-export-settings update`。テーマの適用で消える |
| `mtb sites themes apply` が `--yes` だけで通らない | テンプレートの総入れ替えなので `--confirm <サイト名>` も要る | 対象を確かめてから名前を打つ |
| アップロードが `500 unable to write to the "Upload Destination"` | サイトの配置先がウェブサーバから書けない | `mtb sites get <ID> --fields sitePath` の場所の権限を見る |

## 拡張（MT プラグインが足したエンドポイントのコマンド）

**拡張を入れると、MT プラグインが足したエンドポイントにもコマンドが生える。**
生成コマンドと同じ組み立てなので `--set` `--fields` `--dry-run` `--all` が効き、
名前も同じ規則（「コマンド名とエンドポイントの対応」）で決まる。

```bash
mtb extensions list                       # 入っている拡張と、読み込めたか
mtb extensions show mtb-pro              # 足したコマンド・要る MT プラグイン
mtb extensions check mtb-pro             # その MT に要るエンドポイントがあるか
mtb extensions install ./mtb-plugin-pro   # 入れる（ディレクトリを写す）
```

- **拡張ごとの使い方は `mtb llm` の「拡張: <名前>」の章にある。** 拡張のコマンドは
  コマンドカタログには載らないので、まずこの章を読む
- **拡張は2つで1組。** mtb 側の拡張と、MT 側のプラグイン。片方だけでは動かない。
  MT 側が無いと拡張のコマンドは **404 `Unknown endpoint`** になり、エラーに「プラグインが
  入っていない可能性」と `mtb extensions check` が添えられる。プラグインが返す 404
  （`File not found` など）には添えない（対象が無いだけ）。check は足りなければ終了コード 1
- **「プラグイン」と「拡張」は別物。** `mtb plugins` は MT のプラグインを操作する
  生成コマンド、`mtb extensions` は mtb 自身の拡張の管理
- 名前がぶつかる拡張は**丸ごと読み込まない**（警告が標準エラーに毎回出る）。
  一部だけ入ることはない
- 開発中の拡張は `MTB_EXTENSIONS_PATH`（`:` 区切りのディレクトリ）で入れずに読める。
  インストール済みの同名のものより優先される。`mtb extensions validate <dir>` で検査できる
  （`--strict` で同梱の Claude Code のプラグインの版と書式も失敗にする）
- 拡張は JSON と Markdown だけで、**コードは走らない**。拡張に Claude Code の
  プラグインが同梱されていれば `mtb extensions show` がその場所を示す

拡張の作り方（ディレクトリの形・定義の書き方）は README の「拡張」にある。

## 定義に無いエンドポイントを叩く

埋め込んでいる定義は取り込み時点の v7。プラグインが足したエンドポイントや、
MT 側の新しい追加は生のリクエストで叩く（拡張が入っていればそちらのコマンドを使う）。

```bash
mtb endpoints list --all --fields component,verb,route   # そのMTが実際に持つもの
mtb api GET /sites/1/things --all --fields id,label
mtb api POST /sites/1/things --set thing.label=ラベル --set thing.note=@note.txt
mtb api POST /sites/1/assets/upload --file ./photo.jpg --form path=/2026
```

**パスを組み立てずにエンドポイント ID で呼ぶなら `mtb call`。** 経路とメソッドはサーバの
`/endpoints` から引く。拡張の無いプラグインのエンドポイントは、まずこちら。

```bash
mtb call <エンドポイントID> --describe              # パラメータをサーバの OpenAPI 定義から表示
mtb call <エンドポイントID> 1 42 --form name=value  # パスの値は経路に出てくる順（--path 名前=値 も可）
```

- `--describe` はサーバが返す説明をそのまま出す（プラグインが openapi_handler を持っていれば）
- 同じ ID が版違いで並ぶときは新しい版を採る。メソッドや経路が違うものが並ぶと断る（そのときは `mtb api`）
- `site_id` が足りなければプロファイルの既定サイトを使う。`--dry-run` でも経路を引くために1回問い合わせる
- DELETE は `mtb api` と同じく確認が要る（`--yes`）

**`--all` `--fields` `--set` `--set-json` は生成コマンドと同じように使える。**
プラグインが足したエンドポイントこそ、ページ繰りと本体の組み立てが要る。
`--set` はここでは名前を検査しない（どんな本体を取るか事前に分からないため）。

`mtb endpoints list` の `component` が提供元。検証した MT では 246 が Core、
20 がプラグイン由来だった（Professional Pack / Comments / FormattedText）。
**カスタムフィールドもコメントも、実はプラグインが足したエンドポイント。**

`mtb endpoints list` はサーバ自身に問い合わせるので、**そのMTで本当に使える
エンドポイント**が分かる。定義との差分を疑ったらこれを見る。
