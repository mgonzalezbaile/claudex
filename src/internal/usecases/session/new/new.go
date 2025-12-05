// Package new provides the use case for creating new sessions.
// It orchestrates session name generation, directory creation, and metadata initialization.
package new

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claudex/internal/services/clock"
	"claudex/internal/services/commander"
	"claudex/internal/services/session"
	"claudex/internal/services/uuid"

	"github.com/spf13/afero"
)

// UseCase handles the creation of new sessions
type UseCase struct {
	fs          afero.Fs
	cmd         commander.Commander
	uuidGen     uuid.UUIDGenerator
	clock       clock.Clock
	sessionsDir string
}

// New creates a new session creation use case
func New(fs afero.Fs, cmd commander.Commander, uuidGen uuid.UUIDGenerator, clk clock.Clock, sessionsDir string) *UseCase {
	return &UseCase{
		fs:          fs,
		cmd:         cmd,
		uuidGen:     uuidGen,
		clock:       clk,
		sessionsDir: sessionsDir,
	}
}

// Execute creates a new session by:
// 1. Generating a UUID for the session
// 2. Generating session name from description (via Claude CLI or manual slug)
// 3. Creating session directory with metadata files
// 4. Returning session info for launching Claude
func (uc *UseCase) Execute(description string) (sessionName, sessionPath, claudeSessionID string, err error) {
	description = strings.TrimSpace(description)
	if description == "" {
		return "", "", "", fmt.Errorf("description cannot be empty")
	}

	// Generate UUID for the session upfront
	claudeSessionID = uc.uuidGen.New()

	// Generate session name using Claude CLI or fallback to manual slug
	baseSessionName, err := session.GenerateNameWithCmd(uc.cmd, description)
	if err != nil {
		baseSessionName = session.CreateManualSlug(description)
	}

	// Create final session name with Claude session ID
	sessionName = fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)

	// Ensure unique (in case of collision)
	originalName := sessionName
	counter := 1
	sessionPath = filepath.Join(uc.sessionsDir, sessionName)
	for {
		if _, err := uc.fs.Stat(sessionPath); os.IsNotExist(err) {
			break
		}
		sessionName = fmt.Sprintf("%s-%d", originalName, counter)
		sessionPath = filepath.Join(uc.sessionsDir, sessionName)
		counter++
	}

	// Create session directory
	if err := uc.fs.MkdirAll(sessionPath, 0755); err != nil {
		return "", "", "", err
	}

	// Write description file
	if err := afero.WriteFile(uc.fs, filepath.Join(sessionPath, ".description"), []byte(description), 0644); err != nil {
		return "", "", "", err
	}

	// Write created timestamp
	created := uc.clock.Now().UTC().Format(time.RFC3339)
	if err := afero.WriteFile(uc.fs, filepath.Join(sessionPath, ".created"), []byte(created), 0644); err != nil {
		return "", "", "", err
	}

	return sessionName, sessionPath, claudeSessionID, nil
}
