package posttooluse

import (
	"path/filepath"
	"strings"
	"testing"

	"claudex/internal/doc"
	"claudex/internal/hooks/shared"
	"claudex/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockUpdater captures the config passed to RunBackground for testing
type MockUpdater struct {
	capturedConfig *doc.UpdaterConfig
	runError       error
}

func (m *MockUpdater) RunBackground(config doc.UpdaterConfig) error {
	m.capturedConfig = &config
	return m.runError
}

func (m *MockUpdater) Run(config doc.UpdaterConfig) error {
	m.capturedConfig = &config
	return m.runError
}

// TestAutoDocHandler_UsesAbsoluteTemplatePath tests that the handler passes an absolute path for PromptTemplate
func TestAutoDocHandler_UsesAbsoluteTemplatePath(t *testing.T) {
	h := testutil.NewTestHarness()
	mockUpdater := &MockUpdater{}
	logger := shared.NewLogger(h.FS, h.Env, "autodoc-test")

	// Create a session folder with counter one below threshold
	// Handler increments first, so 4 -> 5 will reach threshold
	sessionPath := "/Users/test/.claudex/sessions/test-session-abc"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".doc-update-counter":           "4", // Will increment to 5 and trigger (frequency=5)
		".last-processed-line-overview": "0",
	})

	// Create .claude directory structure for findProjectRoot to work
	projectRoot := "/Users/test"
	claudeDir := filepath.Join(projectRoot, ".claude", "hooks", "prompts")
	h.CreateDir(claudeDir)
	h.WriteFile(filepath.Join(claudeDir, "session-overview-documenter.md"), "# Template\n$RELEVANT_CONTENT\n$DOC_CONTEXT\n$SESSION_FOLDER")

	// Create a transcript file
	transcriptPath := "/tmp/transcript.jsonl"
	h.WriteFile(transcriptPath, `{"type":"message","message":{"role":"assistant","content":"test"}}`)

	// Setup environment - set CLAUDEX_SESSION_PATH so session can be found
	h.Env.Set("CLAUDEX_SESSION_PATH", sessionPath)
	h.Env.Set("CLAUDE_CONFIG_DIR", "/Users/test/.claude")
	h.Env.Set("HOME", "/Users/test")

	// Create handler with frequency=5
	handler := NewAutoDocHandler(h.FS, h.Env, mockUpdater, logger, 5)

	// Create input
	input := &shared.PostToolUseInput{
		HookInput: shared.HookInput{
			SessionID:      "test-session-abc",
			TranscriptPath: transcriptPath,
			CWD:            sessionPath,
		},
		ToolName: "Write",
		Status:   "success",
	}

	// Execute handler
	_, err := handler.Handle(input)
	require.NoError(t, err)

	// Verify updater was called
	require.NotNil(t, mockUpdater.capturedConfig, "Expected updater to be called")

	// BUG: This assertion will FAIL because PromptTemplate is set to "session-overview-documenter.md" (relative path)
	// but LoadPromptTemplate expects an absolute path
	assert.True(t, filepath.IsAbs(mockUpdater.capturedConfig.PromptTemplate),
		"Expected PromptTemplate to be an absolute path, got: %s", mockUpdater.capturedConfig.PromptTemplate)
}

// TestAutoDocHandler_PopulatesSessionContext tests that SessionContext is populated with existing docs
func TestAutoDocHandler_PopulatesSessionContext(t *testing.T) {
	h := testutil.NewTestHarness()
	mockUpdater := &MockUpdater{}
	logger := shared.NewLogger(h.FS, h.Env, "autodoc-test")

	// Create a session folder with counter one below threshold and existing markdown files
	// Handler increments first, so 4 -> 5 will reach threshold
	sessionPath := "/Users/test/.claudex/sessions/test-session-xyz"
	h.CreateSessionWithFiles(sessionPath, map[string]string{
		".doc-update-counter":           "4", // Will increment to 5 and trigger (frequency=5)
		".last-processed-line-overview": "0",
		"session-overview.md":           "# Session Overview\nCurrent progress...",
		"research-notes.md":             "# Research\nSome findings...",
		"implementation-plan.md":        "# Plan\nSteps to take...",
	})

	// Create .claude directory structure for findProjectRoot to work
	projectRoot := "/Users/test"
	claudeDir := filepath.Join(projectRoot, ".claude", "hooks", "prompts")
	h.CreateDir(claudeDir)
	h.WriteFile(filepath.Join(claudeDir, "session-overview-documenter.md"), "# Template\n$RELEVANT_CONTENT\n$DOC_CONTEXT\n$SESSION_FOLDER")

	// Create a transcript file
	transcriptPath := "/tmp/transcript2.jsonl"
	h.WriteFile(transcriptPath, `{"type":"message","message":{"role":"assistant","content":"test"}}`)

	// Setup environment - set CLAUDEX_SESSION_PATH so session can be found
	h.Env.Set("CLAUDEX_SESSION_PATH", sessionPath)
	h.Env.Set("CLAUDE_CONFIG_DIR", "/Users/test/.claude")
	h.Env.Set("HOME", "/Users/test")

	// Create handler with frequency=5
	handler := NewAutoDocHandler(h.FS, h.Env, mockUpdater, logger, 5)

	// Create input
	input := &shared.PostToolUseInput{
		HookInput: shared.HookInput{
			SessionID:      "test-session-xyz",
			TranscriptPath: transcriptPath,
			CWD:            sessionPath,
		},
		ToolName: "Edit",
		Status:   "success",
	}

	// Execute handler
	_, err := handler.Handle(input)
	require.NoError(t, err)

	// Verify updater was called
	require.NotNil(t, mockUpdater.capturedConfig, "Expected updater to be called")

	// BUG: This assertion will FAIL because SessionContext is never populated
	// The field remains empty even though there are .md files in the session folder
	assert.NotEmpty(t, mockUpdater.capturedConfig.SessionContext,
		"Expected SessionContext to be populated with existing documentation")

	// Additional verification: SessionContext should contain references to the existing files
	if mockUpdater.capturedConfig.SessionContext != "" {
		context := mockUpdater.capturedConfig.SessionContext
		assert.True(t,
			strings.Contains(context, "session-overview.md") ||
				strings.Contains(context, "research-notes.md") ||
				strings.Contains(context, "implementation-plan.md"),
			"Expected SessionContext to mention existing markdown files")
	}
}
