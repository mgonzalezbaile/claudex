package setup

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// AssembleEngineerAgent creates a principal-engineer-{stack} agent from role + skill templates.
// It reads the engineer.md role template and the stack-specific skill file (e.g., typescript.md),
// combines them with frontmatter, and writes to both agents/ and commands/agents/ directories.
//
// Parameters:
//   - fs: Filesystem abstraction for reading/writing files
//   - stack: Stack identifier (e.g., "typescript", "go", "python")
//   - agentsDir: Target directory for agent profiles (.claude/agents)
//   - commandsAgentsDir: Target directory for command agents (.claude/commands/agents)
//   - rolesDir: Source directory for role templates
//   - skillsDir: Source directory for skill templates
//   - noOverwrite: If true, existing files will not be overwritten
//
// Returns an error if assembly fails.
func AssembleEngineerAgent(fs afero.Fs, stack, agentsDir, commandsAgentsDir, rolesDir, skillsDir string, noOverwrite bool) error {
	roleFile := filepath.Join(rolesDir, "engineer.md")
	skillFile := filepath.Join(skillsDir, stack+".md")

	// Read role template
	roleContent, err := afero.ReadFile(fs, roleFile)
	if err != nil {
		return fmt.Errorf("failed to read role file: %w", err)
	}

	// Capitalize stack name for display
	stackDisplay := formatStackName(stack)

	// Generate frontmatter
	frontmatter := fmt.Sprintf(`---
name: principal-engineer-%s
description: Use this agent when you need a Principal %s Engineer for code implementation, debugging, refactoring, and development best practices. This agent executes stories by reading execution plans and implementing tasks sequentially with comprehensive testing and documentation lookup.
model: sonnet
color: blue
---

`, stack, stackDisplay)

	// Replace {Stack} placeholder in role content
	roleStr := strings.ReplaceAll(string(roleContent), "{Stack}", stackDisplay)

	// Read skill content if it exists
	var skillStr string
	if skillContent, err := afero.ReadFile(fs, skillFile); err == nil {
		skillStr = "\n" + string(skillContent)
	}

	// Combine all parts
	agentContent := frontmatter + roleStr + skillStr

	// Write to agents/ directory
	agentPath := filepath.Join(agentsDir, fmt.Sprintf("principal-engineer-%s.md", stack))
	if err := writeAgentFile(fs, agentPath, []byte(agentContent), noOverwrite); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	// Copy to commands/agents/
	commandPath := filepath.Join(commandsAgentsDir, fmt.Sprintf("principal-engineer-%s.md", stack))
	if err := writeAgentFile(fs, commandPath, []byte(agentContent), noOverwrite); err != nil {
		return fmt.Errorf("failed to write command file: %w", err)
	}

	return nil
}

// formatStackName returns the properly capitalized display name for a stack
func formatStackName(stack string) string {
	switch stack {
	case "react-native":
		return "React Native"
	case "typescript":
		return "TypeScript"
	case "go":
		return "Go"
	case "python":
		return "Python"
	default:
		return strings.Title(stack)
	}
}

// writeAgentFile writes an agent file, respecting the noOverwrite flag
func writeAgentFile(fs afero.Fs, path string, content []byte, noOverwrite bool) error {
	if noOverwrite {
		if _, err := fs.Stat(path); err == nil {
			// File exists, skip writing
			return nil
		}
	}
	return afero.WriteFile(fs, path, content, 0644)
}
