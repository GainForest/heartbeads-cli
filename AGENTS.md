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
hb close <id>         # Complete work
hb sync               # Sync with git
```

Note: hb automatically sets --assignee to your ATProto handle on create, update, and close commands.

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   hb sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
