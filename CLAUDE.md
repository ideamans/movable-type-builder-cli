# CLAUDE.md — movable-type-builder-cli

Movable Type のサイトを組み立てる CLI（Movable Type Builder CLI）。
土台は Data API (v7)。**バイナリ名は `mtb`**（リポジトリ名と
リリースアーカイブ名は `movable-type-builder-cli`。main パッケージは
`cmd/mtb` にあるので `go install .../cmd/mtb@latest` は `mtb` を作る）。

**リポジトリは private。** context7 は使えないので、`internal/llmdocs/00-guide.md`
を厚くして `mtb llm` 一本で完結させている。プラグインの登録先も
`ideamans/claude-private-plugins`。

実在の MT サイトの記事・ページ・テンプレートを書き換える。DELETE は取り消せない。

## この CLI の要（先に知っておくこと）

**コマンドツリーは手書きではない。** 公式の OpenAPI 定義（固定したコミットから
取ってくる。リポジトリには置かない）を `internal/gen-spec` が
`internal/spec/spec.json` に絞り込み、`internal/cmd/generated.go` が実行時に
cobra のツリーを組み立てる。生成されるのは 220 コマンド / 226 オペレーション。

**Six Apart の文章を持ち込まない。** 公式定義にも movabletype.jp のマニュアルにも
再配布の許諾が無い。spec.json に残すのは事実（パス・パラメータ・型・選択肢）と
操作の短い題名だけで、説明は公式リファレンスへのリンク（`docUrl`）にする。
テンプレートタグの資料も、事実と自前の要約（`internal/mtdocs/summaries.json`）と
マニュアルへのリンクだけ。**マニュアルや定義の文を写して説明を補わないこと**
（要約は自分の言葉で書く）。OSS として公開するための前提。

手書きなのは `auth` / `profile` / `api` / `call` / `llm` / `dump` / `extensions` と、生成された
`sites templates` に足した `pull` / `push` だけ。**後者はエンドポイントでは
ない**——Data API に一括の取り出し／書き戻しが無いので、mtb が list と update
を回している（`--all` や `--dir` と同じ立場）。利用者はテンプレートの操作を
`sites templates` で探すので、別の場所に置くより見つけやすい。MT 側が同名の
エンドポイントを足せば `TestNoDuplicateNamesAmongSiblings` が落ちる。

つまり **エンドポイントに対応するコマンドを足す作業は無い**。MT 側が増えたら
定義を取り込み直す。名前が気に入らないなら `internal/gen-spec/main.go` の
規則か overrides を直す。

## 拡張（`mtb extensions`）

**MT プラグインが足したエンドポイントは、この CLI に組み込まず拡張で足す。**
拡張は別リポジトリのディレクトリ（`mtb-extension.json` ＋ `spec.json` ＋ `llm.md`
＋ Claude Code のプラグイン）で、`internal/extension` が読み、`internal/cmd/extensions.go`
が**生成コマンドと同じ `newOperationCmd` で**ツリーに足す。だから `--set` などが
そのまま効く。実例は `../mtb-plugin-pro`（MT 側は `../mt-plugin-mtb-pro`）。

- **拡張は宣言だけ。コードを読み込む口は作らない。** 他人の拡張を入れて任意の
  コードが走るようになると、この CLI の性格が変わる
- **「プラグイン」と呼ばない。** トップレベルの `plugins`（別名 `plugin`）が
  MT のプラグインを操作する生成コマンドとしてすでにある
- **ぶつかる拡張は丸ごと落とす。** 一部だけ入ると、どれが拡張のものか分からない。
  拡張は最後に足すので、手書き・生成・別名のどれとも突き合わせられる
- **拡張の章（`mtb llm`）は日本語で固定**（`internal/llmdocs/extension.go`）。
  コマンドカタログは拡張を含まないツリーから描くので、拡張の生成物は無い
- 拡張のコマンドの 404 には「MT 側にプラグインが無いかも」を足す。拡張の
  エンドポイントは MT 本体に無いので、404 の多くはそれが原因
- **file と path を取るエンドポイントには `--dir` が付く**（拡張でも同じ）。
  `--replace`（同名の**アセット**を消す）はアセットのアップロードにだけ付ける
  （`hasUploadPath`）。テーマやサイトのファイルに付くとアセットを消しに行く
