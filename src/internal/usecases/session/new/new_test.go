package new

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"claudex/internal/testutil"

	"github.com/stretchr/testify/require"
)

// Test_Execute_CreatesSessionWithMetadata tests basic session creation workflow
// Creates session directory with .description and .created files
func Test_Execute_CreatesSessionWithMetadata(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.FixedTime = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	sessionsDir := "/project/sessions"
	h.CreateDir(sessionsDir)

	// Mock commander to return slug
	h.Commander.OnPattern("claude", "-p").Return([]byte("implement-auth"), nil)
	h.UUIDs = []string{"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"}

	// Create usecase and execute
	uc := New(h.FS, h.Commander, h, h, sessionsDir)
	sessionName, sessionPath, claudeSessionID, err := uc.Execute("Add user authentication")

	// Verify success
	require.NoError(t, err)
	require.Equal(t, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", claudeSessionID)
	require.Equal(t, "implement-auth-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", sessionName)
	require.Equal(t, filepath.Join(sessionsDir, sessionName), sessionPath)

	// Verify directory created
	testutil.AssertDirExists(t, h.FS, sessionPath)

	// Verify .description file
	testutil.AssertFileExists(t, h.FS, filepath.Join(sessionPath, ".description"))
	testutil.AssertFileContains(t, h.FS, filepath.Join(sessionPath, ".description"), "Add user authentication")

	// Verify .created file with timestamp
	testutil.AssertFileExists(t, h.FS, filepath.Join(sessionPath, ".created"))
	testutil.AssertFileContains(t, h.FS, filepath.Join(sessionPath, ".created"), "2024-01-15T10:30:00Z")
}

// Test_Execute_FallsBackToManualSlug tests fallback when Claude CLI fails
// Should generate slug from description words when API unavailable
func Test_Execute_FallsBackToManualSlug(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	sessionsDir := "/project/sessions"
	h.CreateDir(sessionsDir)

	// Mock commander to fail (simulates Claude CLI unavailable)
	h.Commander.OnPattern("claude").Return(nil, fmt.Errorf("command not found"))
	h.UUIDs = []string{"11111111-2222-3333-4444-555555555555"}

	// Create usecase and execute
	uc := New(h.FS, h.Commander, h, h, sessionsDir)
	sessionName, sessionPath, _, err := uc.Execute("Fix login bug in dashboard")

	// Verify success with manual slug fallback
	require.NoError(t, err)
	// Manual slug takes first 3 words, lowercased, hyphenated
	require.Contains(t, sessionName, "fix-login-bug")
	require.Contains(t, sessionName, "11111111-2222-3333-4444-555555555555")
	testutil.AssertDirExists(t, h.FS, sessionPath)
}

// Test_Execute_HandlesCollision tests unique name generation on collision
// Should append counter when session name already exists
func Test_Execute_HandlesCollision(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	sessionsDir := "/project/sessions"

	// Pre-create existing session with same name pattern
	existingSessionPath := filepath.Join(sessionsDir, "my-task-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	h.CreateSessionWithFiles(existingSessionPath, map[string]string{
		".description": "Existing task",
	})

	// Mock commander to return same slug
	h.Commander.OnPattern("claude", "-p").Return([]byte("my-task"), nil)
	h.UUIDs = []string{"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"}

	// Create usecase and execute
	uc := New(h.FS, h.Commander, h, h, sessionsDir)
	sessionName, sessionPath, _, err := uc.Execute("My task description")

	// Verify collision handling - should append counter
	require.NoError(t, err)
	require.Equal(t, "my-task-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee-1", sessionName)
	testutil.AssertDirExists(t, h.FS, sessionPath)

	// Original still exists
	testutil.AssertDirExists(t, h.FS, existingSessionPath)
}

// Test_Execute_RejectsEmptyDescription tests validation of description input
// Should return error when description is empty or whitespace only
func Test_Execute_RejectsEmptyDescription(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	sessionsDir := "/project/sessions"
	h.CreateDir(sessionsDir)

	uc := New(h.FS, h.Commander, h, h, sessionsDir)

	// Test empty string
	_, _, _, err := uc.Execute("")
	require.Error(t, err)
	require.Contains(t, err.Error(), "description cannot be empty")

	// Test whitespace only
	_, _, _, err = uc.Execute("   ")
	require.Error(t, err)
	require.Contains(t, err.Error(), "description cannot be empty")
}

