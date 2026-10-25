# コマンドカタログ

cobra のコマンド定義から `go generate ./...` で生成しています。
リソース操作のコマンドは internal/spec の OpenAPI 定義から作られるため、
直すのは定義か生成器であって、このファイルではありません。

## Global flags

| flag | type | default | description |
| --- | --- | --- | --- |
| `--api-version` | string | — | エンドポイントのバージョン（既定 v7） |
| `--audit-note` | string | — | 何の作業かの説明を X-MTB-Audit-Note ヘッダで送る（サーバが監査ログを取っていれば一緒に残る。既定は MTB_AUDIT_NOTE） |
| `--base-url` | string | — | mt-data-api.cgi までの URL（プロファイルより優先） |
| `--client-id` | string | — | 認証時に名乗る clientId（既定 mtb） |
| `--compact` | bool | `false` | JSON を1行で出す |
| `--dry-run` | bool | `false` | 送信せずに、送るはずだった内容を表示する |
| `--extract` | string | — | 応答から1つの項目だけを生の値で取り出す（text、items.0.name のように辿れます） |
| `--format` | string | `auto` | 出力形式 auto\|json\|raw |
| `--insecure` | bool | `false` | TLS 証明書を検証しない |
| `--lang` | string | — | 表示言語 ja\|en（既定は環境変数から判定） |
| `--method-override` | bool | `false` | PUT/DELETE を POST + __method で送る |
| `--out` | string | — | 応答をファイルに書き出す |
| `--profile` | string | — | 使うプロファイル名（既定は config.json の current） |
| `--timeout` | duration | `1m0s` | 1リクエストのタイムアウト |
| `-v`, `--verbose` | bool | `false` | 送受信の要約を標準エラーに出す |
| `--verify-publish` | bool | `false` | 保存できたかだけでなく、静的ファイルを書き出せたかまで MT のログで確かめる（失敗なら終了コード 8） |
| `-y`, `--yes` | bool | `false` | 破壊的な操作の確認を省く |
- `mtb api <メソッド> <パス>` `[--all --fields --file --form --no-auth --query --set --set-json]` — 任意のエンドポイントを直接叩く

## `mtb assets`

Assets

- `mtb assets upload` `[--auto-rename-if-exists --auto-rename-non-ascii --dir --fields --file --form --normalize-orientation --overwrite-once --path --query --replace --site-id]` — Upload a file

## `mtb auth`

認証（サインイン・状態確認・サインアウト）

- `mtb auth login` `[--no-remember --password --password-stdin --username]` — ユーザー名とパスワードでサインインする
- `mtb auth logout` — サインアウトしてセッションを無効にする
- `mtb auth refresh` — アクセストークンを取り直す
- `mtb auth status` — いまの接続先と認証状態を表示する
- `mtb call <エンドポイントID> [パスの値...]` `[--all --describe --fields --file --form --no-auth --path --query --set --set-json]` — エンドポイント ID で任意のエンドポイントを呼ぶ（拡張の無いプラグイン向け）
- `mtb dump` `[--asset-hash --dir --sites]` — MT の中身をディレクトリに YAML で書き出す（差分を取るため）

## `mtb endpoints`

Endpoints

- `mtb endpoints list` `[--exclude-components --fields --form --include-components --query]` — Retrieve a list of endpoints

## `mtb extensions`

mtb の拡張（MT プラグインが足したエンドポイントのコマンド）の管理

- `mtb extensions check [name]` — 拡張が要るエンドポイントが MT にあるか確かめる
- `mtb extensions install <directory>` `[--force]` — 拡張を入れる
- `mtb extensions list` — 拡張の一覧
- `mtb extensions show <name>` — 拡張の中身（コマンド・要る MT プラグイン・Claude Code のプラグイン）
- `mtb extensions uninstall <name>` — 拡張を外す
- `mtb extensions validate <directory>` `[--strict]` — 拡張のディレクトリを検査する（入れはしない）

## `mtb groups`

Groups

- `mtb groups create` `[--fields --form --group --query --set --set-json]` — Create a new group
- `mtb groups delete [group-id]` `[--fields --form --group-id --query]` — Delete a group
- `mtb groups get [group-id]` `[--fields --form --group-id --query]` — Retrieve single group by its ID
- `mtb groups list` `[--all --date-field --date-from --date-to --fields --filter-keys --form --limit --offset --query --search --search-fields --sort-by --sort-order]` — Retrieve a list of groups

### `mtb groups members`

Groups

- `mtb groups members create [group-id]` `[--fields --form --group-id --member --query --set --set-json]` — Add a user to a group
- `mtb groups members delete [group-id] [member-id]` `[--fields --form --group-id --member-id --query]` — Remove a user from a group
- `mtb groups members get [group-id] [member-id]` `[--fields --form --group-id --member-id --query]` — Retrieve single member by its ID for specified group
- `mtb groups members list [group-id]` `[--all --date-field --date-from --date-to --fields --filter-keys --form --group-id --limit --offset --query --search --search-fields --sort-by --sort-order]` — Retrieve a list of members for specified group

### `mtb groups permissions`

Permissions

