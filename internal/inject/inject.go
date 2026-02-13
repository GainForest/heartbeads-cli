package inject

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// HasFlag returns true if any of the given flag names appear in args.
// Checks --flag, -f, and --flag=value forms.
func HasFlag(args []string, flags ...string) bool {
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

// GetLatestGitCommit runs `git log -1 --format=%s` to get the latest commit
// subject line. Returns "" on any error (no repo, no commits).
func GetLatestGitCommit() string {
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// GetSessionID checks CLAUDE_SESSION_ID env first, then OPENCODE_SESSION.
// Returns "" if neither is set.
func GetSessionID() string {
	if id := os.Getenv("CLAUDE_SESSION_ID"); id != "" {
		return id
	}
	return os.Getenv("OPENCODE_SESSION")
}

// GetFlagValue extracts the value of a flag from args.
// Handles both "--flag value" and "--flag=value" forms.
// Returns "" if the flag is not found or has no value.
func GetFlagValue(args []string, flags ...string) string {
	for i, arg := range args {
		for _, flag := range flags {
			// --flag=value form
			if strings.HasPrefix(arg, flag+"=") {
				return strings.TrimPrefix(arg, flag+"=")
			}
			// --flag value form (next arg is the value)
			if arg == flag && i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return ""
}

// reasonPattern matches "<7+ hex chars> <message>" — a git commit hash followed by text.
var reasonPattern = regexp.MustCompile(`^[0-9a-f]{7,40}\s+.+`)

// RequireReason checks that "close" commands include --reason/-r with a valid
// commit reference in the format "<hash> <message>".
// Returns an error if missing or malformed. Non-close commands always pass.
func RequireReason(args []string) error {
	if len(args) == 0 || args[0] != "close" {
		return nil
	}
	if !HasFlag(args, "--reason", "-r") {
		return fmt.Errorf("hb close requires --reason \"<commit-hash> <message>\"\n  example: hb close abc123 --reason \"a1b2c3d fix: resolve login timeout\"")
	}
	reason := GetFlagValue(args, "--reason", "-r")
	if reason == "" || !reasonPattern.MatchString(reason) {
		return fmt.Errorf("invalid --reason format: must be \"<commit-hash> <message>\"\n  got:      %q\n  expected: \"a1b2c3d fix: resolve login timeout\"", reason)
	}
	return nil
}

// InjectFlags appends flags (actor, assignee, session) to args based
// on the subcommand and logged-in handle. args[0] is the bd subcommand.
//
// Injection rules:
//   - --actor <handle>: ALL commands (global flag, controls created_by)
//   - --assignee <handle>: update ONLY (NOT create, NOT close, NOT q)
//   - --session <id>: close, update ONLY (from CLAUDE_SESSION_ID or OPENCODE_SESSION env)
//
// If handle is empty, skip --assignee and --actor (return args unchanged).
// If no session env var set, skip --session silently.
//
// Note: --reason is NOT auto-injected. Use RequireReason to enforce it on close.
func InjectFlags(args []string, handle string) []string {
	if len(args) == 0 {
		return args
	}

	subcommand := args[0]
	result := make([]string, len(args))
	copy(result, args)

	// --actor (ALL commands) — only if handle is non-empty and not already present
	if handle != "" && !HasFlag(result, "--actor") {
		result = append(result, "--actor", handle)
	}

	// --assignee (update ONLY) — only if handle is non-empty and not already present
	// NOT on create: auto-assignee breaks "update --claim" on newly created issues
	if handle != "" && subcommand == "update" && !HasFlag(result, "--assignee", "-a") {
		result = append(result, "--assignee", handle)
	}

	// --session (close, update ONLY) — from env, only if not already present
	if (subcommand == "close" || subcommand == "update") && !HasFlag(result, "--session") {
		if sessionID := GetSessionID(); sessionID != "" {
			result = append(result, "--session", sessionID)
		}
	}

	return result
}
