package setupmcp

import (
	"encoding/json"
	"os/exec"
	"testing"

	"claudex/internal/services/globalprefs"
	"claudex/internal/services/mcpconfig"
	"github.com/spf13/afero"
)

func TestShouldPrompt(t *testing.T) {
	// Check if npx is available for this test
	_, npxErr := exec.LookPath("npx")
	hasNpx := npxErr == nil

	tests := []struct {
		name           string
		setupConfig    *mcpconfig.ClaudeConfig
		setupPrefs     *globalprefs.MCPPreferences
		expectedResult Result
		skipIfNoNpx    bool
	}{
		{
			name:           "should prompt when nothing configured",
			setupConfig:    nil,
			setupPrefs:     nil,
			expectedResult: ResultPromptUser,
			skipIfNoNpx:    true,
		},
		{
			name: "already configured",
			setupConfig: &mcpconfig.ClaudeConfig{
				MCPServers: map[string]mcpconfig.MCPServer{
					"sequential-thinking": mcpconfig.GetSequentialThinkingMCP(),
					"context7":            mcpconfig.GetContext7MCP(""),
				},
			},
			setupPrefs:     nil,
			expectedResult: ResultAlreadyConfigured,
			skipIfNoNpx:    true,
		},
		{
			name:        "user previously declined",
			setupConfig: nil,
			setupPrefs: &globalprefs.MCPPreferences{
				MCPSetupDeclined: true,
				DeclinedAt:       "2024-01-01T00:00:00Z",
			},
			expectedResult: ResultUserDeclined,
			skipIfNoNpx:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip tests that need npx if it's not available
			if tt.skipIfNoNpx && !hasNpx && tt.expectedResult != ResultUserDeclined {
				t.Skip("npx not available, skipping test")
			}

			fs := afero.NewMemMapFs()
			uc := New(fs)

			// Setup config if provided
			if tt.setupConfig != nil {
				mcpSvc := mcpconfig.New(fs)
				configPath, _ := mcpSvc.GetConfigPath()
				data, _ := json.Marshal(tt.setupConfig)
				afero.WriteFile(fs, configPath, data, 0644)
			}

			// Setup preferences if provided
			if tt.setupPrefs != nil {
				prefsSvc := globalprefs.New(fs)
				prefsSvc.Save(*tt.setupPrefs)
			}

			result := uc.ShouldPrompt()

			// If npx is not available, we expect ResultNodeMissing
			if !hasNpx && tt.skipIfNoNpx {
				if result != ResultNodeMissing {
					t.Errorf("expected ResultNodeMissing when npx unavailable, got %v", result)
				}
				return
			}

			if result != tt.expectedResult {
				t.Errorf("expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestInstall(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectToken bool
	}{
		{
			name:        "install without token",
			token:       "",
			expectToken: false,
		},
		{
			name:        "install with token",
			token:       "my-api-token",
			expectToken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			uc := New(fs)

			err := uc.Install(tt.token)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify configuration was created
			mcpSvc := mcpconfig.New(fs)
			configured, err := mcpSvc.IsConfigured()
			if err != nil {
				t.Fatalf("failed to check configuration: %v", err)
			}
			if !configured {
				t.Error("MCPs should be configured after install")
			}

			// Verify token if provided
			if tt.expectToken {
				configPath, _ := mcpSvc.GetConfigPath()
				data, _ := afero.ReadFile(fs, configPath)
				var config mcpconfig.ClaudeConfig
				json.Unmarshal(data, &config)

				ctx7MCP := config.MCPServers["context7"]
				hasToken := false
				for _, arg := range ctx7MCP.Args {
					if arg == "--api-key" {
						hasToken = true
						break
					}
				}
				if !hasToken {
					t.Error("expected token in context7 config")
				}
			}
		})
	}
}

func TestSaveDeclined(t *testing.T) {
	fs := afero.NewMemMapFs()
	uc := New(fs)

	err := uc.SaveDeclined()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify preference was saved
	prefsSvc := globalprefs.New(fs)
	prefs, err := prefsSvc.Load()
	if err != nil {
		t.Fatalf("failed to load preferences: %v", err)
	}

	if !prefs.MCPSetupDeclined {
		t.Error("expected MCPSetupDeclined to be true")
	}

	if prefs.DeclinedAt == "" {
		t.Error("expected DeclinedAt to be set")
	}
}
