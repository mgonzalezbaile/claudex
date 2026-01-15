package mcpconfig

import (
	"encoding/json"
	"testing"

	"github.com/spf13/afero"
)

func TestIsConfigured(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    *ClaudeConfig
		expectedResult bool
	}{
		{
			name:           "no config file",
			setupConfig:    nil,
			expectedResult: false,
		},
		{
			name: "empty mcpServers",
			setupConfig: &ClaudeConfig{
				MCPServers: map[string]MCPServer{},
			},
			expectedResult: false,
		},
		{
			name: "only sequential-thinking",
			setupConfig: &ClaudeConfig{
				MCPServers: map[string]MCPServer{
					"sequential-thinking": GetSequentialThinkingMCP(),
				},
			},
			expectedResult: false,
		},
		{
			name: "only context7",
			setupConfig: &ClaudeConfig{
				MCPServers: map[string]MCPServer{
					"context7": GetContext7MCP(""),
				},
			},
			expectedResult: false,
		},
		{
			name: "both configured",
			setupConfig: &ClaudeConfig{
				MCPServers: map[string]MCPServer{
					"sequential-thinking": GetSequentialThinkingMCP(),
					"context7":            GetContext7MCP("test-token"),
				},
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			svc := New(fs)

			// Setup config file if provided
			if tt.setupConfig != nil {
				configPath, _ := svc.GetConfigPath()
				data, _ := json.Marshal(tt.setupConfig)
				afero.WriteFile(fs, configPath, data, 0644)
			}

			result, err := svc.IsConfigured()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expectedResult {
				t.Errorf("expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestConfigure(t *testing.T) {
	tests := []struct {
		name          string
		existingMCPs  map[string]MCPServer
		context7Token string
		expectToken   bool
	}{
		{
			name:          "configure from scratch without token",
			existingMCPs:  nil,
			context7Token: "",
			expectToken:   false,
		},
		{
			name:          "configure from scratch with token",
			existingMCPs:  nil,
			context7Token: "my-api-token",
			expectToken:   true,
		},
		{
			name: "preserve existing MCPs",
			existingMCPs: map[string]MCPServer{
				"custom-mcp": {
					Command: "node",
					Args:    []string{"custom.js"},
				},
			},
			context7Token: "",
			expectToken:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			svc := New(fs)

			// Setup existing config if provided
			if tt.existingMCPs != nil {
				configPath, _ := svc.GetConfigPath()
				config := &ClaudeConfig{MCPServers: tt.existingMCPs}
				data, _ := json.Marshal(config)
				afero.WriteFile(fs, configPath, data, 0644)
			}

			// Configure
			err := svc.Configure(tt.context7Token)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify configuration
			configPath, _ := svc.GetConfigPath()
			data, err := afero.ReadFile(fs, configPath)
			if err != nil {
				t.Fatalf("failed to read config: %v", err)
			}

			var config ClaudeConfig
			if err := json.Unmarshal(data, &config); err != nil {
				t.Fatalf("failed to unmarshal config: %v", err)
			}

			// Check sequential-thinking
			seqMCP, ok := config.MCPServers["sequential-thinking"]
			if !ok {
				t.Error("sequential-thinking not configured")
			}
			if seqMCP.Command != "npx" {
				t.Errorf("expected npx command, got %s", seqMCP.Command)
			}

			// Check context7
			ctx7MCP, ok := config.MCPServers["context7"]
			if !ok {
				t.Error("context7 not configured")
			}
			if ctx7MCP.Command != "npx" {
				t.Errorf("expected npx command, got %s", ctx7MCP.Command)
			}

			// Check token presence
			hasToken := false
			for _, arg := range ctx7MCP.Args {
				if arg == "--api-key" {
					hasToken = true
					break
				}
			}
			if hasToken != tt.expectToken {
				t.Errorf("expected token=%v, got token=%v", tt.expectToken, hasToken)
			}

			// Check preserved MCPs
			if tt.existingMCPs != nil {
				for key := range tt.existingMCPs {
					if _, ok := config.MCPServers[key]; !ok {
						t.Errorf("existing MCP %s was not preserved", key)
					}
				}
			}
		})
	}
}

func TestGetContext7MCP(t *testing.T) {
	tests := []struct {
		name         string
		token        string
		expectToken  bool
		expectedArgs int
	}{
		{
			name:         "without token",
			token:        "",
			expectToken:  false,
			expectedArgs: 2, // -y, @upstash/context7-mcp@latest
		},
		{
			name:         "with token",
			token:        "my-token",
			expectToken:  true,
			expectedArgs: 4, // -y, @upstash/context7-mcp@latest, --api-key, token
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mcp := GetContext7MCP(tt.token)

			if mcp.Command != "npx" {
				t.Errorf("expected npx, got %s", mcp.Command)
			}

			if len(mcp.Args) != tt.expectedArgs {
				t.Errorf("expected %d args, got %d", tt.expectedArgs, len(mcp.Args))
			}

			hasToken := false
			for i, arg := range mcp.Args {
				if arg == "--api-key" && i+1 < len(mcp.Args) {
					hasToken = true
					if mcp.Args[i+1] != tt.token {
						t.Errorf("expected token %s, got %s", tt.token, mcp.Args[i+1])
					}
				}
			}

			if hasToken != tt.expectToken {
				t.Errorf("expected token presence=%v, got=%v", tt.expectToken, hasToken)
			}
		})
	}
}
