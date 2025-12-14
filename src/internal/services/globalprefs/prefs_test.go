package globalprefs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/spf13/afero"
)

func TestLoadPreferences(t *testing.T) {
	tests := []struct {
		name         string
		setupPrefs   *MCPPreferences
		expectZero   bool
	}{
		{
			name:       "no preferences file",
			setupPrefs: nil,
			expectZero: true,
		},
		{
			name: "existing preferences",
			setupPrefs: &MCPPreferences{
				MCPSetupDeclined: true,
				DeclinedAt:       "2024-01-01T00:00:00Z",
			},
			expectZero: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			svc := New(fs)

			// Setup preferences file if provided
			if tt.setupPrefs != nil {
				prefsPath, _ := svc.(*FileService).getPrefsPath()
				data, _ := json.Marshal(tt.setupPrefs)
				fs.MkdirAll(configDir, 0755)
				afero.WriteFile(fs, prefsPath, data, 0644)
			}

			prefs, err := svc.Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectZero {
				if prefs.MCPSetupDeclined {
					t.Error("expected zero value, got declined=true")
				}
			} else {
				if prefs.MCPSetupDeclined != tt.setupPrefs.MCPSetupDeclined {
					t.Errorf("expected declined=%v, got %v",
						tt.setupPrefs.MCPSetupDeclined, prefs.MCPSetupDeclined)
				}
				if prefs.DeclinedAt != tt.setupPrefs.DeclinedAt {
					t.Errorf("expected declinedAt=%s, got %s",
						tt.setupPrefs.DeclinedAt, prefs.DeclinedAt)
				}
			}
		})
	}
}

func TestSavePreferences(t *testing.T) {
	fs := afero.NewMemMapFs()
	svc := New(fs)

	prefs := MCPPreferences{
		MCPSetupDeclined: true,
		DeclinedAt:       time.Now().Format(time.RFC3339),
	}

	err := svc.Save(prefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify saved preferences
	loaded, err := svc.Load()
	if err != nil {
		t.Fatalf("failed to load saved preferences: %v", err)
	}

	if loaded.MCPSetupDeclined != prefs.MCPSetupDeclined {
		t.Errorf("expected declined=%v, got %v",
			prefs.MCPSetupDeclined, loaded.MCPSetupDeclined)
	}

	if loaded.DeclinedAt != prefs.DeclinedAt {
		t.Errorf("expected declinedAt=%s, got %s",
			prefs.DeclinedAt, loaded.DeclinedAt)
	}
}

func TestSavePreferencesCreatesDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	svc := New(fs)

	// Verify directory doesn't exist
	prefsPath, _ := svc.(*FileService).getPrefsPath()
	_, err := fs.Stat(prefsPath)
	if err == nil {
		t.Fatal("preferences file should not exist yet")
	}

	// Save should create directory
	prefs := MCPPreferences{
		MCPSetupDeclined: true,
		DeclinedAt:       time.Now().Format(time.RFC3339),
	}

	err = svc.Save(prefs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	_, err = fs.Stat(prefsPath)
	if err != nil {
		t.Errorf("preferences file should exist: %v", err)
	}
}