- 本体の説明（00-guide・スキル）で「API ではできない」と書いた箇所には、
  mtb-pro でできるなら**その旨を添える**。本体の事実は消さない（素の MT ではできない）

## Builder であって Data API のラッパーではない

名前が示すとおり、この道具の目的は **MT のサイトを組み立てること**であって、
Data API を薄く包むことではない。**API の形をそのまま利用者に渡さない**、
というのが一貫した立場で、いま入っているものは全部その現れ:

| | API に無いもの | なぜ mtb が持つか |
| --- | --- | --- |
| `--all` | 全ページの取得 | ページ繰りを利用者に書かせない |
| `--fields` | GET 以外での属性の絞り込み | 同じフラグが使えたり使えなかったりしない |
| `--dir` / `--replace` | 一括アップロードと差し替え | デザインは1枚ずつ上げるものではない |
| `--set` / `--extract` | 本体の組み立てと取り出し | JSON を手で書かせない。jq も要らない |
| `templates pull` / `push` | テンプレートの往復 | Git で持つのに要る |
| `--confirm` | 入れ物ごと壊す操作の防護 | 取り違えると戻せない |
| `dump` | 環境どうしの差分 | どこが違うかが分からないと同期は始まらない |

**判断の基準は「サイトを1つ作り切れるか」。** API に無いという理由だけで
断らない。逆に、サイト構築に関係ないもの（画像処理など）は、できても入れない
——依存と性格が変わる。

**足すなら生成されたツリーの中に置く。** 利用者はその操作を探しに来る場所で
探す（`sites templates pull` を別の場所に置かない）。名前が MT 側の新しい
エンドポイントとぶつかれば `TestNoDuplicateNamesAmongSiblings` が落ちる。

## 変更時の必須手順

**機能を追加した、フラグを増やした、既存の挙動を変えた — このいずれかをしたら、
3か所すべてを更新してから終わること。**

| 更新先 | 対象 | やり方 |
| --- | --- | --- |
| ① ドキュメント | `README.md` | 使い方・共通フラグ・構成が変わったとき |
| ② ヘルプ | cobra の `Short` / `Long` / フラグ説明 | 手書きコマンドはコード内。生成コマンドは題名・パラメータ・公式リファレンスへのリンクだけで、補足は `paramNotes` に自前で書く |
| ③ **LLMナレッジ** | `internal/llmdocs/00-guide.md` | 認証・接続先・出力・落とし穴・ワークフローが変わったら |
| | `internal/llmdocs/10-endpoints.md` | **命名規則と overrides を変えたら必ず** |
| | `internal/llmdocs/90-commands.md` | **生成物。手編集しない** → `go generate ./...` |
| | `plugins/*/skills/*/references/tags-*.md` `modifiers.md` | **生成物。手編集しない** → 同上。原本は `internal/mtdocs/source/` |
| | `internal/spec/spec.json` | **生成物。手編集しない** → 同上 |
| | `plugins/movable-type-builder-cli/skills/*/SKILL.md` | 手順や前提が変わったとき |

③ を忘れやすい。ドキュメントとヘルプは人間が読んで気づくが、**LLMナレッジが
古いことには誰も気づかない**（エージェントが黙って間違えるだけ）。

判断に迷ったときの目安:

- **命名規則を変えた** → 220 コマンドの名前が動く。`10-endpoints.md` の対応表と
  `TestRepresentativeCommandNames` の期待値を必ず揃える
- **overrides を1行足した** → `10-endpoints.md` の例外表に理由つきで書く。
  「なぜ規則どおりではまずいのか」が残っていないと、次に誰かが外す
- **認証の扱いを変えた** → `00-guide.md` の「認証」と `mtb-usage` の SKILL.md。
  **パスワードを保存しない**という約束は破らないこと（テストが見張っている）
- **破壊的操作を追加した** → `--yes` の対象になっているか確認し、`mtb-usage` に書く
- **人に見せる文言を足した** → `i18n.T(ja, en)` で書く。片方だけ書くと
  `internal/i18n` の coverage テストが落ちる。用語は MT のマニュアル
  （<https://www.movabletype.jp/documentation/mt9/>）に合わせる
  （サインイン／Webサービスパスワード／アーカイブマッピング）
