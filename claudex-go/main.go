package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

//go:embed profiles
var profilesFS embed.FS

// buildSystemPromptWithSessionContext injects session context into the system prompt
// to ensure all agents follow session folder documentation rules.
func buildSystemPromptWithSessionContext(profileContent []byte, sessionPath string) (string, error) {
	// Skip injection for ephemeral sessions (empty sessionPath)
	if sessionPath == "" {
		return string(profileContent), nil
	}

	// List files in session directory (excluding hidden files starting with '.')
	entries, err := os.ReadDir(sessionPath)
	if err != nil {
		return "", fmt.Errorf("failed to read session directory: %w", err)
	}

	// Build file listing
	var files []string
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files (starting with '.')
		if !strings.HasPrefix(name, ".") {
			files = append(files, name)
		}
	}

	var filesDisplay string
	if len(files) == 0 {
		filesDisplay = "- (No files yet - you'll be the first to create documentation!)"
	} else {
		// Format as bullet list
		for _, f := range files {
			filesDisplay += fmt.Sprintf("- %s\n", f)
		}
		filesDisplay = strings.TrimSuffix(filesDisplay, "\n")
	}

	// Build session context template (from pre-tool-use.sh hook)
	sessionContext := fmt.Sprintf(`## SESSION CONTEXT (CRITICAL)

You are working within an active Claudex session. ALL documentation, plans, and artifacts MUST be created in the session folder.

**Session Folder (Absolute Path)**: `+"`%s`"+`

### MANDATORY RULES for Documentation:
1. ✅ ALWAYS save documentation to the session folder above
2. ✅ Use absolute paths when creating files (Write/Edit tools)
3. ✅ Before exploring the codebase, check the session folder for existing context
4. ❌ NEVER save documentation to project root or arbitrary locations
5. ❌ NEVER use relative paths for documentation files

### Session Folder Contents:
%s

### Recommended File Names:
- Research documents: `+"`research-{topic}.md`"+`
- Execution plans: `+"`execution-plan-{feature}.md`"+`
- Analysis reports: `+"`analysis-{component}.md`"+`
- Technical specs: `+"`technical-spec-{feature}.md`"+`

---

`, sessionPath, filesDisplay)

	// Concatenate session context with profile content
	combinedPrompt := sessionContext + "\n" + string(profileContent)
	return combinedPrompt, nil
}

