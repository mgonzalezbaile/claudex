package app

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"claudex"
	"claudex/internal/services/config"
	"claudex/internal/services/profile"
	"claudex/internal/services/session"
	setupuc "claudex/internal/usecases/setup"
	setuphookuc "claudex/internal/usecases/setuphook"
	setupmcpuc "claudex/internal/usecases/setupmcp"
	updatecheckuc "claudex/internal/usecases/updatecheck"
	updatedocsuc "claudex/internal/usecases/updatedocs"
	"claudex/internal/services/mcpconfig"
	"github.com/spf13/afero"
)

// LaunchMode represents how Claude should be started
type LaunchMode string

const (
	LaunchModeNew       LaunchMode = "new"
	LaunchModeResume    LaunchMode = "resume"
	LaunchModeFork      LaunchMode = "fork"
	LaunchModeFresh     LaunchMode = "fresh"
	LaunchModeEphemeral LaunchMode = "ephemeral"
)

// SessionInfo holds session state passed between methods
type SessionInfo struct {
	Name         string
	Path         string
	ClaudeID     string
	Mode         LaunchMode
	OriginalName string // For fork/fresh operations
}

// App is the main application container
type App struct {
	deps            *Dependencies
	cfg             *config.Config
	projectDir      string
	sessionsDir     string
	docPaths        []string
	noOverwrite     bool
	updateDocs      bool
	setupMCP        bool
	logFile         afero.File
	logFilePath     string
	version         string
	showVersion     *bool
	noOverwriteFlag *bool
	updateDocsFlag  *bool
	setupMCPFlag    *bool
	docPathsFlag    []string
}

// New creates a new App instance with production dependencies
func New(version string, showVersion *bool, noOverwrite *bool, updateDocs *bool, setupMCP *bool, docPaths []string) *App {
	return &App{
		deps:            NewDependencies(),
		version:         version,
		showVersion:     showVersion,
		noOverwriteFlag: noOverwrite,
		updateDocsFlag:  updateDocs,
		setupMCPFlag:    setupMCP,
		docPathsFlag:    docPaths,
	}
}

// Init initializes the application (parse flags, load config, setup logging)
func (a *App) Init() error {
	// Load config file (before flag.Parse)
	cfg, err := config.Load(a.deps.FS, ".claudex.toml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		cfg = &config.Config{Doc: []string{}, NoOverwrite: false}
	}
	a.cfg = cfg

	flag.Parse()

	if *a.showVersion {
		fmt.Printf("claudex %s\n", a.version)
		os.Exit(0)
	}

	// Apply precedence: CLI flags > config > defaults
	if !isFlagSet("doc") && len(cfg.Doc) > 0 {
		a.docPaths = cfg.Doc
	} else {
		a.docPaths = a.docPathsFlag
	}
	if !isFlagSet("no-overwrite") && cfg.NoOverwrite {
		a.noOverwrite = cfg.NoOverwrite
	} else {
		a.noOverwrite = *a.noOverwriteFlag
	}
	a.updateDocs = *a.updateDocsFlag
	a.setupMCP = *a.setupMCPFlag

	projectDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	a.projectDir = projectDir
	a.sessionsDir = filepath.Join(projectDir, "sessions")

	// Ensure .claude directory is set up using setup usecase
	setupUC := setupuc.New(a.deps.FS, a.deps.Env)
	if err := setupUC.Execute(projectDir, a.noOverwrite); err != nil {
		return fmt.Errorf("failed to setup .claude directory: %w", err)
	}

	// Setup centralized logging
	logsDir := filepath.Join(projectDir, "logs")
	if err := a.deps.FS.MkdirAll(logsDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not create logs directory: %v\n", err)
	}

	// Create unique log file for this execution
	timestamp := a.deps.Clock.Now().Format("20060102-150405")
	logFileName := fmt.Sprintf("claudex-%s.log", timestamp)
	logFilePath := filepath.Join(logsDir, logFileName)

	// Open log file
	logFile, err := a.deps.FS.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not open log file: %v\n", err)
	} else {
		a.logFile = logFile
		a.logFilePath = logFilePath
		// Configure Go logger with [claudex] prefix
		log.SetOutput(logFile)
		log.SetPrefix("[claudex] ")
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

		// Set environment variable for hooks
		a.deps.Env.Set("CLAUDEX_LOG_FILE", logFilePath)

		log.Printf("Claudex started (log file: %s)", logFileName)
	}

	// Create sessions directory
	if err := a.deps.FS.MkdirAll(a.sessionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create sessions directory: %w", err)
	}

	return nil
}

// Close cleans up resources (close log file)
func (a *App) Close() {
	if a.logFile != nil {
		a.logFile.Close()
	}
}

