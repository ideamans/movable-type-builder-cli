# コマンド名とエンドポイントの対応

コマンドは公式の OpenAPI 定義（v7）から機械的に作られている。
**規則を覚えれば、リファレンスのパスからコマンド名を導ける**（逆も同じ）。

## 規則

1. パスの**リテラルなセグメント**がそのままコマンドのグループになる。
   キャメルケースとスネークケースはケバブケースにする
   （`contentTypes` → `content-types`、`formatted_texts` → `formatted-texts`）。
2. パスパラメータ（`{site_id}` など）はコマンド名に**現れない**。
   フラグ（`--site-id`）か位置引数で渡す。
3. 末尾が動詞なら、その動詞がコマンド名になる。
   （`upload` `export` `import` `preview` `permutate` `grant` `revoke` `clone`
   `publish` `refresh` `apply` `unlock` `enable` `disable` `synchronize`
   `recover_password` `backup` `thumbnail` `search` など）
4. そうでなければ HTTP メソッドから動詞を決める。

| パスの末尾 | GET | POST | PUT | DELETE |
| --- | --- | --- | --- | --- |
| リテラル（コレクション） | `list` | `create` | `update` | `delete` |
| パラメータ（個別） | `get` | `create` | `update` | `delete` |

```
GET    /v7/sites/{site_id}/entries                        mtb sites entries list
POST   /v7/sites/{site_id}/entries                        mtb sites entries create
GET    /v7/sites/{site_id}/entries/{entry_id}             mtb sites entries get
PUT    /v7/sites/{site_id}/entries/{entry_id}             mtb sites entries update
DELETE /v7/sites/{site_id}/entries/{entry_id}             mtb sites entries delete
GET    /v7/sites/{site_id}/entries/export                 mtb sites entries export
POST   /v7/assets/upload                                  mtb assets upload
GET    /v7/sites/{site_id}/categories/{id}/children       mtb sites categories children list
POST   /v7/sites/{site_id}/themes/{theme_id}/apply        mtb sites themes apply
POST   /v7/groups/{group_id}/permissions/grant            mtb groups permissions grant
```

## エンドポイントではないコマンド

以下は**定義に無い**。Data API に無い操作を mtb が組み立てているもので、
リファレンスを探しても対応するパスは見つからない。

| コマンド | 中身 |
| --- | --- |
| `mtb auth login` / `logout` / `refresh` / `status` | `/authentication` `/token` を包んだもの |
| `mtb profile ...` | 接続先の設定。MT には送らない |
| `mtb api <メソッド> <パス>` | 任意のエンドポイントを直接叩く逃げ道 |
| `mtb call <エンドポイントID>` | エンドポイント ID で呼ぶ（経路はサーバの `/endpoints` から。`--describe` でパラメータ） |
| `mtb llm` | 埋め込みリファレンスの表示 |
| `mtb sites templates pull` / `push` | 一覧と更新を回してディレクトリと往復させる |
| `mtb extensions ...` | mtb の拡張の管理。MT には送らない（`check` だけが `/endpoints` を読む） |

## 拡張が足すコマンド

拡張（`mtb extensions`）のコマンドも**同じ規則で名前を付ける**。拡張の定義は
手で書くので規則を強制はしないが、外れると利用者が名前を導けなくなる。

```
PUT    /v7/sites/{site_id}/entries/{entry_id}/author      mtb sites entries author update
PUT    /v7/sites/{site_id}/pages/{page_id}/author         mtb sites pages author update
```

既存のコマンド（生成・手書き・別名のどれでも）とぶつかる拡張は読み込まない。

## 例外

規則どおりだと意味を取り違える、または名前が衝突する7つだけ明示している。

| エンドポイント | コマンド | 理由 |
| --- | --- | --- |
| `GET /` | `mtb openapi` | サーバ自身の OpenAPI 定義を返す |
| `GET /version` | `mtb server-version` | CLI 自身の `--version` と紛らわしい |
| `POST /sites/{site_id}` | `mtb sites create-blog` | ブログ作成。`POST /sites`（ウェブサイト作成）と衝突する |
| `DELETE /sites/{site_id}/logs` | `mtb sites logs reset` | 全ログ削除。1件削除と同名になると事故る |
| `GET /publish/entries` | `mtb publish entries` | GET だが再構築の実行。一覧ではない |
| `GET /publish/contentData` | `mtb publish content-data` | 同上 |
| `GET /sites/{site_id}/stats/provider` | `mtb sites stats provider` | 1つ返す。コレクションではない |

## 末尾のIDがあってもなくても同じ操作

MT には「IDを取る版と取らない版」が対になっているエンドポイントがある。
これらは**1つのコマンドに畳んである**。IDを渡せばID付きの方、省けば無い方を呼ぶ。

| コマンド | IDあり | IDなし |
| --- | --- | --- |
| `mtb plugins enable [plugin-id]` | そのプラグインを有効化 | 全プラグインを有効化 |
| `mtb plugins disable [plugin-id]` | そのプラグインを無効化 | 全プラグインを無効化 |
| `mtb sites entries preview [site-id] [entry-id]` | 既存記事のプレビュー | 未保存の記事のプレビュー |
| `mtb sites pages preview [site-id] [page-id]` | 同上（ページ） | 同上 |
| `mtb sites templates preview [site-id] [template-id]` | 同上（テンプレート） | 同上 |
| `mtb sites content-types data preview [site-id] [content-type-id] [content-data-id]` | 同上 | 同上 |

**IDを省くと挙動が変わるものがある**（`plugins enable` は全プラグインが対象に
なる）。省略する意図がないなら必ずIDを渡すこと。

## 認証エンドポイントは生成していない

`/authentication` `/token` `/authorization` は `mtb auth` が受け持つ。
生のリクエストで叩くと、保存済みの認証情報と食い違うので使わないこと。

## 単数形の別名

複数形のグループには単数形の別名がある。打ちやすさのためだけのもので、
意味は同じ。

```bash
mtb sites entries list --site-id 1
mtb site entry list --site-id 1        # 同じ
```
