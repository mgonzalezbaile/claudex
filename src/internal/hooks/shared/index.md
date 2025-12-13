# hooks/shared

Hook framework providing shared types, JSON parsing/building, and logging infrastructure.

## Key Files

- **types.go** - Hook input/output type definitions for all event types
- **parser.go** - JSON parsing from stdin for hook inputs
- **builder.go** - JSON building to stdout for hook outputs
- **logger.go** - Centralized logging to CLAUDEX_LOG_FILE
- **test_helpers.go** - Mock implementations for testing

## Key Types

- `HookInput` - Common fields for all hook events (session_id, transcript_path, cwd, permission_mode, hook_event_name)
- `PreToolUseInput` - Extends HookInput with tool_name, tool_input, tool_use_id
- `PostToolUseInput` - Extends HookInput with tool_name, tool_input, tool_response, status
- `SessionEndInput` - Extends HookInput with optional reason
- `NotificationInput` - Extends HookInput with message, notification_type
- `SubagentStopInput` - Extends HookInput with agent_id, agent_transcript_path, completion_reason
- `HookOutput` - Standard response structure with hookSpecificOutput
- `HookSpecificOutput` - Response fields: hookEventName, permissionDecision, permissionDecisionReason, updatedInput

## Parser Functions

- `ParsePreToolUse()` - Parse PreToolUse input with validation
- `ParsePostToolUse()` - Parse PostToolUse input with validation
- `ParseNotification()` - Parse Notification input with validation
- `ParseSessionEnd()` - Parse SessionEnd input with validation
- `ParseSubagentStop()` - Parse SubagentStop input with validation

## Builder Functions

- `BuildAllow()` - Build simple "allow" response
- `BuildAllowWithReason()` - Build "allow" response with reason
- `BuildDeny()` - Build "deny" response with reason
- `BuildWithUpdatedInput()` - Build response with modified tool input
- `BuildEmpty()` - Build empty response for notification hooks
- `BuildCustom()` - Build response with custom output

## Logger Functions

- `Log()` - Write timestamped log entry to CLAUDEX_LOG_FILE
- `Logf()` - Write formatted log entry
- `LogError()` - Log error message
- `LogInfo()` - Log informational message
- `LogDebug()` - Log debug message

## Usage

```go
// Parse input
parser := shared.NewParser(os.Stdin)
input, err := parser.ParsePreToolUse()

// Build output
builder := shared.NewBuilder(os.Stdout)
err = builder.BuildAllow("PreToolUse")

// Log actions
logger := shared.NewLogger(fs, env, "PreToolUse")
logger.LogInfo("Processing tool invocation")
```
