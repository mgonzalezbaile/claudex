// Package updatedocs provides the usecase for updating index.md documentation
// based on git history changes. It orchestrates git operations, locking,
// tracking, and Claude invocations to keep documentation current.
package updatedocs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"claudex/internal/doc/rangeupdater"
	"claudex/internal/services/commander"
	"claudex/internal/services/doctracking"
	"claudex/internal/services/env"
	"claudex/internal/services/git"
	"claudex/internal/services/lock"

	"github.com/spf13/afero"
)

// UpdateDocsUseCase orchestrates the documentation update workflow
type UpdateDocsUseCase struct {
	fs  afero.Fs
	cmd commander.Commander
	env env.Environment
}

// New creates a new UpdateDocsUseCase instance with the given dependencies
func New(fs afero.Fs, cmd commander.Commander, env env.Environment) *UpdateDocsUseCase {
	return &UpdateDocsUseCase{
		fs:  fs,
		cmd: cmd,
		env: env,
	}
}

// Execute runs the documentation update workflow.
// It computes changed files from git history, maps them to affected index.md
// files, and invokes Claude to regenerate the documentation.
//
// Parameters:
//   - projectDir: The project directory to update documentation for
//
// Returns an error if the update fails.
func (uc *UpdateDocsUseCase) Execute(projectDir string) error {
	// Use sessions directory for tracking state
	sessionPath := filepath.Join(projectDir, "sessions")

	// Create services
	gitSvc := git.New(uc.cmd)
	lockSvc := lock.New(uc.fs)
	trackingSvc := doctracking.New(uc.fs, sessionPath)

	// Configure updater
	config := rangeupdater.RangeUpdaterConfig{
		SessionPath:   sessionPath,
		DefaultBranch: "main",
		SkipPatterns:  []string{"*.md", "docs/**"},
	}

	// Create updater
	updater := rangeupdater.New(
		config,
		gitSvc,
		lockSvc,
		trackingSvc,
		uc.cmd,
		uc.fs,
		uc.env,
	)

	// Run update
	result, err := updater.Run()
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// Display result
	displayResult(result)

	return nil
}

// displayResult prints the update result to stdout
func displayResult(result *rangeupdater.UpdateResult) {
	switch result.Status {
	case "success":
		if len(result.AffectedIndexes) == 0 {
			fmt.Printf("✓ Documentation update completed\n")
			if result.Reason != "" {
				fmt.Printf("  %s\n", result.Reason)
			}
		} else {
			fmt.Printf("✓ Documentation update completed (%s)\n", result.ProcessedRange)
			fmt.Printf("  Updated %d index.md file(s):\n", len(result.AffectedIndexes))
			for _, idx := range result.AffectedIndexes {
				// Make path relative to current directory for cleaner output
				rel, err := filepath.Rel(".", idx)
				if err != nil {
					rel = idx
				}
				fmt.Printf("    - %s\n", rel)
			}
		}

	case "skipped":
		fmt.Printf("○ Documentation update skipped\n")
		if result.Reason != "" {
			fmt.Printf("  Reason: %s\n", result.Reason)
		}
		if result.ProcessedRange != "" {
			fmt.Printf("  Range: %s\n", result.ProcessedRange)
		}

	case "locked":
		fmt.Printf("⊙ Documentation update already in progress\n")
		if result.Reason != "" {
			fmt.Printf("  %s\n", result.Reason)
		}

	case "error":
		log.Printf("✗ Documentation update failed: %s\n", result.Reason)
		os.Exit(1)

	default:
		log.Printf("? Unknown update status: %s\n", result.Status)
		os.Exit(1)
	}
}