// renameLogFileForSession renames the log file to match the session name.
// For ephemeral sessions (empty path), the timestamp-based name is preserved.
func (a *App) renameLogFileForSession(si SessionInfo) {
	// Skip for ephemeral sessions
	if si.Path == "" || a.logFilePath == "" {
		return
	}

	// Build new log file path: logs/{session-name}.log
	logsDir := filepath.Dir(a.logFilePath)
	newLogFileName := si.Name + ".log"
	newLogFilePath := filepath.Join(logsDir, newLogFileName)

	// Close current log file first
	if a.logFile != nil {
		a.logFile.Close()
	}

	// Check if we need to rename or if target already exists
	if a.logFilePath != newLogFilePath {
		// Check if target log already exists (resume scenario)
		if _, err := a.deps.FS.Stat(newLogFilePath); os.IsNotExist(err) {
			// Rename current log file to session-named log
			if err := a.deps.FS.Rename(a.logFilePath, newLogFilePath); err != nil {
				// Rename failed, try to reopen original
				log.Printf("Warning: Could not rename log file: %v", err)
				a.logFile, _ = a.deps.FS.OpenFile(a.logFilePath,
					os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
				return
			}
		} else {
			// Target exists (resume scenario), remove the timestamp log
			a.deps.FS.Remove(a.logFilePath)
		}
	}

	// Open the session-named log file (append mode)
	logFile, err := a.deps.FS.OpenFile(newLogFilePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Warning: Could not open renamed log file: %v", err)
		return
	}

	// Update App state
	a.logFile = logFile
	a.logFilePath = newLogFilePath

	// Reconfigure logger
	log.SetOutput(logFile)

	// Update environment variable
	a.deps.Env.Set("CLAUDEX_LOG_FILE", newLogFilePath)

	log.Printf("Log file associated with session: %s", si.Name)
}

// Run executes the main application logic
func (a *App) Run() error {
	// Check if Claude CLI is installed
	if !a.isClaudeInstalled() {
		fmt.Println("\n❌ Claude Code CLI not found")
		fmt.Println("\nClaudex requires Claude Code CLI to be installed.")
		fmt.Println("\n⚠️  Note: Claude Code requires a Claude Pro ($20/mo), Max ($100/mo),")
		fmt.Println("   or Team subscription. The free tier does not include Claude Code.")
		fmt.Print("\nInstall Claude Code now? [y/n]: ")

		var response string
		fmt.Scanln(&response)

		switch strings.ToLower(strings.TrimSpace(response)) {
		case "y", "yes":
			fmt.Println("\nInstalling Claude Code CLI...")
			if err := a.deps.Cmd.Start("npm", os.Stdin, os.Stdout, os.Stderr, "install", "-g", "@anthropic-ai/claude-code"); err != nil {
				fmt.Fprintf(os.Stderr, "\nInstallation failed: %v\n", err)
				fmt.Println("You can install manually with: npm install -g @anthropic-ai/claude-code")
				return fmt.Errorf("failed to install claude CLI")
			}
			fmt.Println("\n✓ Claude Code CLI installed successfully!")
			fmt.Println("Please run 'claudex' again to continue.")
			return nil
		default:
			fmt.Println("\nYou can install manually with: npm install -g @anthropic-ai/claude-code")
			fmt.Println("More info: https://docs.anthropic.com/en/docs/claude-code")
			return fmt.Errorf("claude CLI not installed")
		}
	}

	// Early exit for --update-docs mode
	if a.updateDocs {
		uc := updatedocsuc.New(a.deps.FS, a.deps.Cmd, a.deps.Env)
		return uc.Execute(a.projectDir)
	}

	// Early exit for --setup-mcp mode
	if a.setupMCP {
		a.promptMCPSetup()
		return nil
	}

	// Check for updates first (before other prompts)
	a.promptUpdateCheck()

	// Check if user wants to enable git hook integration
	a.promptHookSetup()

	// Check if user wants to configure recommended MCPs
	a.promptMCPSetup()

	// Load team-lead profile directly (skip profile selection menu)
	_, err := profile.LoadComposed(claudex.Profiles, "team-lead")
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	// Show session selector TUI
	fm, err := a.showSessionSelector()
	if err != nil {
		return err
	}
	if fm.Quitting {
		return nil
	}

	// Handle session selection
	var si SessionInfo
	switch fm.Choice {
	case "new":
		si, err = a.handleNewSession()
	case "ephemeral":
		si = SessionInfo{
			Name: fm.SessionName,
			Path: fm.SessionPath,
			Mode: LaunchModeEphemeral,
		}
	case "session":
		// Check if selected session has a Claude session ID (for resume/fork choice)
		if session.HasClaudeSessionID(fm.SessionName) {
			si, err = a.handleResumeOrFork(fm)
		} else {
			// Session without Claude ID - treat as ephemeral
			si = SessionInfo{
				Name: fm.SessionName,
				Path: fm.SessionPath,
				Mode: LaunchModeEphemeral,
			}
		}
	default:
		// fm.Choice is a Claude session ID from new session creation
		si = SessionInfo{
			Name:     fm.SessionName,
			Path:     fm.SessionPath,
			ClaudeID: fm.Choice,
			Mode:     LaunchModeNew,
		}
	}
	if err != nil {
		return err
	}

	// Rename log file to match session (skip for ephemeral)
	a.renameLogFileForSession(si)

	// Set environment and launch
	a.setEnvironment(si, a.cfg)
	return a.launch(si)
}

// promptHookSetup checks if we should offer git hook integration
func (a *App) promptHookSetup() {
	uc := setuphookuc.New(a.deps.FS, a.projectDir, a.deps.Cmd)

	result := uc.ShouldPrompt()
	if result != setuphookuc.ResultPromptUser {
		return // Nothing to prompt
	}

	// Simple prompt using fmt (not TUI - keep it lightweight)
	fmt.Print("\n📝 Enable auto-docs update after git commits? [y/n/never]: ")

	var response string
	fmt.Scanln(&response)

	switch strings.ToLower(strings.TrimSpace(response)) {
	case "y", "yes":
		if err := uc.Install(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not install hook: %v\n", err)
		} else {
			fmt.Println("✓ Git hook installed. Docs will auto-update after commits.")
		}
	case "never":
		if err := uc.SaveDeclined(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not save preference: %v\n", err)
		}
		fmt.Println("○ Won't ask again. Run 'claudex --setup-hook' to enable later.")
	default:
		fmt.Println("○ Skipped for now.")
	}
	fmt.Println()
}

// promptMCPSetup checks if we should offer MCP configuration
func (a *App) promptMCPSetup() {
	uc := setupmcpuc.New(a.deps.FS)

	result := uc.ShouldPrompt()
	if result != setupmcpuc.ResultPromptUser {
		return // Nothing to prompt
	}

	// Simple prompt using fmt (not TUI - keep it lightweight)
	fmt.Print("\nConfigure recommended MCPs (sequential-thinking, context7)? [y/n/never]: ")

	var response string
	fmt.Scanln(&response)

	switch strings.ToLower(strings.TrimSpace(response)) {
	case "y", "yes":
		// Prompt for optional Context7 API token
		fmt.Println("\nContext7 requires an API token for higher rate limits (optional).")
		fmt.Printf("Generate one at: %s\n", mcpconfig.Context7TokenURL)
		fmt.Print("Enter token (or press Enter to skip): ")

		var token string
		fmt.Scanln(&token)
		token = strings.TrimSpace(token)

		if err := uc.Install(token); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not configure MCPs: %v\n", err)
		} else {
			fmt.Println("✓ MCP configuration added to ~/.claude.json")
			if token == "" {
				fmt.Println("  Note: Context7 running in rate-limited mode (60 req/hour)")
			}
		}
	case "never":
		if err := uc.SaveDeclined(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not save preference: %v\n", err)
		}
		fmt.Println("○ Won't ask again. Run 'claudex --setup-mcp' to configure later.")
	default:
		fmt.Println("○ Skipped for now.")
	}
	fmt.Println()
}

