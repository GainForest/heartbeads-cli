package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/gainforest/heartbeads-cli/internal/executor"
	"github.com/urfave/cli/v3"
)

// DetectBackend inspects the beadsDir and returns the storage backend in use.
// Returns "dolt" | "sqlite" | "jsonl" | "unknown".
func DetectBackend(beadsDir string) string {
	// Check for dolt directory
	doltDir := filepath.Join(beadsDir, "dolt")
	if info, err := os.Stat(doltDir); err == nil && info.IsDir() {
		return "dolt"
	}

	// Check for sqlite db file
	dbFile := filepath.Join(beadsDir, "beads.db")
	if _, err := os.Stat(dbFile); err == nil {
		return "sqlite"
	}

	// Check for jsonl-only (no-db mode)
	jsonlFile := filepath.Join(beadsDir, "issues.jsonl")
	if _, err := os.Stat(jsonlFile); err == nil {
		return "jsonl"
	}

	return "unknown"
}

// NeedsMigration returns true if the backend is sqlite or jsonl (known old backends).
// "unknown" backends do not trigger migration.
func NeedsMigration(beadsDir string) bool {
	backend := DetectBackend(beadsDir)
	return backend == "sqlite" || backend == "jsonl"
}

// RunMigration performs the migration from the current backend to dolt.
// It prints status, runs the appropriate bd commands, and verifies the result.
func RunMigration(ctx context.Context, w io.Writer, beadsDir string, dryRun bool) error {
	// Change to the project root (parent of .beads/) so that bd commands run
	// in the correct directory regardless of the caller's working directory.
	projectRoot := filepath.Dir(beadsDir)
	origDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		return fmt.Errorf("cannot change to project directory %q: %w", projectRoot, err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	backend := DetectBackend(beadsDir)
	_, _ = fmt.Fprintf(w, "Current backend: %s\n", backend)

	switch backend {
	case "dolt":
		_, _ = fmt.Fprintln(w, "Already using dolt backend, no migration needed.")
		return nil

	case "sqlite":
		_, _ = fmt.Fprintln(w, "Migration plan: run `bd migrate --to-dolt` to convert SQLite database to Dolt.")
		if dryRun {
			_, _ = fmt.Fprintln(w, "[dry-run] Would run: bd migrate --to-dolt")
			return nil
		}
		_, _ = fmt.Fprintln(w, "Running: bd migrate --to-dolt")
		stdout, stderr, exitCode, err := executor.RunBd(ctx, []string{"migrate", "--to-dolt"}, "")
		if len(stdout) > 0 {
			_, _ = fmt.Fprint(w, string(stdout))
		}
		if len(stderr) > 0 {
			_, _ = fmt.Fprint(w, string(stderr))
		}
		if err != nil {
			return fmt.Errorf("bd migrate --to-dolt failed: %w", err)
		}
		if exitCode != 0 {
			return fmt.Errorf("bd migrate --to-dolt exited with code %d", exitCode)
		}

	case "jsonl":
		_, _ = fmt.Fprintln(w, "Migration plan: run `bd init` to create dolt db, then `bd import -i .beads/issues.jsonl` to import existing issues.")
		if dryRun {
			_, _ = fmt.Fprintln(w, "[dry-run] Would run: bd init")
			_, _ = fmt.Fprintln(w, "[dry-run] Would run: bd import -i .beads/issues.jsonl")
			return nil
		}
		_, _ = fmt.Fprintln(w, "Running: bd init")
		stdout, stderr, exitCode, err := executor.RunBd(ctx, []string{"init"}, "")
		if len(stdout) > 0 {
			_, _ = fmt.Fprint(w, string(stdout))
		}
		if len(stderr) > 0 {
			_, _ = fmt.Fprint(w, string(stderr))
		}
		if err != nil {
			return fmt.Errorf("bd init failed: %w", err)
		}
		if exitCode != 0 {
			return fmt.Errorf("bd init exited with code %d", exitCode)
		}

		jsonlPath := filepath.Join(beadsDir, "issues.jsonl")
		_, _ = fmt.Fprintf(w, "Running: bd import -i %s\n", jsonlPath)
		stdout, stderr, exitCode, err = executor.RunBd(ctx, []string{"import", "-i", jsonlPath}, "")
		if len(stdout) > 0 {
			_, _ = fmt.Fprint(w, string(stdout))
		}
		if len(stderr) > 0 {
			_, _ = fmt.Fprint(w, string(stderr))
		}
		if err != nil {
			return fmt.Errorf("bd import failed: %w", err)
		}
		if exitCode != 0 {
			return fmt.Errorf("bd import exited with code %d", exitCode)
		}

	default:
		return fmt.Errorf("unknown backend %q — cannot migrate", backend)
	}

	// Verify migration by running bd list --json and checking for valid JSON
	_, _ = fmt.Fprintln(w, "Verifying migration...")
	stdout, stderr, exitCode, err := executor.RunBd(ctx, []string{"list", "--json"}, "")
	if len(stderr) > 0 {
		_, _ = fmt.Fprint(w, string(stderr))
	}
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("verification: bd list --json exited with code %d", exitCode)
	}
	var result interface{}
	if err := json.Unmarshal(stdout, &result); err != nil {
		return fmt.Errorf("verification failed: bd list --json did not return valid JSON: %w", err)
	}
	_, _ = fmt.Fprintln(w, "Migration successful.")
	return nil
}

