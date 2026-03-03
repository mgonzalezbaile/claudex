package sessionend

import (
	"fmt"
	"path/filepath"

	"claudex/internal/doc"
	"claudex/internal/hooks/shared"
	"claudex/internal/services/env"
	"claudex/internal/services/session"

	"github.com/spf13/afero"
)

// Handler implements final documentation update on session end
type Handler struct {
	fs      afero.Fs
	env     env.Environment
	updater doc.DocumentationUpdater
	logger  *shared.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(fs afero.Fs, env env.Environment, updater doc.DocumentationUpdater, logger *shared.Logger) *Handler {
	return &Handler{
		fs:      fs,
		env:     env,
		updater: updater,
		logger:  logger,
	}
}

// Handle triggers final documentation update when session ends.
// Returns nil on success (no JSON output is needed for SessionEnd hooks).
func (h *Handler) Handle(input *shared.SessionEndInput) error {
	// Skip processing for internal Claude invocations (e.g., from doc-update subprocess)
	// Only the main user session should trigger documentation updates
	if h.env.Get("CLAUDE_HOOK_INTERNAL") == "1" {
		return nil
	}

	_ = h.logger.LogInfo(fmt.Sprintf("Session ending: %s", input.Reason))

	// Find session folder
	sessionPath, err := session.FindSessionFolderWithCwd(h.fs, h.env, input.SessionID, input.CWD)
	if err != nil {
		// Log error but allow execution to continue
		_ = h.logger.LogError(fmt.Errorf("failed to find session folder: %w", err))
		return nil
	}

	_ = h.logger.LogInfo("Triggering final documentation update")

	// Read last processed line for incremental updates
	startLine, err := session.ReadLastProcessedLine(h.fs, sessionPath)
	if err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to read last processed line: %w", err))
		startLine = 0 // Start from beginning if we can't read the marker
	}

	// Find project root to build absolute template path
	projectRoot, err := h.findProjectRoot(sessionPath)
	if err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to find project root: %w", err))
		return nil
	}

	// Build absolute path to template
	templatePath := filepath.Join(projectRoot, ".claude", "hooks", "prompts", "session-overview-documenter.md")

	// Trigger documentation update (background, non-blocking)
	// This is the final update, so we always run it
	config := doc.UpdaterConfig{
		SessionPath:    sessionPath,
		TranscriptPath: input.TranscriptPath,
		OutputFile:     "session-overview.md",
		PromptTemplate: templatePath,
		Model:          "haiku",
		StartLine:      startLine + 1, // Start from next line (1-indexed)
	}

	if err := h.updater.RunBackground(config); err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to start background doc update: %w", err))
		// Don't fail - log and continue
	}

	return nil
}

// findProjectRoot walks up from sessionPath to find the project root (where .claude directory exists)
func (h *Handler) findProjectRoot(sessionPath string) (string, error) {
	current := sessionPath
	for {
		claudeDir := filepath.Join(current, ".claude")
		exists, err := afero.DirExists(h.fs, claudeDir)
		if err == nil && exists {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root
			return "", fmt.Errorf("could not find .claude directory in any parent of %s", sessionPath)
		}
		current = parent
	}
}
