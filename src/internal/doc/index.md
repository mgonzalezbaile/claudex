# Documentation Updater

Background documentation generation system that invokes Claude to update session documentation based on transcript increments.

## Key Files

- **interface.go** - DocumentationUpdater interface
- **updater.go** - Updater implementation with background/synchronous execution
- **transcript.go** - JSONL transcript parsing and formatting
- **prompts.go** - Prompt template loading and variable substitution

## Key Types

- `DocumentationUpdater` - Interface for synchronous and background doc updates
- `Updater` - Concrete implementation using filesystem, commander, and environment services
- `UpdaterConfig` - Configuration for doc updates (paths, model, start line)
- `TranscriptEntry` - Parsed JSONL entry (assistant messages and agent results)

## Usage

The updater parses transcript JSONL files incrementally, formats relevant entries (assistant messages and completed agent results) into markdown, and invokes Claude CLI with a prompt template to update documentation. It uses a recursion guard (`CLAUDE_HOOK_INTERNAL=1`) to prevent infinite loops when Claude invokes hooks.

See [../services/hooks/](../services/hooks/) for integration with session hooks.
