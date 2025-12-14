// Package setup provides the setup usecase for initializing .claude directory
// structure with hooks, agents, commands, and project-specific configuration.
// It orchestrates services for filesystem operations, environment access,
// stack detection, and profile generation.
package setup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"claudex"
	"claudex/internal/services/env"
	"claudex/internal/services/filesystem"
	"claudex/internal/services/settings"
	"claudex/internal/services/stackdetect"

	"github.com/spf13/afero"
)

// SetupUseCase orchestrates the .claude directory setup workflow
type SetupUseCase struct {
	fs  afero.Fs
	env env.Environment
}

// New creates a new SetupUseCase instance with the given dependencies
func New(fs afero.Fs, environment env.Environment) *SetupUseCase {
	return &SetupUseCase{
		fs:  fs,
		env: environment,
	}
}

// Execute runs the complete .claude directory setup workflow.
// It creates the directory structure, copies hooks and agents, detects
// project stacks, generates engineer profiles, and creates settings.local.json.
//
// Parameters:
//   - projectDir: The project directory where .claude should be created
//   - noOverwrite: If true, existing files will not be overwritten
//
// Returns an error if setup fails.
func (uc *SetupUseCase) Execute(projectDir string, noOverwrite bool) error {
	claudeDir := filepath.Join(projectDir, ".claude")

	// Handle existing .claude directory with user choice
	proceed, err := HandleExistingClaudeDirectory(projectDir, claudeDir)
	if err != nil {
		return err
	}
	if !proceed {
		return fmt.Errorf("installation cancelled by user")
	}

	// Get config dir (~/.config/claudex) for optional hooks
	configDir := uc.env.Get("XDG_CONFIG_HOME")
	if configDir == "" {
		home := uc.env.Get("HOME")
		if home == "" {
			return fmt.Errorf("HOME environment variable not set")
		}
		configDir = filepath.Join(home, ".config")
	}
	claudexConfigDir := filepath.Join(configDir, "claudex")

	// Create .claude directory structure
	hooksDir := filepath.Join(claudeDir, "hooks")
	agentsDir := filepath.Join(claudeDir, "agents")
	commandsAgentsDir := filepath.Join(claudeDir, "commands", "agents")

	if err := uc.fs.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("failed to create hooks directory: %w", err)
	}
	if err := uc.fs.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}
	if err := uc.fs.MkdirAll(commandsAgentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create commands/agents directory: %w", err)
	}

	// Copy hooks from ~/.config/claudex/hooks/
	sourceHooksDir := filepath.Join(claudexConfigDir, "hooks")
	if _, err := uc.fs.Stat(sourceHooksDir); err == nil {
		if err := filesystem.CopyDir(uc.fs, sourceHooksDir, hooksDir, noOverwrite); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to copy hooks: %v\n", err)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Warning: Hooks directory not found at %s\n", sourceHooksDir)
	}

	// Copy agent profiles to both agents/ and commands/agents/ from embedded FS
	if err := uc.copyAgentProfiles(agentsDir, commandsAgentsDir, noOverwrite); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to copy agent profiles: %v\n", err)
	}

	// Generate settings.local.json
	if err := uc.generateSettings(claudeDir, noOverwrite); err != nil {
		return err
	}

	// Detect project stack and generate principal-engineer agents
	stacks := stackdetect.Detect(uc.fs, projectDir)
	if len(stacks) == 0 {
		// Default to all stacks if none detected
		stacks = []string{"typescript", "python", "go"}
	}

	// Generate principal-engineer-{stack} agents from embedded profiles
	for _, stack := range stacks {
		if err := AssembleEngineerAgent(uc.fs, stack, agentsDir, commandsAgentsDir, noOverwrite); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to assemble principal-engineer-%s: %v\n", stack, err)
		}
	}

	// Create principal-engineer alias by copying the primary stack's agent
	if err := uc.createEngineerAlias(stacks, agentsDir, commandsAgentsDir, noOverwrite); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create principal-engineer alias: %v\n", err)
	}

	fmt.Printf("✓ Created .claude directory with %d engineer profile(s)\n", len(stacks))
	return nil
}

