// Package rangeupdater provides range-based documentation update orchestration.
// It coordinates Git operations, locking, tracking, and Claude invocations
// to update index.md files based on commit range changes.
package rangeupdater

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"claudex/internal/services/commander"
	"claudex/internal/services/doctracking"
	"claudex/internal/services/env"
	"claudex/internal/services/git"
	"claudex/internal/services/lock"

	"github.com/spf13/afero"
)

// UpdateResult represents the outcome of a range update operation
type UpdateResult struct {
	// Status indicates the outcome: "success", "skipped", "locked", or "error"
	Status string

	// Reason provides context for skipped or error statuses
	Reason string

	// AffectedIndexes lists the index.md files that were updated
	AffectedIndexes []string

	// ProcessedRange indicates the commit range that was processed
	ProcessedRange string
}

// RangeUpdater orchestrates range-based documentation updates
type RangeUpdater struct {
	config      RangeUpdaterConfig
	gitSvc      git.GitService
	lockSvc     lock.LockService
	trackingSvc doctracking.TrackingService
	cmdr        commander.Commander
	fs          afero.Fs
	env         env.Environment
}

// New creates a new RangeUpdater instance
func New(
	config RangeUpdaterConfig,
	gitSvc git.GitService,
	lockSvc lock.LockService,
	trackingSvc doctracking.TrackingService,
	cmdr commander.Commander,
	fs afero.Fs,
	env env.Environment,
) *RangeUpdater {
	return &RangeUpdater{
		config:      config,
		gitSvc:      gitSvc,
		lockSvc:     lockSvc,
		trackingSvc: trackingSvc,
		cmdr:        cmdr,
		fs:          fs,
		env:         env,
	}
}

// Run executes the main update flow
func (ru *RangeUpdater) Run() (*UpdateResult, error) {
	// Step 1: Acquire lock (skip if locked)
	lockPath := filepath.Join(ru.config.SessionPath, "doc_update.lock")
	isLocked, err := ru.lockSvc.IsLocked(lockPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check lock status: %w", err)
	}
	if isLocked {
		return &UpdateResult{
			Status: "locked",
			Reason: "another update process is running",
		}, nil
	}

	lock, err := ru.lockSvc.Acquire(lockPath)
	if err != nil {
		return &UpdateResult{
			Status: "locked",
			Reason: "failed to acquire lock",
		}, nil
	}
	defer lock.Release()

	// Step 2: Read tracking to get base SHA
	tracking, err := ru.trackingSvc.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read tracking: %w", err)
	}

	// Get current HEAD SHA
	headSHA, err := ru.gitSvc.GetCurrentSHA()
	if err != nil {
		return nil, fmt.Errorf("failed to get current SHA: %w", err)
	}

	// Initialize tracking if this is the first run
	if tracking.LastProcessedCommit == "" {
		log.Printf("No tracking found, initializing with HEAD: %s", headSHA)
		if err := ru.trackingSvc.Initialize(headSHA); err != nil {
			return nil, fmt.Errorf("failed to initialize tracking: %w", err)
		}
		return &UpdateResult{
			Status: "success",
			Reason: "initialized tracking",
		}, nil
	}

	// Check if HEAD has changed
	if tracking.LastProcessedCommit == headSHA {
		return &UpdateResult{
			Status: "skipped",
			Reason: "no new commits since last update",
		}, nil
	}

	baseSHA := tracking.LastProcessedCommit

	// Step 3: Validate SHA reachability (fallback if unreachable)
	valid, err := ru.gitSvc.ValidateCommit(baseSHA)
	if err != nil || !valid {
		log.Printf("Base commit %s is unreachable, attempting fallback", baseSHA)
		fallbackSHA, err := HandleUnreachableBase(ru.gitSvc, ru.config.DefaultBranch)
		if err != nil {
			return nil, fmt.Errorf("failed to handle unreachable base: %w", err)
		}
		baseSHA = fallbackSHA
		log.Printf("Using fallback base: %s", baseSHA)
	}

	// Step 4: Get changed files for base..HEAD
	changedFiles, err := ru.gitSvc.GetChangedFiles(baseSHA, headSHA)
	if err != nil {
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}

	if len(changedFiles) == 0 {
		return &UpdateResult{
			Status: "skipped",
			Reason: "no files changed",
		}, nil
	}

	// Step 5: Apply skip rules
	shouldSkip, reason := ShouldSkip(changedFiles, "", ru.env)
	if shouldSkip {
		return &UpdateResult{
			Status:         "skipped",
			Reason:         reason,
			ProcessedRange: fmt.Sprintf("%s..%s", shortSHA(baseSHA), shortSHA(headSHA)),
		}, nil
	}

	// Step 6: Map files to affected index.md
	affectedIndexes, err := ResolveAffectedIndexes(ru.fs, changedFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve affected indexes: %w", err)
	}

	if len(affectedIndexes) == 0 {
		// No indexes affected, but still update tracking
		if err := ru.updateTracking(headSHA); err != nil {
			return nil, err
		}
		return &UpdateResult{
			Status:         "success",
			Reason:         "no indexes affected by changes",
			ProcessedRange: fmt.Sprintf("%s..%s", shortSHA(baseSHA), shortSHA(headSHA)),
		}, nil
	}

	// Step 7: Update each index via Claude
	log.Printf("Updating %d index.md files", len(affectedIndexes))
	for _, indexPath := range affectedIndexes {
		if err := ru.updateIndex(indexPath, changedFiles); err != nil {
			log.Printf("Warning: failed to update %s: %v", indexPath, err)
			// Continue with other indexes even if one fails
		}
	}

	// Step 8: Write tracking with new HEAD
	if err := ru.updateTracking(headSHA); err != nil {
		return nil, err
	}

	return &UpdateResult{
		Status:          "success",
		AffectedIndexes: affectedIndexes,
		ProcessedRange:  fmt.Sprintf("%s..%s", shortSHA(baseSHA), shortSHA(headSHA)),
	}, nil
}

