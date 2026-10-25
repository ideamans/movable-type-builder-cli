# 生成物 — 手編集しないこと

| 生成物 | 原本 |
| --- | --- |
| `internal/spec/spec.json` | 公式定義（`internal/gen-spec` の `sourceCommit` から取得。リポジトリには無い）＋ `internal/gen-spec/main.go`（命名規則）。`make spec` |
| `internal/llmdocs/90-commands.md` | cobra のコマンドツリー（`internal/gen-llmdocs` が描画） |
| `plugins/*/skills/mt-template-tags/references/*.md` | `internal/mtdocs/source/tags.json` `modifiers.json`（`internal/gen-skilldocs` が描画） |
| `plugins/*/skills/mt-cms-design/references/permissions.md` の**ロール表** | `internal/mtdocs/source/roles.json`（目印 `<!-- roles:begin -->` の中だけ） |
| `plugins/*/skills/mt-cms-design/references/archives.md` の**種別一覧** | `internal/mtdocs/source/archives.json`（目印 `<!-- archives:begin -->` の中だけ） |

`spec.json` がコマンドツリーの素なので、**コマンド名を変えたいときに直すのは
`internal/gen-spec/main.go` の規則か overrides の表**であって、生成物ではない。

手書き（編集してよい）:

- `internal/llmdocs/00-guide.md` — 鉄則・認証・ワークフロー・失敗モード
- `internal/llmdocs/10-endpoints.md` — コマンド名とエンドポイントの対応規則
- `internal/cmd/auth.go` `profile.go` `api.go` — 手書きのコマンド
- `plugins/*/skills/*/SKILL.md`
- `plugins/*/skills/mt-cms-design/references/*.md` の**目印の外**（自前の文章。マニュアルを写さない）
- `internal/mtdocs/summaries.json` — タグ・モディファイア・ロールの自前の要約

公式定義と `internal/mtdocs/source/*.json`（事実だけ）は外から取ってきた
原本で、**手で直さない**。取り直す方法は README の「資料の更新」。

**参照ファイルは 250KB ほどある。** 壊れても人は気づかないので、
`internal/cmd/skilldocs_test.go` が原本との対応を検査している
（全タグが載っているか、要約とマニュアルへのリンクがあるか、索引の指す先が実在するか、
非推奨が混ざっていないか）。

生成物を直接編集しても次の `go generate ./...` で消え、それまでは CI が古い差分で
落ちる。再生成は `/regen-ai`、または `go generate ./... && go test ./...`。
