# Session Module

Session management and metadata operations.

## Key Files
- **session.go** - Session retrieval and listing (GetSessions, UpdateLastUsed)
- **naming.go** - Session name generation and Claude session ID utilities
- **finder.go** - Session folder discovery by ID (FindSessionFolder, FindSessionFolderWithCwd)
- **metadata.go** - Session metadata file operations (description, timestamps)
- **counter.go** - Doc update frequency counter (IncrementCounter, ResetCounter)
- **types.go** - SessionItem type for UI display

## Key Types
- `SessionItem` - Session metadata for UI display and operations
- `SessionMetadata` - Metadata files (description, created, last_used)

## Usage

The session module provides all session-related operations: listing sessions, finding session folders by ID, managing metadata files, and tracking autodoc update frequency. Used by app orchestration and hooks for context-aware operations.