// HandleExistingClaudeDirectory checks if .claude exists and handles user choice
func HandleExistingClaudeDirectory(projectDir, claudeDir string) (proceed bool, err error) {
	// Silent merge: always proceed with setup
	return true, nil
}

// copyAgentProfiles copies agent profiles from embedded FS to both agents/ and commands/agents/
func (uc *SetupUseCase) copyAgentProfiles(agentsDir, commandsAgentsDir string, noOverwrite bool) error {
	// Read agent profiles from embedded FS
	entries, err := fs.ReadDir(claudex.Profiles, "profiles/agents")
	if err != nil {
		return fmt.Errorf("could not read embedded agents directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			// Read from embedded FS
			content, err := fs.ReadFile(claudex.Profiles, filepath.Join("profiles/agents", entry.Name()))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to read embedded agent %s: %v\n", entry.Name(), err)
				continue
			}

			// Copy to agents/
			agentTarget := filepath.Join(agentsDir, entry.Name()+".md")
			if err := uc.writeFileIfNeeded(agentTarget, content, noOverwrite); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to copy to agents/%s: %v\n", entry.Name(), err)
			}

			// Copy to commands/agents/
			commandTarget := filepath.Join(commandsAgentsDir, entry.Name()+".md")
			if err := uc.writeFileIfNeeded(commandTarget, content, noOverwrite); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Failed to copy to commands/agents/%s: %v\n", entry.Name(), err)
			}
		}
	}

	return nil
}

// writeFileIfNeeded writes a file only if it doesn't exist (when noOverwrite is true)
func (uc *SetupUseCase) writeFileIfNeeded(path string, content []byte, noOverwrite bool) error {
	if noOverwrite {
		if _, err := uc.fs.Stat(path); err == nil {
			// File exists, skip writing
			return nil
		}
	}
	return afero.WriteFile(uc.fs, path, content, 0644)
}

// generateSettings creates the settings.local.json file with hooks configuration
// using the embedded template. If a settings file already exists, it merges
// missing hooks while preserving user customizations.
func (uc *SetupUseCase) generateSettings(claudeDir string, _ bool) error {
	settingsPath := filepath.Join(claudeDir, "settings.local.json")

	// Check if file exists
	existingContent, err := afero.ReadFile(uc.fs, settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read existing settings: %w", err)
	}

	var finalContent []byte
	if len(existingContent) > 0 {
		// Merge: add missing hooks, preserve user config
		finalContent, err = settings.MergeSettings(claudex.SettingsTemplate, existingContent)
		if err != nil {
			return fmt.Errorf("failed to merge settings: %w", err)
		}
	} else {
		// No existing file, use template directly
		finalContent = claudex.SettingsTemplate
	}

	return afero.WriteFile(uc.fs, settingsPath, finalContent, 0644)
}

// createEngineerAlias creates a principal-engineer.md alias from the primary stack
func (uc *SetupUseCase) createEngineerAlias(stacks []string, agentsDir, commandsAgentsDir string, noOverwrite bool) error {
	if len(stacks) == 0 {
		return nil
	}

	primaryStack := stacks[0]
	aliasSource := filepath.Join(agentsDir, fmt.Sprintf("principal-engineer-%s.md", primaryStack))

	// Read the primary engineer content
	aliasContent, err := afero.ReadFile(uc.fs, aliasSource)
	if err != nil {
		return fmt.Errorf("failed to read source agent: %w", err)
	}

	// Copy to agents/principal-engineer.md
	aliasAgentTarget := filepath.Join(agentsDir, "principal-engineer.md")
	if err := uc.writeFileIfNeeded(aliasAgentTarget, aliasContent, noOverwrite); err != nil {
		return fmt.Errorf("failed to create principal-engineer alias: %w", err)
	}

	// Copy to commands/agents/principal-engineer.md
	aliasCommandTarget := filepath.Join(commandsAgentsDir, "principal-engineer.md")
	if err := uc.writeFileIfNeeded(aliasCommandTarget, aliasContent, noOverwrite); err != nil {
		return fmt.Errorf("failed to create principal-engineer command alias: %w", err)
	}

	return nil
}
