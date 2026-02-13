package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/urfave/cli/v3"
)

// execBd authenticates, injects flags, runs bd, and writes output.
// Returns an error if auth fails, bd fails to execute, or exits non-zero.
func execBd(ctx context.Context, w io.Writer, args []string) error {
	sess, err := requireAuth()
	if err != nil {
		return err
	}

	args = injectFlags(args, sess.Handle)

	stdout, stderr, exitCode, err := runBd(ctx, args)
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

// proxyAction is the unified action for all proxied bd commands.
// It checks auth, builds args with assignee injection, and delegates to bd.
func proxyAction(ctx context.Context, cmd *cli.Command) error {
	args := append([]string{cmd.Name}, cmd.Args().Slice()...)
	return execBd(ctx, cmd.Root().Writer, args)
}

// buildProxyCommands returns cli.Command entries for common bd commands.
// Each command uses SkipFlagParsing=true so bd handles all flag parsing.
func buildProxyCommands() []*cli.Command {
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

		// Sync
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
		{"comments", "View or manage comments"},
		{"label", "Manage issue labels"},

		// Other
		{"delete", "Delete one or more issues"},
		{"reopen", "Reopen closed issues"},
		{"count", "Count issues matching filters"},
		{"stale", "Show stale issues"},
		{"q", "Quick capture: create issue and output only ID"},
		{"rename", "Rename an issue ID"},
		{"todo", "Manage TODO items"},
	}

	result := make([]*cli.Command, 0, len(commands))
	for _, c := range commands {
		result = append(result, &cli.Command{
			Name:            c.name,
			Usage:           c.usage,
			Action:          proxyAction,
			SkipFlagParsing: true,
			HideHelpCommand: true,
		})
	}
	return result
}
