package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"claudex/internal/services/config"
	"claudex/internal/services/session"
)

// setEnvironment sets environment variables needed for Claude session
func (a *App) setEnvironment(si SessionInfo, cfg *config.Config) {
	os.Setenv("CLAUDEX_SESSION", si.Name)
	os.Setenv("CLAUDEX_SESSION_PATH", si.Path)
	if len(a.docPaths) > 0 {
		os.Setenv("CLAUDEX_DOC_PATHS", resolveDocPaths(a.docPaths))
	}

	// Export feature toggles with env var override support
	// Env vars take precedence over config values
	sessionProgress := getEnvBool("CLAUDEX_AUTODOC_SESSION_PROGRESS", cfg.Features.AutodocSessionProgress)
	sessionEnd := getEnvBool("CLAUDEX_AUTODOC_SESSION_END", cfg.Features.AutodocSessionEnd)
	frequency := getEnvInt("CLAUDEX_AUTODOC_FREQUENCY", cfg.Features.AutodocFrequency)

	os.Setenv("CLAUDEX_AUTODOC_SESSION_PROGRESS", strconv.FormatBool(sessionProgress))
	os.Setenv("CLAUDEX_AUTODOC_SESSION_END", strconv.FormatBool(sessionEnd))
	os.Setenv("CLAUDEX_AUTODOC_FREQUENCY", strconv.Itoa(frequency))
}

// getEnvBool returns env var value if set, otherwise returns default
func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		return val == "true"
	}
	return defaultVal
}

// getEnvInt returns env var value if set, otherwise returns default
func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

// launch launches Claude based on the session info and mode
func (a *App) launch(si SessionInfo) error {
	// Update last used timestamp
	if err := session.UpdateLastUsed(a.deps.FS, a.deps.Clock, si.Path); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not update last used timestamp: %v\n", err)
	}

	// Give terminal a moment to settle
	time.Sleep(100 * time.Millisecond)

	// Clear screen and show launching message
	fmt.Print("\033[H\033[2J\033[3J") // Clear screen and scrollback
	fmt.Print("\033[0m")              // Reset all attributes

	switch si.Mode {
	case LaunchModeNew:
		return a.launchNew(si)
	case LaunchModeResume:
		return a.launchResume(si)
	case LaunchModeFork:
		return a.launchFork(si)
	case LaunchModeFresh:
		return a.launchFresh(si)
	case LaunchModeEphemeral:
		return a.launchEphemeral(si)
	default:
		return fmt.Errorf("unknown launch mode: %s", si.Mode)
	}
}

// launchNew launches a new Claude session
func (a *App) launchNew(si SessionInfo) error {
	fmt.Printf("\n✅ Launching new Claude session\n")
	fmt.Printf("📦 Session: %s\n", si.Name)
	fmt.Printf("🔄 Session ID: %s\n\n", si.ClaudeID)

	// Small delay before launching
	time.Sleep(300 * time.Millisecond)

	// Use absolute session path for activation command
	activationPrompt := fmt.Sprintf("/agents:team-lead activate in session %s", si.Path)
	if len(a.docPaths) > 0 {
		activationPrompt += "\n\nIMPORTANT - Required Documentation:\nBefore proceeding, you MUST read these documentation files:"
		for _, docPath := range a.docPaths {
			absPath, _ := filepath.Abs(docPath)
			activationPrompt += fmt.Sprintf("\n- %s", absPath)
		}
	}

	// Launch the Claude session with activation command
	return launchClaude(a.deps, si.ClaudeID, activationPrompt)
}

// launchResume resumes an existing Claude session
func (a *App) launchResume(si SessionInfo) error {
	fmt.Printf("\n✅ Resuming Claude session\n")
	fmt.Printf("📦 Session: %s\n", si.Name)
	fmt.Printf("🔄 Session ID: %s\n\n", si.ClaudeID)

	// Small delay before launching
	time.Sleep(300 * time.Millisecond)

	// For resume, continue existing session
	return resumeClaude(a.deps, si.ClaudeID)
}

// launchFork launches a forked Claude session
func (a *App) launchFork(si SessionInfo) error {
	fmt.Printf("\n✅ Launching forked session\n")
	fmt.Printf("📦 Session: %s\n", si.Name)
	fmt.Printf("🔄 Session ID: %s\n\n", si.ClaudeID)

	// Small delay before launching
	time.Sleep(300 * time.Millisecond)

	// For fork, start a new session with activation command
	activationPrompt := fmt.Sprintf("/agents:team-lead activate in session %s", si.Path)
	if len(a.docPaths) > 0 {
		activationPrompt += "\n\nIMPORTANT - Required Documentation:\nBefore proceeding, you MUST read these documentation files:"
		for _, docPath := range a.docPaths {
			absPath, _ := filepath.Abs(docPath)
			activationPrompt += fmt.Sprintf("\n- %s", absPath)
		}
	}

	return launchClaude(a.deps, si.ClaudeID, activationPrompt)
}

// launchFresh launches a fresh memory session
func (a *App) launchFresh(si SessionInfo) error {
	fmt.Printf("\n🔄 Launching fresh memory session\n")
	fmt.Printf("📦 Session: %s\n", si.Name)
	fmt.Printf("🔄 Session ID: %s\n\n", si.ClaudeID)

	// Small delay before launching
	time.Sleep(300 * time.Millisecond)

	// For fresh, start a new session with activation command
	activationPrompt := fmt.Sprintf("/agents:team-lead activate in session %s", si.Path)
	if len(a.docPaths) > 0 {
		activationPrompt += "\n\nIMPORTANT - Required Documentation:\nBefore proceeding, you MUST read these documentation files:"
		for _, docPath := range a.docPaths {
			absPath, _ := filepath.Abs(docPath)
			activationPrompt += fmt.Sprintf("\n- %s", absPath)
		}
	}

	return launchClaude(a.deps, si.ClaudeID, activationPrompt)
}

// launchEphemeral launches an ephemeral session
func (a *App) launchEphemeral(si SessionInfo) error {
	// Generate new session ID using dependency injection
	claudeSessionID := a.deps.UUID.New()

	// Show launch message
	fmt.Printf("\n✅ Launching ephemeral Claude session\n")
	fmt.Printf("📦 Session: %s\n", si.Name)
	fmt.Printf("🔄 Session ID: %s\n\n", claudeSessionID)
	time.Sleep(500 * time.Millisecond)

	// Launch Claude with NO activation prompt (ephemeral has no session folder)
	return launchClaude(a.deps, claudeSessionID, "")
}

// launchClaude launches a Claude CLI session with the provided session ID and activation prompt
func launchClaude(deps *Dependencies, sessionID string, activationPrompt string) error {
	args := []string{"--session-id", sessionID}
	if activationPrompt != "" {
		args = append(args, activationPrompt)
	}
	return deps.Cmd.Start("claude", os.Stdin, os.Stdout, os.Stderr, args...)
}

// resumeClaude resumes an existing Claude CLI session
func resumeClaude(deps *Dependencies, sessionID string) error {
	return deps.Cmd.Start("claude", os.Stdin, os.Stdout, os.Stderr, "--resume", sessionID)
}
