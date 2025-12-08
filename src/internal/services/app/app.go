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
	logFile         afero.File
	logFilePath     string
	version         string
	showVersion     *bool
	noOverwriteFlag *bool
	docPathsFlag    []string
}

// New creates a new App instance with production dependencies
func New(version string, showVersion *bool, noOverwrite *bool, docPaths []string) *App {
	return &App{
		deps:            NewDependencies(),
		version:         version,
		showVersion:     showVersion,
		noOverwriteFlag: noOverwrite,
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
