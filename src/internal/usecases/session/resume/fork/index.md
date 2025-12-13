# Fork Session Usecase

Forks existing sessions with new descriptions while preserving all session data.

## Key Files

- **fork.go** - Session forking workflow

## Key Types

- `UseCase` - Handles forking of existing sessions

## Usage

The `Execute` method creates a forked session:
1. Generates a new UUID for the forked session
2. Generates new session name from the new description
3. Copies the entire original session directory to new location
4. Updates .description file with new description
5. Returns forked session name, path, and Claude session ID
