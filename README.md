# hb: heartbeads CLI 

Authenticated issue tracking for AI agents. A wrapper around [bd (beads)](https://github.com/gainforest/beads) that requires ATProto identity before any command runs.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/gainforest/heartbeads-cli/main/install.sh | bash
```

Or with Go:

```bash
go install github.com/gainforest/heartbeads-cli/cmd/hb@latest
```

Requires [bd](https://github.com/gainforest/beads) in your PATH.

## Why hb?

`bd` is a local issue tracker for AI agents. `hb` adds one thing: **identity**. Every action is tied to an ATProto account so you know which agent did what.

Without `hb`, agents write to a shared issue database anonymously. With `hb`, every create, update, and close carries a verified identity. This matters when multiple agents collaborate on the same codebase.

## Quick start

```bash
# 1. Login with your ATProto account
hb account login --username alice.bsky.social --password app-password-123

# 2. Initialize in your project
hb init

# 3. Find work
hb ready

# 4. Create an issue
hb create "Fix login timeout" --type bug --priority 1

# 5. Claim it
hb update <id> --claim

# 6. Close when done
hb close <id>

# 7. Sync to git
hb sync
```

## What hb does

`hb` sits between the agent and `bd`. Every command goes through three steps:

```
agent -> hb -> bd
         |
         +-- 1. Check auth (reject if not logged in)
         +-- 2. Inject identity flags
         +-- 3. Rewrite output ("bd" -> "hb")
```

### Flag injection

`hb` automatically adds flags based on your logged-in ATProto handle:

| Flag | Commands | Source | Purpose |
|------|----------|--------|---------|
| `--actor <handle>` | all | ATProto handle | Sets `created_by` field |
| `--assignee <handle>` | `update` | ATProto handle | Auto-assigns work to you |
| `--reason "<hash> <msg>"` | `close` | **Required** (user must provide) | Commit that resolves the issue |
| `--session <id>` | `close`, `update` | `CLAUDE_SESSION_ID` or `OPENCODE_SESSION` env | Links actions to agent sessions |

`--reason` is **mandatory** on `hb close` and must be a commit reference: `"<hash> <message>"`.
Other flags are never doubled — if you pass one explicitly, the auto-inject is skipped.

### Output rewriting

All `bd` output is rewritten so agents see a consistent `hb` interface:

- `` `bd ready` `` becomes `` `hb ready` ``
- `bd (beads)` becomes `hb (heartbeads)`
- Error messages, help text, and prose references are all rewritten
- Issue IDs like `bd-w382l` are preserved (not rewritten)
- JSON output is never modified

### Owner fallback

When git identity is not configured, `hb` sets `GIT_AUTHOR_EMAIL` and `BD_ACTOR` environment variables to your ATProto handle. This ensures the `owner` and `created_by` fields are always populated.

## Commands

### Account management

```bash
hb account login --username <handle> --password <app-password>
hb account logout
hb account status
```

### Issue tracking (proxied to bd)

```bash
# Finding work
hb ready                          # Unblocked issues
hb list --status open             # All open issues
hb show <id>                      # Issue details
hb search "query"                 # Full-text search

# Creating & updating
hb create "Title" --type task --priority 2
hb update <id> --status in_progress
hb update <id> --claim            # Atomic claim (fails if already assigned)
hb close <id> --reason "a1b2c3d fix: resolve the bug"  # commit hash + message required

# Dependencies
hb dep add <issue> <depends-on>
hb blocked
hb graph <id>

# Hierarchy
hb epic create "Epic title"
hb children <parent-id>

# Sync
hb sync                           # Sync with git
hb export                         # Export to JSONL

# Quick capture
hb q "Fix the bug"                # Create + output only the ID
```

Every `bd` command works through `hb`. Run `hb --help` for the full list.

## For AI agents

Add this to your `AGENTS.md`:

```markdown
## Issue Tracking

This project uses **hb** for issue tracking.
Run `hb prime` for workflow context.

Quick reference:
- `hb ready` - Find work
- `hb create "Title" --type task --priority 2` - Create issue
- `hb update <id> --claim` - Claim work
- `hb close <id>` - Complete work
- `hb sync` - Sync with git
```

Or generate it automatically:

```bash
hb onboard    # Prints a ready-to-paste AGENTS.md snippet
hb prime      # Full workflow context for AI agents
```

## Auth storage

Sessions are stored at `~/.local/state/heartbeads/auth-session.json` (XDG state directory). The file is created with `0600` permissions and contains:

- ATProto DID, handle, and PDS URL
- Access and refresh tokens (auto-refreshed)
- App password (for session recovery)

Run `hb account logout` to delete the session file.

## Environment variables

| Variable | Purpose |
|----------|---------|
| `ATP_PLC_HOST` | Override PLC directory URL (default: `https://plc.directory`) |
| `CLAUDE_SESSION_ID` | Auto-injected as `--session` on close/update |
| `OPENCODE_SESSION` | Fallback for `--session` if `CLAUDE_SESSION_ID` is not set |

## Build from source

```bash
git clone https://github.com/gainforest/heartbeads-cli.git
cd heartbeads-cli
make build    # -> ./hb
make test     # Run tests
make lint     # Run linter
```

## Architecture

```
heartbeads-cli/
  cmd/hb/            # Entry point
    main.go          # CLI app, catchall proxy
  internal/
    auth/            # ATProto session management
    account/         # login/logout/status commands
    executor/        # bd binary discovery, output rewriting, process execution
    inject/          # Flag injection (actor, assignee, reason, session)
    proxy/           # Auth guard + flag injection + bd execution pipeline
```

## License

MIT
