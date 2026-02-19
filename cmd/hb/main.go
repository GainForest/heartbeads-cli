package main

import (
	"context"
	"fmt"
	"io"
	"os"

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

func runWithOutput(args []string, w io.Writer) error {
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
