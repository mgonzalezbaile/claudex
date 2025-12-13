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
4. Lists current session folder contents
5. Injects context before original prompt using `UpdatedInput` field
6. Returns "allow" with modified prompt for Task tools

## Context Injection Format

```markdown
## SESSION CONTEXT (CRITICAL)

**Session Folder (Absolute Path)**: `{sessionPath}`

### MANDATORY RULES for Documentation:
1. ✅ ALWAYS save documentation to the session folder above
2. ✅ Use absolute paths when creating files
3. ✅ Check session folder for existing context before exploring codebase
4. ❌ NEVER save documentation to project root
5. ❌ NEVER use relative paths for documentation files

### Session Folder Contents:
- file1.md
- file2.md

### Recommended File Names:
- {CLAUDEX_DOC_PATHS entries}

---

## ORIGINAL REQUEST

{original prompt}
```

## Usage

Hook is invoked automatically by Claude Code before each tool execution. Executable located at `.claude/hooks/PreToolUse`.
