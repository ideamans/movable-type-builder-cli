---
paths:
  - "internal/cmd/*.go"
  - "internal/gen-spec/*.go"
  - "internal/mtdocs/summaries.json"
  - "internal/llmdocs/0*.md"
  - "internal/llmdocs/1*.md"
---

# 生成物の原本に触れました

コマンド・フラグ・ヘルプ文字列・命名規則・API定義のいずれかを変えたなら、
終わる前に `/regen-ai` を実行して `internal/spec/spec.json` と
`internal/llmdocs/90-commands.md` を一致させること。CI が再生成して差分が出たら
落ちます。

**命名規則を変えると 220 コマンドの名前が動きます。** `internal/spec/spec_test.go`
の `TestRepresentativeCommandNames` と `internal/gen-spec/main_test.go` が
落ちるはずなので、期待値を「意図した新しい名前」に直したうえで、
`internal/llmdocs/10-endpoints.md` の対応表も更新すること。

認証・落とし穴・ワークフローを変えたなら `00-guide.md` を手で直すこと。
ここがズレると、エージェントは**エラーにならずに間違った使い方をします**。

`spec.json` と `90-commands.md` は直接編集しないこと。
