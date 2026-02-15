package comments

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/urfave/cli/v3"
)

// TestFallbackNoArgs verifies that running "hb comment" with no args shows help
// and does NOT contain "bd" branding
func TestFallbackNoArgs(t *testing.T) {
	var buf bytes.Buffer
	app := &cli.Command{
		Name:   "hb",
		Writer: &buf,
		Commands: []*cli.Command{
			CmdComment,
		},
	}

	err := app.Run(context.Background(), []string{"hb", "comment"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Verify output contains "get" (the subcommand)
	if !strings.Contains(output, "get") {
		t.Errorf("expected output to contain 'get' subcommand, got:\n%s", output)
	}

	// Verify output does NOT contain "bd"
	if strings.Contains(output, "bd") {
		t.Errorf("expected output to NOT contain 'bd', got:\n%s", output)
	}
}

// TestFallbackHelp verifies that running "hb comment --help" shows help
// and contains the "get" subcommand
func TestFallbackHelp(t *testing.T) {
	var buf bytes.Buffer
	app := &cli.Command{
		Name:   "hb",
		Writer: &buf,
		Commands: []*cli.Command{
			CmdComment,
		},
	}

	err := app.Run(context.Background(), []string{"hb", "comment", "--help"})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	output := buf.String()

	// Verify output contains "get"
	if !strings.Contains(output, "get") {
		t.Errorf("expected output to contain 'get' subcommand, got:\n%s", output)
	}
}
