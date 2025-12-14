package mcpconfig

import (
	"os/exec"
)

// IsNodeAvailable checks if Node.js/npx is available on the system
// Both MCPs require npx to run, so we check for its availability
func IsNodeAvailable() bool {
	_, err := exec.LookPath("npx")
	return err == nil
}
