// Package session provides session management services for Claudex.
// It handles session metadata, storage operations, and naming utilities.
package session

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"claudex/internal/services/clock"

	"github.com/spf13/afero"
)

// GetSessions retrieves all sessions from the sessions directory
func GetSessions(fs afero.Fs, sessionsDir string) ([]SessionItem, error) {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SessionItem{}, nil
		}
		return nil, err
	}

	var sessions []SessionItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		var desc string
		var lastUsedTime time.Time
		var lastUsedStr string

		if data, err := afero.ReadFile(fs, filepath.Join(sessionsDir, entry.Name(), ".description")); err == nil {
			desc = strings.TrimSpace(string(data))
		}

		// Try to read last_used first, fall back to created
		if data, err := afero.ReadFile(fs, filepath.Join(sessionsDir, entry.Name(), ".last_used")); err == nil {
			lastUsedStr = strings.TrimSpace(string(data))
			if t, err := time.Parse(time.RFC3339, lastUsedStr); err == nil {
				lastUsedTime = t
				lastUsedStr = t.Format("2 Jan 2006 15:04:05")
			}
		} else if data, err := afero.ReadFile(fs, filepath.Join(sessionsDir, entry.Name(), ".created")); err == nil {
			// Fall back to created date if no last_used file
			lastUsedStr = strings.TrimSpace(string(data))
			if t, err := time.Parse(time.RFC3339, lastUsedStr); err == nil {
				lastUsedTime = t
				lastUsedStr = t.Format("2 Jan 2006 15:04:05")
			}
		}

		sessions = append(sessions, SessionItem{
			Title:       entry.Name(),
			Description: desc,
			Date:        lastUsedStr,
			Created:     lastUsedTime,
			ItemType:    "session",
		})
	}

	// Sort by last used date in descending order (most recently used first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Created.After(sessions[j].Created)
	})

	return sessions, nil
}

// UpdateLastUsedWithDeps updates the last used timestamp using injected dependencies
func UpdateLastUsedWithDeps(fs afero.Fs, clk clock.Clock, sessionPath string) error {
	if sessionPath == "" {
		// Ephemeral session, no directory to update
		return nil
	}

	lastUsed := clk.Now().UTC().Format(time.RFC3339)
	return afero.WriteFile(fs, filepath.Join(sessionPath, ".last_used"), []byte(lastUsed), 0644)
}

// UpdateLastUsed is a wrapper that uses default dependencies
// Note: This should not be used directly in production code; use UpdateLastUsedWithDeps instead
func UpdateLastUsed(fs afero.Fs, clk clock.Clock, sessionPath string) error {
	return UpdateLastUsedWithDeps(fs, clk, sessionPath)
}
