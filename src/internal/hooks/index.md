# Hooks Layer

Hook system for Claude Code lifecycle events with context injection and autodoc orchestration.

## Hook Directories

- **[shared/](./shared/index.md)** - Hook framework (types, parser, builder, logger)
- **[pretooluse/](./pretooluse/index.md)** - Context injection before tool execution
- **[posttooluse/](./posttooluse/index.md)** - Autodoc progress tracking and logging after tool execution
- **[sessionend/](./sessionend/index.md)** - Final documentation update on session end
- **[notification/](./notification/index.md)** - macOS notification handling
- **[subagent/](./subagent/index.md)** - Agent completion handling

## Hook Event Flow

1. **PreToolUse** - Injects session context into Task tool prompts before execution
2. **PostToolUse** - Logs tool completion, increments counter, triggers autodoc when threshold reached
3. **SessionEnd** - Triggers final documentation update when session terminates
4. **Notification** - Sends macOS notifications with optional voice synthesis
5. **SubagentStop** - Handles agent completion with doc update and notification

## Architecture

All hooks follow a common pattern:
- Parse input JSON via `shared.Parser`
- Process event with handler logic
- Build output JSON via `shared.Builder`
- Log actions via `shared.Logger`

Hooks run as background processes invoked by Claude Code via hook executables in `.claude/hooks/`.