- `mtb groups permissions grant [group-id]` `[--fields --form --group-id --query --role-id --site-id]` — Grant permissions to group
- `mtb groups permissions list [group-id]` `[--all --blog-ids --date-field --date-from --date-to --fields --filter-keys --form --group-id --limit --offset --query --sort-by --sort-order]` — Retrieve a list of permissions for group
- `mtb groups permissions revoke [group-id]` `[--fields --form --group-id --query --role-id --site-id]` — Revoke permissions from group
- `mtb groups synchronize` `[--fields --form --query]` — Synchronize groups
- `mtb groups update [group-id]` `[--fields --form --group --group-id --query --set --set-json]` — Update a group
- `mtb openapi` `[--fields --form --query]` — Retrieve OpenAPI schema

## `mtb permissions`

Permissions

- `mtb permissions list` `[--all --blog-ids --date-field --date-from --date-to --fields --filter-keys --form --limit --offset --query --sort-by --sort-order]` — Retrieve a list of permissions

## `mtb plugins`

Plugins

- `mtb plugins disable [plugin-id]` `[--fields --form --plugin-id --query]` — Disable all plugins
- `mtb plugins enable [plugin-id]` `[--fields --form --plugin-id --query]` — Enable all plugins
- `mtb plugins get [plugin-id]` `[--fields --form --plugin-id --query]` — Retrieve single plugin by its ID
- `mtb plugins list` `[--fields --form --query]` — Retrieve a list of plugins in the specified site

## `mtb profile`

接続先プロファイルの管理

- `mtb profile add <名前>` `[--site-id --use --username]` — プロファイルを追加する
- `mtb profile list` — プロファイルの一覧
- `mtb profile remove <名前>` — プロファイルと、その認証情報を消す
- `mtb profile set` `[--site-id --username]` — いまのプロファイルの設定を変える
- `mtb profile show` — いま使う接続先の解決結果を表示する
- `mtb profile use <名前>` — 既定のプロファイルを切り替える

## `mtb publish`

Publish

- `mtb publish content-data` `[--blog-id --blog-ids --fields --form --ids --query --start-time --wait]` — Publish Content Data
- `mtb publish entries` `[--blog-id --fields --form --ids --query --start-time --wait]` — Rebuild the static archives in relation to specified entries
- `mtb recover-password` `[--email --fields --form --name --query]` — Send the link for password recovery to specified email

## `mtb roles`

Roles

- `mtb roles create` `[--fields --form --query --role --set --set-json]` — Create a new role
- `mtb roles delete [role-id]` `[--fields --form --query --role-id]` — Delete an existing role
- `mtb roles get [role-id]` `[--fields --form --query --role-id]` — Retrieve a single role by its ID
- `mtb roles list` `[--all --date-field --date-from --date-to --fields --form --limit --offset --query --search --search-fields --sort-by --sort-order]` — Retrieve a list of roles

### `mtb roles permissions`

Permissions

- `mtb roles permissions list [role-id]` `[--all --blog-ids --date-field --date-from --date-to --fields --filter-keys --form --limit --offset --query --role-id --sort-by --sort-order]` — Retrieve a list of permissions by role
- `mtb roles update [role-id]` `[--fields --form --query --role --role-id --set --set-json]` — Update an existing role
- `mtb search` `[--all --blog-id --cd-search --class --exclude-blogs --fields --form --include-blogs --limit --offset --query --search --search-max-results --search-result-display --search-sort-by --tag --tag-search]` — Searching entries/pages/content data
- `mtb server-version` `[--fields --form --query]` — Get server API version

## `mtb sites`

Sites


### `mtb sites assets`

Assets

- `mtb sites assets delete [site-id] [asset-id]` `[--asset-id --fields --form --query --site-id]` — Delete an asset

#### `mtb sites assets entries`

Entries

- `mtb sites assets entries list [site-id] [asset-id]` `[--all --asset-id --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --max-comments --max-trackbacks --no-text-filter --offset --query --search --search-fields --site-id --sort-by --sort-order --status]` — Retrieve a list of entries that related with specific asset
- `mtb sites assets get [site-id] [asset-id]` `[--asset-id --fields --form --query --site-id]` — Retrieve single asset by its ID
- `mtb sites assets list [site-id]` `[--all --class --date-field --date-from --date-to --fields --filter-keys --form --limit --offset --query --related-assets --search --search-fields --site-id --sort-by --sort-order]` — Retrieve assets in the specified site

#### `mtb sites assets pages`

Pages

- `mtb sites assets pages list [site-id] [asset-id]` `[--all --asset-id --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --max-comments --max-trackbacks --no-text-filter --offset --query --search --search-fields --site-id --sort-by --sort-order --status]` — Retrieve a list of pages that related with specific asset
- `mtb sites assets thumbnail [site-id] [asset-id]` `[--asset-id --fields --form --height --query --scale --site-id --square --width]` — Get thumbnail of an asset
- `mtb sites assets update [site-id] [asset-id]` `[--asset --asset-id --fields --form --query --set --set-json --site-id]` — Update an asset
- `mtb sites assets upload [site-id]` `[--auto-rename-if-exists --dir --fields --file --form --normalize-orientation --path --query --replace --site-id]` — Upload a file
- `mtb sites backup [site-id]` `[--fields --form --query --site-id]` — Backup specified site

