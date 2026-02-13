package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

// findBdBinary locates the bd binary in PATH
func findBdBinary() (string, error) {
	path, err := exec.LookPath("bd")
	if err != nil {
		return "", fmt.Errorf("bd binary not found in PATH. Install beads: https://github.com/steveyegge/beads")
	}
	return path, nil
}

// rewritePatterns are compiled once for performance
var rewritePatterns = []struct {
	re   *regexp.Regexp
	repl string
}{
	// Backtick-quoted commands: `bd foo` → `hb foo`
	{regexp.MustCompile("`bd "), "`hb "},
	{regexp.MustCompile("`bd`"), "`hb`"},

	// Markdown bold: **bd** → **hb**
	{regexp.MustCompile(`\*\*bd\*\*`), "**hb**"},

	// Branding: "bd (beads)" → "hb (heartbeads)"
	{regexp.MustCompile(`bd \(beads\)`), "hb (heartbeads)"},

	// Prose references: " bd " → " hb " but NOT issue IDs like "bd-xxxx"
	// Match " bd " only when NOT followed by a hyphen (issue ID)
	{regexp.MustCompile(` bd ([^-])`), " hb ${1}"},

	// Start-of-line: "bd " at beginning of line → "hb " (but not "bd-")
	{regexp.MustCompile(`(?m)^bd ([^-])`), "hb ${1}"},

	// Flag references: "bd --" → "hb --"
	{regexp.MustCompile(`bd --`), "hb --"},

	// Quoted references: "bd" surrounded by quotes
	{regexp.MustCompile(`"bd"`), `"hb"`},

	// Usage line: "  bd " (indented command) → "  hb "
	{regexp.MustCompile(`(  )bd `), "${1}hb "},
}

// rewriteOutput replaces "bd" references with "hb" in command output,
// while preserving issue IDs like "bd-w382l"
func rewriteOutput(input []byte) []byte {
	if len(input) == 0 {
		return input
	}

	result := string(input)
	for _, p := range rewritePatterns {
		result = p.re.ReplaceAllString(result, p.repl)
	}

	return []byte(result)
}

// runBd executes the bd binary with the given arguments, setting BD_NAME=hb
// and applying output rewriting. Returns rewritten stdout, stderr, exit code,
// and any execution error.
func runBd(ctx context.Context, args []string, extraEnv ...string) (stdout []byte, stderr []byte, exitCode int, err error) {
	bdPath, err := findBdBinary()
	if err != nil {
		return nil, nil, 1, err
	}

	cmd := exec.CommandContext(ctx, bdPath, args...)

	// Inherit env and add BD_NAME=hb
	cmd.Env = append(os.Environ(), "BD_NAME=hb")
	cmd.Env = append(cmd.Env, extraEnv...)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()

	// Get exit code
	exitCode = 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, nil, 1, fmt.Errorf("failed to run bd: %w", runErr)
		}
	}

	// Apply output rewriting
	stdout = rewriteOutput(stdoutBuf.Bytes())
	stderr = rewriteOutput(stderrBuf.Bytes())

	return stdout, stderr, exitCode, nil
}
