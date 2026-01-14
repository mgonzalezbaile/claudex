package preferences

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"claudex/internal/services/paths"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileService_Load_MissingFile(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	projectDir := "/test/project"
	service := New(fs, projectDir)

	// Execute
	prefs, err := service.Load()

	// Verify
	require.NoError(t, err)
	assert.False(t, prefs.HookSetupDeclined)
	assert.Equal(t, "", prefs.DeclinedAt)
}

func TestFileService_Load_ValidJSON(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	projectDir := "/test/project"
	claudexPath := filepath.Join(projectDir, paths.ClaudexDir)
	require.NoError(t, fs.MkdirAll(claudexPath, 0755))

	expectedPrefs := Preferences{
		HookSetupDeclined: true,
		DeclinedAt:        "2025-12-13T10:00:00Z",
	}

	data, err := json.Marshal(expectedPrefs)
	require.NoError(t, err)

	prefsPath := filepath.Join(projectDir, paths.PreferencesFile)
	require.NoError(t, afero.WriteFile(fs, prefsPath, data, 0644))

	service := New(fs, projectDir)

	// Execute
	prefs, err := service.Load()

	// Verify
	require.NoError(t, err)
	assert.Equal(t, expectedPrefs.HookSetupDeclined, prefs.HookSetupDeclined)
	assert.Equal(t, expectedPrefs.DeclinedAt, prefs.DeclinedAt)
}

func TestFileService_Load_InvalidJSON(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	projectDir := "/test/project"
	claudexPath := filepath.Join(projectDir, paths.ClaudexDir)
	require.NoError(t, fs.MkdirAll(claudexPath, 0755))

	prefsPath := filepath.Join(projectDir, paths.PreferencesFile)
	require.NoError(t, afero.WriteFile(fs, prefsPath, []byte("invalid json"), 0644))

	service := New(fs, projectDir)

	// Execute
	_, err := service.Load()

	// Verify
	require.Error(t, err)
}

func TestFileService_Save_CreatesDirectory(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	projectDir := "/test/project"
	service := New(fs, projectDir)

	prefs := Preferences{
		HookSetupDeclined: true,
		DeclinedAt:        "2025-12-13T11:00:00Z",
	}

	// Execute
	err := service.Save(prefs)

	// Verify
	require.NoError(t, err)

	// Verify .claudex directory was created
	claudexPath := filepath.Join(projectDir, paths.ClaudexDir)
	exists, err := afero.DirExists(fs, claudexPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify file was created
	prefsPath := filepath.Join(projectDir, paths.PreferencesFile)
	exists, err = afero.Exists(fs, prefsPath)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify content
	data, err := afero.ReadFile(fs, prefsPath)
	require.NoError(t, err)

	var readPrefs Preferences
	require.NoError(t, json.Unmarshal(data, &readPrefs))
	assert.Equal(t, prefs.HookSetupDeclined, readPrefs.HookSetupDeclined)
	assert.Equal(t, prefs.DeclinedAt, readPrefs.DeclinedAt)
}

func TestFileService_Save_UpdatesFile(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	projectDir := "/test/project"
	service := New(fs, projectDir)

	// Write initial preferences
	initialPrefs := Preferences{
		HookSetupDeclined: false,
		DeclinedAt:        "",
	}
	require.NoError(t, service.Save(initialPrefs))

	// Update preferences
	updatedPrefs := Preferences{
		HookSetupDeclined: true,
		DeclinedAt:        "2025-12-13T12:00:00Z",
	}

	// Execute
	err := service.Save(updatedPrefs)

	// Verify
	require.NoError(t, err)

	// Verify updated content
	readPrefs, err := service.Load()
	require.NoError(t, err)
	assert.Equal(t, updatedPrefs.HookSetupDeclined, readPrefs.HookSetupDeclined)
	assert.Equal(t, updatedPrefs.DeclinedAt, readPrefs.DeclinedAt)
}

func TestFileService_Save_NoTempFileRemains(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	projectDir := "/test/project"
	service := New(fs, projectDir)

	prefs := Preferences{
		HookSetupDeclined: true,
		DeclinedAt:        "2025-12-13T10:00:00Z",
	}

	// Execute
	err := service.Save(prefs)
	require.NoError(t, err)

	// Verify temp file doesn't exist
	tempPath := filepath.Join(projectDir, paths.PreferencesFile+".tmp")
	exists, err := afero.Exists(fs, tempPath)
	require.NoError(t, err)
	assert.False(t, exists, "temporary file should not remain after successful write")
}

func TestFileService_JSONRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		prefs Preferences
	}{
		{
			name: "hooks declined",
			prefs: Preferences{
				HookSetupDeclined: true,
				DeclinedAt:        "2025-12-13T10:00:00Z",
			},
		},
		{
			name: "hooks not declined",
			prefs: Preferences{
				HookSetupDeclined: false,
				DeclinedAt:        "",
			},
		},
		{
			name:  "zero value",
			prefs: Preferences{},
		},
		{
			name: "declined without timestamp",
			prefs: Preferences{
				HookSetupDeclined: true,
				DeclinedAt:        "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			fs := afero.NewMemMapFs()
			projectDir := "/test/project"
			service := New(fs, projectDir)

			// Execute write
			err := service.Save(tt.prefs)
			require.NoError(t, err)

			// Execute read
			readPrefs, err := service.Load()
			require.NoError(t, err)

			// Verify
			assert.Equal(t, tt.prefs.HookSetupDeclined, readPrefs.HookSetupDeclined)
			assert.Equal(t, tt.prefs.DeclinedAt, readPrefs.DeclinedAt)
		})
	}
}

func TestFileService_Save_DirectoryAlreadyExists(t *testing.T) {
	// Setup
	fs := afero.NewMemMapFs()
	projectDir := "/test/project"
	claudexPath := filepath.Join(projectDir, paths.ClaudexDir)
	require.NoError(t, fs.MkdirAll(claudexPath, 0755))

	service := New(fs, projectDir)

	prefs := Preferences{
		HookSetupDeclined: true,
		DeclinedAt:        "2025-12-13T11:00:00Z",
	}

	// Execute - should not fail even if directory exists
	err := service.Save(prefs)

	// Verify
	require.NoError(t, err)

	// Verify content
	readPrefs, err := service.Load()
	require.NoError(t, err)
	assert.Equal(t, prefs.HookSetupDeclined, readPrefs.HookSetupDeclined)
	assert.Equal(t, prefs.DeclinedAt, readPrefs.DeclinedAt)
}
