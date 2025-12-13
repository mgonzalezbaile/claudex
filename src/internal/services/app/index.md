# App Module

Application lifecycle and orchestration for Claudex.

## Key Files
- **app.go** - Main App container with initialization and lifecycle management
- **launch.go** - Claude CLI launch orchestration for different session modes
- **session.go** - Session selector TUI and session workflow handlers
- **deps.go** - Dependency injection container for testable external dependencies

## Key Types
- `App` - Main application container managing dependencies and lifecycle
- `LaunchMode` - Enum for session launch modes (new, resume, fork, fresh, ephemeral)
- `SessionInfo` - Session state holder passed between methods
- `Dependencies` - Dependency injection container (FS, Cmd, Clock, UUID, Env)

## Usage

The App module orchestrates the entire application flow: initialization, session selection, and Claude CLI launching. It uses dependency injection for testability and coordinates between UI, usecases, and service layers.