- **共通フラグを足した** → 生成コマンドのフラグと名前が衝突していないか。
  `TestGeneratedFlagsDoNotShadowGlobals` が落ちるはず
- **終了コードを増やした** → `internal/cmd/exit.go` の定数コメント、README、
  `00-guide.md` の表の3つを揃える。エージェントはこの表で分岐する
- **公式定義に足りないパラメータを見つけた** → `internal/gen-spec` の
  `extraBody` に**理由つきで**足す。「実物は受け付けるのに定義に無い」もの
  だけが対象で、推測で足さないこと
- **定義の型が実態と違う** → `paramTypes` で直す。JSON を運ぶのに `string` と
  書かれている項目があり、そのままだと `@ファイル名` が展開されずに
  リテラルが飛ぶ（permutate の4つが実際にそうだった）
- **配列のパラメータを足した** → スカラーの配列は `include=a&include=b` の形で
  送る（`spec.Param.Repeated`）。**JSON の配列を1つの値として送ると MT は
  黙って無視し、しかも success を返す**（テーマのエクスポートが空のテーマを
  作る形で実際に起きた）。オブジェクトの配列は JSON のまま
- **MT がエラーで選択肢を教えてくれない** → `paramNotes` で `--help` に足す。
  **注記は日本語と英語の両方を書く**（`note{ja, en}`）。片方だけだと
  `TestParamNotesAreBilingual` が落ちる

## 設計の約束

**外部依存はすべてインタフェース越し。** `internal/cmd` から `os` や
`net/http` を直接触らないこと。触るとテストがそこだけ実環境に依存する。

| 依存 | 抽象 | テストでの実体 |
| --- | --- | --- |
| ファイルシステム | `config.Storage` | `config.MemStorage` |
| 環境変数 | `config.Environ` | `config.MapEnv` |
| HTTP | `api.Doer` | `fakeDoer`（応答を積む） |
| 認証情報の置き場 | `api.TokenStore` | `api.MemoryTokenStore` |
| 時刻 | `Env.Now` | 固定時刻 |
| 端末入力 | `prompt.Prompter` | `prompt.Scripted` |
| 表示言語 | `Env.Getenv` / `--lang` | `withLang("en")` |
| 入出力・ファイル読み書き | `cmd.Env` | バッファとマップ |

`internal/cmd/harness_test.go` がこれらを一式差し替えた CLI を組む。
新しいコマンドのテストはそこに乗せる。

**.env はプロセスに撒かない。** カレントディレクトリの `.env` は
`config.DotEnv` で `Environ` の後ろに重ねるだけで、`os.Setenv` はしない。
撒くと、この CLI が起動する他のプロセスにまで漏れる。プロセスの環境変数が
勝つ向きも変えないこと（打ったものをファイルが黙って覆してはいけない）。

**ダンプは決定的であること。** `mtb dump` は差分を取るためだけにあるので、
**同じ入力なら必ず同じバイト列**が出ないと意味がない。キーは名前順、並びは
自然キー、ID と時刻と作成者は落とす（`internal/cmd/dumpspec.go` の
`volatileKeys`）。**ID では突き合わせない**——環境が違えば別の番号になる。

複数行はリテラルブロックで書くが、**行末の空白とタブを含む文字列は引用に
落とすこと**（`canBeLiteral`）。ブロックスカラーはそれらを復元できないので、
一律にリテラルにすると「差分は出ないのに実物が違う」という一番まずい壊れ方をする。
`TestDumpYAMLRoundTripsExactly` が見張っている。yaml.v3 を入れているのは
この種の取り決めを自前で間違えないため。

**出力は既定で読める形にする。** MT は日本語を `\uXXXX` でエスケープした
JSON を返す。そのまま出すと人が読めず、エージェントには同じ内容が数倍の
トークンになる。`internal/cmd/jsontext.go` が戻している。`--format raw`
だけは無加工のままにしておくこと（バイト列が要る用途の逃げ道）。

**人に見せる文言は `i18n.T(ja, en)` で書く。** キーの表は持たない。訳を
別ファイルに置くと、消したはずの文言が残り、足した文言が片方だけになる
（そして誰も気づかない）。呼ぶ場所に両方を並べておけば、書き換えるときに
両方が目に入る。パッケージ変数のエラーは `i18n.Error(ja, en)`——先に作られて
言語がまだ決まっていないので、表示のときに選ぶ。

