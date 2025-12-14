package globalprefs

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

const (
	preferencesFileName = "mcp-preferences.json"
	configDir           = ".config/claudex"
)

// FileService is the production implementation of Service
type FileService struct {
	fs afero.Fs
}

// New creates a new Service instance
func New(fs afero.Fs) Service {
	return &FileService{
		fs: fs,
	}
}

// getPrefsPath returns the path to global preferences file
func (fs *FileService) getPrefsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDir, preferencesFileName), nil
}

// Load reads preferences from global storage
// Returns zero-value MCPPreferences if file doesn't exist
func (fs *FileService) Load() (MCPPreferences, error) {
	prefsPath, err := fs.getPrefsPath()
	if err != nil {
		return MCPPreferences{}, err
	}

	data, err := afero.ReadFile(fs.fs, prefsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return zero value if file doesn't exist
			return MCPPreferences{}, nil
		}
		return MCPPreferences{}, err
	}

	var prefs MCPPreferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return MCPPreferences{}, err
	}

	return prefs, nil
}

// Save persists preferences to global storage atomically
func (fs *FileService) Save(prefs MCPPreferences) error {
	prefsPath, err := fs.getPrefsPath()
	if err != nil {
		return err
	}

	prefsDir := filepath.Dir(prefsPath)
	tempPath := prefsPath + ".tmp"

	// Ensure ~/.config/claudex directory exists
	if err := fs.fs.MkdirAll(prefsDir, 0755); err != nil {
		return err
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first
	if err := afero.WriteFile(fs.fs, tempPath, data, 0644); err != nil {
		return err
	}

	// Atomic rename
	return fs.fs.Rename(tempPath, prefsPath)
}
