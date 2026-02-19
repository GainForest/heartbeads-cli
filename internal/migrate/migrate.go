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

// NeedsMigration returns true if the backend is not "dolt".
func NeedsMigration(beadsDir string) bool {
	return DetectBackend(beadsDir) != "dolt"
}

// RunMigration performs the migration from the current backend to dolt.
// It prints status, runs the appropriate bd commands, and verifies the result.
func RunMigration(ctx context.Context, w io.Writer, beadsDir string, dryRun bool) error {
	backend := DetectBackend(beadsDir)
	fmt.Fprintf(w, "Current backend: %s\n", backend)

	switch backend {
	case "dolt":
		fmt.Fprintln(w, "Already using dolt backend, no migration needed.")
		return nil

	case "sqlite":
		fmt.Fprintln(w, "Migration plan: run `bd migrate --to-dolt` to convert SQLite database to Dolt.")
		if dryRun {
			fmt.Fprintln(w, "[dry-run] Would run: bd migrate --to-dolt")
			return nil
		}
		fmt.Fprintln(w, "Running: bd migrate --to-dolt")
		stdout, stderr, exitCode, err := executor.RunBd(ctx, []string{"migrate", "--to-dolt"}, "")
		if len(stdout) > 0 {
			fmt.Fprint(w, string(stdout))
		}
		if len(stderr) > 0 {
			fmt.Fprint(w, string(stderr))
		}
		if err != nil {
			return fmt.Errorf("bd migrate --to-dolt failed: %w", err)
		}
		if exitCode != 0 {
			return fmt.Errorf("bd migrate --to-dolt exited with code %d", exitCode)
		}

	case "jsonl":
		fmt.Fprintln(w, "Migration plan: run `bd init` to create dolt db, then `bd import -i .beads/issues.jsonl` to import existing issues.")
		if dryRun {
			fmt.Fprintln(w, "[dry-run] Would run: bd init")
			fmt.Fprintln(w, "[dry-run] Would run: bd import -i .beads/issues.jsonl")
			return nil
		}
		fmt.Fprintln(w, "Running: bd init")
		stdout, stderr, exitCode, err := executor.RunBd(ctx, []string{"init"}, "")
		if len(stdout) > 0 {
			fmt.Fprint(w, string(stdout))
		}
		if len(stderr) > 0 {
			fmt.Fprint(w, string(stderr))
		}
		if err != nil {
			return fmt.Errorf("bd init failed: %w", err)
		}
		if exitCode != 0 {
			return fmt.Errorf("bd init exited with code %d", exitCode)
		}

		jsonlPath := filepath.Join(beadsDir, "issues.jsonl")
		fmt.Fprintf(w, "Running: bd import -i %s\n", jsonlPath)
		stdout, stderr, exitCode, err = executor.RunBd(ctx, []string{"import", "-i", jsonlPath}, "")
		if len(stdout) > 0 {
			fmt.Fprint(w, string(stdout))
		}
		if len(stderr) > 0 {
			fmt.Fprint(w, string(stderr))
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
	fmt.Fprintln(w, "Verifying migration...")
	stdout, stderr, exitCode, err := executor.RunBd(ctx, []string{"list", "--json"}, "")
	if len(stderr) > 0 {
		fmt.Fprint(w, string(stderr))
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
	fmt.Fprintln(w, "Migration successful.")
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
		fmt.Fprintf(w, "Backend: %s\n", backend)
		if backend == "dolt" {
			fmt.Fprintln(w, "No migration needed.")
			return nil
		}
		fmt.Fprintln(w, "Migration needed.")
		// Exit 1 to signal migration is needed
		return cli.Exit("migration needed", 1)
	}

	dryRun := cmd.Bool("dry-run")
	return RunMigration(ctx, w, beadsDir, dryRun)
}
