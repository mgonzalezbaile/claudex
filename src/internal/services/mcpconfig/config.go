package mcpconfig

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// Service abstracts MCP configuration operations for testability
type Service interface {
	// IsConfigured checks if recommended MCPs are already configured
	IsConfigured() (bool, error)

	// Configure adds recommended MCPs to ~/.claude.json
	// context7Token is optional - pass empty string for rate-limited mode
	Configure(context7Token string) error

	// GetConfigPath returns the path to Claude Code's config file
	GetConfigPath() (string, error)
}

// FileService is the production implementation of Service
type FileService struct {
	fs afero.Fs
}

// New creates a new Service instance
func New(fs afero.Fs) Service {
	return &FileService{
		fs: fs,
	}
}

// GetConfigPath returns ~/.claude.json path
func (s *FileService) GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude.json"), nil
}

// IsConfigured checks if both sequential-thinking and context7 are already configured
func (s *FileService) IsConfigured() (bool, error) {
	config, err := s.loadOrCreate()
	if err != nil {
		return false, err
	}

	if config.MCPServers == nil {
		return false, nil
	}

	_, hasSequential := config.MCPServers["sequential-thinking"]
	_, hasContext7 := config.MCPServers["context7"]

	return hasSequential && hasContext7, nil
}

// Configure adds recommended MCPs, preserving existing entries
// context7Token is optional - pass empty string for rate-limited mode
func (s *FileService) Configure(context7Token string) error {
	config, err := s.loadOrCreate()
	if err != nil {
		return err
	}

	if config.MCPServers == nil {
		config.MCPServers = make(map[string]MCPServer)
	}

	// Add sequential-thinking if not exists
	if _, exists := config.MCPServers["sequential-thinking"]; !exists {
		config.MCPServers["sequential-thinking"] = GetSequentialThinkingMCP()
	}

	// Add context7 if not exists (with optional token)
	if _, exists := config.MCPServers["context7"]; !exists {
		config.MCPServers["context7"] = GetContext7MCP(context7Token)
	}

	return s.save(config)
}

// loadOrCreate loads existing config or returns empty config if file doesn't exist
func (s *FileService) loadOrCreate() (*ClaudeConfig, error) {
	configPath, err := s.GetConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := afero.ReadFile(s.fs, configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &ClaudeConfig{}, nil
		}
		return nil, err
	}

	var config ClaudeConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// save persists config to ~/.claude.json atomically
func (s *FileService) save(config *ClaudeConfig) error {
	configPath, err := s.GetConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	tempPath := configPath + ".tmp"

	// Ensure parent directory exists
	if err := s.fs.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file first
	if err := afero.WriteFile(s.fs, tempPath, data, 0644); err != nil {
		return err
	}

	// Atomic rename
	return s.fs.Rename(tempPath, configPath)
}