func main() {
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Setup centralized logging
	logsDir := filepath.Join(projectDir, "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create logs directory: %v\n", err)
	}

	// Create unique log file for this execution
	timestamp := time.Now().Format("20060102-150405")
	logFileName := fmt.Sprintf("claudex-%s.log", timestamp)
	logFilePath := filepath.Join(logsDir, logFileName)

	// Open log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not open log file: %v\n", err)
	} else {
		defer logFile.Close()
		// Configure Go logger with [claudex] prefix
		log.SetOutput(logFile)
		log.SetPrefix("[claudex] ")
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

		// Set environment variable for hooks
		os.Setenv("CLAUDEX_LOG_FILE", logFilePath)

		log.Printf("Claudex started (log file: %s)", logFileName)
	}

	sessionsDir := filepath.Join(projectDir, "sessions")

	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Get sessions
	sessions, err := getSessions(sessionsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Build items
	items := []list.Item{
		sessionItem{title: "Create New Session", description: "Start a fresh working session", itemType: "new"},
		sessionItem{title: "Ephemeral", description: "Work without saving session data", itemType: "ephemeral"},
	}

	for _, s := range sessions {
		items = append(items, s)
	}

	// Create list
	delegate := itemDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.Title = "Claudex Session Manager"
	l.Styles.Title = titleStyle
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

	m := model{
		list:        l,
		stage:       "session",
		projectDir:  projectDir,
		sessionsDir: sessionsDir,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fm := finalModel.(model)
	if fm.quitting {
		return
	}

	// Handle "Create New Session" - select profile first
	var profileContent []byte
	if fm.choice == "new" {
		// First, select a profile
		profiles, err := getProfiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		delegate := itemDelegate{}
		profileItems := make([]list.Item, len(profiles))
		for i, profile := range profiles {
			fullPath := resolveProfilePath(profile)
			desc := extractProfileDescription(fullPath)
			profileItems[i] = sessionItem{
				title:       profile,
				description: desc,
				itemType:    "profile",
			}
		}

		pl := list.New(profileItems, delegate, 0, 0)
		pl.Title = "Select Profile for New Session"
		pl.Styles.Title = titleStyle
		pl.SetShowStatusBar(false)
		pl.SetFilteringEnabled(true)
		pl.SetShowHelp(true)

		pm := model{
			list:        pl,
			stage:       "profile",
			projectDir:  projectDir,
			sessionsDir: sessionsDir,
		}

		p2 := tea.NewProgram(pm, tea.WithAltScreen())
		finalProfileModel, err := p2.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		pm2 := finalProfileModel.(model)
		if pm2.quitting {
			return
		}

		profileName := pm2.choice
		// profilePath := filepath.Join(profilesDir, profileName) // No longer used directly

		profileContent, err = loadProfile(profileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Now create the session with the selected profile
		sessionName, sessionPath, claudeSessionID, err := createNewSessionParallel(sessionsDir, profileContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fm.sessionName = sessionName
		fm.sessionPath = sessionPath
		fm.choice = claudeSessionID // Store session ID for later use
	}

	// Check if selected session has a Claude session ID (for resume/fork choice)
	var resumeOrForkChoice string
	var isFreshMemory bool // Track if "fresh memory" was chosen
	if fm.choice == "session" && hasClaudeSessionID(fm.sessionName) {
		// Show resume/fork menu
		resumeOrForkItems := []list.Item{
			sessionItem{title: "Resume Session", description: "Continue with existing context", itemType: "resume"},
			sessionItem{title: "Fork Session", description: "Start fresh with copied files", itemType: "fork"},
		}

		delegate := itemDelegate{}
		rfList := list.New(resumeOrForkItems, delegate, 0, 0)
		rfList.Title = fmt.Sprintf("Resume or Fork • Session: %s", fm.sessionName)
		rfList.Styles.Title = titleStyle
		rfList.SetShowStatusBar(false)
		rfList.SetFilteringEnabled(false)
		rfList.SetShowHelp(true)

		rfModel := model{
			list:        rfList,
			stage:       "resume_or_fork",
			sessionName: fm.sessionName,
			sessionPath: fm.sessionPath,
			projectDir:  projectDir,
			sessionsDir: sessionsDir,
		}

		rfProgram := tea.NewProgram(rfModel, tea.WithAltScreen())
		finalRfModel, err := rfProgram.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		rfm := finalRfModel.(model)
		if rfm.quitting {
			return
		}

		resumeOrForkChoice = rfm.choice

		// Add variable to track resume submenu choice
		var resumeSubmenuChoice string

		// If user chose "Resume Session", show submenu: Continue vs Fresh Memory
		if resumeOrForkChoice == "resume" {
			// Show resume submenu: Continue with context vs Fresh memory
			resumeSubmenuItems := []list.Item{
				sessionItem{title: "Continue with context", description: "Resume with full conversation history", itemType: "continue"},
				sessionItem{title: "Fresh memory", description: "Start fresh, keep files, delete original", itemType: "fresh"},
			}

			delegate := itemDelegate{}
			rsMenu := list.New(resumeSubmenuItems, delegate, 0, 0)
			rsMenu.Title = fmt.Sprintf("Resume Options • Session: %s", fm.sessionName)
			rsMenu.Styles.Title = titleStyle
			rsMenu.SetShowStatusBar(false)
			rsMenu.SetFilteringEnabled(false)
			rsMenu.SetShowHelp(true)

			rsModel := model{
				list:        rsMenu,
				stage:       "resume_submenu",
				sessionName: fm.sessionName,
				sessionPath: fm.sessionPath,
				projectDir:  projectDir,
				sessionsDir: sessionsDir,
			}

			rsProgram := tea.NewProgram(rsModel, tea.WithAltScreen())
			finalRsModel, err := rsProgram.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			rsm := finalRsModel.(model)
			if rsm.quitting {
				return
			}

			resumeSubmenuChoice = rsm.choice

			// Handle "Fresh Memory" choice
			if resumeSubmenuChoice == "fresh" {
				newSessionName, newSessionPath, newClaudeSessionID, err := freshMemorySession(sessionsDir, fm.sessionName)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating fresh session: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("\n🔄 Fresh memory: %s → %s (original deleted)\n", fm.sessionName, newSessionName)
				fm.sessionName = newSessionName
				fm.sessionPath = newSessionPath
				fm.choice = newClaudeSessionID
				isFreshMemory = true        // Track that this is a fresh memory session
				resumeOrForkChoice = "fork" // Reuse fork launch path (--session-id)
			}
			// else: resumeSubmenuChoice == "continue" -> proceed with existing resume logic
		}

		// Handle fork choice (but not for fresh memory - already processed above)
		if resumeOrForkChoice == "fork" && !isFreshMemory {
			// Prompt for new description (similar to createNewSessionParallel)
			fmt.Print("\033[H\033[2J") // Clear screen
			fmt.Println()
			fmt.Println("\033[1;36m Fork Session \033[0m")
			fmt.Printf("  Original: %s\n", fm.sessionName)
			fmt.Println()

			fmt.Print("  Description for fork: ")
			reader := bufio.NewReader(os.Stdin)
			forkDescription, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading description: %v\n", err)
				os.Exit(1)
			}
			forkDescription = strings.TrimSpace(forkDescription)

			if forkDescription == "" {
				fmt.Fprintf(os.Stderr, "Error: description cannot be empty\n")
				os.Exit(1)
			}

			newSessionName, newSessionPath, newClaudeSessionID, err := forkSessionWithDescription(sessionsDir, fm.sessionName, forkDescription)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error forking session: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("\n✅ Forked session: %s → %s\n", fm.sessionName, newSessionName)
			fm.sessionName = newSessionName
			fm.sessionPath = newSessionPath
			fm.choice = newClaudeSessionID // Store the new session ID
		}
	}

	// Set environment
	os.Setenv("CLAUDEX_SESSION", fm.sessionName)
	os.Setenv("CLAUDEX_SESSION_PATH", fm.sessionPath)

	// Handle resume vs new/fork session
	var claudeSessionID string
	var isNewSessionAlreadyInitialized bool

	// Check if we just created a new session (session ID stored in fm.choice)
	if fm.choice != "new" && fm.choice != "session" && fm.choice != "ephemeral" && len(fm.choice) > 30 {
		// This is a Claude session ID from createNewSessionParallel
		claudeSessionID = fm.choice
		isNewSessionAlreadyInitialized = true
	}

	if isNewSessionAlreadyInitialized {
		// New session created, launch it with --session-id
		// Update last used timestamp
		if err := updateLastUsed(fm.sessionPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not update last used timestamp: %v\n", err)
		}

		// Give terminal a moment to settle
		time.Sleep(100 * time.Millisecond)

		// Clear screen and show launching message
		fmt.Print("\033[H\033[2J\033[3J") // Clear screen and scrollback
		fmt.Print("\033[0m")              // Reset all attributes
		fmt.Printf("\n✅ Launching new Claude session\n")
		fmt.Printf("📦 Session: %s\n", fm.sessionName)
		fmt.Printf("🔄 Session ID: %s\n\n", claudeSessionID)

		// Small delay before launching
		time.Sleep(300 * time.Millisecond)

		// Construct relative session path for activation command
		relativeSessionPath := filepath.Join("sessions", filepath.Base(fm.sessionPath))
		activationPrompt := fmt.Sprintf("/agents:team-lead-new activate in session %s", relativeSessionPath)

		// Launch the Claude session with activation command
		launchCmd := exec.Command("claude", "--session-id", claudeSessionID, activationPrompt)
		launchCmd.Stdin = os.Stdin
		launchCmd.Stdout = os.Stdout
		launchCmd.Stderr = os.Stderr
		launchCmd.Env = os.Environ() // Ensure environment variables are inherited

		if err := launchCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
			os.Exit(1)
		}
	} else if resumeOrForkChoice == "resume" || resumeOrForkChoice == "fork" {
		// For resume or fork, get the Claude session ID
		if resumeOrForkChoice == "fork" {
			// For fork, we already have the new session ID in fm.choice
			claudeSessionID = fm.choice
		} else {
			// For resume, extract from session name
			claudeSessionID = extractClaudeSessionID(fm.sessionName)
			if claudeSessionID == "" {
				fmt.Fprintf(os.Stderr, "\n❌ Could not extract session ID for resume\n")
				os.Exit(1)
			}
		}

		// Update last used timestamp
		if err := updateLastUsed(fm.sessionPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not update last used timestamp: %v\n", err)
		}

		// Give terminal a moment to settle
		time.Sleep(100 * time.Millisecond)

		// Clear screen and show launching message
		fmt.Print("\033[H\033[2J\033[3J") // Clear screen and scrollback
		fmt.Print("\033[0m")              // Reset all attributes

		if isFreshMemory {
			fmt.Printf("\n🔄 Launching fresh memory session\n")
		} else if resumeOrForkChoice == "fork" {
			fmt.Printf("\n✅ Launching forked session\n")
		} else {
			fmt.Printf("\n✅ Resuming Claude session\n")
		}
		fmt.Printf("📦 Session: %s\n", fm.sessionName)
		fmt.Printf("🔄 Session ID: %s\n\n", claudeSessionID)

		// Small delay before launching
		time.Sleep(300 * time.Millisecond)

		if resumeOrForkChoice == "fork" {
			// For fork, start a new session with activation command
			relativeSessionPath := filepath.Join("sessions", filepath.Base(fm.sessionPath))
			activationPrompt := fmt.Sprintf("/agents:team-lead-new activate in session %s", relativeSessionPath)

			launchCmd := exec.Command("claude", "--session-id", claudeSessionID, activationPrompt)
			launchCmd.Stdin = os.Stdin
			launchCmd.Stdout = os.Stdout
			launchCmd.Stderr = os.Stderr
			launchCmd.Env = os.Environ()

			if err := launchCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
				os.Exit(1)
			}
		} else {
			// For resume, continue existing session
			resumeCmd := exec.Command("claude", "--resume", claudeSessionID)
			resumeCmd.Stdin = os.Stdin
			resumeCmd.Stdout = os.Stdout
			resumeCmd.Stderr = os.Stderr
			resumeCmd.Env = os.Environ()

			if err := resumeCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		// For new/fork sessions, show profile selector
		profiles, err := getProfiles()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		profileItems := make([]list.Item, len(profiles))
		for i, profile := range profiles {
			fullPath := resolveProfilePath(profile)
			desc := extractProfileDescription(fullPath)
			profileItems[i] = sessionItem{
				title:       profile,
				description: desc,
				itemType:    "profile",
			}
		}

		pl := list.New(profileItems, delegate, 0, 0)
		pl.Title = fmt.Sprintf("Select Profile • Session: %s", fm.sessionName)
		pl.Styles.Title = titleStyle
		pl.SetShowStatusBar(false)
		pl.SetFilteringEnabled(true)
		pl.SetShowHelp(true)

		pm := model{
			list:        pl,
			stage:       "profile",
			sessionName: fm.sessionName,
			sessionPath: fm.sessionPath,
			projectDir:  projectDir,
			sessionsDir: sessionsDir,
		}

		p2 := tea.NewProgram(pm, tea.WithAltScreen())
		finalProfileModel, err := p2.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		pm2 := finalProfileModel.(model)
		if pm2.quitting {
			return
		}

		// Now launch Claude - terminal is properly restored
		profileName := pm2.choice
		// profilePath := filepath.Join(profilesDir, profileName) // Not used

		profileContent, err = loadProfile(profileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Give terminal a moment to settle
		time.Sleep(100 * time.Millisecond)

		// Clear screen and show launching message
		fmt.Print("\033[H\033[2J\033[3J") // Clear screen and scrollback
		fmt.Print("\033[0m")              // Reset all attributes
		fmt.Printf("\n✅ Launching Claude with %s\n", profileName)
		fmt.Printf("📦 Session: %s\n", fm.sessionName)

		// Generate new Claude session ID
		claudeSessionID = uuid.New().String()

		// Rename session directory to include Claude session ID
		newSessionPath, err := renameSessionWithClaudeID(fm.sessionPath, fm.sessionName, claudeSessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error renaming session directory: %v\n", err)
			os.Exit(1)
		}

		// Update environment variable with new path
		if newSessionPath != "" {
			os.Setenv("CLAUDEX_SESSION_PATH", newSessionPath)
			fmt.Printf("📁 Session directory: %s\n", filepath.Base(newSessionPath))
			fmt.Printf("🔄 Session ID: %s\n\n", claudeSessionID)

			// Update last used timestamp
			if err := updateLastUsed(newSessionPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not update last used timestamp: %v\n", err)
			}
		}

		// Small delay before launching
		time.Sleep(300 * time.Millisecond)

		// Construct relative session path for activation command
		relativeSessionPath := filepath.Join("sessions", filepath.Base(newSessionPath))
		activationPrompt := fmt.Sprintf("/agents:team-lead-new activate in session %s", relativeSessionPath)

		// Launch the Claude session with activation command
		launchCmd := exec.Command("claude", "--session-id", claudeSessionID, activationPrompt)
		launchCmd.Stdin = os.Stdin
		launchCmd.Stdout = os.Stdout
		launchCmd.Stderr = os.Stderr
		launchCmd.Env = os.Environ() // Ensure environment variables are inherited

		if err := launchCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
			os.Exit(1)
		}
	}
}

func createNewSessionParallel(sessionsDir string, profileContent []byte) (string, string, string, error) {
	fmt.Print("\033[H\033[2J") // Clear screen
	fmt.Println()
	fmt.Println("\033[1;36m Create New Session \033[0m")
	fmt.Println()

	// Generate UUID for the session upfront
	claudeSessionID := uuid.New().String()

	// Get description from user
	fmt.Print("  Description: ")
	reader := bufio.NewReader(os.Stdin)
	description, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", err
	}
	description = strings.TrimSpace(description)

	if description == "" {
		return "", "", "", fmt.Errorf("description cannot be empty")
	}

	fmt.Println()
	fmt.Println("\033[90m  🤖 Generating session name...\033[0m")

	sessionName, err := generateSessionName(description)
	if err != nil {
		sessionName = createManualSlug(description)
	}

	// Create final session name with Claude session ID
	baseSessionName := sessionName
	sessionName = fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)

	// Ensure unique (in case of collision)
	originalName := sessionName
	counter := 1
	sessionPath := filepath.Join(sessionsDir, sessionName)
	for {
		if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
			break
		}
		sessionName = fmt.Sprintf("%s-%d", originalName, counter)
		sessionPath = filepath.Join(sessionsDir, sessionName)
		counter++
	}

	if err := os.MkdirAll(sessionPath, 0755); err != nil {
		return "", "", "", err
	}

	if err := os.WriteFile(filepath.Join(sessionPath, ".description"), []byte(description), 0644); err != nil {
		return "", "", "", err
	}

	created := time.Now().UTC().Format(time.RFC3339)
	if err := os.WriteFile(filepath.Join(sessionPath, ".created"), []byte(created), 0644); err != nil {
		return "", "", "", err
	}

	fmt.Println()
	fmt.Println("\033[1;32m  ✓ Created: " + sessionName + "\033[0m")
	fmt.Println()
	time.Sleep(500 * time.Millisecond)

	return sessionName, sessionPath, claudeSessionID, nil
}

