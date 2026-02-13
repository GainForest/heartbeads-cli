# heartbeads-cli (hb)

Authenticated beads CLI for AI agents.

## Overview

`hb` wraps the [bd (beads)](https://github.com/steveyegge/beads) CLI with ATProto authentication. AI agents use `hb` instead of `bd` to ensure identity-aware issue tracking.

Key features:
- **Auth required**: agents must login with `hb account login` before running any commands
- **Auto-identity**: `--actor` injected on all commands (sets created_by), `--assignee` on `update` only, `--reason` from git on `close`, `--session` from env on `close`/`update`
- **Transparent proxy**: all bd commands work through hb — output is rewritten so agents never see "bd"

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [bd (beads)](https://github.com/steveyegge/beads) CLI installed and in PATH

## Install

```bash
go install github.com/gainforest/heartbeads-cli@latest
```

Or build from source:

```bash
git clone https://github.com/gainforest/heartbeads-cli.git
cd heartbeads-cli
make build
```

## Quick Start

```bash
# Login with your ATProto account
hb account login --username alice.bsky.social --password app-password-123

# Initialize issue tracking in a project
hb init

# Create issues (--assignee auto-set to your handle)
hb create "Implement feature X" --type task --priority 2

# Find available work
hb ready

# Claim and work on an issue
hb update <id> --status in_progress

# Close when done
hb close <id>

# Sync with git
hb sync
```

## How It Works

1. **Auth layer**: `hb` stores ATProto session in `~/.local/state/heartbeads/auth-session.json` (XDG state dir)
2. **Proxy**: all commands are forwarded to `bd` with `BD_NAME=hb` environment variable
3. **Output rewriting**: stdout/stderr text is post-processed to replace "bd" references with "hb"
4. **Flag injection**: `--actor` on all commands, `--assignee` on `update`, `--reason` (from latest git commit) on `close`, `--session` (from `CLAUDE_SESSION_ID`/`OPENCODE_SESSION` env) on `close`/`update`

## Commands

### Account management (native)
- `hb account login` — login with ATProto credentials
- `hb account logout` — delete current session
- `hb account status` — check login status

### All bd commands (proxied)
Every bd command works through hb. Examples:
- `hb list`, `hb ready`, `hb show <id>`
- `hb create`, `hb update`, `hb close`
- `hb sync`, `hb dep`, `hb graph`
- `hb prime`, `hb onboard`, `hb doctor`

Run `hb --help` for the full list.

## For AI Agents

See [AGENTS.md](AGENTS.md) for agent-specific instructions.

## Development

```bash
make test    # Run tests
make lint    # Run linter
make build   # Build binary
```

## License

MIT
