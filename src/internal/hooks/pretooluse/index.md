# hooks/pretooluse

Context injection hook that modifies Task tool prompts with session folder information.

## Key Files

- **context_injector.go** - Handler for PreToolUse events with session context injection
- **context_injector_test.go** - Test suite for context injection logic

## Key Types

- `Handler` - Processes PreToolUse events and injects session context into Task tool prompts

## Behavior

1. Only modifies `Task` tool invocations (all other tools pass through unchanged)
2. Finds session folder using `session.FindSessionFolder()`
3. Builds markdown context block with session path and mandatory rules
4. Uses pointer-based approach: references `session-overview.md` if available; falls back to file enumeration
5. Detects index.md files in project and adds navigation hint if found
6. Injects context before original prompt using `UpdatedInput` field
7. Returns "allow" with modified prompt for Task tools

## Context Injection Format

```markdown
## SESSION CONTEXT (CRITICAL)

You are working within an active Claudex session.
ALL documentation, plans, and artifacts MUST be created in the session folder.

**Session Folder (Absolute Path)**: `{sessionPath}`

### MANDATORY RULES for Documentation:
1. ✅ ALWAYS save documentation to the session folder above
2. ✅ Use absolute paths when creating files (Write/Edit tools)
3. ✅ Before exploring the codebase, check the session folder for existing context
4. ❌ NEVER save documentation to project root or arbitrary locations
5. ❌ NEVER use relative paths for documentation files

### Session Folder Contents:
(Pointer-based approach when session-overview.md exists)
- {sessionPath}/session-overview.md

OR (File enumeration fallback):
- file1.md
- file2.md

### Codebase Navigation:
(Optional - included if index.md files exist in project)
This project contains index.md files. Use them for quick codebase understanding instead of extensive Glob/Grep searches.

### Recommended File Names:
- {CLAUDEX_DOC_PATHS entries}

---

## ORIGINAL REQUEST

{original prompt}
```

## Usage

Hook is invoked automatically by Claude Code before each tool execution. Executable located at `.claude/hooks/PreToolUse`.
