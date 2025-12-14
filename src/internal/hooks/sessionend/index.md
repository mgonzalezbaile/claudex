# hooks/sessionend

Final documentation update hook triggered when Claude Code session terminates.

## Key Files

- **autodoc.go** - Handler for SessionEnd events with final doc update

## Key Types

- `Handler` - Processes SessionEnd events and triggers final documentation update

## Behavior

1. Logs session end reason (if provided)
2. Finds session folder using `session.FindSessionFolderWithCwd()`
3. Reads last processed transcript line via `session.ReadLastProcessedLine()`
4. Triggers final background doc update via `doc.Updater.RunBackground()`
5. Uses incremental transcript parsing (startLine to current)
6. Returns nil on success (no JSON output needed)

## Doc Update Configuration

Final update always runs regardless of autodoc counter. Uses same configuration as PostToolUse:
- Model: haiku
- Template: session-overview-documenter.md
- Output: session-overview.md
- Incremental: Yes (startLine from last processed marker)

## Purpose

Ensures session documentation is complete and up-to-date when user exits Claude Code, capturing all remaining transcript content not processed by PostToolUse autodoc.

## Usage

Hook is invoked automatically by Claude Code when session ends. Executable located at `.claude/hooks/SessionEnd`.