### `mtb sites categories`

Categories


#### `mtb sites categories children`

Categories

- `mtb sites categories children list [site-id] [category-id]` `[--category-id --fields --form --include-current --max-depth --query --site-id]` — Retrieve child categories from the specified category
- `mtb sites categories create [site-id]` `[--category --fields --form --query --set --set-json --site-id]` — Create a new category
- `mtb sites categories delete [site-id] [category-id]` `[--category-id --fields --form --query --site-id]` — Delete an existing category

#### `mtb sites categories entries`

Entries

- `mtb sites categories entries list [site-id] [category-id]` `[--all --category-id --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --max-comments --max-trackbacks --no-text-filter --offset --query --search --search-fields --site-id --sort-by --sort-order --status]` — Retrieve a list of entries by specific category
- `mtb sites categories get [site-id] [category-id]` `[--category-id --fields --form --query --site-id]` — Retrieve single category by its ID
- `mtb sites categories list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order --top]` — Retrieve categories in the specified site

#### `mtb sites categories parents`

Categories

- `mtb sites categories parents list [site-id] [category-id]` `[--category-id --fields --form --include-current --max-depth --query --site-id]` — Retrieve parent categories from the specified category
- `mtb sites categories permutate [site-id]` `[--categories --fields --form --query --set --set-json --site-id]` — Rearrange existing categories in a new order

#### `mtb sites categories siblings`

Categories

- `mtb sites categories siblings list [site-id] [category-id]` `[--all --category-id --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order --top]` — Retrieve siblings categories from the specified category
- `mtb sites categories update [site-id] [category-id]` `[--category --category-id --fields --form --query --set --set-json --site-id]` — Update an existing category

### `mtb sites category-sets`

Category Sets


#### `mtb sites category-sets categories`

Categories


##### `mtb sites category-sets categories children`

Categories

- `mtb sites category-sets categories children list [site-id] [category-set-id] [category-id]` `[--category-id --category-set-id --fields --form --include-current --max-depth --query --site-id]` — Category Collection of child categories for category set
- `mtb sites category-sets categories create [site-id] [category-set-id]` `[--category --category-set-id --fields --form --query --set --set-json --site-id]` — Create a new category for category set
- `mtb sites category-sets categories delete [site-id] [category-set-id] [category-id]` `[--category-id --category-set-id --fields --form --query --site-id]` — Delete single category in category set
- `mtb sites category-sets categories get [site-id] [category-set-id] [category-id]` `[--category-id --category-set-id --fields --form --query --site-id]` — Fetch single category in category set
- `mtb sites category-sets categories list [site-id] [category-set-id]` `[--all --category-set-id --date-field --date-from --date-to --fields --form --include-current --limit --max-depth --offset --query --search-fields --site-id --sort-by --sort-order]` — Category Collection for category set

##### `mtb sites category-sets categories parents`

Categories

- `mtb sites category-sets categories parents list [site-id] [category-set-id] [category-id]` `[--category-id --category-set-id --fields --form --include-current --max-depth --query --site-id]` — Category Collection of parent categories for category set
- `mtb sites category-sets categories permutate [site-id] [category-set-id]` `[--categories --category-set-id --fields --form --query --set --set-json --site-id]` — Save hierarchical categories order in category set

##### `mtb sites category-sets categories siblings`

Categories

- `mtb sites category-sets categories siblings list [site-id] [category-set-id] [category-id]` `[--all --category-id --category-set-id --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order --top]` — Category Collection of sibling categories for category set
- `mtb sites category-sets categories update [site-id] [category-set-id] [category-id]` `[--category --category-id --category-set-id --fields --form --query --set --set-json --site-id]` — Update single category in category set
- `mtb sites category-sets create [site-id]` `[--category-set --fields --form --query --set --set-json --site-id]` — Create a new category set
- `mtb sites category-sets delete [site-id] [category-set-id]` `[--category-set-id --fields --form --query --site-id]` — Delete category set
- `mtb sites category-sets get [site-id] [category-set-id]` `[--category-set-id --fields --form --query --site-id]` — Fetch single category set
- `mtb sites category-sets list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order]` — Category Set Collection
- `mtb sites category-sets update [site-id] [category-set-id]` `[--category-set --category-set-id --fields --form --query --set --set-json --site-id]` — Update category set

### `mtb sites children`

Sites

- `mtb sites children list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order]` — Retrieve sites by parent ID

### `mtb sites comments`

Comments

- `mtb sites comments delete [site-id] [comment-id]` `[--comment-id --fields --form --query --site-id]` — Delete a comment
- `mtb sites comments get [site-id] [comment-id]` `[--comment-id --fields --form --query --site-id]` — Retrieve a single comment by its ID
- `mtb sites comments list [site-id]` `[--all --date-field --date-from --date-to --entry-status --exclude-ids --fields --filter-keys --form --include-ids --limit --offset --query --search-fields --site-id --sort-by --sort-order --status]` — Retrieve a list of comments
- `mtb sites comments update [site-id] [comment-id]` `[--comment --comment-id --fields --form --query --set --set-json --site-id]` — Update a comment

