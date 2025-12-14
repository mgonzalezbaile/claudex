# New Session Usecase

Creates new sessions with generated names, unique UUIDs, and metadata files.

## Key Files

- **new.go** - New session creation workflow

## Key Types

- `UseCase` - Handles creation of new sessions

## Usage

The `Execute` method creates a new session directory with metadata:
1. Generates a UUID for the Claude session
2. Generates session name from description (via Claude CLI or manual slug)
3. Creates session directory with UUID suffix
4. Writes .description and .created timestamp files
5. Auto-creates initial session-overview.md with session summary and timeline
6. Returns session name, path, and Claude session ID
