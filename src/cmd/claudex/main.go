package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"claudex"
	"claudex/internal/services/config"
	"claudex/internal/services/profile"
	"claudex/internal/services/session"
	"claudex/internal/ui"
	setupuc "claudex/internal/usecases/setup"

	newuc "claudex/internal/usecases/session/new"
	forkuc "claudex/internal/usecases/session/resume/fork"
	freshuc "claudex/internal/usecases/session/resume/fresh"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

// Version is set at build time via -ldflags
var Version = "dev"

// stringSlice implements flag.Value to allow multiple --doc flags
type stringSlice []string

func (s *stringSlice) String() string     { return strings.Join(*s, ":") }
func (s *stringSlice) Set(v string) error { *s = append(*s, v); return nil }

var noOverwrite = flag.Bool("no-overwrite", false, "skip overwriting existing .claude files")
var showVersion = flag.Bool("version", false, "print version and exit")
var docPaths stringSlice

func init() {
	flag.Var(&docPaths, "doc", "documentation path for agent context (can be specified multiple times)")
}

// isFlagSet checks if a flag was explicitly set by the user
func isFlagSet(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// resolveDocPaths converts a list of documentation paths to absolute paths
// and joins them with colon separators (Unix PATH convention)
func resolveDocPaths(paths []string) string {
	var resolved []string
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			absPath = p
		}
		resolved = append(resolved, absPath)
	}
	return strings.Join(resolved, ":")
}