### `mtb sites content-types`

Content Types

- `mtb sites content-types create [site-id]` `[--content-type --fields --form --query --set --set-json --site-id]` — Create Content Type

#### `mtb sites content-types data`

Content Types

- `mtb sites content-types data create [site-id] [content-type-id]` `[--content-data --content-type-id --fields --form --query --save-revision --set --set-json --site-id]` — Create Content Data
- `mtb sites content-types data delete [site-id] [content-type-id] [content-data-id]` `[--content-data-id --content-type-id --fields --form --query --site-id]` — Delete single content data
- `mtb sites content-types data get [site-id] [content-type-id] [content-data-id]` `[--content-data-id --content-type-id --fields --form --no-text-filter --query --site-id]` — Fetch single Content Data
- `mtb sites content-types data list [site-id] [content-type-id]` `[--all --content-type-id --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --no-text-filter --offset --query --search-fields --site-id --sort-by --sort-order]` — Content Data Collection
- `mtb sites content-types data preview [site-id] [content-type-id] [content-data-id]` `[--content-data --content-data-id --content-type-id --fields --form --query --raw --set --set-json --site-id]` — Make a preview by data
- `mtb sites content-types data update [site-id] [content-type-id] [content-data-id]` `[--content-data --content-data-id --content-type-id --fields --form --query --save-revision --set --set-json --site-id]` — Update Content Data
- `mtb sites content-types delete [site-id] [content-type-id]` `[--content-type-id --fields --form --query --site-id]` — Delete Content Type

#### `mtb sites content-types fields`

Content Types

- `mtb sites content-types fields create [site-id] [content-type-id]` `[--content-field --content-type-id --fields --form --query --set --set-json --site-id]` — Create Content Field
- `mtb sites content-types fields delete [site-id] [content-type-id] [content-field-id]` `[--content-field-id --content-type-id --fields --form --query --site-id]` — Delete Content Field
- `mtb sites content-types fields get [site-id] [content-type-id] [content-field-id]` `[--content-field-id --content-type-id --fields --form --query --site-id]` — Fetch single Content Field
- `mtb sites content-types fields list [site-id] [content-type-id]` `[--all --content-type-id --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search-fields --site-id --sort-by --sort-order]` — Content Field Collection
- `mtb sites content-types fields permutate [site-id] [content-type-id]` `[--content-fields --content-type-id --fields --form --query --set --set-json --site-id]` — Save hierarchical content field order
- `mtb sites content-types fields update [site-id] [content-type-id] [content-field-id]` `[--content-field --content-field-id --content-type-id --fields --form --query --set --set-json --site-id]` — Update Content Field
- `mtb sites content-types get [site-id] [content-type-id]` `[--content-type-id --fields --form --query --site-id]` — Fetch single Content Type
- `mtb sites content-types list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search-fields --site-id --sort-by --sort-order]` — Content Type Collection
- `mtb sites content-types update [site-id] [content-type-id]` `[--content-type --content-type-id --fields --form --query --set --set-json --site-id]` — Update Content Type
- `mtb sites create` `[--fields --form --query --set --set-json --website]` — Create a new website
- `mtb sites create-blog [site-id]` `[--blog --fields --form --query --set --set-json --site-id]` — Create a new blog
- `mtb sites delete [site-id]` `[--confirm --fields --form --query --site-id]` — Delete an existing blog or website

### `mtb sites entries`

Entries


#### `mtb sites entries assets`

Assets

- `mtb sites entries assets list [site-id] [entry-id]` `[--all --class --date-field --date-from --date-to --entry-id --fields --filter-keys --form --limit --offset --query --search-fields --site-id --sort-by --sort-order]` — Retrieve assets that related with specified entry

#### `mtb sites entries categories`

Categories

- `mtb sites entries categories list [site-id] [entry-id]` `[--all --date-field --date-from --date-to --entry-id --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order --top --type]` — Retrieve categories in the specified entry

#### `mtb sites entries comments`

Comments

- `mtb sites entries comments create [site-id] [entry-id]` `[--comment --entry-id --fields --form --query --set --set-json --site-id]` — Create a new comment on an entry
- `mtb sites entries comments list [site-id] [entry-id]` `[--all --date-field --date-from --date-to --entry-id --fields --filter-keys --form --limit --offset --query --search-fields --site-id --sort-by --sort-order]` — Retrieve a list of comments for an entry

##### `mtb sites entries comments replies`

Comments

