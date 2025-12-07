package main

import (
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
	setupuc "claudex/internal/usecases/setup"
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
	deps        *Dependencies
	cfg         *config.Config
	projectDir  string
	sessionsDir string
	docPaths    []string
	noOverwrite bool
	logFile     *os.File
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

// NewApp creates a new App instance with production dependencies
func NewApp() *App {
	return &App{
		deps: NewDependencies(),
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

	if *showVersion {
		fmt.Printf("claudex %s\n", Version)
		os.Exit(0)
	}

	// Apply precedence: CLI flags > config > defaults
	if !isFlagSet("doc") && len(cfg.Doc) > 0 {
		a.docPaths = cfg.Doc
	} else {
		a.docPaths = docPaths
	}
	if !isFlagSet("no-overwrite") && cfg.NoOverwrite {
		a.noOverwrite = cfg.NoOverwrite
	} else {
		a.noOverwrite = *noOverwrite
	}

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
		a.logFile = logFile
		// Configure Go logger with [claudex] prefix
		log.SetOutput(logFile)
		log.SetPrefix("[claudex] ")
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

		// Set environment variable for hooks
		os.Setenv("CLAUDEX_LOG_FILE", logFilePath)

		log.Printf("Claudex started (log file: %s)", logFileName)
	}

	// Create sessions directory
	if err := os.MkdirAll(a.sessionsDir, 0755); err != nil {
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

// Run executes the main application logic
func (a *App) Run() error {
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

	// Set environment and launch
	a.setEnvironment(si)
	return a.launch(si)
}
