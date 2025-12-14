// Package setuphook provides the usecase for prompting users about git hook
// integration and managing their preferences for automatic documentation updates.
package setuphook

import (
	"time"

	"claudex/internal/services/commander"
	"claudex/internal/services/hooksetup"
	"claudex/internal/services/preferences"
	"github.com/spf13/afero"
)

// Result represents the outcome of the setup hook check
type Result int

const (
	ResultNotGitRepo       Result = iota // Not a git repository
	ResultAlreadyInstalled               // Hook already installed
	ResultUserDeclined                   // User previously declined
	ResultPromptUser                     // Should prompt the user
)

// UseCase orchestrates hook setup detection and preference checking
type UseCase struct {
	hookSvc hooksetup.Service
	prefSvc preferences.Service
}

// New creates a new SetupHook usecase
func New(fs afero.Fs, projectDir string, cmdr commander.Commander) *UseCase {
	return &UseCase{
		hookSvc: hooksetup.New(fs, projectDir, cmdr),
		prefSvc: preferences.New(fs, projectDir),
	}
}

// ShouldPrompt checks if we should prompt the user about hook setup
func (uc *UseCase) ShouldPrompt() Result {
	// Check if git repo
	if !uc.hookSvc.IsGitRepo() {
		return ResultNotGitRepo
	}

	// Check if already installed
	if uc.hookSvc.IsInstalled() {
		return ResultAlreadyInstalled
	}

	// Check if user declined
	prefs, err := uc.prefSvc.Load()
	if err == nil && prefs.HookSetupDeclined {
		return ResultUserDeclined
	}

	return ResultPromptUser
}

// Install installs the hook
func (uc *UseCase) Install() error {
	return uc.hookSvc.Install()
}

// SaveDeclined saves the user's "never ask again" preference
func (uc *UseCase) SaveDeclined() error {
	prefs, _ := uc.prefSvc.Load() // Ignore error, start fresh if needed
	prefs.HookSetupDeclined = true
	prefs.DeclinedAt = time.Now().Format(time.RFC3339)
	return uc.prefSvc.Save(prefs)
}
