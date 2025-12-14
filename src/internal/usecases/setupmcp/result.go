package setupmcp

// Result represents the outcome of the MCP setup check
type Result int

const (
	ResultNodeMissing       Result = iota // Node.js/npx not available
	ResultAlreadyConfigured               // MCPs already configured
	ResultUserDeclined                    // User previously declined
	ResultPromptUser                      // Should prompt the user
)
