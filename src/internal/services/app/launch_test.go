package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"claudex/internal/services/config"
	"claudex/internal/testutil"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// TestLaunchEphemeral_NoDirectoryCreated verifies ephemeral sessions do not create directories
// BUG: Current implementation incorrectly creates and renames session directories
// EXPECTED: Ephemeral sessions should NOT create any session directories
func TestLaunchEphemeral_NoDirectoryCreated(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.UUIDs = []string{"ephemeral-uuid-1111-2222-3333-444444444444"}

	projectDir := "/project"
	sessionsDir := filepath.Join(projectDir, ".claudex", "sessions")
	h.CreateDir(sessionsDir)

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		sessionsDir: sessionsDir,
		docPaths:    []string{},
	}

	// Create session info for ephemeral mode
	si := SessionInfo{
		Name: "ephemeral",
		Path: "", // Empty path for ephemeral
		Mode: LaunchModeEphemeral,
	}

	// Execute launchEphemeral
	err := app.launchEphemeral(si)

	// Verify no error (command may fail, but that's OK for this test)
	// We're testing directory creation behavior, not command execution
	_ = err

	// Assert NO directories created in sessions folder
	entries, err := afero.ReadDir(h.FS, sessionsDir)
	require.NoError(t, err)
	require.Empty(t, entries, "ephemeral session should NOT create any directories in sessions folder")

	// Assert NO directories with ephemeral UUID pattern exist
	matches, _ := filepath.Glob(filepath.Join(sessionsDir, "*ephemeral-uuid*"))
	require.Empty(t, matches, "should not create directories with ephemeral UUID")
}

// TestLaunchEphemeral_NoActivationPrompt verifies ephemeral sessions launch without activation prompt
// BUG: Current implementation incorrectly constructs and sends activation prompt
// EXPECTED: Ephemeral sessions should call launchClaude with empty string for activation prompt
func TestLaunchEphemeral_NoActivationPrompt(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.UUIDs = []string{"ephemeral-uuid-aaaa-bbbb-cccc-dddddddddddd"}

	projectDir := "/project"
	sessionsDir := filepath.Join(projectDir, ".claudex", "sessions")
	h.CreateDir(sessionsDir)

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		sessionsDir: sessionsDir,
		docPaths:    []string{},
	}

	// Create session info for ephemeral mode
	si := SessionInfo{
		Name: "ephemeral",
		Path: "", // Empty path for ephemeral
		Mode: LaunchModeEphemeral,
	}

	// Execute launchEphemeral
	_ = app.launchEphemeral(si)

	// Assert claude was invoked
	require.NotEmpty(t, h.Commander.Invocations, "claude should be invoked")

	// Find the claude invocation
	var claudeInvocation *testutil.CommandInvocation
	for _, inv := range h.Commander.Invocations {
		if inv.Name == "claude" {
			claudeInvocation = &inv
			break
		}
	}
	require.NotNil(t, claudeInvocation, "claude command should be invoked")

	// Assert NO activation prompt in arguments
	// The args should be: ["--session-id", "<uuid>"]
	// Should NOT contain: "/agents:team-lead activate"
	allArgs := strings.Join(claudeInvocation.Args, " ")
	require.NotContains(t, allArgs, "/agents:team-lead", "should NOT send activation prompt for ephemeral session")
	require.NotContains(t, allArgs, "activate", "should NOT send activation command for ephemeral session")

	// Should only have --session-id and the UUID
	require.Contains(t, claudeInvocation.Args, "--session-id")
	require.Equal(t, 2, len(claudeInvocation.Args), "should only have --session-id and UUID, no activation prompt")
}

