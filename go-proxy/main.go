package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// ANSI color codes for logging
const (
	colorGray = "\x1b[90m"
)

// ProxyState holds the state of the proxy
type ProxyState struct {
	sessionName   string
	sessionPath   string
	logFile       *os.File
	originalState *term.State
	interceptor   *Interceptor
}

func main() {
	state := &ProxyState{
		sessionName: getEnvOrDefault("CLAUDEX_SESSION", "no-session"),
		sessionPath: os.Getenv("CLAUDEX_SESSION_PATH"),
	}

	fmt.Fprintf(os.Stderr, "%s[Claudex Proxy]%s Session: %s\n",
		colorCyan, colorReset, state.sessionName)
	fmt.Fprintf(os.Stderr, "%s[Claudex Proxy]%s Starting interactive session...%s\n\n",
		colorGray, colorReset, colorReset)

	// Setup logging
	if err := state.setupLogging(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logging: %v\n", err)
		os.Exit(1)
	}
	defer state.cleanup()

	// Create interceptor with pattern rules
	state.interceptor = NewInterceptor(state.logFile)
	if err := SetupPatterns(state.interceptor); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup patterns: %v\n", err)
		os.Exit(1)
	}

	// Get the real claude command path
	claudePath := "/opt/homebrew/bin/claude"

	// Get all command line arguments (excluding the program name)
	args := os.Args[1:]

	fmt.Fprintf(os.Stderr, "%s[Claudex Proxy]%s Executing: claude %s%s\n\n",
		colorGray, colorReset, strings.Join(args, " "), colorReset)

	// Create and start the command with a PTY
	cmd := exec.Command(claudePath, args...)
	cmd.Env = os.Environ()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start claude: %v\n", err)
		os.Exit(1)
	}
	defer ptmx.Close()

	// Set stdin to raw mode so we can intercept keypresses
	if term.IsTerminal(int(os.Stdin.Fd())) {
		state.originalState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to set raw mode: %v\n", err)
			os.Exit(1)
		}
	}

	// Set initial PTY size
	if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to set pty size: %v\n", err)
	}

	// Handle terminal resize
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				// Silently ignore resize errors
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize

	// Handle termination signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		if sig == syscall.SIGTERM {
			cmd.Process.Kill()
			os.Exit(0)
		}
		// For SIGINT, let the PTY handle it naturally
	}()

	// Set the PTY writer so output rules can send input to Claude
	state.interceptor.SetPtyWriter(ptmx)

	// Create error channel for goroutines
	errCh := make(chan error, 2)

	// Handle stdin -> pty with interception
	go func() {
		errCh <- state.interceptor.HandleInput(os.Stdin, ptmx)
	}()

	// Handle pty -> stdout with interception and logging
	go func() {
		errCh <- state.interceptor.HandleOutput(ptmx, os.Stdout)
	}()

	// Wait for command to finish
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode := exitErr.ExitCode()
			fmt.Fprintf(os.Stderr, "\n%s[Claudex Proxy]%s Claude exited (code: %d)%s\n",
				colorGray, colorReset, exitCode, colorReset)
			state.logExit(exitCode, 0)
			os.Exit(exitCode)
		}
	}

	fmt.Fprintf(os.Stderr, "\n%s[Claudex Proxy]%s Claude exited (code: 0)%s\n",
		colorGray, colorReset, colorReset)
	state.logExit(0, 0)
}

func (s *ProxyState) setupLogging() error {
	if s.sessionPath == "" {
		return nil
	}

	logFilePath := filepath.Join(s.sessionPath, "conversation.log")
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	s.logFile = f

	separator := strings.Repeat("=", 80)
	fmt.Fprintf(s.logFile, "\n%s\n", separator)
	fmt.Fprintf(s.logFile, "Session: %s\n", s.sessionName)
	fmt.Fprintf(s.logFile, "Started: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(s.logFile, "%s\n\n", separator)

	fmt.Fprintf(os.Stderr, "%s[Claudex Proxy]%s Logging to: %s\n\n",
		colorGreen, colorReset, logFilePath)

	return nil
}

func (s *ProxyState) logExit(exitCode, signal int) {
	if s.logFile == nil {
		return
	}

	separator := strings.Repeat("=", 80)
	fmt.Fprintf(s.logFile, "\n%s\n", separator)
	fmt.Fprintf(s.logFile, "Ended: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(s.logFile, "Exit code: %d\n", exitCode)
	fmt.Fprintf(s.logFile, "%s\n\n", separator)
}

func (s *ProxyState) cleanup() {
	// Restore terminal state
	if s.originalState != nil {
		term.Restore(int(os.Stdin.Fd()), s.originalState)
	}

	// Close log file
	if s.logFile != nil {
		s.logFile.Close()
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