- `mtb sites entries comments replies create [site-id] [entry-id] [comment-id]` `[--comment --comment-id --entry-id --fields --form --query --set --set-json --site-id]` — Reply to specified comment
- `mtb sites entries create [site-id]` `[--entry --fields --form --publish --query --save-revision --set --set-json --site-id]` — Create a new entry
- `mtb sites entries delete [site-id] [entry-id]` `[--entry-id --fields --form --query --site-id]` — Delete an entry
- `mtb sites entries export [site-id]` `[--fields --form --query --site-id]` — Export entries
- `mtb sites entries get [site-id] [entry-id]` `[--entry-id --fields --form --max-comments --max-trackbacks --no-text-filter --query --site-id]` — Retrieve a single entry by its ID
- `mtb sites entries import [site-id]` `[--convert-breaks --default-cat-id --default-status --dir --encoding --fields --file --form --import-as-me --import-type --password --query --site-id]` — Import entries
- `mtb sites entries list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --max-comments --max-trackbacks --no-text-filter --offset --query --search --search-fields --site-id --sort-by --sort-order --status]` — Retrieve a list of entries
- `mtb sites entries preview [site-id] [entry-id]` `[--entry --entry-id --fields --form --query --raw --set --set-json --site-id]` — Make a preview for a entry
- `mtb sites entries update [site-id] [entry-id]` `[--entry --entry-id --fields --form --publish --query --save-revision --set --set-json --site-id]` — Update an entry
- `mtb sites export-theme [site-id]` `[--description --fields --form --include --include-all --output --overwrite-yes --query --set --set-json --site-id --theme-author-link --theme-author-name --theme-id --theme-name --theme-version]` — Export site's theme

### `mtb sites fields`

CustomFields

- `mtb sites fields create [site-id]` `[--field --fields --form --query --set --set-json --site-id]` — Create a new field
- `mtb sites fields delete [site-id] [field-id]` `[--field-id --fields --form --query --site-id]` — Delete a field
- `mtb sites fields get [site-id] [field-id]` `[--field-id --fields --form --query --site-id]` — Retrieve single field by its ID
- `mtb sites fields list [site-id]` `[--all --date-field --date-from --date-to --fields --filter-keys --form --limit --offset --query --search --search-fields --site-id --sort-by --sort-order]` — Retrieve a list of fields in the specified site
- `mtb sites fields update [site-id] [field-id]` `[--field --field-id --fields --form --query --set --set-json --site-id]` — Update a field

### `mtb sites folders`

Folders


#### `mtb sites folders children`

Folders

- `mtb sites folders children list [site-id] [folder-id]` `[--fields --folder-id --form --include-current --max-depth --query --site-id]` — Retrieve a list of child folders of the requested folder
- `mtb sites folders create [site-id]` `[--fields --folder --form --query --set --set-json --site-id]` — Create a new folder.
- `mtb sites folders delete [site-id] [folder-id]` `[--fields --folder-id --form --query --site-id]` — Delete an existing folder
- `mtb sites folders get [site-id] [folder-id]` `[--fields --folder-id --form --query --site-id]` — Retrieve single folder by its ID
- `mtb sites folders list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order --top]` — Retrieve a list of folders

#### `mtb sites folders pages`

Pages

- `mtb sites folders pages list [site-id] [folder-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --filter-keys --folder-id --form --include-ids --limit --max-comments --max-trackbacks --no-text-filter --offset --query --search --search-fields --site-id --sort-by --sort-order --status]` — Retrieve a list of pages by specific folder

#### `mtb sites folders parents`

Folders

- `mtb sites folders parents list [site-id] [folder-id]` `[--fields --folder-id --form --include-current --max-depth --query --site-id]` — Retrieve a list of parent folders of the requested folder
- `mtb sites folders permutate [site-id]` `[--fields --folders --form --query --set --set-json --site-id]` — Rearrange existing folders in a new order

#### `mtb sites folders siblings`

Folders

- `mtb sites folders siblings list [site-id] [folder-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --folder-id --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order --top]` — Retrieve a list of sibling folders of the requested folder
- `mtb sites folders update [site-id] [folder-id]` `[--fields --folder --folder-id --form --query --set --set-json --site-id]` — Update an existing folder

### `mtb sites formatted-texts`

FormattedText

