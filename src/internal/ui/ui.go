// Package ui provides terminal UI components for Claudex session management.
// It uses the Bubble Tea framework to provide interactive session selection,
// profile selection, and other UI workflows.
package ui

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"claudex/internal/services/session"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chzyer/readline"
)

// InputReader abstracts input reading for testability.
type InputReader interface {
	Readline() (string, error)
	Close() error
}

// ReadlineReader wraps readline.Instance to implement InputReader.
type ReadlineReader struct {
	instance *readline.Instance
}

// NewReadlineReader creates a new readline-based input reader with the given prompt.
func NewReadlineReader(prompt string) (InputReader, error) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:            prompt,
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		return nil, err
	}
	return &ReadlineReader{instance: rl}, nil
}

// Readline reads a line of input and returns it trimmed.
func (r *ReadlineReader) Readline() (string, error) {
	line, err := r.instance.Readline()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// Close closes the readline instance.
func (r *ReadlineReader) Close() error {
	return r.instance.Close()
}

// Styles
var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Bold(true).
			Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF87")).
				Bold(true).
				PaddingLeft(2)

	normalItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	dimmedItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			PaddingLeft(4)
)

// SessionItem is now defined in internal/services/session package
type SessionItem = session.SessionItem

type Model struct {
	List        list.Model
	SessionName string
	SessionPath string
	ProjectDir  string
	SessionsDir string
	Stage       string
	Quitting    bool
	Choice      string
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.List.SetSize(msg.Width-h, msg.Height-v)

	case SessionChoiceMsg:
		m.SessionName = msg.SessionName
		m.SessionPath = msg.SessionPath
		m.Choice = msg.ItemType
		return m, tea.Quit

	case ProfileChoiceMsg:
		m.Choice = msg.ProfileName
		return m, tea.Quit

	case ResumeOrForkChoiceMsg:
		m.Choice = msg.Choice
		return m, tea.Quit

	case ResumeSubmenuChoiceMsg:
		m.Choice = msg.Choice
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.Quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.List.SelectedItem().(SessionItem)
			if ok {
				m.Choice = i.Title
				switch m.Stage {
				case "session":
					return m, m.handleSessionChoice(i)
				case "profile":
					return m, m.handleProfileChoice(i)
				case "resume_or_fork":
					return m, m.handleResumeOrForkChoice(i)
				case "resume_submenu":
					return m, m.handleResumeSubmenuChoice(i)
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

type SessionChoiceMsg struct {
	SessionName string
	SessionPath string
	ItemType    string
}

func (m Model) handleSessionChoice(item SessionItem) tea.Cmd {
	return func() tea.Msg {
		if item.ItemType == "new" {
			// Return message to quit and handle outside TUI
			return SessionChoiceMsg{ItemType: "new"}
		}

		var sessionName, sessionPath string

		switch item.ItemType {
		case "ephemeral":
			sessionName = "ephemeral"
			sessionPath = ""

		case "session":
			sessionName = item.Title
			sessionPath = filepath.Join(m.SessionsDir, item.Title)
		}

		return SessionChoiceMsg{
			SessionName: sessionName,
			SessionPath: sessionPath,
			ItemType:    item.ItemType,
		}
	}
}

type ProfileChoiceMsg struct {
	ProfileName string
}

func (m Model) handleProfileChoice(item SessionItem) tea.Cmd {
	return func() tea.Msg {
		return ProfileChoiceMsg{ProfileName: item.Title}
	}
}

type ResumeOrForkChoiceMsg struct {
	Choice string // "resume" or "fork"
}

type ResumeSubmenuChoiceMsg struct {
	Choice string // "continue" or "fresh"
}

func (m Model) handleResumeOrForkChoice(item SessionItem) tea.Cmd {
	return func() tea.Msg {
		return ResumeOrForkChoiceMsg{Choice: item.ItemType}
	}
}

func (m Model) handleResumeSubmenuChoice(item SessionItem) tea.Cmd {
	return func() tea.Msg {
		return ResumeSubmenuChoiceMsg{Choice: item.ItemType}
	}
}

func (m Model) View() string {
	if m.Quitting {
		return "\n  👋 Goodbye!\n\n"
	}

	return docStyle.Render(m.List.View())
}

// Custom delegate for better item rendering
type ItemDelegate struct{}

func (d ItemDelegate) Height() int                             { return 2 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(SessionItem)
	if !ok {
		return
	}

	var icon string
	switch i.ItemType {
	case "new":
		icon = "➕"
	case "ephemeral":
		icon = "⚡"
	case "session":
		icon = "📁"
	case "profile":
		icon = "🎭"
	case "continue":
		icon = "▶"
	case "fresh":
		icon = "🔄"
	}

	str := fmt.Sprintf("%s %s", icon, i.Title)
	if i.Description != "" {
		str = fmt.Sprintf("%s\n   %s", str, dimmedItemStyle.Render(i.Description))
	}

	if index == m.Index() {
		fmt.Fprint(w, selectedItemStyle.Render("▶ "+str))
	} else {
		fmt.Fprint(w, normalItemStyle.Render(str))
	}
}

// Exported styles for external use
func TitleStyle() lipgloss.Style {
	return titleStyle
}

// UI Functions for Session Flow
// These functions handle pure UI concerns - rendering prompts, collecting input, displaying results

// PromptDescriptionWithReader shows a prompt screen and collects user input using the provided reader.
// Parameters: title (e.g., "Create New Session" or "Fork Session"), originalSession (optional, for fork context), reader (InputReader interface)
// Returns: description string, error
func PromptDescriptionWithReader(title string, originalSession string, reader InputReader) (string, error) {
	defer reader.Close()

	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println()
	fmt.Printf("\033[1;36m %s \033[0m\n", title)
	if originalSession != "" {
		fmt.Printf("  Original: %s\n", originalSession)
	}
	fmt.Println()

	description, err := reader.Readline()
	if err != nil {
		return "", err
	}

	description = strings.TrimSpace(description)
	if description == "" {
		return "", fmt.Errorf("description cannot be empty")
	}

	return description, nil
}

// PromptDescription shows a prompt screen and collects user input.
// Parameters: title (e.g., "Create New Session" or "Fork Session"), originalSession (optional, for fork context)
// Returns: description string, error
func PromptDescription(title string, originalSession string) (string, error) {
	promptText := "  Description: "
	if originalSession != "" {
		promptText = "  Description for fork: "
	}

	reader, err := NewReadlineReader(promptText)
	if err != nil {
		return "", err
	}

	return PromptDescriptionWithReader(title, originalSession, reader)
}

// ShowGenerating displays "Generating session name..." message
func ShowGenerating() {
	fmt.Println()
	fmt.Println("\033[90m  Generating session name...\033[0m")
}

// ShowSessionCreated displays success message for new session
// Parameters: sessionName
func ShowSessionCreated(sessionName string) {
	fmt.Println()
	fmt.Printf("\033[1;32m  Created: %s\033[0m\n", sessionName)
	fmt.Println()
}

// ShowSessionForked displays success message for forked session
// Parameters: originalName, newName
func ShowSessionForked(originalName, newName string) {
	fmt.Printf("\n\033[1;32m✅ Forked session: %s → %s\033[0m\n", originalName, newName)
}

// ShowFreshMemory displays success message for fresh memory
// Parameters: originalName, newName
func ShowFreshMemory(originalName, newName string) {
	fmt.Printf("\n\033[1;32m🔄 Fresh memory: %s → %s (original deleted)\033[0m\n", originalName, newName)
}
