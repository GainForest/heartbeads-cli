# Agent Instructions

This project uses **hb** (heartbeads) for issue tracking — an authenticated wrapper around beads.

## Setup

First, login with your ATProto account:
```bash
hb account login --username <your-handle> --password <app-password>
```

## Quick Reference

```bash
hb ready              # Find available work
hb show <id>          # View issue details
hb update <id> --status in_progress  # Claim work
hb close <id> --reason "<commit-hash> <message>"  # Complete work (reason required)
hb sync               # Sync with git (no-op with dolt backend)
```

> **Note:** If this repo uses old beads (SQLite), run `hb migrate` first to migrate to the dolt backend.

Note: hb automatically sets --actor and --assignee (on update) to your ATProto handle. `hb close` requires `--reason` with a commit reference (`"<hash> <message>"`).

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   git push
   git status  # MUST show "up to date with origin"
   ```
   > **Note:** `hb sync` is no longer required — changes are persisted automatically by the dolt backend. `git push` is still needed for JSONL portability (exported via pre-commit/post-merge hooks).
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
