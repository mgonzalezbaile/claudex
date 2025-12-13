# Fresh Memory Session Usecase

Creates fresh memory sessions by copying session data, clearing history, and deleting the original.

## Key Files

- **fresh.go** - Fresh memory session workflow

## Key Types

- `UseCase` - Handles creating fresh memory sessions from existing sessions

## Usage

The `Execute` method creates a fresh memory session:
1. Generates a new UUID for the fresh session
2. Strips Claude session ID from original name to preserve base slug
3. Copies session directory with new UUID suffix
4. Removes tracking files (.last-processed-line, .last-processed-line-overview)
5. Resets .doc-update-counter to 0
6. Deletes the original session directory
7. Returns fresh session name, path, and Claude session ID
