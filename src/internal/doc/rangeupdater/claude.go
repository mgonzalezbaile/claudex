package rangeupdater

import (
	"fmt"
	"log"
	"os/exec"

	"claudex/internal/services/commander"
	"claudex/internal/services/env"
)

// InvokeClaudeForIndex invokes Claude to regenerate an index.md file.
// Claude uses its Edit tool to update the file directly.
// The recursion guard (CLAUDE_HOOK_INTERNAL=1) prevents infinite loops.
func InvokeClaudeForIndex(cmdr commander.Commander, env env.Environment, indexPath, listing, modifiedFiles string) error {
	// Recursion guard: check if we're already inside a hook invocation
	if env.Get("CLAUDE_HOOK_INTERNAL") == "1" {
		log.Printf("Skipping index update for %s: recursion guard triggered", indexPath)
		return nil
	}

	log.Printf("Spawning background process to regenerate %s", indexPath)

	// Build Claude prompt with context
	prompt := buildPrompt(indexPath, listing, modifiedFiles)

	// Create a detached background process using bash
	// This ensures the process survives even after the calling process exits
	// Claude will use its Edit tool to update the file directly
	// Using --model haiku for cost efficiency (index updates are simple tasks)
	bashScript := fmt.Sprintf(`
export CLAUDE_HOOK_INTERNAL=1
claude -p %q --model haiku 2>/dev/null
`, prompt)

	cmd := exec.Command("bash", "-c", bashScript)

	// Detach the process so it survives after we exit
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start background Claude process for %s: %v", indexPath, err)
		return fmt.Errorf("failed to start background Claude process: %w", err)
	}

	log.Printf("Background process started (PID: %d) for %s", cmd.Process.Pid, indexPath)
	return nil
}

// buildPrompt constructs the Claude prompt for index.md regeneration
func buildPrompt(indexPath, listing, modifiedFiles string) string {
	return fmt.Sprintf(`Update the index.md file at %s.

MODIFIED FILES:
%s

FILES IN DIRECTORY:
%s

INSTRUCTIONS:
1. Read the existing index.md to understand the current structure and style
2. Update it to reflect all files in the directory
3. Use minimal pointer style: brief one-line descriptions
4. Group files logically if patterns exist
5. Keep descriptions concise (one line per file)
6. Use the Edit tool to update the file directly`, indexPath, modifiedFiles, listing)
}