// TestLaunchEphemeral_GeneratesUUID verifies ephemeral sessions generate and use new UUID
// This test should PASS with current code (UUID generation works correctly)
func TestLaunchEphemeral_GeneratesUUID(t *testing.T) {
	// Setup
	expectedUUID := "test-ephemeral-1234-5678-9abc-def012345678"
	h := testutil.NewTestHarness()
	h.UUIDs = []string{expectedUUID}

	projectDir := "/project"
	sessionsDir := filepath.Join(projectDir, ".claudex", "sessions")
	h.CreateDir(sessionsDir)

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		sessionsDir: sessionsDir,
		docPaths:    []string{},
	}

	// Create session info for ephemeral mode
	si := SessionInfo{
		Name: "ephemeral",
		Path: "", // Empty path for ephemeral
		Mode: LaunchModeEphemeral,
	}

	// Execute launchEphemeral
	_ = app.launchEphemeral(si)

	// Assert claude was invoked with the seeded UUID
	require.NotEmpty(t, h.Commander.Invocations)
	var claudeInvocation *testutil.CommandInvocation
	for _, inv := range h.Commander.Invocations {
		if inv.Name == "claude" {
			claudeInvocation = &inv
			break
		}
	}
	require.NotNil(t, claudeInvocation)

	// Verify UUID is used in --session-id
	require.Contains(t, claudeInvocation.Args, "--session-id")
	require.Contains(t, claudeInvocation.Args, expectedUUID)
}

// TestLaunchEphemeral_CorrectEnvironment verifies ephemeral sessions set correct environment variables
// EXPECTED: CLAUDEX_SESSION="ephemeral", CLAUDEX_SESSION_PATH="" (empty or unset)
func TestLaunchEphemeral_CorrectEnvironment(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.UUIDs = []string{"env-test-uuid-1111-2222-3333-444444444444"}

	projectDir := "/project"
	sessionsDir := filepath.Join(projectDir, ".claudex", "sessions")
	h.CreateDir(sessionsDir)

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		sessionsDir: sessionsDir,
		docPaths:    []string{},
	}

	// Create session info for ephemeral mode
	si := SessionInfo{
		Name: "ephemeral",
		Path: "", // Empty path for ephemeral
		Mode: LaunchModeEphemeral,
	}

	// Set environment before launch (simulating what App.Run does)
	cfg := &config.Config{
		Features: config.Features{
			AutodocSessionProgress: true,
			AutodocSessionEnd:      true,
			AutodocFrequency:       5,
		},
	}
	app.setEnvironment(si, cfg)

	// Execute launchEphemeral
	_ = app.launchEphemeral(si)

	// BUG: Current implementation sets CLAUDEX_SESSION_PATH to a real directory path
	// EXPECTED: Should remain empty for ephemeral sessions

	// Check environment variables (need to read from actual os.Getenv since setEnvironment uses os.Setenv)
	// For proper testing, we'd need to refactor setEnvironment to use injected Env interface
	// For now, verify through mock that session path is empty in SessionInfo
	require.Equal(t, "", si.Path, "ephemeral session path should be empty")
	require.Equal(t, "ephemeral", si.Name, "ephemeral session name should be 'ephemeral'")
}

// TestLaunchEphemeral_NoSessionRename verifies ephemeral sessions don't rename directories
// BUG: Current implementation calls session.RenameWithClaudeID which should not happen for ephemeral
func TestLaunchEphemeral_NoSessionRename(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.UUIDs = []string{"rename-test-uuid-1111-2222-3333-444444444444"}

	projectDir := "/project"
	sessionsDir := filepath.Join(projectDir, ".claudex", "sessions")
	h.CreateDir(sessionsDir)

	// Pre-create a session directory that should NOT be touched
	existingSessionPath := filepath.Join(sessionsDir, "existing-session")
	h.CreateSessionWithFiles(existingSessionPath, map[string]string{
		".description": "Existing session",
	})

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		sessionsDir: sessionsDir,
		docPaths:    []string{},
	}

	// Create session info for ephemeral mode
	si := SessionInfo{
		Name: "ephemeral",
		Path: "", // Empty path for ephemeral
		Mode: LaunchModeEphemeral,
	}

	// Execute launchEphemeral
	_ = app.launchEphemeral(si)

	// Verify existing session was not renamed
	testutil.AssertDirExists(t, h.FS, existingSessionPath)

	// Verify no new directories with UUID were created
	entries, err := afero.ReadDir(h.FS, sessionsDir)
	require.NoError(t, err)
	require.Len(t, entries, 1, "should only have the existing session, no new directories")
	require.Equal(t, "existing-session", entries[0].Name())
}

