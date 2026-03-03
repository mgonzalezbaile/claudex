package setup

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"claudex"

	"github.com/spf13/afero"
)

// AssembleEngineerAgent creates a principal-engineer-{stack} agent from role + skill templates.
// It reads the engineer.md role template and the stack-specific skill file (e.g., typescript.md)
// from the embedded profiles, combines them with frontmatter, and writes to both agents/ and
// commands/agents/ directories.
//
// Parameters:
//   - afs: Filesystem abstraction for writing files
//   - stack: Stack identifier (e.g., "typescript", "go", "python", "php")
//   - agentsDir: Target directory for agent profiles (.claude/agents)
//   - commandsAgentsDir: Target directory for command agents (.claude/commands/agents)
//   - noOverwrite: If true, existing files will not be overwritten
//
// Returns an error if assembly fails.
func AssembleEngineerAgent(afs afero.Fs, stack, agentsDir, commandsAgentsDir string, noOverwrite bool) error {
	// Read role template from embedded FS
	roleContent, err := fs.ReadFile(claudex.Profiles, "profiles/roles/engineer.md")
	if err != nil {
		return fmt.Errorf("failed to read embedded role file: %w", err)
	}

	// Capitalize stack name for display
	stackDisplay := formatStackName(stack)

	// Generate frontmatter
	frontmatter := fmt.Sprintf(`---
name: principal-engineer-%s
description: Use this agent when you need a Principal %s Engineer for code implementation, debugging, refactoring, and development best practices. This agent executes stories by reading execution plans and implementing tasks sequentially with comprehensive testing and documentation lookup.
model: sonnet
color: blue
permissionMode: bypassPermissions
---

`, stack, stackDisplay)

	// Replace {Stack} placeholder in role content
	roleStr := strings.ReplaceAll(string(roleContent), "{Stack}", stackDisplay)

	// Read skill content from embedded FS if it exists
	var skillStr string
	skillPath := filepath.Join("profiles/skills", stack+".md")
	if skillContent, err := fs.ReadFile(claudex.Profiles, skillPath); err == nil {
		skillStr = "\n" + string(skillContent)
	}

	// Read common design principles skill (always loaded for engineers)
	var commonSkillStr string
	commonSkillPath := "profiles/skills/software-design-principles.md"
	if commonContent, err := fs.ReadFile(claudex.Profiles, commonSkillPath); err == nil {
		commonSkillStr = "\n" + string(commonContent)
	}

	// Combine all parts: frontmatter + role + stack-skill + common-skill
	agentContent := frontmatter + roleStr + skillStr + commonSkillStr

	// Write to agents/ directory
	agentPath := filepath.Join(agentsDir, fmt.Sprintf("principal-engineer-%s.md", stack))
	if err := writeAgentFile(afs, agentPath, []byte(agentContent), noOverwrite); err != nil {
		return fmt.Errorf("failed to write agent file: %w", err)
	}

	// Copy to commands/agents/
	commandPath := filepath.Join(commandsAgentsDir, fmt.Sprintf("principal-engineer-%s.md", stack))
	if err := writeAgentFile(afs, commandPath, []byte(agentContent), noOverwrite); err != nil {
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
	case "php":
		return "PHP"
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
