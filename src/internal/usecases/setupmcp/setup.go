// Package setupmcp provides the usecase for prompting users about MCP
// configuration and managing their preferences for MCP setup.
package setupmcp

import (
	"time"

	"claudex/internal/services/globalprefs"
	"claudex/internal/services/mcpconfig"
	"github.com/spf13/afero"
)

// UseCase orchestrates MCP setup detection and preference checking
type UseCase struct {
	mcpSvc   mcpconfig.Service
	prefsSvc globalprefs.Service
}

// New creates a new SetupMCP usecase
func New(fs afero.Fs) *UseCase {
	return &UseCase{
		mcpSvc:   mcpconfig.New(fs),
		prefsSvc: globalprefs.New(fs),
	}
}

// ShouldPrompt checks if we should prompt the user about MCP setup
func (uc *UseCase) ShouldPrompt() Result {
	// Check if Node.js/npx is available
	if !mcpconfig.IsNodeAvailable() {
		return ResultNodeMissing
	}

	// Check if already configured
	configured, err := uc.mcpSvc.IsConfigured()
	if err == nil && configured {
		return ResultAlreadyConfigured
	}

	// Check if user declined
	prefs, err := uc.prefsSvc.Load()
	if err == nil && prefs.MCPSetupDeclined {
		return ResultUserDeclined
	}

	return ResultPromptUser
}

// Install configures the recommended MCPs with optional Context7 API token
func (uc *UseCase) Install(context7Token string) error {
	return uc.mcpSvc.Configure(context7Token)
}

// SaveDeclined saves the user's "never ask again" preference
func (uc *UseCase) SaveDeclined() error {
	prefs, _ := uc.prefsSvc.Load() // Ignore error, start fresh if needed
	prefs.MCPSetupDeclined = true
	prefs.DeclinedAt = time.Now().Format(time.RFC3339)
	return uc.prefsSvc.Save(prefs)
}
