package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// InterceptAction defines what to do when a pattern matches
type InterceptAction func(input string, writer io.Writer) bool

// PatternRule defines a pattern and its associated action
type PatternRule struct {
	Pattern *regexp.Regexp
	Action  InterceptAction
}

// Interceptor handles I/O interception with pattern matching
type Interceptor struct {
	logFile       *os.File
	inputBuffer   *bytes.Buffer
	outputBuffer  *bytes.Buffer
	inputRules    []PatternRule // Rules checked on ENTER
	outputRules   []PatternRule // Rules checked continuously on output
	inEscapeSeq   bool          // Track if we're in the middle of an ANSI escape sequence
	lastEnterByte byte          // Store the last ENTER byte pressed (for replaying)
	ptyWriter     io.Writer     // Writer to send input to Claude's PTY
}

// NewInterceptor creates a new interceptor
func NewInterceptor(logFile *os.File) *Interceptor {
	return &Interceptor{
		logFile:      logFile,
		inputBuffer:  new(bytes.Buffer),
		outputBuffer: new(bytes.Buffer),
		inputRules:   make([]PatternRule, 0),
		outputRules:  make([]PatternRule, 0),
	}
}

// GetLastEnterByte returns the last ENTER byte that was pressed
func (i *Interceptor) GetLastEnterByte() byte {
	return i.lastEnterByte
}

// GetLogFile returns the log file for debugging
func (i *Interceptor) GetLogFile() *os.File {
	return i.logFile
}

// GetPtyWriter returns the PTY writer for sending input to Claude
func (i *Interceptor) GetPtyWriter() io.Writer {
	return i.ptyWriter
}

// SetPtyWriter sets the PTY writer
func (i *Interceptor) SetPtyWriter(writer io.Writer) {
	i.ptyWriter = writer
}

// AddInputRule adds a pattern matching rule for user input (checked on ENTER)
func (i *Interceptor) AddInputRule(pattern string, action InterceptAction) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}

	i.inputRules = append(i.inputRules, PatternRule{
		Pattern: regex,
		Action:  action,
	})

	return nil
}

// AddOutputRule adds a pattern matching rule for output (checked continuously)
func (i *Interceptor) AddOutputRule(pattern string, action InterceptAction) error {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}

	i.outputRules = append(i.outputRules, PatternRule{
		Pattern: regex,
		Action:  action,
	})

	return nil
}

// HandleInput processes input from src to dst with pattern matching
// Patterns are only checked when Enter is pressed
func (i *Interceptor) HandleInput(src io.Reader, dst io.Writer) error {
	buf := make([]byte, 1)

	for {
		n, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if n == 0 {
			continue
		}

		b := buf[0]
		char := string(b)

		// Track ANSI escape sequences
		if b == 27 { // ESC character starts an escape sequence
			i.inEscapeSeq = true
		} else if i.inEscapeSeq {
			// Skip all characters until we hit a letter (end of escape sequence)
			if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') {
				i.inEscapeSeq = false
			}
			// Forward escape sequences immediately and continue
			if _, err := dst.Write(buf[:n]); err != nil {
				return err
			}
			continue
		}

		// Check for Enter key (newline or carriage return) - only when NOT in escape seq
		if char == "\r" || char == "\n" {
			// Store the ENTER byte for pattern rules to use
			i.lastEnterByte = b

			// Get the accumulated input
			inputStr := i.inputBuffer.String()
			trimmedInput := strings.TrimSpace(inputStr)

			// Log the user input
			if i.logFile != nil {
				fmt.Fprintf(i.logFile, "\n[CLAUDEX DEBUG] User input: \"%s\"\n", trimmedInput)
			}

			// Check all INPUT rules against the complete input
			matched := false
			for _, rule := range i.inputRules {
				if rule.Pattern.MatchString(trimmedInput) {
					// Execute the action (it will handle its own output)
					shouldBlock := rule.Action(trimmedInput, dst)

					if i.logFile != nil {
						if shouldBlock {
							fmt.Fprintf(i.logFile, "[CLAUDEX BLOCKED] Pattern matched: %s\n", rule.Pattern.String())
						} else {
							fmt.Fprintf(i.logFile, "[CLAUDEX INTERCEPTED] Pattern matched: %s (shouldForwardEnter=%v, byte=%d)\n",
								rule.Pattern.String(), !shouldBlock, b)
						}
					}

					// If not blocked, forward the Enter key
					if !shouldBlock {
						if _, err := dst.Write(buf[:n]); err != nil {
							return err
						}
					}

					matched = true
					break
				}
			}

			// If no rule matched, forward the Enter key
			if !matched {
				if _, err := dst.Write(buf[:n]); err != nil {
					return err
				}
			}

			// Clear buffer
			i.inputBuffer.Reset()
			continue
		}

		// Handle backspace/delete keys - remove from buffer
		if b == 0x7F || b == 0x08 { // DEL (127) or Backspace (8)
			// Remove the last character from the buffer if it exists
			if i.inputBuffer.Len() > 0 {
				bufBytes := i.inputBuffer.Bytes()
				i.inputBuffer.Reset()
				if len(bufBytes) > 0 {
					i.inputBuffer.Write(bufBytes[:len(bufBytes)-1])
				}
			}
		}

		// Accumulate printable ASCII in buffer for pattern matching
		if b >= 32 && b <= 126 {
			i.inputBuffer.WriteByte(b)
			if i.logFile != nil && (b == ':' || b == '/') {
				fmt.Fprintf(i.logFile, "[CLAUDEX CHAR DEBUG] Captured special char: '%c' (byte %d)\n", b, b)
			}
		}

		// Forward ALL keystrokes immediately for proper display/echo
		if _, err := dst.Write(buf[:n]); err != nil {
			return err
		}
	}
}

// HandleOutput processes output from src to dst with pattern detection
func (i *Interceptor) HandleOutput(src io.Reader, dst io.Writer) error {
	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		data := buf[:n]

		// Accumulate data for pattern detection
		i.outputBuffer.Write(data)

		// Keep buffer manageable (last 1000 chars)
		if i.outputBuffer.Len() > 1000 {
			outputBytes := i.outputBuffer.Bytes()
			i.outputBuffer.Reset()
			i.outputBuffer.Write(outputBytes[len(outputBytes)-1000:])
		}

		// Check for patterns in output
		outputStr := i.outputBuffer.String()
		for _, rule := range i.outputRules {
			if rule.Pattern.MatchString(outputStr) {
				// Execute action for output patterns
				rule.Action(outputStr, dst)

				if i.logFile != nil {
					fmt.Fprintf(i.logFile, "\n[CLAUDEX OUTPUT] Pattern detected: %s\n", rule.Pattern.String())
				}

				// Clear the pattern from buffer to avoid re-triggering
				i.outputBuffer.Reset()
				break
			}
		}

		// Write to terminal
		if _, err := dst.Write(data); err != nil {
			return err
		}

		// Log to file
		if i.logFile != nil {
			i.logFile.Write(data)
		}
	}
}
