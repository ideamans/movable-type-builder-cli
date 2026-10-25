---
name: regen-ai
description: 埋め込みAPI定義とLLMリファレンスを再生成して検証する。コマンド・フラグ・ヘルプ文・命名規則・OpenAPI定義・手書き章を変更した後に使う。
allowed-tools: Bash(make spec:*) Bash(go generate:*) Bash(go test:*) Bash(go build:*) Bash(git status:*) Bash(git diff:*) Read
---

# regen-ai

`internal/spec/spec.json` と `internal/llmdocs/90-commands.md` を原本に合わせ直します。

1. `git status --short` — 既存の差分を把握する。
2. `make spec` — `spec.json` → `90-commands.md` の順で作り直す
   （後者は前者から組み立てたコマンドツリーを見るので、順番が意味を持つ）。
   spec.json は公式定義を固定したコミットから取る（2回目からはキャッシュ）。
   ネットワークが無ければ `go generate ./...`（spec.json 以外）だけを流す。
3. `go build ./... && go test ./...`。
4. `git diff --stat internal/spec/spec.json` — **コマンド名が動いていないか**を確認する。
   増減があれば、どのエンドポイントがどう変わったかを報告する。

生成器が見ているのは OpenAPI 定義とコマンドツリーだけです。**公式の説明文は
spec.json に入れません**（再配布の許諾が無い。CLAUDE.md の「Six Apart の文章を持ち込まない」）。認証の落とし穴や
ワークフローを変えた場合は、`internal/llmdocs/00-guide.md` を手で直す必要があります。

このスキルは Claude Code ローカル用で、配布プラグインには含みません。
