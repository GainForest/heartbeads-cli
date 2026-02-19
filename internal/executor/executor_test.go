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
		// New dolt-backend commands
		{
			name:  "bd vc command start of line",
			input: "bd vc status",
			want:  "hb vc status",
		},
		{
			name:  "bd vc in backticks",
			input: "Run `bd vc` to check version control",
			want:  "Run `hb vc` to check version control",
		},
		{
			name:  "bd dolt command start of line",
			input: "bd dolt log",
			want:  "hb dolt log",
		},
		{
			name:  "bd dolt in backticks",
			input: "Use `bd dolt` for dolt operations",
			want:  "Use `hb dolt` for dolt operations",
		},
		{
			name:  "bd mol command start of line",
			input: "bd mol list",
			want:  "hb mol list",
		},
		{
			name:  "bd mol in backticks",
			input: "Run `bd mol` to manage molecules",
			want:  "Run `hb mol` to manage molecules",
		},
		{
			name:  "bd sql command start of line",
			input: "bd sql query",
			want:  "hb sql query",
		},
		{
			name:  "bd sql in backticks",
			input: "Use `bd sql` to run queries",
			want:  "Use `hb sql` to run queries",
		},
		{
			name:  "bd gate command start of line",
			input: "bd gate check",
			want:  "hb gate check",
		},
		{
			name:  "bd gate in backticks",
			input: "Run `bd gate` to check gates",
			want:  "Run `hb gate` to check gates",
		},
		{
			name:  "bd migrate command start of line",
			input: "bd migrate --from sqlite",
			want:  "hb migrate --from sqlite",
		},
		{
			name:  "bd migrate in backticks",
			input: "Run `bd migrate` to migrate your data",
			want:  "Run `hb migrate` to migrate your data",
		},
		{
			name:  "bd migrate in prose",
			input: "use bd migrate to upgrade",
			want:  "use hb migrate to upgrade",
		},
		{
			name:  "issue ID preserved with new commands nearby",
			input: "bd vc shows bd-w382l status",
			want:  "hb vc shows bd-w382l status",
		},
		{
			name:  "JSON with new command names preserved",
			input: `{"command":"bd vc","result":"ok"}`,
			want:  `{"command":"bd vc","result":"ok"}`,
		},
		{
			name:  "indented bd vc usage line",
			input: "  bd vc [options]",
			want:  "  hb vc [options]",
		},
		{
			name:  "indented bd migrate usage line",
			input: "  bd migrate [flags]",
			want:  "  hb migrate [flags]",
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
