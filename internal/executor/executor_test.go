package executor

import (
	"context"
	"strings"
	"testing"
)

func TestFindBdBinary(t *testing.T) {
	path, err := FindBdBinary()
	if err != nil {
		t.Skipf("bd not installed: %v", err)
	}
	if path == "" {
		t.Error("FindBdBinary returned empty path")
	}
}

func TestRewriteOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "backtick command",
			input: "Run `bd ready` to find work",
			want:  "Run `hb ready` to find work",
		},
		{
			name:  "backtick standalone",
			input: "Use `bd` for tracking",
			want:  "Use `hb` for tracking",
		},
		{
			name:  "backtick create",
			input: "`bd create` - New issue",
			want:  "`hb create` - New issue",
		},
		{
			name:  "flag reference",
			input: "bd --help",
			want:  "hb --help",
		},
		{
			name:  "branding",
			input: "bd (beads) is great",
			want:  "hb (heartbeads) is great",
		},
		{
			name:  "bold markdown",
			input: "**bd** is the tool",
			want:  "**hb** is the tool",
		},
		{
			name:  "issue ID preserved",
			input: "bd-w382l is an issue",
			want:  "bd-w382l is an issue",
		},
		{
			name:  "start of line command",
			input: "bd sync runs sync",
			want:  "hb sync runs sync",
		},
		{
			name:  "prose with space",
			input: "use bd for tracking",
			want:  "use hb for tracking",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "indented usage line",
			input: "  bd create [title]",
			want:  "  hb create [title]",
		},
		{
			name:  "quoted bd",
			input: `run "bd" for help`,
			want:  `run "hb" for help`,
		},
		{
			name:  "multiple bd references",
			input: "`bd ready` and `bd close`",
			want:  "`hb ready` and `hb close`",
		},
		{
			name:  "issue ID in middle preserved",
			input: "see bd-abc123 for details",
			want:  "see bd-abc123 for details",
		},
		{
			name:  "JSON output preserved",
			input: `{"close_reason":"fix bd rewriting"}`,
			want:  `{"close_reason":"fix bd rewriting"}`,
		},
		{
			name:  "JSON array preserved",
			input: `[{"title":"fix bd bug"}]`,
			want:  `[{"title":"fix bd bug"}]`,
		},
		{
			name:  "path-prefixed bd",
			input: "BEADS_DB=/tmp/test.db ./bd create",
			want:  "BEADS_DB=/tmp/test.db ./hb create",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(RewriteOutput([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("RewriteOutput(%q)\n  got:  %q\n  want: %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRunBd(t *testing.T) {
	// Skip if bd is not installed
	if _, err := FindBdBinary(); err != nil {
		t.Skipf("bd not installed: %v", err)
	}

	ctx := context.Background()
	stdout, _, exitCode, err := RunBd(ctx, []string{"version"}, "")
	if err != nil {
		t.Fatalf("RunBd failed: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}

	out := string(stdout)
	if !strings.Contains(out, "hb version") {
		t.Errorf("expected stdout to contain 'hb version', got: %q", out)
	}
}

func TestRunBdFallbackEnv(t *testing.T) {
	if _, err := FindBdBinary(); err != nil {
		t.Skipf("bd not installed: %v", err)
	}

	// Unset GIT_AUTHOR_EMAIL to test fallback
	t.Setenv("GIT_AUTHOR_EMAIL", "")

	ctx := context.Background()
	// Run a command that echoes env — "version" is safe
	_, _, exitCode, err := RunBd(ctx, []string{"version"}, "test.handle.social")
	if err != nil {
		t.Fatalf("RunBd failed: %v", err)
	}
	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d", exitCode)
	}
}
