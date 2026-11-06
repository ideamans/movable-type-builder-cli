# Publishing the movable-type-builder-cli plugin

## Before every release

1. `go generate ./...` — regenerates `internal/spec/spec.json` and
   `internal/llmdocs/90-commands.md`. Commit any diff.
2. `go test ./...` — includes `TestPluginSkills`, which enforces that
   `plugin.json.version` equals `PluginVersion` in `internal/cmd/root.go` and
   that the SKILL.md frontmatter stays within the Agent Skills standard.
3. `claude plugin validate plugins/movable-type-builder-cli`.
4. Bump `PluginVersion` and `plugin.json.version` together, in the same commit
   as the release tag. The release workflow refuses a mismatched tag.

## Registering in the marketplace (first release only)

The repository is public, so the entry goes in the public marketplace
`ideamans/claude-public-plugins` (marketplace name `ideamans-plugins`).
It used to be in `ideamans/claude-private-plugins` while the repository was
private; that entry was removed when the repository went public (2026-11-06).

```json
{
  "name": "movable-type-builder-cli",
  "source": {
    "source": "git-subdir",
    "url": "https://github.com/ideamans/movable-type-builder-cli.git",
    "path": "plugins/movable-type-builder-cli"
  }
}
```

## Verifying the published result

```
/plugin marketplace add ideamans/claude-public-plugins
/plugin install movable-type-builder-cli@ideamans-plugins
/mtb-usage
```

スキルは4本。

| スキル | 役割 |
| --- | --- |
| `mtb-install` | `mtb` を PATH に用意する |
| `mtb-usage` | CLI で MT を読み書きする手順 |
| `mt-cms-design` | 要件からサイトを設計して組み立てる。オブジェクト・権限・アーカイブ |
| `mt-template-tags` | テンプレートを書く。全614タグと48モディファイアの参照つき |

`mt-cms-design` と `mt-template-tags` は `references/` に資料を同梱している
（合計 600KB 弱）。**SKILL.md 本文には仕様を書かず、参照させる**方針。
資料は movabletype.jp の公式ドキュメントと、実機の MT 9.0.9 から起こしてある。

```bash
gh skill install ideamans/movable-type-builder-cli/plugins/movable-type-builder-cli/skills/mtb-usage --agent copilot
```

Neither route needs a GitHub login.

## The binary

GitHub Releases (goreleaser, on a `v*` tag) is the source. bin.ideamans.com
(category `mtb`, key `mtb`) syncs it every 10 minutes and serves the install
script that `mtb-install` and the README use:
`curl -fsSL https://bin.ideamans.com/install/mtb.sh | bash`.