// promptUpdateCheck checks if we should offer to update claudex
func (a *App) promptUpdateCheck() {
	uc := updatecheckuc.New(a.deps.FS, a.version)

	result := uc.ShouldPrompt()
	if result != updatecheckuc.ResultPromptUser {
		return // Nothing to prompt
	}

	// Prompt user
	fmt.Printf("\nNew version available: %s (current: %s)\n", uc.GetLatestVersion(), uc.GetCurrentVersion())
	fmt.Print("Update now? [y/n/never]: ")

	var response string
	fmt.Scanln(&response)

	switch strings.ToLower(strings.TrimSpace(response)) {
	case "y", "yes":
		fmt.Println("Updating claudex...")
		if err := a.deps.Cmd.Start("npm", os.Stdin, os.Stdout, os.Stderr, "install", "-g", "@claudex/cli@latest"); err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			fmt.Println("You can update manually with: npm install -g @claudex/cli@latest")
		} else {
			fmt.Printf("✓ Updated to %s\n", uc.GetLatestVersion())
		}
	case "never":
		if err := uc.SaveNeverAsk(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Could not save preference: %v\n", err)
		}
		fmt.Println("○ Won't ask again. Run 'npm install -g @claudex/cli@latest' to update manually.")
	default:
		fmt.Println("○ Skipped for now.")
	}
	fmt.Println()
}

// isClaudeInstalled checks if the Claude CLI is available in PATH
func (a *App) isClaudeInstalled() bool {
	_, err := a.deps.Cmd.Run("claude", "--version")
	return err == nil
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
