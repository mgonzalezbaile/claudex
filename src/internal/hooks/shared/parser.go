package shared

import (
	"encoding/json"
	"fmt"
	"io"
)

// Parser handles parsing of hook input JSON from stdin
type Parser struct {
	reader io.Reader
}

// NewParser creates a new Parser instance
func NewParser(reader io.Reader) *Parser {
	return &Parser{reader: reader}
}

// ParsePreToolUse parses PreToolUse input from JSON
func (p *Parser) ParsePreToolUse() (*PreToolUseInput, error) {
	var input PreToolUseInput
	if err := json.NewDecoder(p.reader).Decode(&input); err != nil {
		return nil, fmt.Errorf("failed to parse PreToolUse input: %w", err)
	}

	// Validate required fields
	if input.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	if input.ToolName == "" {
		return nil, fmt.Errorf("tool_name is required")
	}

	return &input, nil
}

// ParsePostToolUse parses PostToolUse input from JSON
func (p *Parser) ParsePostToolUse() (*PostToolUseInput, error) {
	var input PostToolUseInput
	if err := json.NewDecoder(p.reader).Decode(&input); err != nil {
		return nil, fmt.Errorf("failed to parse PostToolUse input: %w", err)
	}

	// Validate required fields
	if input.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	if input.ToolName == "" {
		return nil, fmt.Errorf("tool_name is required")
	}

	return &input, nil
}

// ParseNotification parses Notification input from JSON
func (p *Parser) ParseNotification() (*NotificationInput, error) {
	var input NotificationInput
	if err := json.NewDecoder(p.reader).Decode(&input); err != nil {
		return nil, fmt.Errorf("failed to parse Notification input: %w", err)
	}

	// Validate required fields
	if input.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	if input.Message == "" {
		return nil, fmt.Errorf("message is required")
	}

	return &input, nil
}

// ParseSessionEnd parses SessionEnd input from JSON
func (p *Parser) ParseSessionEnd() (*SessionEndInput, error) {
	var input SessionEndInput
	if err := json.NewDecoder(p.reader).Decode(&input); err != nil {
		return nil, fmt.Errorf("failed to parse SessionEnd input: %w", err)
	}

	// Validate required fields
	if input.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}

	return &input, nil
}

// ParseSubagentStop parses SubagentStop input from JSON
func (p *Parser) ParseSubagentStop() (*SubagentStopInput, error) {
	var input SubagentStopInput
	if err := json.NewDecoder(p.reader).Decode(&input); err != nil {
		return nil, fmt.Errorf("failed to parse SubagentStop input: %w", err)
	}

	// Validate required fields
	if input.SessionID == "" {
		return nil, fmt.Errorf("session_id is required")
	}
	if input.AgentID == "" {
		return nil, fmt.Errorf("agent_id is required")
	}

	return &input, nil
}

// ParseDocUpdate parses DocUpdate input from JSON
func (p *Parser) ParseDocUpdate() (*DocUpdateInput, error) {
	var input DocUpdateInput
	if err := json.NewDecoder(p.reader).Decode(&input); err != nil {
		return nil, fmt.Errorf("failed to parse DocUpdate input: %w", err)
	}

	// Validate required fields
	if input.SessionPath == "" {
		return nil, fmt.Errorf("session_path is required")
	}
	if input.TranscriptPath == "" {
		return nil, fmt.Errorf("transcript_path is required")
	}
	if input.PromptTemplate == "" {
		return nil, fmt.Errorf("prompt_template is required")
	}
	if input.Model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if input.StartLine < 1 {
		return nil, fmt.Errorf("start_line must be >= 1")
	}

	return &input, nil
}
