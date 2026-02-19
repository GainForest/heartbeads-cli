package proxy

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gainforest/heartbeads-cli/internal/auth"
	"github.com/gainforest/heartbeads-cli/internal/executor"
	"github.com/gainforest/heartbeads-cli/internal/inject"
	"github.com/gainforest/heartbeads-cli/internal/migrate"
	"github.com/urfave/cli/v3"
)

// ExecBd authenticates, validates required flags, injects flags, runs bd, and writes output.
// Returns an error if auth fails, validation fails, bd fails to execute, or exits non-zero.
func ExecBd(ctx context.Context, w io.Writer, args []string) error {
	// Validate required flags before auth (fast-fail on bad input)
	if err := inject.RequireReason(args); err != nil {
		return err
	}

	sess, err := auth.RequireAuth()
	if err != nil {
		return err
	}

	args = inject.InjectFlags(args, sess.Handle)

	stdout, stderr, exitCode, err := executor.RunBd(ctx, args, sess.Handle)
	if err != nil {
		return err
	}

	if len(stdout) > 0 {
		_, _ = w.Write(stdout)
	}
	if len(stderr) > 0 {
		_, _ = os.Stderr.Write(stderr)
	}

	if exitCode != 0 {
		return fmt.Errorf("hb %s failed with exit code %d", args[0], exitCode)
	}

	return nil
}

// ProxyAction is the unified action for all proxied bd commands.
// It checks auth, builds args with assignee injection, and delegates to bd.
func ProxyAction(ctx context.Context, cmd *cli.Command) error {
	args := append([]string{cmd.Name}, cmd.Args().Slice()...)
	return ExecBd(ctx, cmd.Root().Writer, args)
}

// findBeadsDirBestEffort walks up from cwd to find a .beads directory.
// Returns "" if not found or on any error.
func findBeadsDirBestEffort() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, ".beads")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// SyncAction proxies bd sync and prints an informational note about
// the no-op behavior only when using the dolt backend.
func SyncAction(ctx context.Context, cmd *cli.Command) error {
	beadsDir := findBeadsDirBestEffort()
	if beadsDir != "" && migrate.DetectBackend(beadsDir) == "dolt" {
		_, _ = fmt.Fprintln(os.Stderr, "Note: hb sync is a no-op with dolt backend. Changes are persisted automatically.")
	}
	args := append([]string{cmd.Name}, cmd.Args().Slice()...)
	return ExecBd(ctx, cmd.Root().Writer, args)
}

// BuildProxyCommands returns cli.Command entries for common bd commands.
// Each command uses SkipFlagParsing=true so bd handles all flag parsing.
func BuildProxyCommands() []*cli.Command {
	commands := []struct {
		name  string
		usage string
	}{
		// Core workflow
		{"init", "Initialize hb in the current directory"},
		{"list", "List issues"},
		{"ready", "Show issues ready to work (no blockers)"},
		{"show", "Show issue details"},
		{"create", "Create a new issue"},
		{"update", "Update one or more issues"},
		{"close", "Close one or more issues"},
		{"search", "Search issues by text query"},
		{"blocked", "Show blocked issues"},

		// Dependencies
		{"dep", "Manage dependencies"},

		// Sync (note: bd sync is a no-op with dolt backend — changes are persisted automatically)
		{"sync", "Sync with git"},
		{"export", "Export issues to JSONL"},
		{"import", "Import issues from JSONL"},

		// Setup
		{"onboard", "Display minimal snippet for AGENTS.md"},
		{"prime", "Output AI-optimized workflow context"},
		{"quickstart", "Quick start guide"},
		{"setup", "Setup integration with AI editors"},
		{"config", "Manage configuration settings"},
		{"info", "Show database and daemon information"},
		{"status", "Show issue database overview"},
		{"hooks", "Manage git hooks"},
		{"doctor", "Check for issues"},

		// Structure
		{"epic", "Epic management commands"},
		{"children", "List child beads of a parent"},
		{"graph", "Display issue dependency graph"},
		{"label", "Manage issue labels"},

		// Other
		{"delete", "Delete one or more issues"},
		{"reopen", "Reopen closed issues"},
		{"count", "Count issues matching filters"},
		{"stale", "Show stale issues"},
		{"q", "Quick capture: create issue and output only ID"},
		{"rename", "Rename an issue ID"},
		{"todo", "Manage TODO items"},

		// Dolt-era commands (beads v0.50+)
		{"vc", "Dolt version control (log, diff, commit)"},
		{"sql", "Raw SQL access to dolt database"},
		{"dolt", "Dolt management (show, set, test, start, stop, commit, push, pull)"},
		{"mol", "Molecule workflows"},
		{"gate", "Gate management"},
		{"where", "Show where beads dir is"},
		{"validate", "Validate issue data"},
		{"duplicate", "Mark issues as duplicates"},
		{"supersede", "Supersede an issue"},
	}

	result := make([]*cli.Command, 0, len(commands))
	for _, c := range commands {
		action := ProxyAction
		if c.name == "sync" {
			action = SyncAction
		}
		result = append(result, &cli.Command{
			Name:            c.name,
			Usage:           c.usage,
			Action:          action,
			SkipFlagParsing: true,
			HideHelpCommand: true,
		})
	}
	return result
}
