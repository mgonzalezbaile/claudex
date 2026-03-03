package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"claudex/internal/testutil"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// Test_HasClaudeSessionID tests session ID detection in session names
func Test_HasClaudeSessionID(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		want        bool
	}{
		{
			name:        "Valid UUID in session name",
			sessionName: "feature-login-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			want:        true,
		},
		{
			name:        "No UUID in session name",
			sessionName: "feature-login",
			want:        false,
		},
		{
			name:        "Invalid UUID format",
			sessionName: "feature-login-invalid-uuid-format",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasClaudeSessionID(tt.sessionName)
			require.Equal(t, tt.want, result)
		})
	}
}

// Test_ExtractClaudeSessionID tests extracting Claude session ID from session names
func Test_ExtractClaudeSessionID(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		want        string
	}{
		{
			name:        "Extract UUID from end of session name",
			sessionName: "feature-login-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			want:        "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		},
		{
			name:        "No UUID returns empty string",
			sessionName: "feature-login",
			want:        "",
		},
		{
			name:        "Complex slug with UUID",
			sessionName: "implement-auth-flow-11112222-3333-4444-5555-666666666666",
			want:        "11112222-3333-4444-5555-666666666666",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractClaudeSessionID(tt.sessionName)
			require.Equal(t, tt.want, result)
		})
	}
}

// Test_StripClaudeSessionID tests removing Claude session ID from session names
func Test_StripClaudeSessionID(t *testing.T) {
	tests := []struct {
		name        string
		sessionName string
		want        string
	}{
		{
			name:        "Strip UUID from session name",
			sessionName: "feature-login-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			want:        "feature-login",
		},
		{
			name:        "No UUID returns original",
			sessionName: "feature-login",
			want:        "feature-login",
		},
		{
			name:        "Complex slug with UUID",
			sessionName: "implement-auth-flow-11112222-3333-4444-5555-666666666666",
			want:        "implement-auth-flow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripClaudeSessionID(tt.sessionName)
			require.Equal(t, tt.want, result)
		})
	}
}

// Test_GenerateNameWithCmd tests slug generation using Claude CLI
func Test_GenerateNameWithCmd(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()

	// Test slug generation with Commander mock
	h.Commander.OnPattern("claude", "-p").Return([]byte("feature-login"), nil)

	description := "Implement login feature"
	slug, err := GenerateNameWithCmd(h.Commander, description)

	// Verify slug generation
	require.NoError(t, err)
	require.Equal(t, "feature-login", slug)

	// Verify Commander was invoked
	require.Len(t, h.Commander.Invocations, 1)
	invocation := h.Commander.Invocations[0]
	require.Equal(t, "claude", invocation.Name)
	require.Contains(t, invocation.Args, "-p")
}

// Test_UpdateLastUsedWithDeps tests updating last used timestamp
func Test_UpdateLastUsedWithDeps(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()
	h.FixedTime = time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)

	sessionDir := "/project/.claudex/sessions/feature-login-aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	h.CreateSessionWithFiles(sessionDir, map[string]string{
		".description": "Login feature",
		".created":     "2024-01-15T10:30:00Z",
	})

	// Exercise - Update last used
	err := UpdateLastUsedWithDeps(h.FS, h, sessionDir)

	// Verify
	require.NoError(t, err)
	testutil.AssertFileExists(t, h.FS, filepath.Join(sessionDir, ".last_used"))
	testutil.AssertFileContains(t, h.FS, filepath.Join(sessionDir, ".last_used"), "2024-01-15T14:00:00Z")
}

// Test_UpdateLastUsedWithDeps_EphemeralSession tests that ephemeral sessions (empty path) are handled
func Test_UpdateLastUsedWithDeps_EphemeralSession(t *testing.T) {
	// Setup
	h := testutil.NewTestHarness()

	// Exercise - Update last used with empty path (ephemeral session)
	err := UpdateLastUsedWithDeps(h.FS, h, "")

	// Verify - no error for ephemeral sessions
	require.NoError(t, err)
}

// Test_GetSessions verifies that Description and Date fields are populated separately
func Test_GetSessions(t *testing.T) {
	// Setup - create real directories since GetSessions uses os.ReadDir
	sessionsDir := t.TempDir()
	sessionDir := filepath.Join(sessionsDir, "feature-login")
	require.NoError(t, os.MkdirAll(sessionDir, 0755))

	fs := afero.NewOsFs()
	require.NoError(t, afero.WriteFile(fs, filepath.Join(sessionDir, ".description"), []byte("Login feature"), 0644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(sessionDir, ".last_used"), []byte("2024-06-15T10:30:00Z"), 0644))

	// Exercise
	sessions, err := GetSessions(fs, sessionsDir)

	// Verify
	require.NoError(t, err)
	require.Len(t, sessions, 1)
	require.Equal(t, "feature-login", sessions[0].Title)
	require.Equal(t, "Login feature", sessions[0].Description)
	require.Equal(t, "15 Jun 2024 10:30:00", sessions[0].Date)
}
