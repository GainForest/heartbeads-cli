package comments

import (
	"context"
	"fmt"
	"strings"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/gainforest/heartbeads-cli/internal/auth"
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
		{
			Name:      "add",
			Usage:     "Add a comment to a beads issue",
			ArgsUsage: "<beads-id> <text>",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "reply-to",
					Usage: "AT-URI of parent comment to reply to",
				},
			},
			Action: runCommentAdd,
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

// runCommentAdd posts a new comment to a beads issue
func runCommentAdd(ctx context.Context, cmd *cli.Command) error {
	// Validate args
	if cmd.Args().Len() < 2 {
		return fmt.Errorf("usage: hb comment add [--reply-to <at-uri>] <beads-id> <text>")
	}

	// Extract beads-id and text
	beadsID := cmd.Args().First()
	text := strings.Join(cmd.Args().Slice()[1:], " ")

	// Get --reply-to flag
	replyTo := cmd.String("reply-to")

	// Load authenticated client
	client, err := auth.LoadClient(ctx)
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	// Get session info to get the DID
	sess, err := comatproto.ServerGetSession(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Create the comment
	output, err := CreateComment(ctx, client, sess.Did, CreateCommentInput{
		BeadsID: beadsID,
		Text:    text,
		ReplyTo: replyTo,
	})
	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	// Print success message
	fmt.Fprintf(cmd.Root().Writer, "Comment posted: %s\n", output.URI)
	return nil
}

// fallbackAction shows help when no subcommand is provided
func fallbackAction(ctx context.Context, cmd *cli.Command) error {
	return cli.ShowSubcommandHelp(cmd)
}
