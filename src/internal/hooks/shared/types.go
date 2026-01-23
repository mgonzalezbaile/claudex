// Package shared provides hook implementations for Claude Code events.
// It defines types, parsers, and builders for processing hook inputs and outputs.
package shared

// HookInput represents common fields present in all hook inputs
type HookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	CWD            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"`
	HookEventName  string `json:"hook_event_name"`
}

// PreToolUseInput extends HookInput for PreToolUse events
type PreToolUseInput struct {
	HookInput
	ToolName  string                 `json:"tool_name"`
	ToolInput map[string]interface{} `json:"tool_input"`
	ToolUseID string                 `json:"tool_use_id"`
	AgentID   string                 `json:"agent_id,omitempty"`
}

// PostToolUseInput extends HookInput for PostToolUse events
type PostToolUseInput struct {
	HookInput
	ToolName     string                 `json:"tool_name"`
	ToolInput    map[string]interface{} `json:"tool_input"`
	ToolResponse interface{}            `json:"tool_response"`
	ToolUseID    string                 `json:"tool_use_id"`
	Status       string                 `json:"status"`
	AgentID      string                 `json:"agent_id,omitempty"`
}

// NotificationInput represents input for notification hook
type NotificationInput struct {
	HookInput
	Message          string `json:"message"`
	NotificationType string `json:"notification_type"`
}

// SessionEndInput extends HookInput for SessionEnd events
type SessionEndInput struct {
	HookInput
	Reason string `json:"reason,omitempty"`
}

// SubagentStopInput extends HookInput for SubagentStop events
type SubagentStopInput struct {
	HookInput
	AgentID             string `json:"agent_id"`
	AgentTranscriptPath string `json:"agent_transcript_path"`
	CompletionReason    string `json:"completion_reason,omitempty"`
}

// DocUpdateInput represents input for the doc-update command
// This is used to pass configuration to the detached subprocess
type DocUpdateInput struct {
	SessionPath    string `json:"session_path"`
	TranscriptPath string `json:"transcript_path"`
	OutputFile     string `json:"output_file"`
	PromptTemplate string `json:"prompt_template"`
	SessionContext string `json:"session_context"`
	Model          string `json:"model"`
	StartLine      int    `json:"start_line"`
}

// HookOutput represents the response structure for all hooks
type HookOutput struct {
	HookSpecificOutput HookSpecificOutput `json:"hookSpecificOutput"`
}

// HookSpecificOutput contains hook-specific response fields
type HookSpecificOutput struct {
	HookEventName            string                 `json:"hookEventName"`
	PermissionDecision       string                 `json:"permissionDecision,omitempty"`
	PermissionDecisionReason string                 `json:"permissionDecisionReason,omitempty"`
	UpdatedInput             map[string]interface{} `json:"updatedInput,omitempty"`
}