func main() {
	// Initialize dependencies
	deps := NewDependencies()

	// Load config file (before flag.Parse)
	cfg, err := config.Load(deps.FS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		cfg = &config.Config{Doc: []string{}, NoOverwrite: false}
	}

	flag.Parse()

	if *showVersion {
		fmt.Printf("claudex %s\n", Version)
		os.Exit(0)
	}

	// Apply precedence: CLI flags > config > defaults
	if !isFlagSet("doc") && len(cfg.Doc) > 0 {
		docPaths = cfg.Doc
	}
	if !isFlagSet("no-overwrite") && cfg.NoOverwrite {
		*noOverwrite = cfg.NoOverwrite
	}

	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Ensure .claude directory is set up using setup usecase
	setupUC := setupuc.New(deps.FS, deps.Env)
	if err := setupUC.Execute(projectDir, *noOverwrite); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up .claude directory: %v\n", err)
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
	sessions, err := session.GetSessions(deps.FS, sessionsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
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
		ProjectDir:  projectDir,
		SessionsDir: sessionsDir,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fm := finalModel.(ui.Model)
	if fm.Quitting {
		return
	}

	// Handle "Create New Session" - use team-lead profile directly
	if fm.Choice == "new" {
		// Load team-lead profile directly (skip profile selection menu)
		_, err := profile.LoadComposed(claudex.Profiles, "team-lead")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading profile: %v\n", err)
			os.Exit(1)
		}

		// Create the session using new usecase
		newSessionUC := newuc.New(deps.FS, deps.Cmd, deps.UUID, deps.Clock, sessionsDir)
		sessionName, sessionPath, claudeSessionID, err := newSessionUC.Execute()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fm.SessionName = sessionName
		fm.SessionPath = sessionPath
		fm.Choice = claudeSessionID // Store session ID for later use
	}

	// Check if selected session has a Claude session ID (for resume/fork choice)
	var resumeOrForkChoice string
	var isFreshMemory bool // Track if "fresh memory" was chosen
	if fm.Choice == "session" && session.HasClaudeSessionID(fm.SessionName) {
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
			ProjectDir:  projectDir,
			SessionsDir: sessionsDir,
		}

		rfProgram := tea.NewProgram(rfModel, tea.WithAltScreen())
		finalRfModel, err := rfProgram.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		rfm := finalRfModel.(ui.Model)
		if rfm.Quitting {
			return
		}

		resumeOrForkChoice = rfm.Choice

		// Add variable to track resume submenu choice
		var resumeSubmenuChoice string

		// If user chose "Resume Session", show submenu: Continue vs Fresh Memory
		if resumeOrForkChoice == "resume" {
			// Show resume submenu: Continue with context vs Fresh memory
			resumeSubmenuItems := []list.Item{
				session.SessionItem{Title: "Continue with context", Description: "Resume with full conversation history", ItemType: "continue"},
				session.SessionItem{Title: "Fresh memory", Description: "Start fresh, keep files, delete original", ItemType: "fresh"},
			}

			delegate := ui.ItemDelegate{}
			rsMenu := list.New(resumeSubmenuItems, delegate, 0, 0)
			rsMenu.Title = fmt.Sprintf("Resume Options • Session: %s", fm.SessionName)
			rsMenu.Styles.Title = ui.TitleStyle()
			rsMenu.SetShowStatusBar(false)
			rsMenu.SetFilteringEnabled(false)
			rsMenu.SetShowHelp(true)

			rsModel := ui.Model{
				List:        rsMenu,
				Stage:       "resume_submenu",
				SessionName: fm.SessionName,
				SessionPath: fm.SessionPath,
				ProjectDir:  projectDir,
				SessionsDir: sessionsDir,
			}

			rsProgram := tea.NewProgram(rsModel, tea.WithAltScreen())
			finalRsModel, err := rsProgram.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			rsm := finalRsModel.(ui.Model)
			if rsm.Quitting {
				return
			}

			resumeSubmenuChoice = rsm.Choice

			// Handle "Fresh Memory" choice using fresh usecase
			if resumeSubmenuChoice == "fresh" {
				freshUC := freshuc.New(deps.FS, deps.UUID, sessionsDir)
				newSessionName, newSessionPath, newClaudeSessionID, err := freshUC.Execute(fm.SessionName)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error creating fresh session: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("\n🔄 Fresh memory: %s → %s (original deleted)\n", fm.SessionName, newSessionName)
				fm.SessionName = newSessionName
				fm.SessionPath = newSessionPath
				fm.Choice = newClaudeSessionID
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
			fmt.Printf("  Original: %s\n", fm.SessionName)
			fmt.Println()

			fmt.Print("  Description for fork: ")
			reader := bufio.NewReader(os.Stdin)
			forkDescription, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading Description: %v\n", err)
				os.Exit(1)
			}
			forkDescription = strings.TrimSpace(forkDescription)

			if forkDescription == "" {
				fmt.Fprintf(os.Stderr, "Error: description cannot be empty\n")
				os.Exit(1)
			}

			// Use fork usecase with description
			forkUC := forkuc.New(deps.FS, deps.Cmd, deps.UUID, sessionsDir)
			newSessionName, newSessionPath, newClaudeSessionID, err := forkUC.ExecuteWithDescription(fm.SessionName, forkDescription)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error forking session: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("\n✅ Forked session: %s → %s\n", fm.SessionName, newSessionName)
			fm.SessionName = newSessionName
			fm.SessionPath = newSessionPath
			fm.Choice = newClaudeSessionID // Store the new session ID
		}
	}

	// Set environment
	os.Setenv("CLAUDEX_SESSION", fm.SessionName)
	os.Setenv("CLAUDEX_SESSION_PATH", fm.SessionPath)
	if len(docPaths) > 0 {
		os.Setenv("CLAUDEX_DOC_PATHS", resolveDocPaths(docPaths))
	}

	// Handle resume vs new/fork session
	var claudeSessionID string
	var isNewSessionAlreadyInitialized bool

	// Check if we just created a new session (session ID stored in fm.Choice)
	if fm.Choice != "new" && fm.Choice != "session" && fm.Choice != "ephemeral" && len(fm.Choice) > 30 {
		// This is a Claude session ID from createNewSessionParallel
		claudeSessionID = fm.Choice
		isNewSessionAlreadyInitialized = true
	}

	if isNewSessionAlreadyInitialized {
		// New session created, launch it with --session-id
		// Update last used timestamp
		if err := session.UpdateLastUsed(deps.FS, deps.Clock, fm.SessionPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not update last used timestamp: %v\n", err)
		}

		// Give terminal a moment to settle
		time.Sleep(100 * time.Millisecond)

		// Clear screen and show launching message
		fmt.Print("\033[H\033[2J\033[3J") // Clear screen and scrollback
		fmt.Print("\033[0m")              // Reset all attributes
		fmt.Printf("\n✅ Launching new Claude session\n")
		fmt.Printf("📦 Session: %s\n", fm.SessionName)
		fmt.Printf("🔄 Session ID: %s\n\n", claudeSessionID)

		// Small delay before launching
		time.Sleep(300 * time.Millisecond)

		// Construct relative session path for activation command
		relativeSessionPath := filepath.Join("sessions", filepath.Base(fm.SessionPath))
		activationPrompt := fmt.Sprintf("/agents:team-lead activate in session %s", relativeSessionPath)
		if len(docPaths) > 0 {
			activationPrompt += fmt.Sprintf(" with documentation %s", strings.Join(docPaths, ", "))
		}

		// Launch the Claude session with activation command
		if err := launchClaude(deps, claudeSessionID, activationPrompt); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
			os.Exit(1)
		}
	} else if resumeOrForkChoice == "resume" || resumeOrForkChoice == "fork" {
		// For resume or fork, get the Claude session ID
		if resumeOrForkChoice == "fork" {
			// For fork, we already have the new session ID in fm.Choice
			claudeSessionID = fm.Choice
		} else {
			// For resume, extract from session name
			claudeSessionID = session.ExtractClaudeSessionID(fm.SessionName)
			if claudeSessionID == "" {
				fmt.Fprintf(os.Stderr, "\n❌ Could not extract session ID for resume\n")
				os.Exit(1)
			}
		}

		// Update last used timestamp
		if err := session.UpdateLastUsed(deps.FS, deps.Clock, fm.SessionPath); err != nil {
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
		fmt.Printf("📦 Session: %s\n", fm.SessionName)
		fmt.Printf("🔄 Session ID: %s\n\n", claudeSessionID)

		// Small delay before launching
		time.Sleep(300 * time.Millisecond)

		if resumeOrForkChoice == "fork" {
			// For fork, start a new session with activation command
			relativeSessionPath := filepath.Join("sessions", filepath.Base(fm.SessionPath))
			activationPrompt := fmt.Sprintf("/agents:team-lead activate in session %s", relativeSessionPath)
			if len(docPaths) > 0 {
				activationPrompt += fmt.Sprintf(" with documentation %s", strings.Join(docPaths, ", "))
			}

			if err := launchClaude(deps, claudeSessionID, activationPrompt); err != nil {
				fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
				os.Exit(1)
			}
		} else {
			// For resume, continue existing session
			if err := resumeClaude(deps, claudeSessionID); err != nil {
				fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
				os.Exit(1)
			}
		}
	} else {
		// Load team-lead profile directly (skip profile selection menu)
		_, err := profile.LoadComposed(claudex.Profiles, "team-lead")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading profile: %v\n", err)
			os.Exit(1)
		}
		profileName := "team-lead"

		// Give terminal a moment to settle
		time.Sleep(100 * time.Millisecond)

		// Clear screen and show launching message
		fmt.Print("\033[H\033[2J\033[3J") // Clear screen and scrollback
		fmt.Print("\033[0m")              // Reset all attributes
		fmt.Printf("\n✅ Launching Claude with %s\n", profileName)
		fmt.Printf("📦 Session: %s\n", fm.SessionName)

		// Generate new Claude session ID
		claudeSessionID = uuid.New().String()

		// Rename session directory to include Claude session ID
		err = session.RenameWithClaudeID(deps.FS, fm.SessionPath, claudeSessionID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error renaming session directory: %v\n", err)
			os.Exit(1)
		}

		// Calculate new session path after rename
		sessionName := filepath.Base(fm.SessionPath)
		baseSessionName := session.StripClaudeSessionID(sessionName)
		newDirName := fmt.Sprintf("%s-%s", baseSessionName, claudeSessionID)
		newSessionPath := filepath.Join(filepath.Dir(fm.SessionPath), newDirName)

		// Update environment variable with new path
		os.Setenv("CLAUDEX_SESSION_PATH", newSessionPath)
		fmt.Printf("📁 Session directory: %s\n", filepath.Base(newSessionPath))
		fmt.Printf("🔄 Session ID: %s\n\n", claudeSessionID)

		// Update last used timestamp
		if err := session.UpdateLastUsed(deps.FS, deps.Clock, newSessionPath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not update last used timestamp: %v\n", err)
		}

		// Small delay before launching
		time.Sleep(300 * time.Millisecond)

		// Construct relative session path for activation command
		relativeSessionPath := filepath.Join("sessions", filepath.Base(newSessionPath))
		activationPrompt := fmt.Sprintf("/agents:team-lead activate in session %s", relativeSessionPath)
		if len(docPaths) > 0 {
			activationPrompt += fmt.Sprintf(" with documentation %s", strings.Join(docPaths, ", "))
		}

		// Launch the Claude session with activation command
		if err := launchClaude(deps, claudeSessionID, activationPrompt); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ Error running Claude session: %v\n", err)
			os.Exit(1)
		}
	}
}

// launchClaude launches a Claude CLI session with the provided session ID and activation prompt
func launchClaude(deps *Dependencies, sessionID string, activationPrompt string) error {
	args := []string{"--session-id", sessionID}
	if activationPrompt != "" {
		args = append(args, activationPrompt)
	}
	return deps.Cmd.Start("claude", os.Stdin, os.Stdout, os.Stderr, args...)
}

// resumeClaude resumes an existing Claude CLI session
func resumeClaude(deps *Dependencies, sessionID string) error {
	return deps.Cmd.Start("claude", os.Stdin, os.Stdout, os.Stderr, "--resume", sessionID)
}