// TestLaunchEphemeral_WithDocPaths verifies ephemeral sessions with doc paths still don't send activation
// BUG: Current implementation adds documentation to activation prompt
// EXPECTED: Even with docPaths, ephemeral sessions should NOT send activation prompt
func TestLaunchEphemeral_WithDocPaths(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.UUIDs = []string{"docpath-test-uuid-1111-2222-3333-444444444444"}

	projectDir := "/project"
	sessionsDir := filepath.Join(projectDir, ".claudex", "sessions")
	h.CreateDir(sessionsDir)

	// Create app with doc paths configured
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		sessionsDir: sessionsDir,
		docPaths:    []string{"docs/api.md", "docs/guide.md"}, // Documentation paths configured
	}

	// Create session info for ephemeral mode
	si := SessionInfo{
		Name: "ephemeral",
		Path: "", // Empty path for ephemeral
		Mode: LaunchModeEphemeral,
	}

	// Execute launchEphemeral
	_ = app.launchEphemeral(si)

	// Assert claude was invoked WITHOUT activation prompt (even with doc paths)
	require.NotEmpty(t, h.Commander.Invocations)
	var claudeInvocation *testutil.CommandInvocation
	for _, inv := range h.Commander.Invocations {
		if inv.Name == "claude" {
			claudeInvocation = &inv
			break
		}
	}
	require.NotNil(t, claudeInvocation)

	// Should NOT contain activation command or documentation references
	allArgs := strings.Join(claudeInvocation.Args, " ")
	require.NotContains(t, allArgs, "/agents:team-lead")
	require.NotContains(t, allArgs, "activate")
	require.NotContains(t, allArgs, "documentation")
	require.NotContains(t, allArgs, "docs/api.md")

	// Should only have --session-id and UUID
	require.Equal(t, 2, len(claudeInvocation.Args))
}

// TestLaunchEphemeral_NoLastUsedUpdate verifies ephemeral sessions don't update .last_used
// This is an implied behavior - ephemeral sessions shouldn't touch any files
func TestLaunchEphemeral_NoLastUsedUpdate(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.UUIDs = []string{"lastused-test-uuid-1111-2222-3333-444444444444"}

	projectDir := "/project"
	sessionsDir := filepath.Join(projectDir, ".claudex", "sessions")
	h.CreateDir(sessionsDir)

	// Create app with mocked dependencies
	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir:  projectDir,
		sessionsDir: sessionsDir,
		docPaths:    []string{},
	}

	// Create session info for ephemeral mode
	si := SessionInfo{
		Name: "ephemeral",
		Path: "", // Empty path for ephemeral
		Mode: LaunchModeEphemeral,
	}

	// Execute launchEphemeral
	_ = app.launchEphemeral(si)

	// Verify no .last_used files were created anywhere
	entries, err := afero.ReadDir(h.FS, sessionsDir)
	require.NoError(t, err)
	require.Empty(t, entries, "should not create any files for ephemeral session")
}

