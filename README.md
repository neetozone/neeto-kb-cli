# NeetoKB CLI

A command-line interface for NeetoKB.

## Installation

### macOS / Linux

**Homebrew (recommended on macOS):**

```bash
brew install neetozone/tap/neetokb
```

**Shell script:**

```bash
curl -fsSL https://neetokb.com/cli/install.sh | sh
```

### Windows

**PowerShell:**

```powershell
irm https://neetokb.com/cli/install.ps1 | iex
```

**Command Prompt (CMD):**

```cmd
curl -fsSL https://neetokb.com/cli/install.cmd -o install.cmd && install.cmd
```

### Verify installation

```bash
neetokb --help
```

## Prerequisites (development)

- [Go](https://go.dev/dl/) 1.26.1+
- Access to a NeetoKB organization

## Development

```bash
git clone https://github.com/neetozone/neetokb-cli.git
cd neetokb-cli
bin/setup
```

This installs Go dependencies, golangci-lint, configures git hooks, and builds the binary.

### Make targets

```bash
make build          # Builds ./neetokb
make test           # Run tests
make lint           # golangci-lint
make fmt            # gofmt -w
make vet            # go vet
make check          # fmt + vet + test
make install        # Installs to /usr/local/bin
make clean          # Remove built binary
```

### Pointing to a local or staging server

Set `NEETOKB_BASE_URL` to override the default `https://<subdomain>.neetokb.com`:

```bash
export NEETOKB_BASE_URL=http://acme.lvh.me:8980
neetokb login --subdomain acme
```

## Global flags

Every command accepts:

| Flag | Description |
|---|---|
| `--subdomain <name>` | Which logged-in subdomain to use (required when multiple are logged in). |
| `--json` | Force JSON envelope output. |
| `--quiet` | Emit raw data only. Action commands print just the identifier; `delete` prints `success`. |
| `--toon` | TOON (Token-Optimized Output Notation) — compact format for LLMs. |

## Adding product-specific commands

See [`docs/adding-commands.md`](docs/adding-commands.md) for the step-by-step
workflow for adding new resource commands that use the built-in auth, HTTP
client, and output helpers.

Quick API wrapper reference: [`docs/api-wrapper-reference.md`](docs/api-wrapper-reference.md).

## Release

Releases are cut by BigBinary's CI pipeline defined in
`.neetoci/release.yml`. Merging a PR with a `major` / `minor` / `patch`
label to `main` triggers `.scripts/release.sh`, which tags the current
VERSION, runs GoReleaser, uploads artifacts to
`s3://neeto-downloads/cli/NeetoKB/`, updates the Homebrew tap
(`neetozone/homebrew-tap`), and opens the next-version bump PR.

## AI coding assistants

```bash
neetokb setup claude      # Register plugin with Claude Code
neetokb setup cursor      # Write .cursor/rules/neetokb.mdc
neetokb setup windsurf    # Write .windsurf/rules/neetokb.md
neetokb setup copilot     # Append to .github/copilot-instructions.md
neetokb setup gemini      # Append to GEMINI.md
neetokb setup codex       # Append to AGENTS.md
```
