package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

func TestPersistSession(t *testing.T) {
	tmpDir := setupTestXDG(t)

	sess := &Session{
		DID:          syntax.DID("did:plc:testuser123"),
		PDS:          "https://bsky.social",
		Handle:       "test.bsky.social",
		Password:     "app-password-123",
		AccessToken:  "eyJ0eXAi...",
		RefreshToken: "eyJhbGci...",
	}

	err := PersistSession(sess)
	if err != nil {
		t.Fatalf("PersistSession failed: %v", err)
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

	var loaded Session
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

func TestLoadSessionFile(t *testing.T) {
	t.Run("returns session when file exists", func(t *testing.T) {
		tmpDir := setupTestXDG(t)

		// Create session file
		sessDir := filepath.Join(tmpDir, "heartbeads")
		if err := os.MkdirAll(sessDir, 0700); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		sess := Session{
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

		loaded, err := LoadSessionFile()
		if err != nil {
			t.Fatalf("LoadSessionFile failed: %v", err)
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

		_, err := LoadSessionFile()
		if err != ErrNoAuthSession {
			t.Errorf("expected ErrNoAuthSession, got %v", err)
		}
	})
}

func TestWipeSession(t *testing.T) {
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

		err := WipeSession()
		if err != nil {
			t.Fatalf("WipeSession failed: %v", err)
		}

		// Verify file is gone
		if _, err := os.Stat(sessFile); !os.IsNotExist(err) {
			t.Error("session file should be deleted")
		}
	})

	t.Run("handles missing file gracefully", func(t *testing.T) {
		setupTestXDG(t)

		err := WipeSession()
		if err != nil {
			t.Errorf("WipeSession should not error on missing file: %v", err)
		}
	})
}

func TestConfigDirectory(t *testing.T) {
	dir := ConfigDirectory()
	if dir == nil {
		t.Fatal("ConfigDirectory returned nil")
	}
}

func TestGetLoggedInHandle(t *testing.T) {
	t.Run("returns handle when session exists", func(t *testing.T) {
		tmpDir := setupTestXDG(t)

		// Create session file
		sessDir := filepath.Join(tmpDir, "heartbeads")
		if err := os.MkdirAll(sessDir, 0700); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}

		sess := Session{
			DID:    syntax.DID("did:plc:handletest"),
			Handle: "alice.bsky.social",
		}
		data, _ := json.Marshal(sess)
		if err := os.WriteFile(filepath.Join(sessDir, "auth-session.json"), data, 0600); err != nil {
			t.Fatalf("failed to write session: %v", err)
		}

		handle, err := GetLoggedInHandle()
		if err != nil {
			t.Fatalf("GetLoggedInHandle failed: %v", err)
		}

		if handle != "alice.bsky.social" {
			t.Errorf("handle mismatch: got %s, want alice.bsky.social", handle)
		}
	})

	t.Run("returns ErrNoAuthSession when no session", func(t *testing.T) {
		setupTestXDG(t)

		_, err := GetLoggedInHandle()
		if err != ErrNoAuthSession {
			t.Errorf("expected ErrNoAuthSession, got %v", err)
		}
	})
}

func TestRequireAuth_NotLoggedIn(t *testing.T) {
	setupTestXDG(t)

	_, err := RequireAuth()
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

	sess := Session{
		DID:    syntax.DID("did:plc:guardtest"),
		Handle: "guard.bsky.social",
	}
	data, _ := json.Marshal(sess)
	if err := os.WriteFile(filepath.Join(sessDir, "auth-session.json"), data, 0600); err != nil {
		t.Fatalf("failed to write session: %v", err)
	}

	loaded, err := RequireAuth()
	if err != nil {
		t.Fatalf("RequireAuth should succeed: %v", err)
	}

	if loaded.Handle != "guard.bsky.social" {
		t.Errorf("handle mismatch: got %s, want guard.bsky.social", loaded.Handle)
	}
}
