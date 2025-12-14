# doc

Documentation generation and update services for Claudex sessions.

## Core Files

- `interface.go` - DocumentationUpdater interface definition
- `updater.go` - Background Claude invocation for documentation updates
- `transcript.go` - JSONL transcript parsing and formatting
- `prompts.go` - Prompt template loading and building

## Subdirectories

- `rangeupdater/` - Range-based documentation updates using Git commit ranges

## Tests

- `transcript_test.go` - Tests for transcript parsing
- `prompts_test.go` - Tests for prompt template handling
- `updater_test.go` - Tests for the documentation updater