// shortSHA returns a short version of the SHA (first 7 chars) or the full SHA if shorter
func shortSHA(sha string) string {
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

// updateIndex updates a single index.md file via Claude
func (ru *RangeUpdater) updateIndex(indexPath string, changedFiles []string) error {
	indexDir := filepath.Dir(indexPath)

	// Get directory listing for context
	listing, err := ru.getDirectoryListing(indexDir)
	if err != nil {
		return fmt.Errorf("failed to get directory listing: %w", err)
	}

	// Format changed files for context
	filesContext := formatChangedFilesContext(changedFiles, indexDir)

	// Invoke Claude to update the index file directly
	return InvokeClaudeForIndex(ru.cmdr, ru.env, indexPath, listing, filesContext)
}

// getDirectoryListing returns a formatted listing of files in the directory
func (ru *RangeUpdater) getDirectoryListing(dir string) (string, error) {
	entries, err := afero.ReadDir(ru.fs, dir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files and directories (except .claude)
		if name[0] == '.' && name != ".claude" {
			continue
		}
		if entry.IsDir() {
			files = append(files, name+"/")
		} else {
			files = append(files, name)
		}
	}

	result := ""
	for i, file := range files {
		if i > 0 {
			result += "\n"
		}
		result += file
	}
	return result, nil
}

// formatChangedFilesContext formats changed files relative to index directory
func formatChangedFilesContext(changedFiles []string, indexDir string) string {
	result := ""
	for i, file := range changedFiles {
		if i > 0 {
			result += "\n"
		}
		// Try to make path relative to index dir if possible
		rel, err := filepath.Rel(indexDir, file)
		if err == nil && rel != "" && rel[0] != '.' {
			result += rel
		} else {
			result += file
		}
	}
	return result
}

// updateTracking updates the tracking state with the new HEAD SHA
func (ru *RangeUpdater) updateTracking(headSHA string) error {
	tracking := doctracking.DocUpdateTracking{
		LastProcessedCommit: headSHA,
		UpdatedAt:           time.Now().Format(time.RFC3339),
		StrategyVersion:     "v1",
	}
	if err := ru.trackingSvc.Write(tracking); err != nil {
		return fmt.Errorf("failed to write tracking: %w", err)
	}
	return nil
}
