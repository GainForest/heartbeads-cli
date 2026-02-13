package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/bluesky-social/indigo/atproto/syntax"
)

// setupTestXDG configures xdg to use a temporary directory for tests
func setupTestXDG(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)
	xdg.Reload()
	return tmpDir
}

func TestPersistAuthSession(t *testing.T) {
	tmpDir := setupTestXDG(t)

	sess := &AuthSession{
		DID:          syntax.DID("did:plc:testuser123"),
		PDS:          "https://bsky.social",
		Handle:       "test.bsky.social",
		Password:     "app-password-123",
		AccessToken:  "eyJ0eXAi...",
		RefreshToken: "eyJhbGci...",
	}

	err := persistAuthSession(sess)
	if err != nil {
		t.Fatalf("persistAuthSession failed: %v", err)
	}

	// Verify file exists with correct permissions
	fPath := filepath.Join(tmpDir, "heartbeads", "auth-session.json")
	info, err := os.Stat(fPath)
	if err != nil {
		t.Fatalf("session file not created: %v", err)
	}

	// Check file permissions (0600)
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("expected permissions 0600, got %o", perm)
	}

	// Verify content
	data, err := os.ReadFile(fPath)
	if err != nil {
		t.Fatalf("failed to read session file: %v", err)
	}

	var loaded AuthSession
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal session: %v", err)
	}

	if loaded.DID != sess.DID {
		t.Errorf("DID mismatch: got %s, want %s", loaded.DID, sess.DID)
	}
	if loaded.Handle != sess.Handle {
		t.Errorf("Handle mismatch: got %s, want %s", loaded.Handle, sess.Handle)
	}
	if loaded.Password != sess.Password {
		t.Errorf("Password not persisted")
	}
}

func TestLoadAuthSessionFile(t *testing.T) {
	t.Run("returns session when file exists", func(t *testing.T) {
		tmpDir := setupTestXDG(t)

		// Create session file
		sessDir := filepath.Join(tmpDir, "heartbeads")
		if err := os.MkdirAll(sessDir, 0700); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		sess := AuthSession{
			DID:          syntax.DID("did:plc:loadtest"),
			PDS:          "https://pds.example.com",
			Handle:       "user.example.com",
			Password:     "secret",
			AccessToken:  "access123",
			RefreshToken: "refresh456",
		}
		data, _ := json.Marshal(sess)
		if err := os.WriteFile(filepath.Join(sessDir, "auth-session.json"), data, 0600); err != nil {
			t.Fatalf("failed to write session: %v", err)
		}

		loaded, err := loadAuthSessionFile()
		if err != nil {
			t.Fatalf("loadAuthSessionFile failed: %v", err)
		}

		if loaded.DID != sess.DID {
			t.Errorf("DID mismatch: got %s, want %s", loaded.DID, sess.DID)
		}
		if loaded.Handle != sess.Handle {
			t.Errorf("Handle mismatch: got %s, want %s", loaded.Handle, sess.Handle)
		}
	})

	t.Run("returns ErrNoAuthSession when file missing", func(t *testing.T) {
		setupTestXDG(t)

		_, err := loadAuthSessionFile()
		if err != ErrNoAuthSession {
			t.Errorf("expected ErrNoAuthSession, got %v", err)
		}
	})
}

func TestWipeAuthSession(t *testing.T) {
	t.Run("deletes existing session", func(t *testing.T) {
		tmpDir := setupTestXDG(t)

		// Create session file
		sessDir := filepath.Join(tmpDir, "heartbeads")
		if err := os.MkdirAll(sessDir, 0700); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
		sessFile := filepath.Join(sessDir, "auth-session.json")
		if err := os.WriteFile(sessFile, []byte("{}"), 0600); err != nil {
			t.Fatalf("failed to write session: %v", err)
		}

		err := wipeAuthSession()
		if err != nil {
			t.Fatalf("wipeAuthSession failed: %v", err)
		}

		// Verify file is gone
		if _, err := os.Stat(sessFile); !os.IsNotExist(err) {
			t.Error("session file should be deleted")
		}
	})

	t.Run("handles missing file gracefully", func(t *testing.T) {
		setupTestXDG(t)

		err := wipeAuthSession()
		if err != nil {
			t.Errorf("wipeAuthSession should not error on missing file: %v", err)
		}
	})
}
