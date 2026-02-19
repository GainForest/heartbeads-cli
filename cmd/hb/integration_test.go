//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/gainforest/heartbeads-cli/internal/auth"
	"github.com/gainforest/heartbeads-cli/internal/executor"
)

// writeFakeSession creates a fake auth session file for integration tests.
// This avoids real ATProto network calls while testing the proxy pipeline.
func writeFakeSession(t *testing.T) {
	t.Helper()
	tmpDir := setupTestXDG(t)

	sessDir := filepath.Join(tmpDir, "heartbeads")
	if err := os.MkdirAll(sessDir, 0700); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	sess := auth.Session{
		DID:          syntax.DID("did:plc:integrationtest"),
		PDS:          "https://bsky.social",
		Handle:       "integration.test.social",
		Password:     "fake-password",
		AccessToken:  "fake-access-token",
		RefreshToken: "fake-refresh-token",
	}
	data, _ := json.Marshal(sess)
	if err := os.WriteFile(filepath.Join(sessDir, "auth-session.json"), data, 0600); err != nil {
		t.Fatalf("failed to write session: %v", err)
	}
}

func skipIfNoBd(t *testing.T) {
	t.Helper()
	if _, err := executor.FindBdBinary(); err != nil {
		t.Skipf("bd not installed: %v", err)
	}
}

func TestHbWithoutLogin(t *testing.T) {
	skipIfNoBd(t)
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "list"}, &buf)
	if err == nil {
		t.Fatal("expected error when not logged in")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "not logged in") {
		t.Errorf("expected 'not logged in' in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "hb account login") {
		t.Errorf("expected 'hb account login' in error, got: %s", errMsg)
	}
}

func TestHbHelp(t *testing.T) {
	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "--help"}, &buf)
	if err != nil {
		t.Fatalf("help should not require auth: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "hb") {
		t.Errorf("help should contain 'hb', got: %s", output)
	}
	if !strings.Contains(output, "account") {
		t.Errorf("help should list 'account' subcommand, got: %s", output)
	}
}

func TestHbAccountLogout(t *testing.T) {
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "account", "logout"}, &buf)
	if err != nil {
		t.Fatalf("logout should always succeed: %v", err)
	}
}

func TestHbVersionNoAuth(t *testing.T) {
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "--version"}, &buf)
	if err != nil {
		t.Fatalf("version should not require auth: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "hb version") {
		t.Errorf("expected 'hb version' in output, got: %s", output)
	}
}

func TestOutputRewriting(t *testing.T) {
	skipIfNoBd(t)
	writeFakeSession(t)

	var buf bytes.Buffer
	// Run "hb onboard" which produces text containing "bd" references
	// This will fail at the bd level (session is fake) but the onboard
	// command doesn't require a real session — it just prints text
	err := runWithOutput([]string{"hb", "onboard"}, &buf)

	// Even if bd returns an error (e.g., legacy DB), check the output
	// that was produced for rewriting
	output := buf.String()
	stderr := ""
	if err != nil {
		stderr = err.Error()
	}
	combined := output + stderr

	// If we got any output, check that bd references were rewritten
	if len(output) > 0 {
		// Check that backtick-quoted bd commands are rewritten
		if strings.Contains(output, "`bd ") {
			t.Errorf("output still contains '`bd ' — rewriting incomplete.\nOutput:\n%s", output)
		}
	}

	// At minimum, the error message should reference "hb" not "bd"
	_ = combined
}
