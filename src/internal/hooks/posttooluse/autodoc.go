package posttooluse

import (
	"fmt"
	"path/filepath"
	"strings"

	"claudex/internal/doc"
	"claudex/internal/hooks/shared"
	"claudex/internal/services/env"
	"claudex/internal/services/session"

	"github.com/spf13/afero"
)

// AutoDocHandler implements frequency-controlled documentation updates
type AutoDocHandler struct {
	fs        afero.Fs
	env       env.Environment
	updater   doc.DocumentationUpdater
	logger    *shared.Logger
	frequency int
}

// NewAutoDocHandler creates a new AutoDocHandler instance
func NewAutoDocHandler(fs afero.Fs, env env.Environment, updater doc.DocumentationUpdater, logger *shared.Logger, frequency int) *AutoDocHandler {
	return &AutoDocHandler{
		fs:        fs,
		env:       env,
		updater:   updater,
		logger:    logger,
		frequency: frequency,
	}
}

// Handle checks counter and triggers doc update if threshold reached
func (h *AutoDocHandler) Handle(input *shared.PostToolUseInput) (*shared.HookOutput, error) {
	// Skip processing for internal Claude invocations (e.g., from doc-update subprocess)
	// Only the main user session should trigger documentation updates
	if h.env.Get("CLAUDE_HOOK_INTERNAL") == "1" {
		return h.allowOutput(), nil
	}

	// Find session folder
	sessionPath, err := session.FindSessionFolderWithCwd(h.fs, h.env, input.SessionID, input.CWD)
	if err != nil {
		// Log error but allow execution to continue
		_ = h.logger.LogError(fmt.Errorf("failed to find session folder: %w", err))
		return h.allowOutput(), nil
	}

	// Increment counter
	newCount, err := session.IncrementCounter(h.fs, sessionPath)
	if err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to increment counter: %w", err))
		return h.allowOutput(), nil
	}

	_ = h.logger.LogInfo(fmt.Sprintf("Auto-doc counter: %d/%d", newCount, h.frequency))

	// Check if we've reached the threshold
	if newCount < h.frequency {
		return h.allowOutput(), nil
	}

	// Reset counter
	if err := session.ResetCounter(h.fs, sessionPath); err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to reset counter: %w", err))
		// Continue anyway - better to update docs than to fail
	}

	_ = h.logger.LogInfo("Auto-doc threshold reached, triggering documentation update")

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
		return h.allowOutput(), nil
	}

	// Build absolute path to template
	templatePath := filepath.Join(projectRoot, ".claude", "hooks", "prompts", "session-overview-documenter.md")

	// Read existing session context
	sessionContext, err := h.readSessionContext(sessionPath)
	if err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to read session context: %w", err))
		sessionContext = "" // Continue with empty context if reading fails
	}

	// Trigger documentation update (background, non-blocking)
	config := doc.UpdaterConfig{
		SessionPath:    sessionPath,
		TranscriptPath: input.TranscriptPath,
		OutputFile:     "session-overview.md",
		PromptTemplate: templatePath,
		SessionContext: sessionContext,
		Model:          "haiku",
		StartLine:      startLine + 1, // Start from next line (1-indexed)
	}

	if err := h.updater.RunBackground(config); err != nil {
		_ = h.logger.LogError(fmt.Errorf("failed to start background doc update: %w", err))
		// Don't fail - log and continue
	}

	return h.allowOutput(), nil
}

// allowOutput creates a standard "allow" response
func (h *AutoDocHandler) allowOutput() *shared.HookOutput {
	return &shared.HookOutput{
		HookSpecificOutput: shared.HookSpecificOutput{
			HookEventName:      "PostToolUse",
			PermissionDecision: "allow",
		},
	}
}

// findProjectRoot walks up from sessionPath to find the project root (where .claude directory exists)
func (h *AutoDocHandler) findProjectRoot(sessionPath string) (string, error) {
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

// readSessionContext reads existing markdown files from session folder and builds context string
func (h *AutoDocHandler) readSessionContext(sessionPath string) (string, error) {
	files, err := afero.ReadDir(h.fs, sessionPath)
	if err != nil {
		return "", fmt.Errorf("failed to read session directory: %w", err)
	}

	var mdFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			mdFiles = append(mdFiles, file.Name())
		}
	}

	if len(mdFiles) == 0 {
		return "", nil
	}

	var context strings.Builder
	context.WriteString("Existing documentation files in session:\n")
	for _, filename := range mdFiles {
		context.WriteString(fmt.Sprintf("- %s\n", filename))
	}

	return context.String(), nil
}