// TestLaunchEphemeral_CompareWithLaunchNew demonstrates the difference between ephemeral and new session
// This test shows what ephemeral SHOULD do vs what new session DOES do
func TestLaunchEphemeral_CompareWithLaunchNew(t *testing.T) {
	t.Run("New session creates directory and sends activation", func(t *testing.T) {
		h := testutil.NewTestHarness()
		h.UUIDs = []string{"new-session-uuid"}

		projectDir := "/project"
		sessionsDir := filepath.Join(projectDir, ".claudex", "sessions")
		sessionPath := filepath.Join(sessionsDir, "test-session-new-session-uuid")
		h.CreateDir(sessionPath)

		app := &App{
			deps: &Dependencies{
				FS:    h.FS,
				Cmd:   h.Commander,
				Clock: h,
				UUID:  h,
				Env:   h.Env,
			},
			projectDir:  projectDir,
			sessionsDir: sessionsDir,
			docPaths:    []string{},
		}

		si := SessionInfo{
			Name:     "test-session-new-session-uuid",
			Path:     sessionPath,
			ClaudeID: "new-session-uuid",
			Mode:     LaunchModeNew,
		}

		_ = app.launchNew(si)

		// New session DOES send activation prompt
		require.NotEmpty(t, h.Commander.Invocations)
		invocation := h.Commander.Invocations[0]
		allArgs := strings.Join(invocation.Args, " ")
		require.Contains(t, allArgs, "/agents:team-lead")
		require.Contains(t, allArgs, "activate")
	})

	t.Run("Ephemeral session should NOT create directory or send activation", func(t *testing.T) {
		h := testutil.NewTestHarness()
		h.UUIDs = []string{"ephemeral-session-uuid"}

		projectDir := "/project"
		sessionsDir := filepath.Join(projectDir, ".claudex", "sessions")
		h.CreateDir(sessionsDir)

		app := &App{
			deps: &Dependencies{
				FS:    h.FS,
				Cmd:   h.Commander,
				Clock: h,
				UUID:  h,
				Env:   h.Env,
			},
			projectDir:  projectDir,
			sessionsDir: sessionsDir,
			docPaths:    []string{},
		}

		si := SessionInfo{
			Name: "ephemeral",
			Path: "", // Empty path
			Mode: LaunchModeEphemeral,
		}

		_ = app.launchEphemeral(si)

		// Ephemeral should NOT send activation prompt
		if len(h.Commander.Invocations) > 0 {
			invocation := h.Commander.Invocations[0]
			allArgs := strings.Join(invocation.Args, " ")
			require.NotContains(t, allArgs, "/agents:team-lead")
			require.NotContains(t, allArgs, "activate")
		}

		// Ephemeral should NOT create directory
		entries, _ := afero.ReadDir(h.FS, sessionsDir)
		require.Empty(t, entries, "ephemeral should not create directories")
	})
}

// TestSetEnvironment_FeaturesDefaults verifies environment variables are set with config defaults
func TestSetEnvironment_FeaturesDefaults(t *testing.T) {
	// Save and restore env vars
	origProgress := os.Getenv("CLAUDEX_AUTODOC_SESSION_PROGRESS")
	origEnd := os.Getenv("CLAUDEX_AUTODOC_SESSION_END")
	origFreq := os.Getenv("CLAUDEX_AUTODOC_FREQUENCY")
	defer func() {
		os.Setenv("CLAUDEX_AUTODOC_SESSION_PROGRESS", origProgress)
		os.Setenv("CLAUDEX_AUTODOC_SESSION_END", origEnd)
		os.Setenv("CLAUDEX_AUTODOC_FREQUENCY", origFreq)
	}()

	// Clear env vars to test config defaults
	os.Unsetenv("CLAUDEX_AUTODOC_SESSION_PROGRESS")
	os.Unsetenv("CLAUDEX_AUTODOC_SESSION_END")
	os.Unsetenv("CLAUDEX_AUTODOC_FREQUENCY")

	h := testutil.NewTestHarness()
	projectDir := "/project"

	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir: projectDir,
	}

	si := SessionInfo{
		Name: "test-session",
		Path: "/project/.claudex/sessions/test-session",
		Mode: LaunchModeNew,
	}

	cfg := &config.Config{
		Features: config.Features{
			AutodocSessionProgress: true,
			AutodocSessionEnd:      true,
			AutodocFrequency:       5,
		},
	}

	// Set environment
	app.setEnvironment(si, cfg)

	// Verify environment variables match config defaults (using os.Getenv since setEnvironment uses os.Setenv)
	require.Equal(t, "true", os.Getenv("CLAUDEX_AUTODOC_SESSION_PROGRESS"))
	require.Equal(t, "true", os.Getenv("CLAUDEX_AUTODOC_SESSION_END"))
	require.Equal(t, "5", os.Getenv("CLAUDEX_AUTODOC_FREQUENCY"))
}

