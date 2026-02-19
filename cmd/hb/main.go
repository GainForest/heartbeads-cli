package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/adrg/xdg"
	"github.com/gainforest/heartbeads-cli/internal/account"
	"github.com/gainforest/heartbeads-cli/internal/comments"
	"github.com/gainforest/heartbeads-cli/internal/migrate"
	"github.com/gainforest/heartbeads-cli/internal/proxy"
	"github.com/urfave/cli/v3"
)

// Version can be set at build time with -ldflags="-X main.Version=X.Y.Z"
var Version = "dev"

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithOutput(args, os.Stdout)
}

// skipNudgeCommands are commands that should not show the migration nudge.
var skipNudgeCommands = map[string]bool{
	"migrate": true,
	"version": true,
	"help":    true,
	"account": true,
}

// maybeNudgeMigrate prints a one-time-per-day warning to stderr if the current
// project uses a legacy beads backend (sqlite or jsonl). It is a best-effort
// check: any error is silently ignored so it never blocks command execution.
func maybeNudgeMigrate(args []string) {
	// Determine the first non-flag argument (the subcommand name).
	subcommand := ""
	for _, a := range args[1:] {
		if len(a) > 0 && a[0] != '-' {
			subcommand = a
			break
		}
	}

	// Skip nudge for exempt commands and global flags like --version / --help.
	if subcommand == "" || skipNudgeCommands[subcommand] {
		return
	}
	// Also skip if the first real arg is a global flag (e.g. --version).
	if len(args) > 1 && len(args[1]) > 0 && args[1][0] == '-' {
		return
	}

	// Find the .beads directory by walking up from cwd.
	beadsDir := findBeadsDirBestEffort()
	if beadsDir == "" {
		return
	}

	backend := migrate.DetectBackend(beadsDir)
	if backend != "sqlite" && backend != "jsonl" {
		return
	}

	// Rate-limit: only show once per day using a timestamp file.
	nudgeTSPath, err := xdg.StateFile("heartbeads/migrate-nudge-ts")
	if err != nil {
		// Can't determine path — skip nudge.
		return
	}

	now := time.Now()
	if data, err := os.ReadFile(nudgeTSPath); err == nil {
		if ts, err := time.Parse(time.RFC3339, string(data)); err == nil {
			if now.Sub(ts) < 24*time.Hour {
				// Already nudged within the last day.
				return
			}
		}
	}

	// Write the current timestamp.
	if err := os.MkdirAll(filepath.Dir(nudgeTSPath), 0700); err == nil {
		_ = os.WriteFile(nudgeTSPath, []byte(now.Format(time.RFC3339)), 0600)
	}

	fmt.Fprintln(os.Stderr, "⚠ This project uses the legacy beads backend. Run `hb migrate` to upgrade to dolt.")
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

func runWithOutput(args []string, w io.Writer) error {
	maybeNudgeMigrate(args)
	app := buildApp(w)
	return app.Run(context.Background(), args)
}

func buildApp(w io.Writer) *cli.Command {
	return &cli.Command{
		Name:    "hb",
		Usage:   "Authenticated beads CLI for AI agents",
		Version: Version,
		Writer:  w,
		ExitErrHandler: func(ctx context.Context, cmd *cli.Command, err error) {
			// Don't call os.Exit, just let the error propagate
		},
		Action: catchallAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "plc-host",
				Usage:   "PLC directory URL",
				Value:   "https://plc.directory",
				Sources: cli.EnvVars("ATP_PLC_HOST"),
			},
		},
		Commands: append([]*cli.Command{
			account.CmdAccount,
			comments.CmdComment,
			migrate.CmdMigrate,
		}, proxy.BuildProxyCommands()...),
	}
}

// catchallAction handles the root command invocation.
// If args look like an unknown subcommand, proxy them to bd.
// Otherwise, show help.
func catchallAction(ctx context.Context, cmd *cli.Command) error {
	args := cmd.Args().Slice()

	// No args — show help
	if len(args) == 0 {
		return cli.ShowAppHelp(cmd)
	}

	// If first arg starts with "-", it's a flag — show help
	if len(args[0]) > 0 && args[0][0] == '-' {
		return cli.ShowAppHelp(cmd)
	}

	// Looks like an unknown subcommand — proxy to bd with auth
	return proxy.ExecBd(ctx, cmd.Root().Writer, args)
}
