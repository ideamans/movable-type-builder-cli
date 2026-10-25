---
name: mtb-install
description: Make the mtb command available, installing it only if it is missing. Use when another skill reports that `mtb` is not on PATH, or when the user asks to install, update or upgrade the ideamans Movable Type Builder CLI. Prefers an already-installed binary, then a release downloaded with the authenticated gh CLI, then a build from source with go install.
license: MIT
compatibility: Requires the GitHub CLI (`gh`) authenticated against the ideamans organisation, plus tar, or a Go toolchain for the source fallback. Standalone — does not need mtb to be present already. The repository github.com/ideamans/movable-type-builder-cli is private, so an unauthenticated download will not work.
allowed-tools: Bash(gh:*) Bash(tar:*) Bash(unzip:*) Bash(go:*) Bash(uname:*) Bash(command:*) Bash(which:*) Bash(mkdir:*) Bash(mv:*) Bash(cp:*) Bash(rm:*) Bash(chmod:*) Bash(ls:*) Bash(test:*) Bash(echo:*) Read
---

# mtb-install

Make the `mtb` command usable, doing the least work that achieves it.

## Route 1 — an existing installation on PATH

```bash
command -v mtb && mtb --version
```

If that resolves, **use it and stop here.** Do not check for a newer release —
it costs an API call and the user did not ask for an upgrade.

Two checks before trusting the hit:

- **It is the right tool.** `mtb llm | head -1` must read
  `# mtb — AIエージェント向けリファレンス`.
- **It is recent enough.** If `mtb llm` is not a known command, the binary
  predates the embedded reference. Say so and continue to route 2.

Continue past this section only when the command is missing, is the wrong tool,
is too old, or the user explicitly asked to update.

## Route 2 — the latest GitHub release

The repository is **private**, so releases must be fetched with `gh`, not
`curl`. Check authentication first:

```bash
gh auth status
```

If `gh` is missing, or is not authenticated with access to the `ideamans`
organisation, **stop and ask the user to run `gh auth login` themselves.** Do
not ask them to paste a token into the conversation — it would be kept in the
transcript.

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')             # darwin | linux
ARCH=$(uname -m); [ "$ARCH" = "x86_64" ] && ARCH=amd64  # amd64 | arm64

gh release download --repo ideamans/movable-type-builder-cli \
  --pattern "*_${OS}_${ARCH}.tar.gz" --dir /tmp --clobber
```

**The archive is named `movable-type-builder-cli_…` but the binary inside is
`mtb`.** Windows ships a `.zip`.

If nothing matches the pattern, list the assets instead of guessing variations:

```bash
gh release view --repo ideamans/movable-type-builder-cli --json assets \
  --jq '.assets[].name'
```

### Install onto PATH

```bash
tar -xzf /tmp/movable-type-builder-cli_*_${OS}_${ARCH}.tar.gz -C /tmp
mkdir -p ~/.local/bin && mv /tmp/mtb ~/.local/bin/ && chmod +x ~/.local/bin/mtb
```

Prefer the first writable directory already on PATH — `~/.local/bin`, then
`/usr/local/bin`. Two things not to do on your own initiative:

- If nothing on PATH is writable, leave the binary in `/tmp`, print the exact
  `sudo mv` command and let the user run it. Do not run `sudo` yourself.
- If `~/.local/bin` is not on PATH, give the user the line to add to their shell
  profile. Do not edit the profile for them.

## Route 3 — build from source

Needs a Go toolchain and a git credential that can read the private repository.
Note the `/cmd/mtb` suffix — the main package is not at the module root.

```bash
GOPRIVATE=github.com/ideamans/* \
  go install github.com/ideamans/movable-type-builder-cli/cmd/mtb@latest
```

This produces a binary named `mtb` (the last path element), so no rename is
needed. If the module download fails with a 404, the Go toolchain is going
through the public proxy — `GOPRIVATE` is what stops that.

## Verify

```bash
mtb --version
mtb llm | head -5
mtb profile list
```

Report which route was taken, the version and the install path.

`mtb profile list` showing nothing is expected on a fresh install — it is not
an error. Point the user at the next step rather than trying to guess their
server:

```bash
mtb profile add prod --base-url https://<their-mt-host>/cgi-bin/mt/mt-data-api.cgi
mtb auth login
```

`mtb auth login` prompts for a password, so the user has to run it. It wants
their MT **Web Services Password**, not the sign-in password.
