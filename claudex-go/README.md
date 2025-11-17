# Claudex Session Manager

A modern, interactive session manager for Claude Code, built in Go with a beautiful TUI.

## Features

- 🎨 **Beautiful TUI** - Modern interface built with Bubble Tea framework
- 📋 **Session Management** - Create, resume, and manage work sessions
- 🤖 **AI-Generated Names** - Claude suggests session names based on descriptions
- ⚡ **Ephemeral Mode** - Work without saving session data
- 🎭 **Profile Selection** - Choose different Claude configurations
- 🔍 **Fuzzy Search** - Filter sessions and profiles by typing `/`
- 📊 **Session Metadata** - Track descriptions and creation dates
- 📱 **Responsive** - Adapts to terminal width automatically
- ⌨️ **Keyboard Controls** - Full keyboard navigation (↑↓, Enter, q, Ctrl+C)

## Installation

```bash
cd claudex-go
make install
```

Or manually:
```bash
go mod tidy
go build -o claudex-session main.go
chmod +x claudex-session
```

## Usage

```bash
./claudex-session
```

The program will guide you through:
1. **Session Selection** - Create new, use ephemeral mode, or resume an existing session
2. **Profile Selection** - Choose a Claude profile/configuration
3. **Launch Claude** - Automatically launches Claude Code with your selections

### Keyboard Controls

**Session/Profile Lists:**
- `↑/↓` - Navigate menu
- `Enter` - Select option
- `/` - Start fuzzy search
- `q` or `Ctrl+C` - Quit

**Create New Session:**
- Type description and press `Enter`
- Claude will generate a slug-based name automatically

## Session Data

Sessions are stored in `./sessions/` with:
- `.description` - Your session description
- `.created` - ISO 8601 timestamp

Example:
```
sessions/
├── auth-refactor/
│   ├── .description
│   └── .created
└── api-performance-fix/
    ├── .description
    └── .created
```

## Environment Variables

The program sets these variables for your Claude session:
- `CLAUDEX_SESSION` - Current session name (or "ephemeral")
- `CLAUDEX_SESSION_PATH` - Full path to session directory

## Directory Structure

```
.
├── claudex-session         # The executable
├── sessions/               # Session storage (auto-created)
└── .profiles/              # Claude profile configurations (required)
    ├── default.md
    ├── architect.md
    └── engineer.md
```

## Session Name Generation

When creating a session:
1. Enter a description (e.g., "Working on user authentication")
2. Claude generates a slug (e.g., "user-auth-module")
3. Falls back to auto-generated slug if Claude unavailable
4. Ensures uniqueness by appending numbers if needed

## Building

Use the Makefile for easy building:

```bash
make              # Build claudex-session
make deps         # Install/update dependencies
make install      # Build and mark executable
make clean        # Remove build artifacts
make run          # Build and run
make help         # Show all targets
```

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/charmbracelet/bubbles` - TUI components

## Example Workflow

```bash
$ ./claudex-session

# Beautiful TUI appears
# Select "➕ Create New Session"
# Enter: "Refactoring authentication module"
# Auto-generated: "auth-refactor"

# Select profile "🎭 principal-architect.md"
# Claude launches with session context

# Later...
$ ./claudex-session

# Select "📁 auth-refactor"
# Continue where you left off
```

## Troubleshooting

**Issue:** Profiles directory not found
```bash
mkdir .profiles
# Add your profile markdown files here
```

**Issue:** Claude command not found
```bash
# Install Claude CLI first
# See: https://claude.ai/cli
```

**Issue:** Terminal artifacts when launching Claude
- This is fixed with proper terminal restoration
- Wait for "Launching Claude..." message before interaction

## Technical Details

- **Framework:** Bubble Tea (Elm architecture for Go)
- **Alt-screen:** Uses alternate screen buffer for clean UI
- **Terminal restoration:** Properly restores terminal state before launching Claude
- **Message passing:** Uses Bubble Tea's message system for state management
- **Fuzzy search:** Built-in filtering for quick navigation

## License

See main project LICENSE

## Credits

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
