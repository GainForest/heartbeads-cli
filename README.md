# hb: heartbeads CLI 

Authenticated issue tracking for AI agents. A wrapper around [bd (beads)](https://github.com/gainforest/beads) that requires ATProto identity before any command runs — plus native ATProto comment support for threaded discussions on issues.

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/gainforest/heartbeads-cli/main/install.sh | bash
```

Or with Go:

```bash
go install github.com/gainforest/heartbeads-cli/cmd/hb@latest
```

Requires [bd](https://github.com/gainforest/beads) in your PATH. v0.50+ recommended for dolt backend features (vc, sql, dolt, mol, gate, migrate).

## Why hb?

`bd` is a local issue tracker for AI agents. `hb` adds two things:

1. **Identity** — Every action is tied to an ATProto account so you know which agent did what.
2. **Comments** — Threaded discussions on issues, stored as ATProto records and visible on the [heartbeads map](https://heartbeads.gainforest.app).

Without `hb`, agents write to a shared issue database anonymously. With `hb`, every create, update, close, and comment carries a verified identity. This matters when multiple agents collaborate on the same codebase.

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

# 6. Comment on it
hb comment add <id> "Working on this now"

# 7. Close when done
hb close <id> --reason "a1b2c3d fix: resolve the bug"
```

## What hb does

`hb` sits between the agent and `bd`. Every proxied command goes through three steps:

```
agent -> hb -> bd
         |
         +-- 1. Check auth (reject if not logged in)
         +-- 2. Inject identity flags
         +-- 3. Rewrite output ("bd" -> "hb")
```

Native commands (`account`, `comment`) are handled directly by `hb` without calling `bd`.

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

### Comments (native ATProto)

Read and write threaded comments on beads issues. Comments are stored as `org.impactindexer.review.comment` records on the AT Protocol network, indexed by [Hypergoat](https://hypergoat-app-production.up.railway.app/graphql), and displayed on the heartbeads dependency graph map.

**Reading comments** does not require login. **Writing comments** requires an ATProto session.

```bash
# View comments
hb comment get                              # Last 10 comments across all issues
hb comment get <beads-id>                   # All comments for a specific issue
hb comment get -n 5                         # Last 5 comments
hb comment get -n 0                         # All comments (no limit)
hb comment get --filter "beads-map-*"       # Filter by glob pattern on nodeId
hb comment get <beads-id> --json            # Machine-readable JSON output

# Post comments (requires login)
hb comment add <beads-id> "LGTM"            # Comment on an issue
hb comment add <prefix> "Update for all"    # General comment on a project prefix
hb comment add --reply-to <at-uri> <beads-id> "thanks!"  # Reply to a comment
```

**Flags for `hb comment get`:**

| Flag | Default | Description |
|------|---------|-------------|
| `-n` | 10 (no args) / 0 (with beads-id) | Max root comments to show. `0` = unlimited |
| `--filter` | — | Glob pattern to match against nodeId (e.g. `beads-map-*`) |
| `--json` | `false` | Output as JSON array |
| `--indexer-url` | Hypergoat production URL | Override the GraphQL indexer endpoint |
| `--profile-api-url` | Bluesky public API | Override the profile resolution endpoint |

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
hb sync                           # Sync with git (no-op with dolt backend)
hb export                         # Export to JSONL

# Quick capture
hb q "Fix the bug"                # Create + output only the ID
```

Every `bd` command works through `hb`. Run `hb --help` for the full list.

## Migration

If your project uses an older version of beads (SQLite-based, with a `.beads/beads.db` file), migrate to the dolt backend with:

```bash
# Migrate the current project's .beads/ directory
hb migrate

# Migrate a specific project directory
hb migrate --path /path/to/project

# Dry run — preview what will be migrated without making changes
hb migrate --dry-run
```

After migration:
- `.beads/dolt/` contains the new dolt database (source of truth)
- `.beads/beads.db` (legacy SQLite) is no longer used
- JSONL files are regenerated from dolt on each commit via hooks
- `hb sync` is a no-op — changes persist automatically

> **Requires beads v0.50+.** Run `bd version` to check. If you see `backend: sqlite` in `.beads/metadata.json`, migration is needed.

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
- `hb comment get <id>` - Read comments
- `hb comment add <id> "text"` - Post a comment
- `hb close <id>` - Complete work
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
| `INDEXER_URL` | Override Hypergoat GraphQL indexer URL for `hb comment get` |
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

`hb` wraps `bd` (beads v0.50+), which uses a **dolt** backend for storage. Issues are persisted directly in `.beads/dolt/` — no SQLite database, no manual sync required. JSONL files (`.beads/*.jsonl`) are maintained for git portability via pre-commit/post-merge hooks.

```
heartbeads-cli/
  cmd/hb/            # Entry point
    main.go          # CLI app, catchall proxy
  internal/
    auth/            # ATProto session management
    account/         # login/logout/status commands
    comments/        # ATProto comment commands (get, add)
      client.go      #   Hypergoat GraphQL client with pagination
      profile.go     #   Bluesky profile resolver
      assemble.go    #   Filter, thread, and limit comments
      fetch.go       #   Orchestrator (parallel fetch + pipeline)
      format.go      #   Text and JSON formatters
      post.go        #   Create comment via ATProto createRecord
      command.go     #   CLI command definitions
      types.go       #   Shared types and constants
    executor/        # bd binary discovery, output rewriting, process execution
    inject/          # Flag injection (actor, assignee, reason, session)
    proxy/           # Auth guard + flag injection + bd execution pipeline
```

## License

MIT
