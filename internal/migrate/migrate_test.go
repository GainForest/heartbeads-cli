package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

// makeBeadsDir creates a temporary .beads directory for testing.
func makeBeadsDir(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	beadsDir := filepath.Join(tmp, ".beads")
	if err := os.MkdirAll(beadsDir, 0o755); err != nil {
		t.Fatalf("failed to create .beads dir: %v", err)
	}
	return beadsDir
}

func TestDetectBackend_Dolt(t *testing.T) {
	beadsDir := makeBeadsDir(t)
	doltDir := filepath.Join(beadsDir, "dolt")
	if err := os.MkdirAll(doltDir, 0o755); err != nil {
		t.Fatalf("failed to create dolt dir: %v", err)
	}

	got := DetectBackend(beadsDir)
	if got != "dolt" {
		t.Errorf("DetectBackend() = %q, want %q", got, "dolt")
	}
}

func TestDetectBackend_SQLite(t *testing.T) {
	beadsDir := makeBeadsDir(t)
	dbFile := filepath.Join(beadsDir, "beads.db")
	if err := os.WriteFile(dbFile, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create beads.db: %v", err)
	}

	got := DetectBackend(beadsDir)
	if got != "sqlite" {
		t.Errorf("DetectBackend() = %q, want %q", got, "sqlite")
	}
}

func TestDetectBackend_JSONL(t *testing.T) {
	beadsDir := makeBeadsDir(t)
	jsonlFile := filepath.Join(beadsDir, "issues.jsonl")
	if err := os.WriteFile(jsonlFile, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create issues.jsonl: %v", err)
	}

	got := DetectBackend(beadsDir)
	if got != "jsonl" {
		t.Errorf("DetectBackend() = %q, want %q", got, "jsonl")
	}
}

func TestDetectBackend_Unknown(t *testing.T) {
	beadsDir := makeBeadsDir(t)

	got := DetectBackend(beadsDir)
	if got != "unknown" {
		t.Errorf("DetectBackend() = %q, want %q", got, "unknown")
	}
}

// TestDetectBackend_DoltTakesPrecedence verifies that dolt is preferred over sqlite.
func TestDetectBackend_DoltTakesPrecedence(t *testing.T) {
	beadsDir := makeBeadsDir(t)

	// Create both dolt dir and beads.db
	doltDir := filepath.Join(beadsDir, "dolt")
	if err := os.MkdirAll(doltDir, 0o755); err != nil {
		t.Fatalf("failed to create dolt dir: %v", err)
	}
	dbFile := filepath.Join(beadsDir, "beads.db")
	if err := os.WriteFile(dbFile, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create beads.db: %v", err)
	}

	got := DetectBackend(beadsDir)
	if got != "dolt" {
		t.Errorf("DetectBackend() = %q, want %q (dolt should take precedence)", got, "dolt")
	}
}

func TestNeedsMigration_Dolt(t *testing.T) {
	beadsDir := makeBeadsDir(t)
	doltDir := filepath.Join(beadsDir, "dolt")
	if err := os.MkdirAll(doltDir, 0o755); err != nil {
		t.Fatalf("failed to create dolt dir: %v", err)
	}

	if NeedsMigration(beadsDir) {
		t.Error("NeedsMigration() = true for dolt backend, want false")
	}
}

func TestNeedsMigration_SQLite(t *testing.T) {
	beadsDir := makeBeadsDir(t)
	dbFile := filepath.Join(beadsDir, "beads.db")
	if err := os.WriteFile(dbFile, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create beads.db: %v", err)
	}

	if !NeedsMigration(beadsDir) {
		t.Error("NeedsMigration() = false for sqlite backend, want true")
	}
}

func TestNeedsMigration_JSONL(t *testing.T) {
	beadsDir := makeBeadsDir(t)
	jsonlFile := filepath.Join(beadsDir, "issues.jsonl")
	if err := os.WriteFile(jsonlFile, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to create issues.jsonl: %v", err)
	}

	if !NeedsMigration(beadsDir) {
		t.Error("NeedsMigration() = false for jsonl backend, want true")
	}
}
