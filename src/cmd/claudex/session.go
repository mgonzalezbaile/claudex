package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"claudex/internal/services/session"
	"claudex/internal/ui"
	newuc "claudex/internal/usecases/session/new"
	forkuc "claudex/internal/usecases/session/resume/fork"
	freshuc "claudex/internal/usecases/session/resume/fresh"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// showSessionSelector displays the session selection UI and returns the user's choice
func (a *App) showSessionSelector() (*ui.Model, error) {
	// Get sessions
	sessions, err := session.GetSessions(a.deps.FS, a.sessionsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	// Build items
	items := []list.Item{
		session.SessionItem{Title: "Create New Session", Description: "Start a fresh working session", ItemType: "new"},
		session.SessionItem{Title: "Ephemeral", Description: "Work without saving session data", ItemType: "ephemeral"},
	}

	for _, s := range sessions {
		items = append(items, s)
	}

	// Create list
	delegate := ui.ItemDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.Title = "Claudex Session Manager"
	l.Styles.Title = ui.TitleStyle()
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Additional keybindings
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("q"),
				key.WithHelp("q", "quit"),
			),
		}
	}

	m := ui.Model{
		List:        l,
		Stage:       "session",
		ProjectDir:  a.projectDir,
		SessionsDir: a.sessionsDir,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run session selector: %w", err)
	}

	fm := finalModel.(ui.Model)
	return &fm, nil
}

// handleNewSession processes the "Create New Session" choice
func (a *App) handleNewSession() (SessionInfo, error) {
	// Prompt for description
	description, err := a.promptNewSessionDescription()
	if err != nil {
		return SessionInfo{}, err
	}

	fmt.Println()
	fmt.Println("\033[90m  Generating session name...\033[0m")

	// Create the session using new usecase
	newSessionUC := newuc.New(a.deps.FS, a.deps.Cmd, a.deps.UUID, a.deps.Clock, a.sessionsDir)
	sessionName, sessionPath, claudeSessionID, err := newSessionUC.Execute(description)
	if err != nil {
		return SessionInfo{}, fmt.Errorf("failed to create new session: %w", err)
	}

	fmt.Println()
	fmt.Printf("\033[1;32m  Created: %s\033[0m\n", sessionName)
	fmt.Println()

	return SessionInfo{
		Name:     sessionName,
		Path:     sessionPath,
		ClaudeID: claudeSessionID,
		Mode:     LaunchModeNew,
	}, nil
}

// promptNewSessionDescription prompts user for new session description
func (a *App) promptNewSessionDescription() (string, error) {
	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println()
	fmt.Println("\033[1;36m Create New Session \033[0m")
	fmt.Println()
	fmt.Print("  Description: ")

	reader := bufio.NewReader(os.Stdin)
	description, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	description = strings.TrimSpace(description)

	if description == "" {
		return "", fmt.Errorf("description cannot be empty")
	}

	return description, nil
}