func generateSessionName(description string) (string, error) {
	prompt := fmt.Sprintf("Generate a short, descriptive slug (2-4 words max, lowercase, hyphen-separated) for a work session based on this description: '%s'. Reply with ONLY the slug, nothing else. Examples: 'auth-refactor', 'api-performance-fix', 'user-dashboard-ui'", description)

	cmd := exec.Command("claude", "-p")
	cmd.Stdin = strings.NewReader(prompt)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`[a-z0-9-]+`)
	matches := re.FindAllString(string(output), -1)

	if len(matches) == 0 {
		return "", fmt.Errorf("no valid slug")
	}

	sessionName := matches[0]
	if len(sessionName) < 3 {
		return "", fmt.Errorf("slug too short")
	}

	return sessionName, nil
}

func createManualSlug(description string) string {
	slug := strings.ToLower(description)
	slug = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if len(slug) > 50 {
		slug = slug[:50]
	}

	return slug
}

func getSessions(sessionsDir string) ([]sessionItem, error) {
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []sessionItem{}, nil
		}
		return nil, err
	}

	var sessions []sessionItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		var desc string
		var lastUsedTime time.Time
		var lastUsedStr string

		if data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name(), ".description")); err == nil {
			desc = strings.TrimSpace(string(data))
		}

		// Try to read last_used first, fall back to created
		if data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name(), ".last_used")); err == nil {
			lastUsedStr = strings.TrimSpace(string(data))
			if t, err := time.Parse(time.RFC3339, lastUsedStr); err == nil {
				lastUsedTime = t
				lastUsedStr = t.Format("2 Jan 2006 15:04:05")
			}
		} else if data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name(), ".created")); err == nil {
			// Fall back to created date if no last_used file
			lastUsedStr = strings.TrimSpace(string(data))
			if t, err := time.Parse(time.RFC3339, lastUsedStr); err == nil {
				lastUsedTime = t
				lastUsedStr = t.Format("2 Jan 2006 15:04:05")
			}
		}

		sessions = append(sessions, sessionItem{
			title:       entry.Name(),
			description: fmt.Sprintf("%s • %s", desc, lastUsedStr),
			created:     lastUsedTime,
			itemType:    "session",
		})
	}

	// Sort by last used date in descending order (most recently used first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].created.After(sessions[j].created)
	})

	return sessions, nil
}

