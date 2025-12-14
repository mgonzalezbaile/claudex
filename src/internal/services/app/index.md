# App Package

Main application container for claudex CLI.

## Core

- `app.go` - App struct with Init/Run/Close lifecycle, config loading, logging setup, hook/MCP setup prompts
- `deps.go` - Dependencies struct for dependency injection (FS, Cmd, Clock, UUID, Env)

## Launch

- `launch.go` - Session launch modes (new, resume, fork, fresh, ephemeral) and Claude CLI invocation
- `session.go` - Session selector TUI and handlers for new/resume/fork workflows

## Setup Flows

- `promptHookSetup()` - Interactive git hook integration setup (auto-docs on commits)
- `promptMCPSetup()` - Interactive MCP configuration for recommended MCPs (sequential-thinking, context7) with optional Context7 API token

## Tests

- `app_test.go` - Tests for App initialization and run logic
- `launch_test.go` - Tests for launch modes and Claude invocation
