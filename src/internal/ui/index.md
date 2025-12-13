# Terminal UI Components

Bubble Tea-based interactive terminal UI for Claudex session management workflows.

## Key Files

- **ui.go** - Bubble Tea models, delegates, and UI workflows

## Key Types

- `Model` - Bubble Tea model for session/profile selection with multi-stage support
- `SessionItem` - List item type (from session package) representing sessions, profiles, or menu options
- `ItemDelegate` - Custom list delegate for rendering items with icons and descriptions
- Message types: `SessionChoiceMsg`, `ProfileChoiceMsg`, `ResumeOrForkChoiceMsg`, `ResumeSubmenuChoiceMsg`

## Usage

The UI module provides interactive selection lists for sessions, profiles, and menu choices. It handles multiple workflow stages (session selection, profile selection, resume-or-fork decision, resume submenu). Also includes non-interactive helper functions for prompting descriptions, showing progress messages, and displaying success confirmations.

See [../services/session/](../services/session/) for session management integration and [../../cmd/](../../cmd/) for CLI entry points.