// handleResumeOrFork processes resume/fork/fresh choices for existing sessions
func (a *App) handleResumeOrFork(fm *ui.Model) (SessionInfo, error) {
	// Show resume/fork menu
	resumeOrForkItems := []list.Item{
		session.SessionItem{Title: "Resume Session", Description: "Continue with existing context", ItemType: "resume"},
		session.SessionItem{Title: "Fork Session", Description: "Start fresh with copied files", ItemType: "fork"},
	}

	delegate := ui.ItemDelegate{}
	rfList := list.New(resumeOrForkItems, delegate, 0, 0)
	rfList.Title = fmt.Sprintf("Resume or Fork • Session: %s", fm.SessionName)
	rfList.Styles.Title = ui.TitleStyle()
	rfList.SetShowStatusBar(false)
	rfList.SetFilteringEnabled(false)
	rfList.SetShowHelp(true)

	rfModel := ui.Model{
		List:        rfList,
		Stage:       "resume_or_fork",
		SessionName: fm.SessionName,
		SessionPath: fm.SessionPath,
		ProjectDir:  a.projectDir,
		SessionsDir: a.sessionsDir,
	}

	rfProgram := tea.NewProgram(rfModel, tea.WithAltScreen())
	finalRfModel, err := rfProgram.Run()
	if err != nil {
		return SessionInfo{}, fmt.Errorf("failed to run resume/fork menu: %w", err)
	}

	rfm := finalRfModel.(ui.Model)
	if rfm.Quitting {
		return SessionInfo{}, fmt.Errorf("user quit")
	}

	resumeOrForkChoice := rfm.Choice

	// If user chose "Resume Session", show submenu: Continue vs Fresh Memory
	if resumeOrForkChoice == "resume" {
		submenuChoice, err := a.showResumeSubmenu(fm.SessionName, fm.SessionPath)
		if err != nil {
			return SessionInfo{}, err
		}

		// Handle "Fresh Memory" choice using fresh usecase
		if submenuChoice == "fresh" {
			freshUC := freshuc.New(a.deps.FS, a.deps.UUID, a.sessionsDir)
			newSessionName, newSessionPath, newClaudeSessionID, err := freshUC.Execute(fm.SessionName)
			if err != nil {
				return SessionInfo{}, fmt.Errorf("failed to create fresh session: %w", err)
			}
			fmt.Printf("\n🔄 Fresh memory: %s → %s (original deleted)\n", fm.SessionName, newSessionName)

			return SessionInfo{
				Name:         newSessionName,
				Path:         newSessionPath,
				ClaudeID:     newClaudeSessionID,
				Mode:         LaunchModeFresh,
				OriginalName: fm.SessionName,
			}, nil
		}
		// else: submenuChoice == "continue" -> proceed with existing resume logic
		claudeSessionID := session.ExtractClaudeSessionID(fm.SessionName)
		if claudeSessionID == "" {
			return SessionInfo{}, fmt.Errorf("could not extract session ID for resume")
		}

		return SessionInfo{
			Name:     fm.SessionName,
			Path:     fm.SessionPath,
			ClaudeID: claudeSessionID,
			Mode:     LaunchModeResume,
		}, nil
	}

	// Handle fork choice
	if resumeOrForkChoice == "fork" {
		forkDescription, err := a.promptForkDescription(fm.SessionName)
		if err != nil {
			return SessionInfo{}, err
		}

		// Use fork usecase with description
		forkUC := forkuc.New(a.deps.FS, a.deps.Cmd, a.deps.UUID, a.sessionsDir)
		newSessionName, newSessionPath, newClaudeSessionID, err := forkUC.Execute(fm.SessionName, forkDescription)
		if err != nil {
			return SessionInfo{}, fmt.Errorf("failed to fork session: %w", err)
		}
		fmt.Printf("\n✅ Forked session: %s → %s\n", fm.SessionName, newSessionName)

		return SessionInfo{
			Name:         newSessionName,
			Path:         newSessionPath,
			ClaudeID:     newClaudeSessionID,
			Mode:         LaunchModeFork,
			OriginalName: fm.SessionName,
		}, nil
	}

	return SessionInfo{}, fmt.Errorf("unknown resume/fork choice: %s", resumeOrForkChoice)
}

// showResumeSubmenu shows the Continue vs Fresh Memory submenu
func (a *App) showResumeSubmenu(sessionName, sessionPath string) (string, error) {
	resumeSubmenuItems := []list.Item{
		session.SessionItem{Title: "Continue with context", Description: "Resume with full conversation history", ItemType: "continue"},
		session.SessionItem{Title: "Fresh memory", Description: "Start fresh, keep files, delete original", ItemType: "fresh"},
	}

	delegate := ui.ItemDelegate{}
	rsMenu := list.New(resumeSubmenuItems, delegate, 0, 0)
	rsMenu.Title = fmt.Sprintf("Resume Options • Session: %s", sessionName)
	rsMenu.Styles.Title = ui.TitleStyle()
	rsMenu.SetShowStatusBar(false)
	rsMenu.SetFilteringEnabled(false)
	rsMenu.SetShowHelp(true)

	rsModel := ui.Model{
		List:        rsMenu,
		Stage:       "resume_submenu",
		SessionName: sessionName,
		SessionPath: sessionPath,
		ProjectDir:  a.projectDir,
		SessionsDir: a.sessionsDir,
	}

	rsProgram := tea.NewProgram(rsModel, tea.WithAltScreen())
	finalRsModel, err := rsProgram.Run()
	if err != nil {
		return "", fmt.Errorf("failed to run resume submenu: %w", err)
	}

	rsm := finalRsModel.(ui.Model)
	if rsm.Quitting {
		return "", fmt.Errorf("user quit")
	}

	return rsm.Choice, nil
}

// promptForkDescription prompts the user for a fork description
func (a *App) promptForkDescription(original string) (string, error) {
	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println()
	fmt.Println("\033[1;36m Fork Session \033[0m")
	fmt.Printf("  Original: %s\n", original)
	fmt.Println()

	fmt.Print("  Description for fork: ")
	reader := bufio.NewReader(os.Stdin)
	forkDescription, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading description: %w", err)
	}
	forkDescription = strings.TrimSpace(forkDescription)

	if forkDescription == "" {
		return "", fmt.Errorf("description cannot be empty")
	}

	return forkDescription, nil
}
