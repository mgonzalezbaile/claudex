package doc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"claudex/internal/services/commander"
	"claudex/internal/services/env"

	"github.com/spf13/afero"
)

// UpdaterConfig holds configuration for documentation updates
type UpdaterConfig struct {
	SessionPath    string // Absolute path to session folder
	TranscriptPath string // Path to transcript JSONL file
	OutputFile     string // Target file (e.g., session-overview.md)
	PromptTemplate string // Path to prompt template file
	SessionContext string // Additional session context to include
	Model          string // Claude model to use (e.g., "haiku")
	StartLine      int    // Line number to start reading transcript (1-indexed)
}

// Updater handles background Claude invocations for doc updates
type Updater struct {
	fs  afero.Fs
	cmd commander.Commander
	env env.Environment
}

// NewUpdater creates a new Updater instance
func NewUpdater(fs afero.Fs, cmd commander.Commander, env env.Environment) *Updater {
	return &Updater{
		fs:  fs,
		cmd: cmd,
		env: env,
	}
}

// RunBackground starts doc update as a detached subprocess
// Returns immediately, update happens asynchronously in a separate process
// that survives the parent process exit
func (u *Updater) RunBackground(config UpdaterConfig) error {
	// Validate configuration
	if err := u.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Prepare input JSON for the subprocess
	input := docUpdateInput{
		SessionPath:    config.SessionPath,
		TranscriptPath: config.TranscriptPath,
		OutputFile:     config.OutputFile,
		PromptTemplate: config.PromptTemplate,
		SessionContext: config.SessionContext,
		Model:          config.Model,
		StartLine:      config.StartLine,
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	// Find the claudex-hooks binary
	hooksBin := u.env.Get("CLAUDEX_HOOKS_BIN")
	if hooksBin == "" {
		hooksBin = "claudex-hooks"
	}

	// Create command for the detached subprocess
	cmd := exec.Command(hooksBin, "doc-update")

	// Set up stdin pipe to pass the config
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	// Inherit environment (recursion guard will be set by invokeClaude, not here)
	cmd.Env = os.Environ()

	// Detach the process so it survives parent exit
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
	}

	// Discard stdout/stderr (subprocess logs to file via logger)
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Start the subprocess (non-blocking)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start doc-update subprocess: %w", err)
	}

	// Write input synchronously and close stdin
	// This must complete before we return, otherwise the subprocess won't receive input
	if _, err := stdin.Write(inputJSON); err != nil {
		return fmt.Errorf("failed to write input to subprocess: %w", err)
	}
	stdin.Close()

	// Don't wait for the subprocess - let it run independently
	// The process will be orphaned and adopted by init/launchd
	go func() {
		_ = cmd.Wait() // Reap the zombie when done
	}()

	return nil
}

// docUpdateInput matches the shared.DocUpdateInput structure
// Defined here to avoid circular imports
type docUpdateInput struct {
	SessionPath    string `json:"session_path"`
	TranscriptPath string `json:"transcript_path"`
	OutputFile     string `json:"output_file"`
	PromptTemplate string `json:"prompt_template"`
	SessionContext string `json:"session_context"`
	Model          string `json:"model"`
	StartLine      int    `json:"start_line"`
}

// Run executes doc update synchronously (for testing)
// This is the main implementation that does the actual work
func (u *Updater) Run(config UpdaterConfig) error {
	// Check recursion guard before doing any work
	if u.env.Get("CLAUDE_HOOK_INTERNAL") == "1" {
		return fmt.Errorf("recursion guard: CLAUDE_HOOK_INTERNAL is set")
	}

	// Parse transcript from startLine
	entries, lastLine, err := ParseTranscript(u.fs, config.TranscriptPath, config.StartLine)
	if err != nil {
		return fmt.Errorf("failed to parse transcript: %w", err)
	}

	// Nothing to process
	if len(entries) == 0 {
		return nil
	}

	// Format transcript for prompt
	transcriptContent := FormatTranscriptForPrompt(entries)

	// Load prompt template
	template, err := LoadPromptTemplate(u.fs, config.PromptTemplate)
	if err != nil {
		return fmt.Errorf("failed to load prompt template: %w", err)
	}

	// Build final prompt
	prompt := BuildDocumentationPrompt(template, transcriptContent, config.SessionContext, config.SessionPath)

	// Invoke Claude with recursion guard
	if err := u.invokeClaude(prompt, config.Model); err != nil {
		return fmt.Errorf("failed to invoke Claude: %w", err)
	}

	// Update last processed line marker
	lastLineFile := fmt.Sprintf("%s/.last-processed-line-overview", config.SessionPath)
	if err := afero.WriteFile(u.fs, lastLineFile, []byte(fmt.Sprintf("%d", lastLine)), 0644); err != nil {
		return fmt.Errorf("failed to update last processed line: %w", err)
	}

	return nil
}

// validateConfig checks that all required configuration fields are present
func (u *Updater) validateConfig(config UpdaterConfig) error {
	if config.SessionPath == "" {
		return fmt.Errorf("SessionPath is required")
	}
	if config.TranscriptPath == "" {
		return fmt.Errorf("TranscriptPath is required")
	}
	if config.PromptTemplate == "" {
		return fmt.Errorf("PromptTemplate is required")
	}
	if config.Model == "" {
		return fmt.Errorf("Model is required")
	}
	if config.StartLine < 1 {
		return fmt.Errorf("StartLine must be >= 1")
	}
	return nil
}

// invokeClaude calls the claude CLI with the given prompt
// Sets CLAUDE_HOOK_INTERNAL=1 to prevent recursion
func (u *Updater) invokeClaude(prompt string, model string) error {
	// Set recursion guard in environment
	originalValue := u.env.Get("CLAUDE_HOOK_INTERNAL")
	u.env.Set("CLAUDE_HOOK_INTERNAL", "1")
	defer func() {
		if originalValue == "" {
			// Restore by setting to empty (best effort, depends on env implementation)
			u.env.Set("CLAUDE_HOOK_INTERNAL", "")
		} else {
			u.env.Set("CLAUDE_HOOK_INTERNAL", originalValue)
		}
	}()

	// Create command with recursion guard via actual exec.Command
	// We need to use exec.Command directly here to set custom environment
	// Note: We don't use --output-format stream-json as it requires --verbose with -p
	cmd := exec.Command("claude", "-p", prompt, "--model", model)

	// Set environment with recursion guard
	cmdEnv := os.Environ()
	cmdEnv = append(cmdEnv, "CLAUDE_HOOK_INTERNAL=1")
	cmd.Env = cmdEnv

	// Capture output for potential logging
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("claude command failed: %w (stderr: %s)", err, stderr.String())
	}

	return nil
}
