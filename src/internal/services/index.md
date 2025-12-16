Now I understand the structure and style. Let me generate the updated content:

```markdown
# Services

Application service layer providing infrastructure and domain-specific operations.

## Core Dependencies

- `app/` - Main application container, session lifecycle, and Claude CLI integration
- `config/` - TOML configuration loading and parsing for .claudex.toml files
- `settings/` - Claude Code settings.json management and smart-merge operations

## MCP Configuration

- `globalprefs/` - Global MCP and update check preferences stored in ~/.config/claudex/mcp-preferences.json
- `mcpconfig/` - MCP server configuration management in Claude Code's ~/.claude.json file

## Package Management

- `npmregistry/` - npm registry client for querying latest package versions

## Infrastructure

- `clock/` - Time abstraction for testability
- `commander/` - Process execution abstraction (Run, Start)
- `env/` - Environment variable access abstraction
- `filesystem/` - Directory copy, file search, and existence checks with afero
- `uuid/` - UUID generation abstraction

## Git & Version Control

- `git/` - Git operations (commit SHA, changed files, merge base, commit validation)
- `hooksetup/` - Post-commit hook installation for documentation updates

## Session & State

- `session/` - Session retrieval, listing, naming, and metadata operations
- `doctracking/` - Documentation update tracking state (last commit, timestamps)
- `lock/` - File-based cross-process locking with atomic acquisition
- `preferences/` - Project preferences storage (.claudex/preferences.json)

## Detection & Profiles

- `profile/` - Agent profile loading and composition from embedded/filesystem sources
- `stackdetect/` - Technology stack detection (TypeScript, Go, Python, React Native, PHP) via marker files
```
