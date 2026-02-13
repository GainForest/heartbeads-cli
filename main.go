package main

import (
	"context"
	"fmt"
	"io"
	"os"

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
	}
}