- `mtb sites formatted-texts create [site-id]` `[--fields --form --formatted-text --query --set --set-json --site-id]` — Create a new formatted_text
- `mtb sites formatted-texts delete [site-id] [formatted-text-id]` `[--fields --form --formatted-text-id --query --site-id]` — Delete a formatted_text
- `mtb sites formatted-texts get [site-id] [formatted-text-id]` `[--fields --form --formatted-text-id --query --site-id]` — Retrieve single formatted_text by its ID
- `mtb sites formatted-texts list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order]` — Retrieve a list of formatted_texts in the specified site
- `mtb sites formatted-texts update [site-id] [formatted-text-id]` `[--fields --form --formatted-text --formatted-text-id --query --set --set-json --site-id]` — Update a formatted_text
- `mtb sites get [site-id]` `[--fields --form --query --site-id]` — Retrieve a single blog by its ID
- `mtb sites list` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --sort-by --sort-order]` — Retrieve sites

### `mtb sites logs`

Logs

- `mtb sites logs create [site-id]` `[--fields --form --log --query --set --set-json --site-id]` — Create a new log
- `mtb sites logs delete [site-id] [log-id]` `[--fields --form --log-id --query --site-id]` — Delete an existing log
- `mtb sites logs export [site-id]` `[--encoding --fields --form --query --site-id]` — Export logs
- `mtb sites logs get [site-id] [log-id]` `[--fields --form --log-id --query --site-id]` — Retrieve a single log by its ID
- `mtb sites logs list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --level --limit --offset --query --search --search-fields --site-id --sort-by --sort-order]` — Retrieve a list of logs in the specified site
- `mtb sites logs reset [site-id]` `[--confirm --fields --form --query --site-id]` — Reset logs
- `mtb sites logs update [site-id] [log-id]` `[--fields --form --log --log-id --query --set --set-json --site-id]` — Update an existing log

### `mtb sites pages`

Pages


#### `mtb sites pages assets`

Assets

- `mtb sites pages assets list [site-id] [page-id]` `[--all --class --date-field --date-from --date-to --fields --filter-keys --form --limit --offset --page-id --query --search-fields --site-id --sort-by --sort-order]` — Retrieve assets that related with specified page

#### `mtb sites pages comments`

Comments

- `mtb sites pages comments create [site-id] [page-id]` `[--comment --fields --form --page-id --query --set --set-json --site-id]` — Create a new comment on a page
- `mtb sites pages comments list [site-id] [page-id]` `[--all --date-field --date-from --date-to --fields --filter-keys --form --limit --offset --page-id --query --search-fields --site-id --sort-by --sort-order]` — Retrieve a list of comments for the specified page

##### `mtb sites pages comments replies`

Comments

- `mtb sites pages comments replies create [site-id] [page-id] [comment-id]` `[--comment --comment-id --fields --form --page-id --query --set --set-json --site-id]` — Reply to specified comment
- `mtb sites pages create [site-id]` `[--fields --form --page --publish --query --save-revision --set --set-json --site-id]` — Create a new page
- `mtb sites pages delete [site-id] [page-id]` `[--fields --form --page-id --query --site-id]` — Delete an existing page
- `mtb sites pages get [site-id] [page-id]` `[--fields --form --no-text-filter --page-id --query --site-id]` — Retrieve a single page by its ID
- `mtb sites pages list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --max-comments --max-trackbacks --no-text-filter --offset --query --search --search-fields --site-id --sort-by --sort-order --status]` — Retrieve a list of pages in the specified site
- `mtb sites pages preview [site-id] [page-id]` `[--fields --form --page --page-id --query --raw --set --set-json --site-id]` — Make a preview for a page
- `mtb sites pages update [site-id] [page-id]` `[--fields --form --page --page-id --publish --query --save-revision --set --set-json --site-id]` — Update an existing page

### `mtb sites permissions`

Permissions

- `mtb sites permissions grant [site-id]` `[--fields --form --query --role-id --site-id --user-id]` — Grant permissions to site
- `mtb sites permissions list [site-id]` `[--all --date-field --date-from --date-to --fields --form --limit --offset --query --site-id --sort-by --sort-order]` — Retrieve a list of permissions for site
- `mtb sites permissions revoke [site-id]` `[--fields --form --query --role-id --site-id --user-id]` — Revoke permissions from site
- `mtb sites refresh-templates [site-id]` `[--confirm --fields --form --query --refresh-type --site-id]` — Reset all templates in the site

### `mtb sites stats`

Statistics


#### `mtb sites stats date`

Statistics


##### `mtb sites stats date pageviews`

Statistics

- `mtb sites stats date pageviews list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date --unique-path]` — Retrieve screenPageViews count for each date from provider (e.g. Google Analytics V4)

##### `mtb sites stats date screen-page-views`

Statistics

- `mtb sites stats date screen-page-views list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date --unique-path]` — Retrieve screenPageViews count for each date from provider (e.g. Google Analytics V4)

##### `mtb sites stats date sessions`

Statistics

- `mtb sites stats date sessions list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date]` — Retrieve sessions count for each date from provider (e.g. Google Analytics V4)

##### `mtb sites stats date visits`

Statistics

- `mtb sites stats date visits list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date]` — Retrieve sessions count for each date from provider (e.g. Google Analytics V4)

#### `mtb sites stats path`

Statistics


##### `mtb sites stats path pageviews`

Statistics

- `mtb sites stats path pageviews list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date --unique-path]` — Retrieve screenPageViews count for each path from provider (e.g. Google Analytics V4)

##### `mtb sites stats path screen-page-views`

Statistics

- `mtb sites stats path screen-page-views list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date --unique-path]` — Retrieve screenPageViews count for each path from provider (e.g. Google Analytics V4)

##### `mtb sites stats path sessions`

Statistics

- `mtb sites stats path sessions list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date]` — Retrieve sessions count for each path from provider (e.g. Google Analytics V4)

##### `mtb sites stats path visits`

Statistics

- `mtb sites stats path visits list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date]` — Retrieve sessions count for each path from provider (e.g. Google Analytics V4)
- `mtb sites stats provider [site-id]` `[--fields --form --query --site-id]` — Retrieve a current effective provider

#### `mtb sites stats year`

Statistics


##### `mtb sites stats year screen-page-views`

Statistics

- `mtb sites stats year screen-page-views list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date --unique-path]` — Retrieve screenPageViews count for each year from provider (e.g. Google Analytics V4)

##### `mtb sites stats year sessions`

Statistics

- `mtb sites stats year sessions list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date]` — Retrieve sessions count for each year from provider (e.g. Google Analytics V4)

#### `mtb sites stats year-month`

Statistics


##### `mtb sites stats year-month screen-page-views`

Statistics

- `mtb sites stats year-month screen-page-views list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date --unique-path]` — Retrieve screenPageViews count for each yearMonth from provider (e.g. Google Analytics V4)

##### `mtb sites stats year-month sessions`

Statistics

- `mtb sites stats year-month sessions list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date]` — Retrieve sessions count for each yearMonth from provider (e.g. Google Analytics V4)

#### `mtb sites stats year-week`

Statistics


##### `mtb sites stats year-week screen-page-views`

Statistics

- `mtb sites stats year-week screen-page-views list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date --unique-path]` — Retrieve screenPageViews count for each yearWeek from provider (e.g. Google Analytics V4)

##### `mtb sites stats year-week sessions`

Statistics

- `mtb sites stats year-week sessions list [site-id]` `[--all --end-date --fields --form --limit --offset --page-path --query --site-id --start-date]` — Retrieve sessions count for each yearWeek from provider (e.g. Google Analytics V4)

### `mtb sites tags`

Tags


#### `mtb sites tags assets`

Assets

- `mtb sites tags assets list [site-id] [tag-id]` `[--all --class --date-field --date-from --date-to --fields --filter-keys --form --limit --offset --query --search-fields --site-id --sort-by --sort-order --tag-id]` — Retrieve assets that related with specified tag
- `mtb sites tags delete [site-id] [tag-id]` `[--fields --form --query --site-id --tag-id]` — Delete an existing tag

#### `mtb sites tags entries`

Entries

- `mtb sites tags entries list [site-id] [tag-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --max-comments --max-trackbacks --no-text-filter --offset --query --search --search-fields --site-id --sort-by --sort-order --status --tag-id]` — Retrieve a list of entries that related with specific tag
- `mtb sites tags get [site-id] [tag-id]` `[--fields --form --query --site-id --tag-id]` — Retrieve a single tag by its ID
- `mtb sites tags list [site-id]` `[--all --date-field --date-from --date-to --fields --form --limit --offset --query --search --search-fields --site-id --sort-by --sort-order]` — Retrieve a list of tags that related with specific site

#### `mtb sites tags pages`

Pages

- `mtb sites tags pages list [site-id] [tag-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --max-comments --max-trackbacks --no-text-filter --offset --query --search --search-fields --site-id --sort-by --sort-order --status --tag-id]` — Retrieve a list of pages that related with specific tag.
- `mtb sites tags update [site-id] [tag-id]` `[--fields --form --query --set --set-json --site-id --tag --tag-id]` — Update an existing tag

### `mtb sites templates`

Templates

- `mtb sites templates clone [site-id] [template-id]` `[--fields --form --query --site-id --template-id]` — Make a clone of a template
- `mtb sites templates create [site-id]` `[--fields --form --query --save-revision --set --set-json --site-id --template]` — Create a new template
- `mtb sites templates delete [site-id] [template-id]` `[--fields --form --query --site-id --template-id]` — Delete a template
- `mtb sites templates get [site-id] [template-id]` `[--fields --form --query --site-id --template-id]` — Retrieve single template by its ID
- `mtb sites templates list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order --type]` — Retrieve a list of templates in the specified site
- `mtb sites templates preview [site-id] [template-id]` `[--fields --form --query --raw --set --set-json --site-id --template --template-id]` — Make a preview for a template
- `mtb sites templates publish [site-id] [template-id]` `[--fields --form --query --site-id --template-id]` — Publish a template
- `mtb sites templates pull <サイトID>` `[--dir]` — テンプレートをディレクトリに書き出す
- `mtb sites templates push <サイトID>` `[--dir]` — ディレクトリのテンプレートを MT に戻す
- `mtb sites templates refresh [site-id] [template-id]` `[--fields --form --query --site-id --template-id]` — Reset template text to theme default or tempalte_set default

#### `mtb sites templates templatemaps`

Templates

- `mtb sites templates templatemaps create [site-id] [template-id]` `[--fields --form --query --set --set-json --site-id --template-id --templatemap]` — Create a new templatemap
- `mtb sites templates templatemaps delete [site-id] [template-id] [templatemap-id]` `[--fields --form --query --site-id --template-id --templatemap-id]` — Delete an existing templatemap
- `mtb sites templates templatemaps get [site-id] [template-id] [templatemap-id]` `[--fields --form --query --site-id --template-id --templatemap-id]` — Retrieve a single templatemap by its ID
- `mtb sites templates templatemaps list [site-id] [template-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order --template-id]` — Retrieve a list of templatemaps in the specified site
- `mtb sites templates templatemaps update [site-id] [template-id] [templatemap-id]` `[--fields --form --query --set --set-json --site-id --template-id --templatemap --templatemap-id]` — Update an existing templatemap
- `mtb sites templates update [site-id] [template-id]` `[--fields --form --query --save-revision --set --set-json --site-id --template --template-id]` — Update a template

### `mtb sites themes`

Themes

- `mtb sites themes apply [site-id] [theme-id]` `[--confirm --fields --form --query --site-id --theme-id]` — Apply a theme to site
- `mtb sites themes get [site-id] [theme-id]` `[--fields --form --query --site-id --theme-id]` — Retrieve a single theme by its ID for site
- `mtb sites themes list [site-id]` `[--fields --form --query --site-id]` — Retrieve a list of themes for site
- `mtb sites update [site-id]` `[--blog --fields --form --query --set --set-json --site-id --website]` — Update an existing blog or website

### `mtb sites widgets`

Widgets

- `mtb sites widgets clone [site-id] [widget-id]` `[--fields --form --query --site-id --widget-id]` — Make a clone of a widget
- `mtb sites widgets create [site-id]` `[--fields --form --query --set --set-json --site-id --widget]` — Create a new widget
- `mtb sites widgets delete [site-id] [widget-id]` `[--fields --form --query --site-id --widget-id]` — Delete a widget
- `mtb sites widgets get [site-id] [widget-id]` `[--fields --form --query --site-id --widget-id]` — Retrieve a single widget by its ID
- `mtb sites widgets list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order]` — Retrieve a list of widgets in the specified site
- `mtb sites widgets refresh [site-id] [widget-id]` `[--fields --form --query --site-id --widget-id]` — Reset widget text to default
- `mtb sites widgets update [site-id] [widget-id]` `[--fields --form --query --set --set-json --site-id --widget --widget-id]` — Update a widget

### `mtb sites widgetsets`

WidgetSets

- `mtb sites widgetsets create [site-id]` `[--fields --form --query --set --set-json --site-id --widgetset]` — Create a new widgetset
- `mtb sites widgetsets delete [site-id] [widgetset-id]` `[--fields --form --query --site-id --widgetset-id]` — Delete an existing widgetset
- `mtb sites widgetsets get [site-id] [widgetset-id]` `[--fields --form --query --site-id --widgetset-id]` — Retrieve a single widgetset by its ID
- `mtb sites widgetsets list [site-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order]` — Retrieve a list of widgetsets in the specified site
- `mtb sites widgetsets update [site-id] [widgetset-id]` `[--fields --form --query --set --set-json --site-id --widgetset --widgetset-id]` — Update an existing widgetset

#### `mtb sites widgetsets widgets`

Widgets

- `mtb sites widgetsets widgets get [site-id] [widgetset-id] [widget-id]` `[--fields --form --query --site-id --widget-id --widgetset-id]` — Retrieve a single widget by widgetset ID
- `mtb sites widgetsets widgets list [site-id] [widgetset-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --search --search-fields --site-id --sort-by --sort-order --widgetset-id]` — Retrieve a list of widgets in the specified widgetset

## `mtb text-filters`

System

- `mtb text-filters list` `[--fields --form --query]` — Retrieve a list of text filters

## `mtb themes`

Themes

- `mtb themes delete [theme-id]` `[--fields --form --query --theme-id]` — Uninstall a specified theme from the MT
- `mtb themes get [theme-id]` `[--fields --form --query --theme-id]` — Retrieve a single theme by its ID
- `mtb themes list` `[--fields --form --query]` — Retrieve a list of themes

## `mtb users`

Users

- `mtb users create` `[--fields --form --query --set --set-json --user]` — Create a new user
- `mtb users delete [user-id]` `[--fields --form --query --user-id]` — Delete user
- `mtb users export` `[--fields --form --query]` — Export users
- `mtb users get [user-id]` `[--fields --form --query --user-id]` — Retrieve a single user by its ID

### `mtb users groups`

Groups

- `mtb users groups list [user-id]` `[--all --date-field --date-from --date-to --fields --filter-keys --form --limit --offset --query --search --search-fields --sort-by --sort-order --user-id]` — Retrieve a list of groups
- `mtb users import` `[--dir --fields --file --form --query]` — Import users
- `mtb users list` `[--all --date-field --date-from --date-to --exclude-ids --fields --filter-keys --form --include-ids --limit --lockout --offset --query --search --search-fields --sort-by --sort-order --status]` — Retrieve a list of users in the specified site

### `mtb users permissions`

Permissions

- `mtb users permissions grant [user-id]` `[--fields --form --query --role-id --site-id --user-id]` — Grant permissions to user
- `mtb users permissions list [user-id]` `[--all --blog-ids --date-field --date-from --date-to --fields --filter-keys --form --limit --offset --query --sort-by --sort-order --user-id]` — Retrieve a list of permissions for user
- `mtb users permissions revoke [user-id]` `[--fields --form --query --role-id --site-id --user-id]` — Revoke permissions from user
- `mtb users recover-password [user-id]` `[--fields --form --query --user-id]` — Send the link for password recovery to specified user by email

### `mtb users sites`

Sites

- `mtb users sites list [user-id]` `[--all --date-field --date-from --date-to --exclude-ids --fields --form --include-ids --limit --offset --query --sort-by --sort-order --user-id]` — Retrieve a list of blogs by user
- `mtb users synchronize` `[--fields --form --query]` — Synchronize users
- `mtb users unlock [user-id]` `[--fields --form --query --user-id]` — Unlock user account
- `mtb users update [user-id]` `[--fields --form --query --set --set-json --user --user-id]` — Update user data