// findBeadsDir walks up from the current directory to find the .beads directory.
func findBeadsDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot determine working directory: %w", err)
	}
	for {
		candidate := filepath.Join(dir, ".beads")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no .beads directory found in current or parent directories")
}

// resolveBeadsDir returns the beads directory to use.
// If pathFlag is non-empty, it is used as the project directory:
//   - if it already ends with ".beads", it is used directly.
//   - otherwise, the ".beads" subdirectory is used.
//
// If pathFlag is empty, the existing walk-up behaviour is used.
func resolveBeadsDir(pathFlag string) (string, error) {
	if pathFlag == "" {
		return findBeadsDir()
	}
	// Normalise the provided path.
	abs, err := filepath.Abs(pathFlag)
	if err != nil {
		return "", fmt.Errorf("invalid --path %q: %w", pathFlag, err)
	}
	if filepath.Base(abs) == ".beads" {
		return abs, nil
	}
	return filepath.Join(abs, ".beads"), nil
}

// CmdMigrate is the migrate subcommand.
var CmdMigrate = &cli.Command{
	Name:  "migrate",
	Usage: "Migrate from old beads backend to dolt",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "path",
			Usage: "Path to the project directory (or its .beads/ subdirectory); defaults to walking up from cwd",
		},
		&cli.BoolFlag{
			Name:  "check",
			Usage: "Only check if migration is needed; exit 1 if migration is needed",
		},
		&cli.BoolFlag{
			Name:  "dry-run",
			Usage: "Show what would happen without executing",
		},
	},
	Action: runMigrate,
}

func runMigrate(ctx context.Context, cmd *cli.Command) error {
	w := cmd.Root().Writer

	beadsDir, err := resolveBeadsDir(cmd.String("path"))
	if err != nil {
		return err
	}

	if cmd.Bool("check") {
		backend := DetectBackend(beadsDir)
		_, _ = fmt.Fprintf(w, "Backend: %s\n", backend)
		if backend == "dolt" {
			_, _ = fmt.Fprintln(w, "No migration needed.")
			return nil
		}
		_, _ = fmt.Fprintln(w, "Migration needed.")
		// Exit 1 to signal migration is needed
		return cli.Exit("migration needed", 1)
	}

	dryRun := cmd.Bool("dry-run")
	return RunMigration(ctx, w, beadsDir, dryRun)
}
