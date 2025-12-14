// Package globalprefs provides services for managing global user preferences.
// It persists preferences to ~/.config/claudex/mcp-preferences.json.
package globalprefs

// MCPPreferences holds global MCP setup preferences
type MCPPreferences struct {
	// MCPSetupDeclined indicates whether user declined MCP setup
	MCPSetupDeclined bool `json:"mcpSetupDeclined,omitempty"`

	// DeclinedAt is the RFC3339 timestamp when MCP setup was declined
	DeclinedAt string `json:"declinedAt,omitempty"`
}

// Service abstracts global preferences persistence for testability
type Service interface {
	// Load reads preferences from global storage
	// Returns zero-value MCPPreferences if file doesn't exist
	Load() (MCPPreferences, error)

	// Save persists preferences to global storage atomically
	Save(prefs MCPPreferences) error
}
