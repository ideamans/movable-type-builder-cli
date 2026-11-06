---
name: mtb-install
description: Make the mtb command available, installing it only if it is missing. Use when another skill reports that `mtb` is not on PATH, or when the user asks to install, update or upgrade the ideamans Movable Type Builder CLI. Prefers an already-installed binary, then the ideamans install script (bin.ideamans.com), then the GitHub release, then a build from source with go install.
license: MIT
compatibility: Requires curl and tar (Linux/macOS), or PowerShell on Windows, or a Go toolchain for the source fallback. No GitHub login is needed — the repository github.com/ideamans/movable-type-builder-cli and its releases are public. Standalone — does not need mtb to be present already.
allowed-tools: Bash(curl:*) Bash(bash:*) Bash(tar:*) Bash(unzip:*) Bash(go:*) Bash(uname:*) Bash(command:*) Bash(which:*) Bash(mkdir:*) Bash(mv:*) Bash(cp:*) Bash(rm:*) Bash(chmod:*) Bash(ls:*) Bash(test:*) Bash(echo:*) Read
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

## Route 2 — the install script (bin.ideamans.com)

The usual route. It picks the archive for the OS and CPU, checks the SHA-256 and
puts `mtb` on PATH.

```bash
curl -fsSL https://bin.ideamans.com/install/mtb.sh | bash -s -- --install-dir "$HOME/.local/bin" --yes
```

Pass `--install-dir "$HOME/.local/bin"` so that it does not need `sudo` (the default
is `/usr/local/bin`, which usually does). **Do not run `sudo` yourself**: if the user
wants `/usr/local/bin`, print the command without `--install-dir` and let them run it.
If `~/.local/bin` is not on PATH, give the user the line to add to their shell profile.
Do not edit the profile for them.

On Windows (PowerShell):

```powershell
irm https://bin.ideamans.com/install/mtb.ps1 | iex
```

The same server also offers apt, yum, Scoop and Chocolatey; the page
<https://bin.ideamans.com/mtb/mtb> lists them. Use those only if the user asks.

## Route 3 — the GitHub release

If bin.ideamans.com cannot be reached, take the archive from the public release.

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')             # darwin | linux
ARCH=$(uname -m); case "$ARCH" in x86_64) ARCH=amd64 ;; aarch64) ARCH=arm64 ;; esac

URL=$(curl -fsSL https://api.github.com/repos/ideamans/movable-type-builder-cli/releases/latest \
  | grep -o "https://[^\"]*_${OS}_${ARCH}\.tar\.gz" | head -1)
echo "$URL"
curl -fsSL -o /tmp/mtb.tar.gz "$URL"
tar -xzf /tmp/mtb.tar.gz -C /tmp mtb
mkdir -p ~/.local/bin && mv /tmp/mtb ~/.local/bin/ && chmod +x ~/.local/bin/mtb
```

**The archive is named `movable-type-builder-cli_<version>_<os>_<arch>.tar.gz` but
the binary inside is `mtb`.** If `URL` comes back empty, print every
`browser_download_url` of that response instead of guessing names.

Windows ships a `.zip`.

## Route 4 — build from source

Needs a Go toolchain. Note the `/cmd/mtb` suffix — the main package is not at the
module root.

```bash
go install github.com/ideamans/movable-type-builder-cli/cmd/mtb@latest
```

This produces a binary named `mtb` (the last path element), so no rename is needed.

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
