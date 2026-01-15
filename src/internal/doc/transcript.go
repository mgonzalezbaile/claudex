// Package doc provides functionality for documentation generation,
// including transcript parsing, prompt template loading, and background
// Claude invocation for documentation updates.
package doc

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/afero"
)

// TranscriptEntry represents a parsed line from JSONL transcript
type TranscriptEntry struct {
	Type      string   `json:"type"`      // "assistant_message" or "agent_result"
	Timestamp string   `json:"timestamp"` // ISO 8601 timestamp
	AgentID   string   `json:"agentId,omitempty"`
	Content   []string `json:"content"` // Text content extracted
}

// rawTranscriptLine represents the raw JSONL structure we're parsing
type rawTranscriptLine struct {
	Type          string            `json:"type"`
	Timestamp     string            `json:"timestamp"`
	Message       *rawMessage       `json:"message,omitempty"`
	ToolUseResult *rawToolUseResult `json:"toolUseResult,omitempty"`
}

type rawMessage struct {
	Content []rawContent `json:"content"`
}

type rawToolUseResult struct {
	Status  string       `json:"status"`
	AgentID string       `json:"agentId"`
	Content []rawContent `json:"content"`
}

type rawContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ParseTranscript reads JSONL transcript and extracts relevant entries.
// It filters for assistant messages and completed agent results.
// startLine: line number to start from (1-indexed)
// Returns entries and the last line number processed
func ParseTranscript(fs afero.Fs, transcriptPath string, startLine int) ([]TranscriptEntry, int, error) {
	file, err := fs.Open(transcriptPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open transcript: %w", err)
	}
	defer file.Close()

	return parseTranscriptFromReader(file, startLine)
}

// parseTranscriptFromReader parses transcript from an io.Reader
// This allows for easier testing with in-memory data
func parseTranscriptFromReader(r io.Reader, startLine int) ([]TranscriptEntry, int, error) {
	scanner := bufio.NewScanner(r)

	// Increase buffer size to handle large lines (default is 64KB)
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	entries := []TranscriptEntry{}
	lineNum := 0

	for scanner.Scan() {
		lineNum++

		// Skip lines before startLine
		if lineNum < startLine {
			continue
		}

		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse the raw JSONL line
		var raw rawTranscriptLine
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			// Skip malformed JSON lines gracefully
			continue
		}

		// Extract relevant entries based on type
		entry := extractEntry(&raw)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, lineNum, fmt.Errorf("error reading transcript: %w", err)
	}

	return entries, lineNum, nil
}

// extractEntry converts a raw transcript line to a TranscriptEntry if relevant
// Returns nil if the line should be filtered out
func extractEntry(raw *rawTranscriptLine) *TranscriptEntry {
	// Filter 1: Assistant messages with content
	if raw.Type == "assistant" && raw.Message != nil && len(raw.Message.Content) > 0 {
		textContent := extractTextContent(raw.Message.Content)
		if len(textContent) == 0 {
			return nil
		}

		return &TranscriptEntry{
			Type:      "assistant_message",
			Timestamp: raw.Timestamp,
			Content:   textContent,
		}
	}

	// Filter 2: Completed tool results with agentId (sub-agent results)
	if raw.Type == "user" && raw.ToolUseResult != nil &&
		raw.ToolUseResult.Status == "completed" &&
		raw.ToolUseResult.AgentID != "" {

		textContent := extractTextContent(raw.ToolUseResult.Content)
		if len(textContent) == 0 {
			return nil
		}

		return &TranscriptEntry{
			Type:      "agent_result",
			Timestamp: raw.Timestamp,
			AgentID:   raw.ToolUseResult.AgentID,
			Content:   textContent,
		}
	}

	return nil
}

// extractTextContent filters content array for text-only items
func extractTextContent(content []rawContent) []string {
	texts := []string{}
	for _, c := range content {
		if c.Type == "text" && strings.TrimSpace(c.Text) != "" {
			texts = append(texts, c.Text)
		}
	}
	return texts
}

// FormatTranscriptForPrompt converts entries to markdown for Claude prompt
func FormatTranscriptForPrompt(entries []TranscriptEntry) string {
	if len(entries) == 0 {
		return "No new transcript content."
	}

	var sb strings.Builder
	sb.WriteString("# Transcript Increment\n\n")

	for _, entry := range entries {
		switch entry.Type {
		case "assistant_message":
			sb.WriteString("## Assistant Message\n")
			sb.WriteString(fmt.Sprintf("**Timestamp**: %s\n\n", entry.Timestamp))
			for _, text := range entry.Content {
				sb.WriteString(text)
				sb.WriteString("\n\n")
			}

		case "agent_result":
			sb.WriteString("## Agent Result\n")
			sb.WriteString(fmt.Sprintf("**Timestamp**: %s\n", entry.Timestamp))
			sb.WriteString(fmt.Sprintf("**Agent ID**: %s\n\n", entry.AgentID))
			for _, text := range entry.Content {
				sb.WriteString(text)
				sb.WriteString("\n\n")
			}
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}
