package comments

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// CmdComment is the "comment" command group.
// Subcommands:
//   - get <beads-id>: fetch and display comments (native, no auth required)
//
// Fallback: proxy unrecognized subcommands to bd via proxy.ExecBd
var CmdComment = &cli.Command{
	Name:   "comment",
	Usage:  "View or manage comments",
	Action: fallbackAction,
	Commands: []*cli.Command{
		{
			Name:      "get",
			Usage:     "Get comments for a beads issue",
			ArgsUsage: "<beads-id>",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "json",
					Usage: "Output as JSON",
				},
				&cli.StringFlag{
					Name:    "indexer-url",
					Usage:   "Hypergoat indexer URL",
					Value:   DefaultIndexerURL,
					Sources: cli.EnvVars("INDEXER_URL"),
				},
				&cli.StringFlag{
					Name:  "profile-api-url",
					Usage: "Bluesky profile API URL",
					Value: DefaultProfileAPIURL,
				},
			},
			Action: runCommentsGet,
		},
	},
}

// runCommentsGet fetches and displays comments for a beads issue
func runCommentsGet(ctx context.Context, cmd *cli.Command) error {
	// Get beads-id from args
	beadsID := cmd.Args().First()
	if beadsID == "" {
		return fmt.Errorf("usage: hb comment get <beads-id>")
	}

	// Get flag values
	jsonOutput := cmd.Bool("json")
	indexerURL := cmd.String("indexer-url")
	profileAPIURL := cmd.String("profile-api-url")

	// Fetch comments
	comments, err := FetchComments(ctx, indexerURL, profileAPIURL, beadsID)
	if err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

	// Format output
	if jsonOutput {
		return FormatJSON(cmd.Root().Writer, comments)
	}

	FormatText(cmd.Root().Writer, comments)
	return nil
}

// fallbackAction shows help when no subcommand is provided
func fallbackAction(ctx context.Context, cmd *cli.Command) error {
	return cli.ShowSubcommandHelp(cmd)
}