func getProfiles() ([]string, error) {
	var profiles []string

	// Look for profiles in profiles/agents/ directory in embedded FS
	entries, err := fs.ReadDir(profilesFS, "profiles/agents")
	if err != nil {
		// If agents directory doesn't exist, return empty list
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		name := entry.Name()
		// Skip directories and hidden files
		if !entry.IsDir() && !strings.HasPrefix(name, ".") {
			profiles = append(profiles, name)
		}
	}

	sort.Strings(profiles)
	return profiles, nil
}

func extractProfileDescription(profilePath string) string {
	file, err := profilesFS.Open(profilePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`(?i)(role:|principal|agent)`)

	for scanner.Scan() {
		line := scanner.Text()
		if re.MatchString(line) {
			desc := strings.TrimLeft(line, "#*- ")
			desc = regexp.MustCompile(`(?i)role:`).ReplaceAllString(desc, "")
			desc = strings.TrimSpace(desc)
			if len(desc) > 60 {
				desc = desc[:60]
			}
			return desc
		}
	}

	return ""
}

func copyDir(src, dst string) error {
	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, 0644); err != nil {
				return err
			}
		}
	}

	return nil
}

func forkSession(sessionsDir, originalSessionName string) (string, string, string, error) {
	// Generate new UUID for the forked session
	claudeSessionID := uuid.New().String()

	// Strip the Claude session ID to get the base session name
	baseSessionName := stripClaudeSessionID(originalSessionName)

	// Also need to strip any existing fork counter (e.g., "my-task-2" -> "my-task")
	// Check if the last segment is a number
	lastHyphen := strings.LastIndex(baseSessionName, "-")
	if lastHyphen != -1 {
		potentialCounter := baseSessionName[lastHyphen+1:]
		// If it's just a number, strip it too
		if regexp.MustCompile(`^\d+$`).MatchString(potentialCounter) {
			baseSessionName = baseSessionName[:lastHyphen]
		}
	}

	// Create session name with new Claude session ID
	newSessionName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newSessionPath := filepath.Join(sessionsDir, newSessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(sessionsDir, originalSessionName)
	if err := copyDir(originalSessionPath, newSessionPath); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	return newSessionName, newSessionPath, claudeSessionID, nil
}

func freshMemorySession(sessionsDir, originalSessionName string) (string, string, string, error) {
	// Generate new UUID for the fresh session
	claudeSessionID := uuid.New().String()

	// Strip the Claude session ID to get the base session name
	baseSessionName := stripClaudeSessionID(originalSessionName)

	// Create session name with new Claude session ID (keep base slug)
	newSessionName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newSessionPath := filepath.Join(sessionsDir, newSessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(sessionsDir, originalSessionName)
	if err := copyDir(originalSessionPath, newSessionPath); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	// DELETE the original folder (key difference from fork)
	if err := os.RemoveAll(originalSessionPath); err != nil {
		return "", "", "", fmt.Errorf("failed to delete original session: %w", err)
	}

	return newSessionName, newSessionPath, claudeSessionID, nil
}

func forkSessionWithDescription(sessionsDir, originalSessionName, description string) (string, string, string, error) {
	// Generate new UUID for the forked session
	claudeSessionID := uuid.New().String()

	// Generate new session name from description (like new session creation)
	baseSessionName, err := generateSessionName(description)
	if err != nil {
		// Fallback to manual slug if Claude API fails
		baseSessionName = createManualSlug(description)
	}

	// Create session name with new Claude session ID
	newSessionName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newSessionPath := filepath.Join(sessionsDir, newSessionName)

	// Copy original session directory to new location
	originalSessionPath := filepath.Join(sessionsDir, originalSessionName)
	if err := copyDir(originalSessionPath, newSessionPath); err != nil {
		return "", "", "", fmt.Errorf("failed to copy session directory: %w", err)
	}

	// Update .description file with new description
	descPath := filepath.Join(newSessionPath, ".description")
	if err := os.WriteFile(descPath, []byte(description), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to write description: %w", err)
	}

	return newSessionName, newSessionPath, claudeSessionID, nil
}

func hasClaudeSessionID(sessionName string) bool {
	// Claude session IDs are UUIDs in format: 8-4-4-4-12 hex digits
	// Example: 33342657-73dc-407d-9aa6-a28f2e619268
	uuidPattern := regexp.MustCompile(`-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidPattern.MatchString(sessionName)
}

func extractClaudeSessionID(sessionName string) string {
	if !hasClaudeSessionID(sessionName) {
		return ""
	}

	// Find the UUID pattern at the end
	uuidPattern := regexp.MustCompile(`-([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})$`)
	matches := uuidPattern.FindStringSubmatch(sessionName)
	if len(matches) > 1 {
		return matches[1] // Return the captured UUID without the leading hyphen
	}
	return ""
}

func stripClaudeSessionID(sessionName string) string {
	// Claude session IDs are UUIDs in format: 8-4-4-4-12 hex digits
	// We want to strip the entire UUID, not just the last segment

	if !hasClaudeSessionID(sessionName) {
		return sessionName
	}

	// Remove the UUID pattern from the end
	uuidPattern := regexp.MustCompile(`-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidPattern.ReplaceAllString(sessionName, "")
}

func renameSessionWithClaudeID(oldPath, sessionName, claudeSessionID string) (string, error) {
	if oldPath == "" {
		// Ephemeral session, no directory to rename
		return "", nil
	}

	// Strip any existing Claude session ID from the session name
	baseSessionName := stripClaudeSessionID(sessionName)

	// Create new directory name with Claude session ID suffix
	parentDir := filepath.Dir(oldPath)
	newDirName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
	newPath := filepath.Join(parentDir, newDirName)

	// Rename the directory
	if err := os.Rename(oldPath, newPath); err != nil {
		return "", fmt.Errorf("failed to rename session directory: %w", err)
	}

	return newPath, nil
}

func loadProfile(profileName string) ([]byte, error) {
	// Look for profile in profiles/agents/ directory
	agentPath := "profiles/agents/" + profileName
	return fs.ReadFile(profilesFS, agentPath)
}

func resolveProfilePath(profileName string) string {
	// Look for profile in profiles/agents/ directory
	agentPath := "profiles/agents/" + profileName
	if _, err := fs.Stat(profilesFS, agentPath); err == nil {
		return agentPath
	}

	return ""
}

func updateLastUsed(sessionPath string) error {
	if sessionPath == "" {
		// Ephemeral session, no directory to update
		return nil
	}

	lastUsed := time.Now().UTC().Format(time.RFC3339)
	return os.WriteFile(filepath.Join(sessionPath, ".last_used"), []byte(lastUsed), 0644)
}
