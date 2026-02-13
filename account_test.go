package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bluesky-social/indigo/atproto/syntax"
)

func TestAccountLogout(t *testing.T) {
	t.Run("logout removes session file", func(t *testing.T) {
		tmpDir := setupTestXDG(t)

		// Create a session file
		sessDir := filepath.Join(tmpDir, "heartbeads")
		if err := os.MkdirAll(sessDir, 0700); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		sess := AuthSession{
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

func TestAccountLogin_MissingCredentials(t *testing.T) {
	setupTestXDG(t)

	var buf bytes.Buffer
	err := runWithOutput([]string{"hb", "account", "login"}, &buf)
	if err == nil {
		t.Fatal("expected error for missing required flags")
	}
}
