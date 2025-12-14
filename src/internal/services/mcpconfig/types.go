// Package mcpconfig provides services for managing MCP server configurations
// in Claude Code's ~/.claude.json file.
package mcpconfig

// MCPServer represents a single MCP server configuration
type MCPServer struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// ClaudeConfig represents the ~/.claude.json structure
type ClaudeConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers,omitempty"`
	// Future fields can be added here and will be preserved during merge
}

// Context7TokenURL is the URL where users can generate their API token
const Context7TokenURL = "https://context7.com/dashboard"

// GetSequentialThinkingMCP returns the sequential-thinking MCP config
func GetSequentialThinkingMCP() MCPServer {
	return MCPServer{
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-sequential-thinking"},
	}
}

// GetContext7MCP returns the context7 MCP config with optional API token
// If apiToken is empty, context7 runs in rate-limited mode (60 requests/hour)
func GetContext7MCP(apiToken string) MCPServer {
	args := []string{"-y", "@upstash/context7-mcp@latest"}
	if apiToken != "" {
		args = append(args, "--api-key", apiToken)
	}
	return MCPServer{
		Command: "npx",
		Args:    args,
	}
}
