package main

import (
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

type sessionItem struct {
	title       string
	description string
	created     time.Time
	itemType    string // "new", "ephemeral", "session"
}

func (i sessionItem) Title() string       { return i.title }
func (i sessionItem) Description() string { return i.description }
func (i sessionItem) FilterValue() string { return i.title }

type model struct {
	list        list.Model
	sessionName string
	sessionPath string
	projectDir  string
	sessionsDir string
	stage       string
	quitting    bool
	choice      string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case sessionChoiceMsg:
		m.sessionName = msg.sessionName
		m.sessionPath = msg.sessionPath
		m.choice = msg.itemType
		return m, tea.Quit

	case profileChoiceMsg:
		m.choice = msg.profileName
		return m, tea.Quit

	case resumeOrForkChoiceMsg:
		m.choice = msg.choice
		return m, tea.Quit

	case resumeSubmenuChoiceMsg:
		m.choice = msg.choice
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(sessionItem)
			if ok {
				m.choice = i.title
				if m.stage == "session" {
					return m, m.handleSessionChoice(i)
				} else if m.stage == "profile" {
					return m, m.handleProfileChoice(i)
				} else if m.stage == "resume_or_fork" {
					return m, m.handleResumeOrForkChoice(i)
				} else if m.stage == "resume_submenu" {
					return m, m.handleResumeSubmenuChoice(i)
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

type sessionChoiceMsg struct {
	sessionName string
	sessionPath string
	itemType    string
}

func (m model) handleSessionChoice(item sessionItem) tea.Cmd {
	return func() tea.Msg {
		if item.itemType == "new" {
			// Return message to quit and handle outside TUI
			return sessionChoiceMsg{itemType: "new"}
		}

		var sessionName, sessionPath string

		switch item.itemType {
		case "ephemeral":
			sessionName = "ephemeral"
			sessionPath = ""

		case "session":
			sessionName = item.title
			sessionPath = filepath.Join(m.sessionsDir, item.title)
		}

		return sessionChoiceMsg{
			sessionName: sessionName,
			sessionPath: sessionPath,
			itemType:    item.itemType,
		}
	}
}

type profileChoiceMsg struct {
	profileName string
}

func (m model) handleProfileChoice(item sessionItem) tea.Cmd {
	return func() tea.Msg {
		return profileChoiceMsg{profileName: item.title}
	}
}

type resumeOrForkChoiceMsg struct {
	choice string // "resume" or "fork"
}

type resumeSubmenuChoiceMsg struct {
	choice string // "continue" or "fresh"
}

func (m model) handleResumeOrForkChoice(item sessionItem) tea.Cmd {
	return func() tea.Msg {
		return resumeOrForkChoiceMsg{choice: item.itemType}
	}
}

func (m model) handleResumeSubmenuChoice(item sessionItem) tea.Cmd {
	return func() tea.Msg {
		return resumeSubmenuChoiceMsg{choice: item.itemType}
	}
}

func (m model) View() string {
	if m.quitting {
		return "\n  👋 Goodbye!\n\n"
	}

	return docStyle.Render(m.list.View())
}

// Custom delegate for better item rendering
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(sessionItem)
	if !ok {
		return
	}

	var icon string
	switch i.itemType {
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

	str := fmt.Sprintf("%s %s", icon, i.title)
	if i.description != "" {
		str = fmt.Sprintf("%s\n   %s", str, dimmedItemStyle.Render(i.description))
	}

	if index == m.Index() {
		fmt.Fprint(w, selectedItemStyle.Render("▶ "+str))
	} else {
		fmt.Fprint(w, normalItemStyle.Render(str))
	}
}
