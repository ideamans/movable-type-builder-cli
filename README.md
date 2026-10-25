# mtb — Movable Type Builder CLI

Movable Type のサイトを、コマンドラインから組み立てる Go 製 CLI です。
接続先はプロファイルで切り替え、認証は一度サインインすれば以後は自動で
トークンを取り直します。

**コマンドは [Data API v7](https://movabletype.github.io/mt-docs-data-api-reference/v7.html)
の公式定義から作られています。** 226 のエンドポイントが最初から全部使え、
`--help` にはパラメータと、公式リファレンスの該当箇所へのリンクが出ます。

**その上に、API に無いものを足してあります。** ディレクトリごとのアップロード、
JSON を書かずにリソースを組み立てる `--set`、1項目だけ取り出す `--extract`、
テンプレートをディレクトリと往復させる `pull` / `push`。**サイトを1つ作り切る
のに要る操作が、API の形のままでは足りないからです。**

```bash
mtb profile add prod --base-url https://example.com/cgi-bin/mt/mt-data-api.cgi
mtb auth login
mtb sites entries list --site-id 1 --status Publish --limit 20
```

## インストール

リポジトリが private のため、リリースの取得には `gh` の認証が要ります。

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64

gh release download --repo ideamans/movable-type-builder-cli \
  --pattern "*_${OS}_${ARCH}.tar.gz" --dir /tmp --clobber
tar -xzf /tmp/movable-type-builder-cli_*_${OS}_${ARCH}.tar.gz -C /tmp
mkdir -p ~/.local/bin && mv /tmp/mtb ~/.local/bin/
```

**アーカイブ名はリポジトリ名、中のバイナリは `mtb`** です。

ソースから入れる場合（main パッケージはモジュール直下ではありません）:

```bash
GOPRIVATE=github.com/ideamans/* \
  go install github.com/ideamans/movable-type-builder-cli/cmd/mtb@latest
```

## 接続先（プロファイル）

MT は複数台あるのが普通なので、接続先に名前を付けて切り替えます。

```bash
mtb profile add prod --base-url https://example.com/cgi-bin/mt/mt-data-api.cgi
mtb profile add stg  --base-url https://stg.example.com/cgi-bin/mt/mt-data-api.cgi
mtb profile use stg                    # 既定を切り替える
mtb profile set --site-id 3            # 既定サイトID（--site-id を省ける）
mtb profile list
mtb profile show                       # いま実際に使う接続先
mtb sites entries list --profile prod  # その場だけ切り替える
```

`mtb profile add` を引数なしで叩くと、いまあるプロファイルを一覧してから
名前と URL を聞きます。同じサーバを二重に登録したり、本番に向いたまま
作業を続けたりしにくくするためです。登録できたら**そのままサインインするか
聞きます**（既定は Y）。

```
$ mtb profile add
プロファイル名（例: prod）: prod
Data API の URL（例: https://example.com/cgi-bin/mt/mt-data-api.cgi）: ...
プロファイル prod を追加しました
このままサインインしますか？ (Y/n):
ユーザー名: alice
Webサービスパスワード:
```

対話できない場面（エージェント・CI）では聞かず、次にやることだけ伝えます。

`--base-url` は `mt-data-api.cgi` までの URL です。リファレンスからコピーして
末尾に `/v7` が付いていても、取り除いてバージョンとして解釈します。

優先順位は **フラグ → 環境変数 → プロファイル → 既定値**。

| 環境変数 | 用途 |
| --- | --- |
| `MTB_PROFILE` | 使うプロファイル名 |
| `MTB_BASE_URL` | 接続先（プロファイル無しでも動かせる） |
| `MTB_SITE_ID` | 既定サイトID |
| `MTB_USERNAME` / `MTB_PASSWORD` | 非対話サインイン |
| `MTB_ACCESS_TOKEN` | 取得済みトークンだけで叩く |
| `MTB_CONFIG_DIR` | 設定ディレクトリ（既定 `~/.config/mtb`）。拡張は `extensions/` に入る |
| `MTB_EXTENSIONS_PATH` | 入れずに読み込む拡張のディレクトリ（`:` 区切り。開発用） |
| `MTB_AUDIT_NOTE` | 何の作業かの説明。全リクエストに `X-MTB-Audit-Note` ヘッダで付ける（`--audit-note` と同じ） |
| `MTB_LANG` | 表示言語 `ja` / `en`（既定は `LC_ALL` → `LC_MESSAGES` → `LANG` から判定。いずれも無ければ英語） |

**カレントディレクトリに `.env` があれば読み込みます。** 上の環境変数を
プロジェクトごとに置いておけます。

```bash
# work/.env
MTB_PROFILE=stg
MTB_SITE_ID=3
```

`export` 付きの行・`#` で始まる行・値を囲む引用符に対応します。**行末コメントは
扱いません**（`MTB_PASSWORD=pa#ss` の `#` から先を落とすと、パスワードが変わった
まま 401 を踏むため）。**プロセスの環境変数が勝ちます**——打ったものをファイルが
黙って覆すことはありません。読み込んだ場所は `mtb profile show` に出ます。

`MTB_PASSWORD` を書くなら `.gitignore` に `.env` を入れてください。

表示言語は `--lang ja` / `--lang en` でその場だけ切り替えられます。切り替わるのは
人に向けた文言（メッセージ・ヘルプ・エラー）で、公式の OpenAPI 定義から来る
エンドポイントの説明は英語のままです。`mtb llm` の埋め込みリファレンスも
日本語で固定です（AIエージェント向けなので言語を揃える意味が薄いため）。

## 認証

```bash
mtb auth login       # ユーザー名とパスワードを聞かれる
mtb auth status
mtb auth logout
```

パスワードは MT の**Webサービスパスワード**です（サインイン用とは別物で、
MT のユーザー設定画面にあります）。ここを取り違えると 401 になります。

保存されるのは**セッションIDとアクセストークンだけ**で、パスワードは保存しません。
アクセストークンは1時間ほどで切れますが、セッションIDから自動で取り直すので
`mtb auth refresh` を明示的に叩く必要はありません。

設定は `~/.config/mtb/config.json`、認証情報は同じディレクトリの
`credentials.json`（0600）に分けて保存されます。

CI やエージェントからは対話なしで:

```bash
# パスワードをパイプで渡す（どこにも残りません）
pass show mt/prod | mtb auth login --username alice --password-stdin

# あるいは環境変数で
MTB_PASSWORD=... mtb auth login --username alice

# あるいは設定ファイルを一切使わずに
MTB_BASE_URL=... MTB_ACCESS_TOKEN=... mtb sites entries list --site-id 1
```

パスワードの渡し方は3つあり、残るものが違います。

| | 残るところ |
| --- | --- |
| `--password-stdin` | どこにも残りません |
| `MTB_PASSWORD` | 同じホストの他のプロセスから見えます |
| `--password` | プロセス一覧に残ります。他に手が無いときだけ |

## リソースの組み立てと取り出し

記事やテンプレートの本体は JSON ですが、**JSON を手で書かなくても組み立てられます。**

```bash
# 項目ごとに指定する。値の @ファイル名 はファイルから読みます
mtb sites templates create --site-id 1 \
  --set template.name=メインページ \
  --set template.type=index \
  --set template.outputFile=index.html \
  --set template.text=@index.html

# 逆に、1つの項目だけを生の値で取り出す
mtb sites templates get 1 42 --extract text --out index.html
```

`--set` の値は文字列です。数値・真偽値・入れ子は `--set-json` を使います
（`--set-json template.contentType='{"id":5}'`）。`--template @base.json` と
併用すると、ファイルを土台にして一部だけ差し替えられます。

`--set` で何を指定できるかは `--help` に出ます。

```
--template string  (object: buildType, contentType, customFields, linkToFile,
                    name, outputFile, templateType, text, type)
```

`--extract` は文字列ならそのまま、それ以外は JSON で書きます（`jq -r` と同じ）。
`items.0.name` のように配列の添字も辿れます。**`--out` と組み合わせるとバイト
そのままファイルになります**（端末に出すときだけ末尾に改行が付きます）。

## 環境どうしの差分を取る

ステージングで組んだデザインを本番へ反映するとき、一番時間を食うのは
反映そのものではなく **どこが違うのかを突き止めること**です。`mtb dump` は
接続先の中身を種類ごとの YAML にして書き出すので、あとは普通の diff で読めます。

```bash
mtb dump --profile stg  --dir ./dump/stg
mtb dump --profile prod --dir ./dump/prod
diff -ru ./dump/prod ./dump/stg
```

```diff
--- ./dump/prod/sites/SaasHub/templates.yaml
+++ ./dump/stg/sites/SaasHub/templates.yaml
@@ -60,6 +60,8 @@
        </mt:If>
+
+    <!-- ステージングで足した行 -->
   type: custom
```

**ID・時刻・作成者は書き出しません。** 環境ごとに違うので、残すと中身が
同じでも全行が差分になります。並べ替えと突き合わせは `basename` や `name`
などの**自然キー**で行い、他のオブジェクトへの参照（記事のカテゴリなど）も
自然キーに畳みます。階層のあるカテゴリ・フォルダは親を根からの道で持ちます。

**複数行はリテラルブロック（`|`）で書きます。** テンプレート本文は千行を
超えることがあり、1行の文字列にすると差分が「1行変わった」としか出ません。
行末の空白やタブを含む文字列は引用に落とすので、**書き出しても中身は変わりません**。

アセットはサイトルートからの相対パスで持ちます。中身まで突き合わせるなら
`--asset-hash` を付けてください（MD5 を記録します。1ファイル1リクエスト増えます）。

**受け側の MT が生成する値は書き出しません。** `uniqueID`（環境ごとに生成される）
と `fileSize`（MT は JPEG を取り込むときに再エンコードする）は、残すと永久に
消えない差分になります。参照している側は意味が残る形に畳みます
（`dataLabel` → `dataLabelField: 氏名`）。

**サイト設定は2つに分かれます。** `site.yaml` は運ぶべき設定、
`site-environment.yaml` は `host` `sitePath` など**運んではいけないもの**です。
混ぜると「差分だが運んではいけない」行が紛れて判断できなくなります。

**サーバの素性も `system/` に記録します。** 運べないものですが、**違っていると
分かるだけで価値があります**。

| | なぜ要るか |
| --- | --- |
| `plugins.yaml` | 無いプラグインに依存したテンプレートは公開で落ちます（ブロックエディタもプラグイン） |
| `text-filters.yaml` | `format: markdown` の記事は、その書式が無い MT では読めません |
| `endpoints.yaml` | その MT に無い機能は、叩いた時点で 404 になります |
| `server.yaml` | API のバージョン |

**`mt-config.cgi` は API から読めません。** `EmailAddressMain`（未設定だとメール
送信が 500）、`DeniedAssetFileExtensions`（`.js` `.html` が上げられない理由）、
`DefaultLanguage` などは、移行の前に人が揃える必要があります。

統計・ログ・コメントは対象外です（環境で違って当たり前のもの）。

## テンプレートを Git で持つ

サイトのテンプレートをディレクトリに書き出し、編集して戻せます。

```bash
mtb sites templates pull 4 --dir ./templates   # 1テンプレート1ファイル＋マニフェスト
$EDITOR ./templates/スタイルシート.mtml
mtb sites templates push 4 --dir ./templates   # 変わったものだけ戻す
```

置き場所は MT 自身のテーマ構成に寄せてあります。本文は `.mtml` ファイル、
種別や出力ファイル名は `mtb-templates.json` が持ちます。

- **対応は ID で取ります。** MT 側で名前を変えても、手元でファイル名を変えても
  切れません
- **変わっていないものは送りません。** 1つ直しただけで全件にリビジョンが
  積まれると、MT の更新履歴が役に立たなくなります
- **MT にあって手元に無いテンプレートには触れません。** 件数と名前を報告するので、
  消すなら `sites templates delete` を明示的に使ってください
- マニフェストに `id` の無い項目を書くと新規に作り、採番された ID を書き戻します
- アーカイブマッピングも一緒に往復します（マニフェストの `archiveMaps`）。
  **マッピングは URL の設計そのもの**なので、これが無いと pull したディレクトリから
  出力先を戻せません。本文が同じでもマッピングだけ変わっていれば送ります

**`pull` は手元のファイルを上書きします。** 先に commit するか `--dry-run` で
確かめてください。`push` は書き換える件数を示して確認を求めます（`--yes` で省けます）。

サーバのファイルシステムに触れるなら、テンプレートを外部ファイルとリンクする
（`--set template.linkToFile=...`）方が MT 公式の仕組みで、pull/push は要りません。

## ファイルのアップロード

```bash
mtb assets upload --site-id 1 --file ./logo.png --path /theme-images

# ディレクトリを丸ごと（下の階層は送り先にも写します）
mtb assets upload --site-id 1 --dir ./dist/theme-images --path /theme-images
```

Data API に一括アップロードは無いので、`--dir` は mtb が1ファイルずつ送ります。
ページを繰る `--all` と同じ立場のものです。隠しファイル（`.DS_Store` など）は
送りません。途中で失敗したらそこで止まり、何件目で止まったかを標準エラーに
出します。何が送られるかは `--dry-run` で先に確認できます。

**同名のファイルは上書きできません。** MT は 409 を返します
（`A file named '...' already exists.`）。Data API に上書きの手段が無いためです。
入れ替えるときは `--replace` を使ってください。送り先にある同名のアセットを
消してから送ります。

```bash
mtb assets upload --site-id 1 --dir ./dist/theme-images --path /theme-images --replace --yes
```

消すのは破壊的な操作なので、確認が入ります（`--yes` で省けます）。**消す対象は
送り先まで一致したものだけ**で、別フォルダの同名ファイルは巻き込みません。

**JS と HTML はアセットにできません。** MT の `DeniedAssetFileExtensions` が
既定で拒否します（`js` `html` `htm` `shtml` ほか）。MT の作法では
**スタイルシートと JavaScript はインデックステンプレート**なので、
`--set template.type=index --set template.outputFile=theme.js` で置いてください。
画像・CSS・SVG・フォント・webp はアセットで上げられます。

## 定義に無いエンドポイントを叩く

**MT はプラグインで Data API を増やせます。** 増えたぶんは公式定義に載らないので
コマンドは生えませんが、`mtb api` から叩けます。何があるかは `mtb endpoints list`
がそのサーバに聞いてくれます。

```bash
# そのMTが実際に持っているもの（提供元つき）
mtb endpoints list --all --fields component,verb,route

# 叩く。--all・--fields・--set は生成コマンドと同じように使えます
mtb api GET /sites/1/things --all --fields id,label
mtb api POST /sites/1/things --set thing.label=ラベル --set thing.note=@note.txt
```

**パスを組み立てずに、エンドポイント ID で呼ぶこともできます**（`mtb call`）。経路とメソッドは
サーバの `/endpoints` から、パラメータはサーバ自身の OpenAPI 定義から引くので、拡張の無い
プラグインのエンドポイントでも `--describe` で何を渡せばよいかが分かります。

```bash
mtb endpoints list --fields id,verb,route,component     # ID を調べる
mtb call update_thing --describe                         # パラメータをサーバに聞く
mtb call update_thing 1 42 --form label=新しい名前        # :site_id :thing_id を順に埋める
mtb call update_thing 42 --path site_id=1 --dry-run      # 名前で渡す・送らずに確かめる
```

検証に使った MT では、246 が Core、20 がプラグイン由来でした
（Professional Pack 5 / Comments 10 / FormattedText 5）。**カスタムフィールドや
コメントも、実はプラグインが足したエンドポイント**です。

## 拡張（MT プラグインのエンドポイントにコマンドを生やす）

`mtb api` で叩けるとはいえ、プラグインのエンドポイントにも `--help` と
フラグが欲しくなります。**拡張**は、その定義を mtb に持ち込むものです。
入れると生成コマンドと同じ形のコマンドが生え、`--set` `--fields` `--dry-run`
もそのまま効きます。

```bash
mtb extensions install ./mtb-plugin-pro    # 入れる（設定ディレクトリの extensions/ に写す）
mtb extensions list                            # 一覧
mtb extensions show mtb-pro                   # 足したコマンド・要る MT プラグイン
mtb extensions check mtb-pro                  # MT 側にエンドポイントがあるか（無ければ終了コード 1）
mtb extensions uninstall mtb-pro

mtb sites entries author update 1 42 --author-name editor   # 拡張が足したコマンド
```

**拡張は MT 側のプラグインと組で使います。** MT 側が無いと拡張のコマンドは 404 に
なり、その旨の助言が出ます。MT のプラグインそのものを操作する `mtb plugins` とは
別物です（名前がぶつかるので、mtb 側は「拡張」と呼んでいます）。

### 拡張の作り方

拡張は**宣言だけ**でできたディレクトリで、コードは走りません。

```
mtb-plugin-pro/
├── mtb-extension.json        マニフェスト（名前・版・要る MT プラグインとエンドポイント）
├── spec.json                 追加するコマンドの定義（internal/spec/spec.json と同じ形）
├── llm.md                    mtb llm に足す章（任意）
└── plugins/mtb-pro/         Claude Code のプラグイン（任意）
```

```json
{
  "name": "mtb-pro",
  "version": "0.1.0",
  "description": "What only the admin screen can do",
  "descriptionJa": "管理画面でしかできない操作",
  "requires": {
    "mtPlugin": { "name": "MtbPro", "version": "0.1.0", "url": "https://..." },
    "endpoints": ["update_entry_author", "upload_theme_file", "..."]
  },
  "claudePlugin": "plugins/mtb-pro"
}
```

- `spec.json` は `groups`（新しい中間ノード）と `commands` を持ちます。パラメータの
  `flag` を省けば `name` のケバブケース、`type` を省けば `string` になります
- **コマンド名は生成コマンドと同じ規則で付けてください**（`internal/llmdocs/10-endpoints.md`）
- 既存のコマンドとぶつかる拡張は丸ごと読み込まれません
- `requires.endpoints` は MT の `/endpoints` に出る ID です。`mtb extensions check` が使います
- 作っている間は `MTB_EXTENSIONS_PATH=./mtb-ext-xxx mtb ...` で入れずに試せます。
  `mtb extensions validate ./mtb-ext-xxx` で検査できます。`--strict` を付けると、同梱の
  Claude Code のプラグインの問題（版の不一致・スキルの書式）も失敗にします（拡張の CI 向け）
- 送るものの確かめ方は `--dry-run` の出力を期待値と突き合わせるのが手堅い（実例は
  mtb-plugin-pro の `tests/dry_run.py`）
- `llm.md` は `mtb llm` の「拡張: <名前>」の章になります（見出しは1段下げて差し込みます）

実例は [`mtb-plugin-pro`](https://github.com/ideamans/mtb-plugin-pro)（MT 側は
[`mt-plugin-mtb-pro`](https://github.com/ideamans/mt-plugin-mtb-pro)）。
管理画面でしかできなかった操作（作成者・テーマの取り込み・書き出し設定・カスタムブロック・
Webサービスパスワード・サイトの Data API 許可・アセットの差し替え・.js の配置）を16のコマンドにしています。

- `file` と `path` を取るエンドポイントには `--dir` が付きます（1ファイルずつ階層を写して送る）

## コマンドの探し方

コマンド名はリファレンスのパスと1対1で対応します。

```
GET    /v7/sites/{site_id}/entries              mtb sites entries list
POST   /v7/sites/{site_id}/entries              mtb sites entries create
GET    /v7/sites/{site_id}/entries/{entry_id}   mtb sites entries get
PUT    /v7/sites/{site_id}/entries/{entry_id}   mtb sites entries update
DELETE /v7/sites/{site_id}/entries/{entry_id}   mtb sites entries delete
POST   /v7/assets/upload                       mtb assets upload
```

規則はこれだけです。

1. パスのリテラルなセグメントがグループ（キャメルケースはケバブケースに）
2. パスパラメータはコマンド名に現れない（フラグか位置引数で渡す）
3. 末尾が動詞（`upload` `export` `preview` `grant` など）ならそれが動詞
4. そうでなければ HTTP メソッドから `list` / `get` / `create` / `update` / `delete`

例外は7つだけで、`mtb llm` の「コマンド名とエンドポイントの対応」に全部載っています。

```bash
mtb sites entries --help          # ぶら下がっている操作
mtb sites entries list --help     # パラメータと公式リファレンスへのリンク
```

複数形のグループには単数形の別名があります（`mtb site entry list` も同じ）。

## 使い方

```bash
# 一覧（既定は10件。--limit と totalResults を見る）
mtb sites entries list --site-id 1 --search "キーワード" --limit 50

# 作成（status を省くと下書き）
mtb sites entries create --site-id 1 --entry '{"title":"タイトル","body":"本文","status":"Publish"}'

# 更新（渡したフィールドだけ書き換わる）
mtb sites entries update 1 42 --entry '{"title":"直したタイトル"}'

# 画像をアップロードして記事に添付
mtb assets upload --site-id 1 --file ./photo.jpg --path /2026
mtb sites entries update 1 42 --entry '{"assets":[{"id":"7"}]}'

# 再構築（静的公開のサイトはこれをしないと反映されない）
mtb publish entries --ids 42,43
```

パスパラメータはフラグでも位置引数でも渡せます（位置引数はパスの並び順）。

```bash
mtb sites entries get --site-id 1 --entry-id 42
mtb sites entries get 1 42          # 同じ
```

JSON はファイルや標準入力からも渡せます。送る前に構文検査します。

```bash
mtb sites entries create --site-id 1 --entry @entry.json
cat entry.json | mtb sites entries create --site-id 1 --entry -
```

### 共通フラグ

| フラグ | 用途 |
| --- | --- |
| `--dry-run` | 送らずに、送るはずだった内容を表示する |
| `-y, --yes` | 破壊的な操作（DELETE）の確認を省く |
| `--compact` | JSON を1行で出す |
| `--extract 項目` | 応答から1つの項目を生の値で取り出す（`jq -r` と同じ。`items.0.name` も辿れます） |
| `--set フラグ.項目=値` | リソース本体を項目ごとに組み立てる（値の `@ファイル名` でファイルから） |
| `--set-json` | `--set` と同じだが値を JSON として解釈する |
| `--lang ja\|en` | 表示言語（既定は環境変数から判定） |
| `--format raw` | 応答を無加工で出す |
| `--out FILE` | 応答をファイルに書き出す |
| `--all` | 全ページを取得して1つにまとめる（一覧コマンドのみ） |
| `--wait` | `restIds` が空になるまで呼び直す（`publish entries` などの再構築のみ） |
| `--verify-publish` | 保存できたかだけでなく、**静的ファイルを書き出せたか**まで MT のログで確かめる |
| `--dir DIR` | ディレクトリを丸ごと送る（アップロードのみ） |
| `--replace` | 送り先の同名アセットを消してから送る（アップロードのみ） |
| `--fields a,b` | 返す属性を絞る（**応答は既定で全属性。一覧では必ず使う**）。MT が未対応のエンドポイントでは手元で絞ります。1つも一致しなければ知らせます |
| `-v, --verbose` | 送受信の要約を標準エラーに出す |
| `--method-override` | PUT/DELETE を POST + `__method` で送る（前段が塞いでいるとき） |
| `--insecure` | TLS 証明書を検証しない（検証中サーバ向け） |
| `--audit-note 説明` | 何の作業かを `X-MTB-Audit-Note` ヘッダで送る（既定は `MTB_AUDIT_NOTE`） |

**DELETE は確認を挟みます。** 端末でない環境（CI・エージェント）では確認できないので、
`--yes` が無ければ実行せずにエラーになります。

**保存できたことと、サイトに出たことは別です。** MT は記事を保存すると同期で
再構築しますが、そこで失敗しても API は 200 を返します。しかも原因は、その記事とは
関係のないテンプレートのことがあります（アーカイブテンプレートが1つ壊れていると、
記事の個別アーカイブごと書き出されません）。手がかりは MT のログにしか残らないので、
**サイトに反映されたことまで確かめたいなら `--verify-publish` を付けてください**。
送信の前後でエラーログを見比べ、増えていればその内容を出して終了コード 8 で終わります。

```bash
mtb sites entries update 6 309 --entry '{"title":"…"}' --verify-publish
# => 保存はできましたが、静的ファイルの書き出しに失敗しています（MT のログに 1 件のエラー）。
#      Error during publishing: ... Cannot find template 'main_index'
```

**誰が何をしたかを、あとで見直せるようにしておけます。** mtb は `User-Agent: mtb/<版>` を
名乗り、`--audit-note`（または `MTB_AUDIT_NOTE`）の説明を `X-MTB-Audit-Note` ヘッダで
送ります（日本語は UTF-8 をパーセントエンコード）。サーバが監査ログを取っていれば
一緒に残ります（商用プラグイン MtbPro は Data API の全呼び出しを記録し、
`mtb audit-logs list --note …` で引けます）。取っていないサーバでは、ヘッダは無視されるだけです。

```bash
export MTB_AUDIT_NOTE="記事の移行（stg→本番）"
mtb sites entries create ...      # 以降の全リクエストに説明が付く
```

日本語はそのまま読める形で出ます。MT は `\uXXXX` にエスケープした JSON を
返しますが、mtb が戻しています（バイト列そのものが要るときは `--format raw`）。

### 終了コード

| コード | 意味 |
| --- | --- |
| 0 | 成功 |
| 1 | 分類できないエラー |
| 2 | 未認証・認証切れ（401） |
| 3 | 権限不足（403）— リトライしても直らない |
| 4 | 対象が無い（404） |
| 5 | 呼び出し方の誤り — **送信していない** |
| 6 | 接続できない・時間切れ — リトライの余地あり |
| 7 | サーバ側のエラー（5xx） |
| 8 | 保存はできたが書き出せなかった（`--verify-publish` / `--wait` のときだけ） |

### 定義に無いエンドポイント

埋め込んでいる定義は取り込み時点の v7 です。プラグインが足したものや MT 側の
新しい追加は生のリクエストで叩けます。

```bash
mtb endpoints list                       # そのサーバが実際に持つ一覧
mtb api GET /sites/1/entries --query limit=5
mtb api POST /sites/1/entries --form 'entry={"title":"x"}'
```

## AIエージェントから使う

`mtb llm` でAIエージェント向けの詳細リファレンスを出力します。認証・接続先の
切り替え・コマンド名の決まり方・代表的なワークフロー・失敗モードまでを網羅し、
コマンドカタログはコマンド定義から生成されるので実装と乖離しません。
バイナリに埋め込まれているのでオフラインでも動きます。

```bash
mtb llm                  # Markdown
mtb llm --format json    # 章ごとのJSON配列
mtb --llm                # 非推奨エイリアス。どの位置でも動作
```

Claude Code ではプラグインを導入すると6つのスキルが使えます。

| スキル | 役割 |
| --- | --- |
| `/mtb-install` | `mtb` を PATH に用意する |
| `/mtb-usage` | CLI で MT を読み書きする手順 |
| `/mtb-theme` | テーマを Git のファイルとして開発し、適用してサイトを起こす |
| `/mtb-migrate` | ある MT から別の MT へ、器を同期する。差分の検出から適用・確認まで |
| `/mt-cms-design` | 要件からサイトを設計して組み立てる。オブジェクトの選び分け・権限・アーカイブ |
| `/mt-template-tags` | テンプレートを書く。**全614タグと48グローバルモディファイア**の参照を同梱 |

後ろ2つが「MT で CMS を作る専門家」の中身です。タグの仕様は SKILL.md に
書かず `references/` に置いてあり、必要なものだけを開く形にしています。

```
/plugin marketplace add ideamans/claude-private-plugins
/plugin install movable-type-builder-cli@ideamans-private-plugins
```

同じスキルは Copilot や Cursor など Agent Skills 対応ホストでも使えます。

```bash
gh skill install ideamans/movable-type-builder-cli/plugins/movable-type-builder-cli/skills/mtb-usage --agent copilot
```

## 開発

```bash
make check      # go generate → gofmt → vet → build → test → git diff
```

構成は外部依存をすべてインタフェース越しにしてあります。

| パッケージ | 役割 | 差し替えるもの |
| --- | --- | --- |
| `internal/spec` | 絞り込み済みの API 定義（埋め込み） | — |
| `internal/gen-spec` | 公式 OpenAPI → `spec.json`。**コマンド名の決定はここだけ** | — |
| `internal/config` | プロファイルと認証情報 | `Storage` / `Environ` |
| `internal/api` | HTTP クライアントと認証 | `Doer` / `TokenStore` / `Now` |
| `internal/prompt` | 対話入力 | `Prompter` |
| `internal/cmd` | コマンドツリー | `Env`（上記すべて＋入出力＋ファイル） |

テストは実ファイル・実ネットワーク・実端末に一切触れません。

### Six Apart の資料の扱い

**このリポジトリは Six Apart の文章を持ちません。** Data API の公式定義
（movabletype/mt-docs-data-api-reference）にも movabletype.jp のマニュアルにも
再配布の許諾が無いためです。持つのは事実（パス・パラメータ・型・タグ名・属性の書き方）
と自前で書いた要約だけで、説明は公式のページへのリンクにしています。

| | 持つもの | 説明の置き場 |
| --- | --- | --- |
| Data API の定義（`internal/spec/spec.json`） | パス・メソッド・パラメータ・型・選択肢・題名 | `--help` から公式リファレンスへ |
| テンプレートタグ（`internal/mtdocs/`） | タグ名・種類・属性の書き方＋自前の要約（`summaries.json`） | 参照ファイルからマニュアルの各ページへ |

### 資料の更新（新しい MT に追随する）

配布スキルのテンプレートタグ資料も生成物です。

```bash
go run ./internal/gen-skilldocs -fetch    # movabletype.jp から事実だけを取り直す（数分）
go generate ./...
go test ./...                             # 新しいタグの要約が無ければ落ちる → summaries.json に書く
```

### API定義の更新

```bash
# internal/gen-spec/main.go の sourceCommit を新しいコミットに書き換えてから
make spec                                   # 公式定義を取って spec.json とカタログを作り直す
go test ./...
git diff --stat internal/spec/spec.json     # コマンドが増減していないか見る
```

公式定義はリポジトリに置かず、固定したコミットから取ってきます（キャッシュは
利用者のキャッシュディレクトリ）。`make check` はネットワークを使わないので、
spec.json の作り直しは `make spec` を別に流します。

`internal/gen-spec` は名前が衝突したら黙って畳まずに失敗します。落ちたら
overrides の表に理由つきで名前を足してください。

## ライセンス

MIT
