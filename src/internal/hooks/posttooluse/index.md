# hooks/posttooluse

Post-tool execution hooks for autodoc progress tracking and logging.

## Key Files

- **autodoc.go** - Frequency-controlled documentation update handler
- **logger.go** - Simple logging handler for tool completion events

## Key Types

- `AutoDocHandler` - Implements frequency-controlled autodoc updates (e.g., every 5 tool uses)
- `Handler` - Logs tool completion information

## AutoDocHandler Behavior

1. Finds session folder using `session.FindSessionFolderWithCwd()`
2. Increments counter via `session.IncrementCounter()`
3. Checks if counter reached frequency threshold (default: 5)
4. When threshold reached:
   - Resets counter via `session.ResetCounter()`
   - Reads last processed transcript line via `session.ReadLastProcessedLine()`
   - Triggers background doc update via `doc.Updater.RunBackground()`
   - Uses incremental transcript parsing (startLine to current)
5. Returns "allow" decision in all cases (non-blocking)

## Logger Handler Behavior

Logs tool completion with status using `shared.Logger` and returns "allow" decision.

## Configuration

Autodoc frequency controlled by config file (default: 5). Background doc updates use:
- Model: haiku
- Template: session-overview-documenter.md
- Output: session-overview.md

## Usage

Hook is invoked automatically by Claude Code after each tool execution. Executable located at `.claude/hooks/PostToolUse`.
