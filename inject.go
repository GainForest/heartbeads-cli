package main

import (
	"os"
	"os/exec"
	"strings"
)

// hasFlag returns true if any of the given flag names appear in args.
// Checks both --flag and -f and --flag=value forms.
func hasFlag(args []string, flags ...string) bool {
	for _, arg := range args {
		for _, flag := range flags {
			// Exact match: --flag or -f
			if arg == flag {
				return true
			}
			// --flag=value form
			if strings.HasPrefix(arg, flag+"=") {
				return true
			}
		}
	}
	return false
}

// getLatestGitCommit runs `git log -1 --format=%s` to get the latest commit
// subject line. Returns "" on any error (no repo, no commits).
func getLatestGitCommit() string {
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// getSessionID checks CLAUDE_SESSION_ID env first, then OPENCODE_SESSION.
// Returns "" if neither is set.
func getSessionID() string {
	if id := os.Getenv("CLAUDE_SESSION_ID"); id != "" {
		return id
	}
	return os.Getenv("OPENCODE_SESSION")
}

// injectFlags appends flags (actor, assignee, reason, session) to args based
// on the subcommand and logged-in handle. args[0] is the bd subcommand.
//
// Injection rules:
//   - --actor <handle>: ALL commands (global flag, controls created_by)
//   - --assignee <handle>: create, update ONLY (NOT close, NOT q)
//   - --reason <commit>: close ONLY (from latest git commit subject)
//   - --session <id>: close, update ONLY (from CLAUDE_SESSION_ID or OPENCODE_SESSION env)
//
// If handle is empty, skip --assignee and --actor (return args unchanged).
// If git log fails, skip --reason silently.
// If no session env var set, skip --session silently.
func injectFlags(args []string, handle string) []string {
	if len(args) == 0 {
		return args
	}

	subcommand := args[0]
	result := make([]string, len(args))
	copy(result, args)

	// --actor (ALL commands) — only if handle is non-empty and not already present
	if handle != "" && !hasFlag(result, "--actor") {
		result = append(result, "--actor", handle)
	}

	// --assignee (create, update ONLY) — only if handle is non-empty and not already present
	if handle != "" && (subcommand == "create" || subcommand == "update") && !hasFlag(result, "--assignee", "-a") {
		result = append(result, "--assignee", handle)
	}

	// --reason (close ONLY) — from latest git commit, only if not already present
	if subcommand == "close" && !hasFlag(result, "--reason", "-r") {
		if reason := getLatestGitCommit(); reason != "" {
			result = append(result, "--reason", reason)
		}
	}

	// --session (close, update ONLY) — from env, only if not already present
	if (subcommand == "close" || subcommand == "update") && !hasFlag(result, "--session") {
		if sessionID := getSessionID(); sessionID != "" {
			result = append(result, "--session", sessionID)
		}
	}

	return result
}