言語は `internal/cmd.New` の先頭で決める。ヘルプの文字列はツリーを組む
時点で確定するため、cobra のフラグ解析より前に `--lang` を先読みしている
（`langFromArgs`）。**訳し漏れは `internal/i18n/coverage_test.go` が落とす**——
`internal/cmd` `internal/api` `internal/config` `internal/prompt` `internal/spec`
`cmd/mtb` の中に、`i18n.T` / `i18n.Error` の外の日本語リテラルがあれば失敗する。

**書式の引数は、言語で順番を変えないこと。** 日本語の語順に合わせて `%s` を
入れ替えたくなるが、渡す引数は1組しかないので必ず壊れる。英語側の語順を
工夫して合わせる（`%s に対して %s %s` → `On %s, about to run %s %s`）。
`%s` 同士だと指定子の並びは一致してしまうので、
`TestFormatVerbsMatchAcrossLanguages` でも捕まらない。**この罠は実際に2回踏んだ。**

**公式定義から来る題名は訳さない。** 操作の題名（summary）は公式定義の英語のまま
出る。説明文は持ち込まない（上の「Six Apart の文章を持ち込まない」）。mtb が補った
注記（`paramNotes`）だけが `spec.json` に `description` と `descriptionEn` の
両方を持ち、`spec.Param.Describe()` が選ぶ。**生成物の言語は生成器が固定する**
（`gen-llmdocs` は日本語）。環境の言語で生成物が入れ替わってはいけない。

**JSON を手で書かせない。** リソース本体は `--set <フラグ>.<項目>=<値>` で
組み立てられる（`internal/cmd/set.go`）。**組み立てた JSON は `<` を
`\uXXXX` に逃がさないこと**——MT のテンプレートは `<` だらけで、逃がすと
`--dry-run` で目視できず、エージェントには同じ内容が数倍のトークンになる
（`marshalNoEscape`）。取り出し側は `--extract`。この2つが無いと、mtb を
使うのに jq が要る——**問い合わせ言語を足すのではなく、要らなくする**方を採った。

**同じフラグが使えたり使えなかったりしないこと。** MT は GET にしか fields を
用意していないが、`--fields` はどのコマンドでも受ける（未対応のぶんは
`internal/cmd/output.go` が手元で絞る）。API の非対称をそのまま利用者に
渡さない、というのがこの CLI の立場。

**配布スキルの参照ファイルも生成物。** 600KB 弱あり、`internal/mtdocs/source/*.json`
から `internal/gen-skilldocs` が作る。MT 側のドキュメントが更新されたら
`go run ./internal/gen-skilldocs -fetch` で原本を取り直してから `go generate`。
**手で書き足したいときは SKILL.md か、mt-cms-design の references（目印の外）に書く。**

**生成の順番に意味がある。** `spec.json` を作り直してからでないと
コマンドカタログが古いツリーから描かれる。spec.json は公式定義をネットワークから
取るので `go generate` には入れず、`make spec`（gen-spec → `go generate ./...`）で
この順に流す。カタログと参照ファイルの go:generate は `cmd/mtb/main.go` に
まとめてある（`go generate ./...` はインポートパス順に回るので `cmd/mtb` が先）。

## リリース

`PluginVersion`（`internal/cmd/root.go`）と `plugin.json` の `version` と git タグ
の3つを揃える。テストとリリースワークフローが不一致を検出する。手順は
`plugins/movable-type-builder-cli/PUBLISH.md`。

## 確認

```bash
make check            # go generate → gofmt → vet → build → test → git diff
go run ./cmd/mtb llm | head
```

## 参照

- 公式リファレンス: <https://movabletype.github.io/mt-docs-data-api-reference/v7.html>
- 定義の原本: <https://github.com/movabletype/mt-docs-data-api-reference>（`src/openapi/v7.json`）
- 標準: <https://github.com/ideamans/go-llm-cli-kit/blob/main/LLM.md>
- 生成物と原本の対応: `.claude/rules/ai-artifacts-policy.md`
- 再生成: `/regen-ai`
