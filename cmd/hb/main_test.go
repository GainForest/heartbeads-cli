package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/gainforest/heartbeads-cli/internal/auth"
)

// setupTestXDG configures xdg to use a temporary directory for tests
func setupTestXDG(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)
	xdg.Reload()
	return tmpDir
}

// TestCLI tests basic CLI functionality (version, help, no-args)
func TestCLI(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "version flag prints hb version",
			args:       []string{"hb", "--version"},
			wantErr:    false,
			wantOutput: "hb version",
		},
		{
			name:       "no args shows help",
			args:       []string{"hb"},
			wantErr:    false,
			wantOutput: "hb",
		},
		{
			name:       "help flag shows usage",
			args:       []string{"hb", "--help"},
			wantErr:    false,
			wantOutput: "Authenticated beads CLI for AI agents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := runWithOutput(tt.args, &buf)

			if (err != nil) != tt.wantErr {
				t.Errorf("runWithOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()
			if !strings.Contains(output, tt.wantOutput) {
				t.Errorf("output %q does not contain %q", output, tt.wantOutput)
			}
		})
	}
}

// TestCatchallNoAuth tests that unknown commands require auth
func TestCatchallNoAuth(t *testing.T) {
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "some-unknown-command"}, &buf)
	if err == nil {
		t.Fatal("expected error for unknown command without auth")
	}

	if !strings.Contains(err.Error(), "Not logged in") {
		t.Errorf("expected 'Not logged in' error, got: %v", err)
	}
}

// TestCatchallHelp tests that help works without auth
func TestCatchallHelp(t *testing.T) {
	var buf bytes.Buffer
	err := runWithOutput([]string{"hb"}, &buf)
	if err != nil {
		t.Fatalf("expected help to succeed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "account") {
		t.Errorf("help should mention 'account', got: %s", output)
	}
}

// TestProxyAction_NotLoggedIn tests that proxy commands require auth
func TestProxyAction_NotLoggedIn(t *testing.T) {
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "list"}, &buf)
	if err == nil {
		t.Fatal("expected error when not logged in")
	}

	if !strings.Contains(err.Error(), "Not logged in") {
		t.Errorf("expected 'Not logged in' error, got: %v", err)
	}
}

// TestProxyHelp_NoAuthRequired tests that help works without auth
func TestProxyHelp_NoAuthRequired(t *testing.T) {
	// Help should always work without auth
	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "--help"}, &buf)
	if err != nil {
		t.Fatalf("help should work without auth: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "account") {
		t.Errorf("help should mention 'account' command, got: %s", output)
	}
	if !strings.Contains(output, "list") {
		t.Errorf("help should mention 'list' command, got: %s", output)
	}
}

// TestAccountLogout tests the logout command
func TestAccountLogout(t *testing.T) {
	t.Run("logout removes session file", func(t *testing.T) {
		tmpDir := setupTestXDG(t)

		// Create a session file
		sessDir := filepath.Join(tmpDir, "heartbeads")
		if err := os.MkdirAll(sessDir, 0700); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		sess := auth.Session{
			DID:    syntax.DID("did:plc:testuser"),
			Handle: "test.bsky.social",
		}
		data, _ := json.Marshal(sess)
		sessFile := filepath.Join(sessDir, "auth-session.json")
		if err := os.WriteFile(sessFile, data, 0600); err != nil {
			t.Fatalf("failed to write session: %v", err)
		}

		// Run logout
		var buf bytes.Buffer
		err := runWithOutput([]string{"hb", "account", "logout"}, &buf)
		if err != nil {
			t.Fatalf("logout failed: %v", err)
		}

		// Verify session file is gone
		if _, err := os.Stat(sessFile); !os.IsNotExist(err) {
			t.Error("session file should be deleted after logout")
		}

		// Verify output
		if !strings.Contains(buf.String(), "Logged out") {
			t.Errorf("expected 'Logged out' message, got: %s", buf.String())
		}
	})

	t.Run("logout succeeds when no session exists", func(t *testing.T) {
		setupTestXDG(t)

		var buf bytes.Buffer
		err := runWithOutput([]string{"hb", "account", "logout"}, &buf)
		if err != nil {
			t.Fatalf("logout should succeed even with no session: %v", err)
		}
	})
}

// TestAccountStatus_NotLoggedIn tests that status requires auth
func TestAccountStatus_NotLoggedIn(t *testing.T) {
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "account", "status"}, &buf)
	if err == nil {
		t.Fatal("expected error when not logged in")
	}

	if !strings.Contains(err.Error(), "not logged in") {
		t.Errorf("expected 'not logged in' error, got: %v", err)
	}
}

// TestAccountLogin_MissingCredentials tests that login requires credentials
func TestAccountLogin_MissingCredentials(t *testing.T) {
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "account", "login"}, &buf)
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
}

// TestCommentGetNoAuth tests that comment get does not require auth
func TestCommentGetNoAuth(t *testing.T) {
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "comment", "get"}, &buf)
	if err == nil {
		t.Fatal("expected error for missing beads-id")
	}

	// Should error about usage, not auth
	if !strings.Contains(err.Error(), "usage") {
		t.Errorf("expected 'usage' error, got: %v", err)
	}
}

// TestCommentGetHelp tests that comment get help works
func TestCommentGetHelp(t *testing.T) {
	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "comment", "get", "--help"}, &buf)
	if err != nil {
		t.Fatalf("help should work: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "beads-id") {
		t.Errorf("help should mention 'beads-id', got: %s", output)
	}
}

// TestCommentInHelp tests that comment appears in main help
func TestCommentInHelp(t *testing.T) {
	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "--help"}, &buf)
	if err != nil {
		t.Fatalf("help should work: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "comment") {
		t.Errorf("help should mention 'comment' command, got: %s", output)
	}
}
