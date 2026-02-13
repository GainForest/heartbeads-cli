package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bluesky-social/indigo/atproto/syntax"
)

func TestRequireAuth_NotLoggedIn(t *testing.T) {
	setupTestXDG(t)

	_, err := requireAuth()
	if err == nil {
		t.Fatal("expected error when not logged in")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Not logged in") {
		t.Errorf("expected 'Not logged in' in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "hb account login") {
		t.Errorf("expected 'hb account login' in error, got: %s", errMsg)
	}
}

func TestRequireAuth_LoggedIn(t *testing.T) {
	tmpDir := setupTestXDG(t)

	// Create session file
	sessDir := filepath.Join(tmpDir, "heartbeads")
	if err := os.MkdirAll(sessDir, 0700); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	sess := AuthSession{
		DID:    syntax.DID("did:plc:guardtest"),
		Handle: "guard.bsky.social",
	}
	data, _ := json.Marshal(sess)
	if err := os.WriteFile(filepath.Join(sessDir, "auth-session.json"), data, 0600); err != nil {
		t.Fatalf("failed to write session: %v", err)
	}

	loaded, err := requireAuth()
	if err != nil {
		t.Fatalf("requireAuth should succeed: %v", err)
	}

	if loaded.Handle != "guard.bsky.social" {
		t.Errorf("handle mismatch: got %s, want guard.bsky.social", loaded.Handle)
	}
}