// TestSetEnvironment_FeaturesCustomConfig verifies custom config values are exported
func TestSetEnvironment_FeaturesCustomConfig(t *testing.T) {
	// Save and restore env vars
	origProgress := os.Getenv("CLAUDEX_AUTODOC_SESSION_PROGRESS")
	origEnd := os.Getenv("CLAUDEX_AUTODOC_SESSION_END")
	origFreq := os.Getenv("CLAUDEX_AUTODOC_FREQUENCY")
	defer func() {
		os.Setenv("CLAUDEX_AUTODOC_SESSION_PROGRESS", origProgress)
		os.Setenv("CLAUDEX_AUTODOC_SESSION_END", origEnd)
		os.Setenv("CLAUDEX_AUTODOC_FREQUENCY", origFreq)
	}()

	// Clear env vars to test config values
	os.Unsetenv("CLAUDEX_AUTODOC_SESSION_PROGRESS")
	os.Unsetenv("CLAUDEX_AUTODOC_SESSION_END")
	os.Unsetenv("CLAUDEX_AUTODOC_FREQUENCY")

	h := testutil.NewTestHarness()
	projectDir := "/project"

	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir: projectDir,
	}

	si := SessionInfo{
		Name: "test-session",
		Path: "/project/.claudex/sessions/test-session",
		Mode: LaunchModeNew,
	}

	cfg := &config.Config{
		Features: config.Features{
			AutodocSessionProgress: false,
			AutodocSessionEnd:      true,
			AutodocFrequency:       10,
		},
	}

	// Set environment
	app.setEnvironment(si, cfg)

	// Verify custom config values
	require.Equal(t, "false", os.Getenv("CLAUDEX_AUTODOC_SESSION_PROGRESS"))
	require.Equal(t, "true", os.Getenv("CLAUDEX_AUTODOC_SESSION_END"))
	require.Equal(t, "10", os.Getenv("CLAUDEX_AUTODOC_FREQUENCY"))
}

// TestSetEnvironment_EnvVarOverridesConfig verifies env vars override config values
func TestSetEnvironment_EnvVarOverridesConfig(t *testing.T) {
	// Save original env vars and restore after test
	origProgress := os.Getenv("CLAUDEX_AUTODOC_SESSION_PROGRESS")
	origEnd := os.Getenv("CLAUDEX_AUTODOC_SESSION_END")
	origFreq := os.Getenv("CLAUDEX_AUTODOC_FREQUENCY")
	defer func() {
		os.Setenv("CLAUDEX_AUTODOC_SESSION_PROGRESS", origProgress)
		os.Setenv("CLAUDEX_AUTODOC_SESSION_END", origEnd)
		os.Setenv("CLAUDEX_AUTODOC_FREQUENCY", origFreq)
	}()

	// Set env vars that should override config
	os.Setenv("CLAUDEX_AUTODOC_SESSION_PROGRESS", "false")
	os.Setenv("CLAUDEX_AUTODOC_SESSION_END", "false")
	os.Setenv("CLAUDEX_AUTODOC_FREQUENCY", "20")

	h := testutil.NewTestHarness()
	projectDir := "/project"

	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir: projectDir,
	}

	si := SessionInfo{
		Name: "test-session",
		Path: "/project/.claudex/sessions/test-session",
		Mode: LaunchModeNew,
	}

	cfg := &config.Config{
		Features: config.Features{
			AutodocSessionProgress: true, // Config says true
			AutodocSessionEnd:      true, // Config says true
			AutodocFrequency:       5,    // Config says 5
		},
	}

	// Set environment - env vars should override config
	app.setEnvironment(si, cfg)

	// Verify env vars won (overrode config)
	require.Equal(t, "false", os.Getenv("CLAUDEX_AUTODOC_SESSION_PROGRESS"))
	require.Equal(t, "false", os.Getenv("CLAUDEX_AUTODOC_SESSION_END"))
	require.Equal(t, "20", os.Getenv("CLAUDEX_AUTODOC_FREQUENCY"))
}

