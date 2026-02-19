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

// TestResolveBeadsDir_PathFlag verifies that --path resolves to the correct beadsDir.
func TestResolveBeadsDir_PathFlag(t *testing.T) {
	tmp := t.TempDir()
	beadsDir := filepath.Join(tmp, ".beads")
	if err := os.MkdirAll(beadsDir, 0o755); err != nil {
		t.Fatalf("failed to create .beads dir: %v", err)
	}

	// --path points to the project root (parent of .beads)
	got, err := resolveBeadsDir(tmp)
	if err != nil {
		t.Fatalf("resolveBeadsDir(%q) error: %v", tmp, err)
	}
	if got != beadsDir {
		t.Errorf("resolveBeadsDir(%q) = %q, want %q", tmp, got, beadsDir)
	}

	// --path points directly to .beads
	got2, err := resolveBeadsDir(beadsDir)
	if err != nil {
		t.Fatalf("resolveBeadsDir(%q) error: %v", beadsDir, err)
	}
	if got2 != beadsDir {
		t.Errorf("resolveBeadsDir(%q) = %q, want %q", beadsDir, got2, beadsDir)
	}
}

// TestResolveBeadsDir_Empty verifies that an empty path falls back to walk-up behaviour.
// We just check that it returns an error when there is no .beads in the tree
// (the temp dir is not in the cwd hierarchy).
func TestResolveBeadsDir_Empty(t *testing.T) {
	// Change cwd to a temp dir that has no .beads ancestor.
	tmp := t.TempDir()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	_, err = resolveBeadsDir("")
	if err == nil {
		t.Error("resolveBeadsDir(\"\") expected error when no .beads found, got nil")
	}
}