// Test_Execute_GeneratesUniqueUUID tests UUID generation for each session
// Each session should have a unique Claude session ID
func Test_Execute_GeneratesUniqueUUID(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	sessionsDir := "/project/sessions"
	h.CreateDir(sessionsDir)

	// Mock commander
	h.Commander.OnPattern("claude", "-p").Return([]byte("task"), nil)
	h.UUIDs = []string{
		"uuid-1111-1111-1111-111111111111",
		"uuid-2222-2222-2222-222222222222",
	}

	uc := New(h.FS, h.Commander, h, h, sessionsDir)

	// Create first session
	_, _, uuid1, err := uc.Execute("First task")
	require.NoError(t, err)
	require.Equal(t, "uuid-1111-1111-1111-111111111111", uuid1)

	// Create second session
	_, _, uuid2, err := uc.Execute("Second task")
	require.NoError(t, err)
	require.Equal(t, "uuid-2222-2222-2222-222222222222", uuid2)

	// UUIDs should be different
	require.NotEqual(t, uuid1, uuid2)
}

// Test_Execute_InvokesClaudeCLI tests that Claude CLI is called for slug generation
// Should call "claude -p" with appropriate prompt
func Test_Execute_InvokesClaudeCLI(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	sessionsDir := "/project/sessions"
	h.CreateDir(sessionsDir)

	h.Commander.OnPattern("claude", "-p").Return([]byte("generated-slug"), nil)
	h.UUIDs = []string{"test-uuid"}

	// Create usecase and execute
	uc := New(h.FS, h.Commander, h, h, sessionsDir)
	_, _, _, err := uc.Execute("My description for testing")

	// Verify Claude CLI was invoked
	require.NoError(t, err)
	require.Len(t, h.Commander.Invocations, 1)
	invocation := h.Commander.Invocations[0]
	require.Equal(t, "claude", invocation.Name)
	require.Contains(t, invocation.Args, "-p")
}

// Test_Execute_CreatesSessionsDirectory tests directory creation
// Should create sessions directory if it doesn't exist
func Test_Execute_CreatesSessionsDirectory(t *testing.T) {
	// Setup - sessions directory does NOT exist
	h := testutil.NewTestHarness()
	sessionsDir := "/project/sessions"
	// Intentionally NOT creating sessionsDir

	h.Commander.OnPattern("claude", "-p").Return([]byte("new-feature"), nil)
	h.UUIDs = []string{"test-uuid"}

	// Create usecase and execute
	uc := New(h.FS, h.Commander, h, h, sessionsDir)
	_, sessionPath, _, err := uc.Execute("New feature description")

	// Should succeed and create the directory structure
	require.NoError(t, err)
	testutil.AssertDirExists(t, h.FS, sessionPath)
}

// Test_Execute_HandlesSpecialCharactersInDescription tests slug sanitization
// Description with special characters should produce clean slug
func Test_Execute_HandlesSpecialCharactersInDescription(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	sessionsDir := "/project/sessions"
	h.CreateDir(sessionsDir)

	// Mock commander to fail so we test manual slug fallback with special chars
	h.Commander.OnPattern("claude").Return(nil, fmt.Errorf("unavailable"))
	h.UUIDs = []string{"test-uuid"}

	uc := New(h.FS, h.Commander, h, h, sessionsDir)
	sessionName, _, _, err := uc.Execute("Fix bug #123 (urgent!)")

	// Verify slug is sanitized (manual fallback)
	require.NoError(t, err)
	// Should not contain special characters in final path
	require.NotContains(t, sessionName, "#")
	require.NotContains(t, sessionName, "(")
	require.NotContains(t, sessionName, ")")
	require.NotContains(t, sessionName, "!")
}

// Test_Execute_SetsCorrectFilePermissions tests metadata file permissions
// .description and .created should have 0644 permissions
func Test_Execute_SetsCorrectFilePermissions(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	sessionsDir := "/project/sessions"
	h.CreateDir(sessionsDir)

	h.Commander.OnPattern("claude", "-p").Return([]byte("test-task"), nil)
	h.UUIDs = []string{"test-uuid"}

	uc := New(h.FS, h.Commander, h, h, sessionsDir)
	_, sessionPath, _, err := uc.Execute("Test task")

	require.NoError(t, err)

	// Check .description permissions
	descInfo, err := h.FS.Stat(filepath.Join(sessionPath, ".description"))
	require.NoError(t, err)
	require.Equal(t, "-rw-r--r--", descInfo.Mode().String())

	// Check .created permissions
	createdInfo, err := h.FS.Stat(filepath.Join(sessionPath, ".created"))
	require.NoError(t, err)
	require.Equal(t, "-rw-r--r--", createdInfo.Mode().String())
}