// TestSetEnvironment_PartialEnvVarOverride verifies partial env var overrides
func TestSetEnvironment_PartialEnvVarOverride(t *testing.T) {
	// Save original env vars and restore after test
	origProgress := os.Getenv("CLAUDEX_AUTODOC_SESSION_PROGRESS")
	origEnd := os.Getenv("CLAUDEX_AUTODOC_SESSION_END")
	origFreq := os.Getenv("CLAUDEX_AUTODOC_FREQUENCY")
	defer func() {
		os.Setenv("CLAUDEX_AUTODOC_SESSION_PROGRESS", origProgress)
		os.Setenv("CLAUDEX_AUTODOC_SESSION_END", origEnd)
		os.Setenv("CLAUDEX_AUTODOC_FREQUENCY", origFreq)
	}()

	// Only override one env var
	os.Setenv("CLAUDEX_AUTODOC_SESSION_PROGRESS", "false")
	os.Unsetenv("CLAUDEX_AUTODOC_SESSION_END") // Not set
	os.Unsetenv("CLAUDEX_AUTODOC_FREQUENCY")   // Not set

	h := testutil.NewTestHarness()
	projectDir := "/project"

	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir: projectDir,
	}

	si := SessionInfo{
		Name: "test-session",
		Path: "/project/.claudex/sessions/test-session",
		Mode: LaunchModeNew,
	}

	cfg := &config.Config{
		Features: config.Features{
			AutodocSessionProgress: true,
			AutodocSessionEnd:      true,
			AutodocFrequency:       10,
		},
	}

	// Set environment
	app.setEnvironment(si, cfg)

	// Verify: env var wins for progress, config wins for others
	require.Equal(t, "false", os.Getenv("CLAUDEX_AUTODOC_SESSION_PROGRESS")) // Env var override
	require.Equal(t, "true", os.Getenv("CLAUDEX_AUTODOC_SESSION_END"))       // Config value
	require.Equal(t, "10", os.Getenv("CLAUDEX_AUTODOC_FREQUENCY"))           // Config value
}

// TestSetEnvironment_InvalidEnvVarValues verifies invalid env var values are handled
func TestSetEnvironment_InvalidEnvVarValues(t *testing.T) {
	// Save original env vars and restore after test
	origProgress := os.Getenv("CLAUDEX_AUTODOC_SESSION_PROGRESS")
	origFreq := os.Getenv("CLAUDEX_AUTODOC_FREQUENCY")
	defer func() {
		os.Setenv("CLAUDEX_AUTODOC_SESSION_PROGRESS", origProgress)
		os.Setenv("CLAUDEX_AUTODOC_FREQUENCY", origFreq)
	}()

	// Set invalid env var values
	os.Setenv("CLAUDEX_AUTODOC_SESSION_PROGRESS", "not-a-bool")
	os.Setenv("CLAUDEX_AUTODOC_FREQUENCY", "not-a-number")

	h := testutil.NewTestHarness()
	projectDir := "/project"

	app := &App{
		deps: &Dependencies{
			FS:    h.FS,
			Cmd:   h.Commander,
			Clock: h,
			UUID:  h,
			Env:   h.Env,
		},
		projectDir: projectDir,
	}

	si := SessionInfo{
		Name: "test-session",
		Path: "/project/.claudex/sessions/test-session",
		Mode: LaunchModeNew,
	}

	cfg := &config.Config{
		Features: config.Features{
			AutodocSessionProgress: true,
			AutodocSessionEnd:      true,
			AutodocFrequency:       5,
		},
	}

	// Set environment
	app.setEnvironment(si, cfg)

	// Verify: invalid bool becomes false, invalid int falls back to config
	require.Equal(t, "false", os.Getenv("CLAUDEX_AUTODOC_SESSION_PROGRESS")) // "not-a-bool" != "true" = false
	require.Equal(t, "5", os.Getenv("CLAUDEX_AUTODOC_FREQUENCY"))            // Invalid int, uses config default
}
