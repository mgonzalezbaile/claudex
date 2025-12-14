# App Package

Main application container for claudex CLI. Handles initialization, session lifecycle, and npm distribution setup.

## Core

- `app.go` - App struct with Init/Run/Close lifecycle, config loading, logging setup, hook/MCP setup prompts
- `deps.go` - Dependencies struct for dependency injection (FS, Cmd, Clock, UUID, Env)

## Startup Validation

- `isClaudeInstalled()` - Checks if Claude CLI is available in PATH
- Claude CLI installation prompt - If missing, prompts user to install `@anthropic-ai/claude-code` via npm; automatically continues with normal flow after successful installation (or returns error if declined)

## Launch

- `launch.go` - Session launch modes (new, resume, fork, fresh, ephemeral) and Claude CLI invocation
- `session.go` - Session selector TUI and handlers for new/resume/fork workflows

## Setup Flows

- `promptUpdateCheck()` - Checks for newer versions of claudex and prompts user to update (with never-ask-again option)
- `promptHookSetup()` - Interactive git hook integration setup (auto-docs on commits)
- `promptMCPSetup()` - Interactive MCP configuration for recommended MCPs (sequential-thinking, context7) with optional Context7 API token

## NPM Distribution

Post-install setup is now handled by Node.js scripts in the npm package (`@claudex/cli`):
- `postinstall.js` - Detects platform, links platform-specific binaries from optional dependencies
- Platform packages: `@claudex/darwin-arm64`, `@claudex/darwin-x64`, `@claudex/linux-x64`, `@claudex/linux-arm64`

## Tests

- `app_test.go` - Tests for App initialization and run logic
- `launch_test.go` - Tests for launch modes and Claude invocation
