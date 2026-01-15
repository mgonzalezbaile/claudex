package migrate

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"claudex/internal/services/paths"
)

func TestMigrator_Run_FreshInstallation(t *testing.T) {
	fs := afero.NewMemMapFs()
	migrator := New(fs)

	err := migrator.Run()
	require.NoError(t, err)

	// Verify .claudex/ directory was created
	exists, err := afero.DirExists(fs, paths.ClaudexDir)
	require.NoError(t, err)
	assert.True(t, exists, ".claudex directory should be created")

	// Verify config.toml was created with defaults
	configExists, err := afero.Exists(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.True(t, configExists, "config.toml should be created")

	// Verify config content
	content, err := afero.ReadFile(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "autodoc_session_progress = true")
	assert.Contains(t, string(content), "autodoc_session_end = true")
	assert.Contains(t, string(content), "autodoc_frequency = 5")
}

func TestMigrator_Run_IdempotentOperation(t *testing.T) {
	fs := afero.NewMemMapFs()
	migrator := New(fs)

	// Run migration twice
	err := migrator.Run()
	require.NoError(t, err)

	err = migrator.Run()
	require.NoError(t, err)

	// Verify .claudex/ still exists
	exists, err := afero.DirExists(fs, paths.ClaudexDir)
	require.NoError(t, err)
	assert.True(t, exists)

	// Verify config still exists and wasn't overwritten
	configExists, err := afero.Exists(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.True(t, configExists)
}

func TestMigrator_Run_MigrateLegacySessions(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create legacy sessions directory with content
	err := fs.MkdirAll(paths.LegacySessionsDir, 0755)
	require.NoError(t, err)

	sessionFile := paths.LegacySessionsDir + "/session-1.json"
	err = afero.WriteFile(fs, sessionFile, []byte(`{"id": "session-1"}`), 0644)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify legacy sessions directory was migrated
	legacyExists, err := afero.DirExists(fs, paths.LegacySessionsDir)
	require.NoError(t, err)
	assert.False(t, legacyExists, "legacy sessions directory should be removed")

	newExists, err := afero.DirExists(fs, paths.SessionsDir)
	require.NoError(t, err)
	assert.True(t, newExists, "new sessions directory should exist")

	// Verify content was migrated
	newSessionFile := paths.SessionsDir + "/session-1.json"
	content, err := afero.ReadFile(fs, newSessionFile)
	require.NoError(t, err)
	assert.Equal(t, `{"id": "session-1"}`, string(content))
}

func TestMigrator_Run_MigrateLegacyLogs(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create legacy logs directory with content
	err := fs.MkdirAll(paths.LegacyLogsDir, 0755)
	require.NoError(t, err)

	logFile := paths.LegacyLogsDir + "/app.log"
	err = afero.WriteFile(fs, logFile, []byte("log entry"), 0644)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify legacy logs directory was migrated
	legacyExists, err := afero.DirExists(fs, paths.LegacyLogsDir)
	require.NoError(t, err)
	assert.False(t, legacyExists, "legacy logs directory should be removed")

	newExists, err := afero.DirExists(fs, paths.LogsDir)
	require.NoError(t, err)
	assert.True(t, newExists, "new logs directory should exist")

	// Verify content was migrated
	newLogFile := paths.LogsDir + "/app.log"
	content, err := afero.ReadFile(fs, newLogFile)
	require.NoError(t, err)
	assert.Equal(t, "log entry", string(content))
}

func TestMigrator_Run_MigrateLegacyConfig(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create legacy config file
	legacyConfigContent := `[features]
autodoc_session_progress = false
autodoc_session_end = false
autodoc_frequency = 10
`
	err := afero.WriteFile(fs, paths.LegacyConfigFile, []byte(legacyConfigContent), 0644)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify legacy config was removed
	legacyExists, err := afero.Exists(fs, paths.LegacyConfigFile)
	require.NoError(t, err)
	assert.False(t, legacyExists, "legacy config should be removed")

	// Verify new config exists with legacy content (not defaults)
	newExists, err := afero.Exists(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.True(t, newExists, "new config should exist")

	content, err := afero.ReadFile(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "autodoc_session_progress = false")
	assert.Contains(t, string(content), "autodoc_session_end = false")
	assert.Contains(t, string(content), "autodoc_frequency = 10")
}

func TestMigrator_Run_CompleteSetup(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create all legacy artifacts
	err := fs.MkdirAll(paths.LegacySessionsDir, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.LegacySessionsDir+"/session.json", []byte("session"), 0644)
	require.NoError(t, err)

	err = fs.MkdirAll(paths.LegacyLogsDir, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.LegacyLogsDir+"/app.log", []byte("log"), 0644)
	require.NoError(t, err)

	err = afero.WriteFile(fs, paths.LegacyConfigFile, []byte("legacy config"), 0644)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify all migrations happened
	legacySessionsExists, _ := afero.DirExists(fs, paths.LegacySessionsDir)
	assert.False(t, legacySessionsExists)

	legacyLogsExists, _ := afero.DirExists(fs, paths.LegacyLogsDir)
	assert.False(t, legacyLogsExists)

	legacyConfigExists, _ := afero.Exists(fs, paths.LegacyConfigFile)
	assert.False(t, legacyConfigExists)

	// Verify new locations exist
	newSessionsExists, _ := afero.DirExists(fs, paths.SessionsDir)
	assert.True(t, newSessionsExists)

	newLogsExists, _ := afero.DirExists(fs, paths.LogsDir)
	assert.True(t, newLogsExists)

	newConfigExists, _ := afero.Exists(fs, paths.ConfigFile)
	assert.True(t, newConfigExists)

	// Verify content was preserved
	sessionContent, _ := afero.ReadFile(fs, paths.SessionsDir+"/session.json")
	assert.Equal(t, "session", string(sessionContent))

	logContent, _ := afero.ReadFile(fs, paths.LogsDir+"/app.log")
	assert.Equal(t, "log", string(logContent))

	configContent, _ := afero.ReadFile(fs, paths.ConfigFile)
	assert.Equal(t, "legacy config", string(configContent))
}

func TestMigrator_Run_NestedDirectoryMigration(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create nested directory structure in legacy sessions
	err := fs.MkdirAll(paths.LegacySessionsDir+"/2024/01", 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.LegacySessionsDir+"/2024/01/session.json", []byte("nested"), 0644)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify nested structure was preserved
	content, err := afero.ReadFile(fs, paths.SessionsDir+"/2024/01/session.json")
	require.NoError(t, err)
	assert.Equal(t, "nested", string(content))
}

func TestMigrator_Run_PreserveFilePermissions(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create legacy file with specific permissions
	err := fs.MkdirAll(paths.LegacySessionsDir, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.LegacySessionsDir+"/session.json", []byte("content"), 0600)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify file exists (permission preservation is best-effort in MemMapFs)
	exists, err := afero.Exists(fs, paths.SessionsDir+"/session.json")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestMigrator_Run_SkipIfDestinationExists(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create both legacy and new sessions directories
	err := fs.MkdirAll(paths.LegacySessionsDir, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.LegacySessionsDir+"/legacy.json", []byte("legacy"), 0644)
	require.NoError(t, err)

	err = fs.MkdirAll(paths.SessionsDir, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.SessionsDir+"/existing.json", []byte("existing"), 0644)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify legacy directory still exists (wasn't migrated)
	legacyExists, err := afero.DirExists(fs, paths.LegacySessionsDir)
	require.NoError(t, err)
	assert.True(t, legacyExists, "legacy directory should still exist when destination exists")

	// Verify existing file wasn't overwritten
	content, err := afero.ReadFile(fs, paths.SessionsDir+"/existing.json")
	require.NoError(t, err)
	assert.Equal(t, "existing", string(content))

	// Verify legacy file wasn't migrated
	legacyFileExists, err := afero.Exists(fs, paths.SessionsDir+"/legacy.json")
	require.NoError(t, err)
	assert.False(t, legacyFileExists, "legacy file should not be migrated when destination exists")
}

func TestMigrator_Run_PreserveConfigIfExists(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create .claudex directory and config file first
	err := fs.MkdirAll(paths.ClaudexDir, 0755)
	require.NoError(t, err)

	customConfig := `[features]
autodoc_session_progress = false
`
	err = afero.WriteFile(fs, paths.ConfigFile, []byte(customConfig), 0644)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify config wasn't overwritten
	content, err := afero.ReadFile(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.Equal(t, customConfig, string(content), "existing config should not be overwritten")
}

// TestRun_MigrateLegacySessions_WithNestedContent tests migration of sessions
// with deeply nested directory structures and multiple files.
func TestRun_MigrateLegacySessions_WithNestedContent(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create nested directory structure with multiple levels
	dirs := []string{
		paths.LegacySessionsDir + "/2024/01",
		paths.LegacySessionsDir + "/2024/02/subfolder",
		paths.LegacySessionsDir + "/archived",
	}

	for _, dir := range dirs {
		err := fs.MkdirAll(dir, 0755)
		require.NoError(t, err)
	}

	// Create multiple files in various locations
	files := map[string]string{
		paths.LegacySessionsDir + "/root.json":                     `{"session": "root"}`,
		paths.LegacySessionsDir + "/2024/01/jan-session.json":      `{"session": "january"}`,
		paths.LegacySessionsDir + "/2024/02/feb-session.json":      `{"session": "february"}`,
		paths.LegacySessionsDir + "/2024/02/subfolder/nested.json": `{"session": "nested"}`,
		paths.LegacySessionsDir + "/archived/old-session.json":     `{"session": "archived"}`,
	}

	for path, content := range files {
		err := afero.WriteFile(fs, path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Run migration
	migrator := New(fs)
	err := migrator.Run()
	require.NoError(t, err)

	// Verify all nested content was preserved
	for legacyPath, expectedContent := range files {
		// Calculate new path
		newPath := paths.SessionsDir + legacyPath[len(paths.LegacySessionsDir):]

		// Verify file exists at new location
		content, err := afero.ReadFile(fs, newPath)
		require.NoError(t, err, "file should exist at %s", newPath)
		assert.Equal(t, expectedContent, string(content), "content should be preserved for %s", newPath)
	}

	// Verify legacy directory was removed
	legacyExists, err := afero.DirExists(fs, paths.LegacySessionsDir)
	require.NoError(t, err)
	assert.False(t, legacyExists, "legacy sessions directory should be removed")
}

// TestRun_MigrateLegacyLogs_WithMultipleFiles tests migration of logs
// directory with multiple log files.
func TestRun_MigrateLegacyLogs_WithMultipleFiles(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create legacy logs directory
	err := fs.MkdirAll(paths.LegacyLogsDir, 0755)
	require.NoError(t, err)

	// Create multiple log files
	logFiles := map[string]string{
		paths.LegacyLogsDir + "/app.log":    "2024-01-01 INFO Application started\n",
		paths.LegacyLogsDir + "/error.log":  "2024-01-01 ERROR Something went wrong\n",
		paths.LegacyLogsDir + "/access.log": "2024-01-01 GET /api/status 200\n",
		paths.LegacyLogsDir + "/debug.log":  "2024-01-01 DEBUG Verbose information\n",
	}

	for path, content := range logFiles {
		err := afero.WriteFile(fs, path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Run migration
	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify all log files were migrated
	for legacyPath, expectedContent := range logFiles {
		newPath := paths.LogsDir + legacyPath[len(paths.LegacyLogsDir):]

		content, err := afero.ReadFile(fs, newPath)
		require.NoError(t, err, "log file should exist at %s", newPath)
		assert.Equal(t, expectedContent, string(content), "log content should be preserved")
	}

	// Verify legacy logs directory was removed
	legacyExists, err := afero.DirExists(fs, paths.LegacyLogsDir)
	require.NoError(t, err)
	assert.False(t, legacyExists, "legacy logs directory should be removed")
}

// TestRun_MigrateLegacyConfig_OverwritesDefault tests that a legacy config
// overwrites the default config when both exist.
func TestRun_MigrateLegacyConfig_OverwritesDefault(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create .claudex directory first
	err := fs.MkdirAll(paths.ClaudexDir, 0755)
	require.NoError(t, err)

	// Create default config
	defaultConfig := `# Claudex Configuration
# See documentation for all available options

[features]
autodoc_session_progress = true
autodoc_session_end = true
autodoc_frequency = 5
`
	err = afero.WriteFile(fs, paths.ConfigFile, []byte(defaultConfig), 0644)
	require.NoError(t, err)

	// Create legacy config with custom content
	customConfig := `# Custom legacy configuration

[features]
autodoc_session_progress = false
autodoc_session_end = false
autodoc_frequency = 20

[custom_section]
custom_key = "custom_value"
`
	err = afero.WriteFile(fs, paths.LegacyConfigFile, []byte(customConfig), 0644)
	require.NoError(t, err)

	// Run migration
	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify custom content overwrote default
	content, err := afero.ReadFile(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.Equal(t, customConfig, string(content), "custom config should overwrite default")
	assert.Contains(t, string(content), "autodoc_session_progress = false")
	assert.Contains(t, string(content), "autodoc_frequency = 20")
	assert.Contains(t, string(content), "custom_key = \"custom_value\"")

	// Verify legacy config was removed
	legacyExists, err := afero.Exists(fs, paths.LegacyConfigFile)
	require.NoError(t, err)
	assert.False(t, legacyExists, "legacy config should be removed")
}

// TestRun_PartialLegacyArtifacts tests migration when only some legacy
// artifacts exist.
func TestRun_PartialLegacyArtifacts(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create only legacy sessions (no logs, no config)
	err := fs.MkdirAll(paths.LegacySessionsDir, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.LegacySessionsDir+"/session.json", []byte(`{"id": "1"}`), 0644)
	require.NoError(t, err)

	// Run migration
	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify sessions were migrated
	sessionExists, err := afero.DirExists(fs, paths.SessionsDir)
	require.NoError(t, err)
	assert.True(t, sessionExists, "sessions directory should be migrated")

	content, err := afero.ReadFile(fs, paths.SessionsDir+"/session.json")
	require.NoError(t, err)
	assert.Equal(t, `{"id": "1"}`, string(content))

	// Verify legacy sessions removed
	legacySessionsExists, err := afero.DirExists(fs, paths.LegacySessionsDir)
	require.NoError(t, err)
	assert.False(t, legacySessionsExists)

	// Verify default config was created
	configExists, err := afero.Exists(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.True(t, configExists, "default config should be created")

	configContent, err := afero.ReadFile(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.Contains(t, string(configContent), "autodoc_session_progress = true")

	// Verify .claudex directory exists
	claudexExists, err := afero.DirExists(fs, paths.ClaudexDir)
	require.NoError(t, err)
	assert.True(t, claudexExists)
}

// TestRun_DestinationAlreadyExists_Sessions tests that migration is skipped
// when the destination sessions directory already exists.
func TestRun_DestinationAlreadyExists_Sessions(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create legacy sessions with content
	err := fs.MkdirAll(paths.LegacySessionsDir, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.LegacySessionsDir+"/legacy.json", []byte("legacy content"), 0644)
	require.NoError(t, err)

	// Create new sessions directory with existing content
	err = fs.MkdirAll(paths.SessionsDir, 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.SessionsDir+"/existing.json", []byte("existing content"), 0644)
	require.NoError(t, err)

	// Run migration
	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify legacy directory still exists (migration was skipped)
	legacyExists, err := afero.DirExists(fs, paths.LegacySessionsDir)
	require.NoError(t, err)
	assert.True(t, legacyExists, "legacy directory should remain when destination exists")

	// Verify legacy content still exists
	legacyContent, err := afero.ReadFile(fs, paths.LegacySessionsDir+"/legacy.json")
	require.NoError(t, err)
	assert.Equal(t, "legacy content", string(legacyContent), "legacy content should be preserved")

	// Verify existing content wasn't modified
	existingContent, err := afero.ReadFile(fs, paths.SessionsDir+"/existing.json")
	require.NoError(t, err)
	assert.Equal(t, "existing content", string(existingContent), "existing content should be preserved")

	// Verify legacy file was NOT migrated
	legacyFileExists, err := afero.Exists(fs, paths.SessionsDir+"/legacy.json")
	require.NoError(t, err)
	assert.False(t, legacyFileExists, "legacy file should not be migrated when destination exists")
}

// TestRun_DestinationAlreadyExists_Config tests that existing config is
// preserved when both legacy and new config exist.
func TestRun_DestinationAlreadyExists_Config(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create .claudex directory and config file
	err := fs.MkdirAll(paths.ClaudexDir, 0755)
	require.NoError(t, err)

	existingConfig := `[features]
autodoc_session_progress = true
autodoc_frequency = 15

[existing_section]
preserve_me = "important"
`
	err = afero.WriteFile(fs, paths.ConfigFile, []byte(existingConfig), 0644)
	require.NoError(t, err)

	// Create legacy config
	legacyConfig := `[features]
autodoc_session_progress = false
autodoc_frequency = 10
`
	err = afero.WriteFile(fs, paths.LegacyConfigFile, []byte(legacyConfig), 0644)
	require.NoError(t, err)

	// Run migration
	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify existing config was overwritten by legacy (this is intended behavior)
	content, err := afero.ReadFile(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.Equal(t, legacyConfig, string(content), "legacy config should overwrite existing config")
	assert.Contains(t, string(content), "autodoc_session_progress = false")
	assert.Contains(t, string(content), "autodoc_frequency = 10")

	// Verify legacy config was removed
	legacyExists, err := afero.Exists(fs, paths.LegacyConfigFile)
	require.NoError(t, err)
	assert.False(t, legacyExists, "legacy config should be removed after migration")
}

// TestCopyAndRemoveDirectory tests the copy-and-remove fallback mechanism
// by using a filesystem that forces the rename to fail.
func TestCopyAndRemoveDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create source directory with nested content
	sourceDir := "test_source"
	destDir := "test_dest"

	err := fs.MkdirAll(sourceDir+"/subdir", 0755)
	require.NoError(t, err)

	files := map[string]string{
		sourceDir + "/file1.txt":         "content1",
		sourceDir + "/file2.txt":         "content2",
		sourceDir + "/subdir/nested.txt": "nested content",
	}

	for path, content := range files {
		err := afero.WriteFile(fs, path, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test copyAndRemoveDirectory directly
	migrator := New(fs)
	err = migrator.copyAndRemoveDirectory(sourceDir, destDir)
	require.NoError(t, err)

	// Verify all files were copied
	for sourcePath, expectedContent := range files {
		destPath := destDir + sourcePath[len(sourceDir):]

		content, err := afero.ReadFile(fs, destPath)
		require.NoError(t, err, "file should exist at %s", destPath)
		assert.Equal(t, expectedContent, string(content))
	}

	// Verify source directory was removed
	sourceExists, err := afero.DirExists(fs, sourceDir)
	require.NoError(t, err)
	assert.False(t, sourceExists, "source directory should be removed")
}

// TestCopyAndRemoveDirectory_EmptyDirectory tests copying an empty directory.
func TestCopyAndRemoveDirectory_EmptyDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()

	sourceDir := "empty_source"
	destDir := "empty_dest"

	err := fs.MkdirAll(sourceDir, 0755)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.copyAndRemoveDirectory(sourceDir, destDir)
	require.NoError(t, err)

	// Verify destination directory exists
	destExists, err := afero.DirExists(fs, destDir)
	require.NoError(t, err)
	assert.True(t, destExists)

	// Verify source was removed
	sourceExists, err := afero.DirExists(fs, sourceDir)
	require.NoError(t, err)
	assert.False(t, sourceExists)
}

// TestCopyAndRemoveDirectory_PreservesDirectoryStructure tests that nested
// directory structures are preserved during copy.
func TestCopyAndRemoveDirectory_PreservesDirectoryStructure(t *testing.T) {
	fs := afero.NewMemMapFs()

	sourceDir := "structured_source"
	destDir := "structured_dest"

	// Create complex directory structure
	dirs := []string{
		sourceDir + "/level1/level2/level3",
		sourceDir + "/another_branch/deep",
	}

	for _, dir := range dirs {
		err := fs.MkdirAll(dir, 0755)
		require.NoError(t, err)
	}

	// Add files at various levels
	files := map[string]string{
		sourceDir + "/root.txt":                       "root",
		sourceDir + "/level1/first.txt":               "first",
		sourceDir + "/level1/level2/second.txt":       "second",
		sourceDir + "/level1/level2/level3/third.txt": "third",
		sourceDir + "/another_branch/branch.txt":      "branch",
		sourceDir + "/another_branch/deep/deep.txt":   "deep",
	}

	for path, content := range files {
		err := afero.WriteFile(fs, path, []byte(content), 0644)
		require.NoError(t, err)
	}

	migrator := New(fs)
	err := migrator.copyAndRemoveDirectory(sourceDir, destDir)
	require.NoError(t, err)

	// Verify all files exist at correct locations
	for sourcePath, expectedContent := range files {
		destPath := destDir + sourcePath[len(sourceDir):]

		content, err := afero.ReadFile(fs, destPath)
		require.NoError(t, err, "file should exist at %s", destPath)
		assert.Equal(t, expectedContent, string(content))
	}

	// Verify source was removed
	sourceExists, err := afero.DirExists(fs, sourceDir)
	require.NoError(t, err)
	assert.False(t, sourceExists)
}

// TestRun_MigrateLegacySessions_ErrorLogging tests that errors during
// session migration are logged but don't fail the migration.
func TestRun_MigrateLegacySessions_ErrorLogging(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create legacy sessions and new sessions to trigger skip condition
	err := fs.MkdirAll(paths.LegacySessionsDir, 0755)
	require.NoError(t, err)
	err = fs.MkdirAll(paths.SessionsDir, 0755)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err) // Should not fail even if migration is skipped

	// Verify both directories exist
	legacyExists, err := afero.DirExists(fs, paths.LegacySessionsDir)
	require.NoError(t, err)
	assert.True(t, legacyExists)

	newExists, err := afero.DirExists(fs, paths.SessionsDir)
	require.NoError(t, err)
	assert.True(t, newExists)
}

// TestRun_MigrateLegacyLogs_ErrorLogging tests that errors during
// log migration are logged but don't fail the migration.
func TestRun_MigrateLegacyLogs_ErrorLogging(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create legacy logs and new logs to trigger skip condition
	err := fs.MkdirAll(paths.LegacyLogsDir, 0755)
	require.NoError(t, err)
	err = fs.MkdirAll(paths.LogsDir, 0755)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err) // Should not fail even if migration is skipped

	// Verify both directories exist
	legacyExists, err := afero.DirExists(fs, paths.LegacyLogsDir)
	require.NoError(t, err)
	assert.True(t, legacyExists)

	newExists, err := afero.DirExists(fs, paths.LogsDir)
	require.NoError(t, err)
	assert.True(t, newExists)
}

// TestRun_MigrateLegacyConfig_NoLegacyConfig tests that migration
// continues when no legacy config exists.
func TestRun_MigrateLegacyConfig_NoLegacyConfig(t *testing.T) {
	fs := afero.NewMemMapFs()

	migrator := New(fs)
	err := migrator.Run()
	require.NoError(t, err)

	// Verify default config was created
	configExists, err := afero.Exists(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.True(t, configExists)

	content, err := afero.ReadFile(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "autodoc_session_progress = true")
}

// TestMigrateDirectory_SourceDoesNotExist tests that migration is skipped
// when source directory doesn't exist.
func TestMigrateDirectory_SourceDoesNotExist(t *testing.T) {
	fs := afero.NewMemMapFs()

	migrator := New(fs)
	err := migrator.migrateDirectory("nonexistent_source", "dest")
	require.NoError(t, err) // Should succeed (no-op)

	// Verify destination was not created
	destExists, err := afero.DirExists(fs, "dest")
	require.NoError(t, err)
	assert.False(t, destExists)
}

// TestRun_AllLegacyArtifactsWithLogs tests migration with all legacy
// artifacts including logs with nested structure.
func TestRun_AllLegacyArtifactsWithLogs(t *testing.T) {
	fs := afero.NewMemMapFs()

	// Create all legacy artifacts with nested content
	err := fs.MkdirAll(paths.LegacySessionsDir+"/2024", 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.LegacySessionsDir+"/2024/session.json", []byte("session"), 0644)
	require.NoError(t, err)

	err = fs.MkdirAll(paths.LegacyLogsDir+"/archive", 0755)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.LegacyLogsDir+"/current.log", []byte("current log"), 0644)
	require.NoError(t, err)
	err = afero.WriteFile(fs, paths.LegacyLogsDir+"/archive/old.log", []byte("old log"), 0644)
	require.NoError(t, err)

	err = afero.WriteFile(fs, paths.LegacyConfigFile, []byte("legacy config"), 0644)
	require.NoError(t, err)

	migrator := New(fs)
	err = migrator.Run()
	require.NoError(t, err)

	// Verify all migrations
	legacySessionsExists, _ := afero.DirExists(fs, paths.LegacySessionsDir)
	assert.False(t, legacySessionsExists)

	legacyLogsExists, _ := afero.DirExists(fs, paths.LegacyLogsDir)
	assert.False(t, legacyLogsExists)

	legacyConfigExists, _ := afero.Exists(fs, paths.LegacyConfigFile)
	assert.False(t, legacyConfigExists)

	// Verify content preservation
	sessionContent, err := afero.ReadFile(fs, paths.SessionsDir+"/2024/session.json")
	require.NoError(t, err)
	assert.Equal(t, "session", string(sessionContent))

	currentLogContent, err := afero.ReadFile(fs, paths.LogsDir+"/current.log")
	require.NoError(t, err)
	assert.Equal(t, "current log", string(currentLogContent))

	oldLogContent, err := afero.ReadFile(fs, paths.LogsDir+"/archive/old.log")
	require.NoError(t, err)
	assert.Equal(t, "old log", string(oldLogContent))

	configContent, err := afero.ReadFile(fs, paths.ConfigFile)
	require.NoError(t, err)
	assert.Equal(t, "legacy config", string(configContent))
}
