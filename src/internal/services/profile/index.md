# Profile Module

Profile loading and composition for Claudex agents.

## Key Files
- **profile.go** - Profile loading from embedded FS and .claude/agents/ directory

## Key Functions
- `GetProfiles` - Returns sorted list of all available profiles
- `LoadComposed` - Loads profile from embedded FS, then fallback to filesystem
- `ExtractDescription` - Extracts role description from profile content

## Usage

The profile module handles agent profile discovery and loading. It supports two sources: embedded FS (profiles/agents/) and filesystem (.claude/agents/), with composition allowing filesystem overrides. Used by App to load team-lead profile during initialization.
