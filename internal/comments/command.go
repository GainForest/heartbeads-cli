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
var CmdComment = &cli.Command{
	Name:  "comment",
	Usage: "View or manage ATProto comments on beads issues",
	Description: `Read and write comments stored on the AT Protocol network.

Comments are org.impactindexer.review.comment records in each user's ATProto
repo, indexed by Hypergoat and displayed on the heartbeads map.

Reading comments does not require login. Writing requires an ATProto session
(run: hb account login).

Examples:
  hb comment get                                Last 10 comments
  hb comment get beads-map-3jy                  View threaded comments
  hb comment get -n 5                           Last 5 comments
  hb comment get --filter "beads-map-*"         Filter by glob pattern
  hb comment get beads-map-3jy --json           Machine-readable output
  hb comment add beads-map-3jy "LGTM"           Post a comment (requires login)
  hb comment add --reply-to at://... beads-map-3jy "thanks!"`,
	Action: fallbackAction,
	Commands: []*cli.Command{
		{
			Name:      "get",
			Usage:     "Get comments for a beads issue or across all issues",
			ArgsUsage: "[beads-id]",
			Description: `Fetch and display threaded comments.

Without arguments, shows the last 10 comments across all issues.
With a beads-id, shows all comments for that specific issue.

Comments are fetched from the Hypergoat GraphQL indexer and Bluesky profiles
are resolved for each commenter. No login required.

Examples:
  hb comment get                        Last 10 comments across all issues
  hb comment get beads-map-3jy          All comments for beads-map-3jy
  hb comment get -n 5                   Last 5 comments across all issues
  hb comment get -n 0                   All comments (no limit)
  hb comment get --filter "beads-map-*" Comments matching glob pattern`,
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
				&cli.IntFlag{
					Name:  "n",
					Usage: "Maximum number of root comments to show (0 = unlimited)",
					Value: 0,
				},
				&cli.StringFlag{
					Name:  "filter",
					Usage: "Glob pattern to match against nodeId (e.g. beads-map-*)",
				},
			},
			Action: runCommentsGet,
		},
		{
			Name:      "add",
			Usage:     "Add a comment to a beads issue",
			ArgsUsage: "<beads-id> <text>",
			Description: `Post a comment to a beads issue via ATProto (requires login).

The comment is written as an org.impactindexer.review.comment record to your
ATProto repo. Use --reply-to to reply to an existing comment by its AT-URI.`,
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
	beadsID := cmd.Args().First() // may be empty now

	jsonOutput := cmd.Bool("json")
	indexerURL := cmd.String("indexer-url")
	profileAPIURL := cmd.String("profile-api-url")
	limit := cmd.Int("n")
	filter := cmd.String("filter")

	// Build fetch options
	opts := FetchOptions{
		BeadsID: beadsID,
		Pattern: filter,
		Limit:   limit,
	}

	// Default limit: 10 when no beads-id and no explicit -n flag was set
	if beadsID == "" && !cmd.IsSet("n") {
		opts.Limit = 10
	}

	comments, err := FetchComments(ctx, indexerURL, profileAPIURL, opts)
	if err != nil {
		return fmt.Errorf("failed to fetch comments: %w", err)
	}

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
	_, _ = fmt.Fprintf(cmd.Root().Writer, "Comment posted: %s\n", output.URI)
	return nil
}

// fallbackAction shows help when no subcommand is provided
func fallbackAction(ctx context.Context, cmd *cli.Command) error {
	return cli.ShowSubcommandHelp(cmd)
}
