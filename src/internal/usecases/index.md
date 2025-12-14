# Use Cases

Core business logic orchestrating features like session management, setup workflows, documentation generation, and git hook integration.

## Modules

- **createindex/** - Generate index.md documentation files for any directory using Claude
- **migrate/** - Migrate legacy Claudex artifacts to .claudex/ directory structure and create defaults
- **session/** - Session lifecycle management (create, resume fresh, resume fork)
- **setup/** - Initialize .claude directory structure with hooks, agents, and configuration
- **setuphook/** - Git hook installation detection and user preference management
- **setupmcp/** - Prompt users about MCP configuration with opt-in flow and preference management
- **updatecheck/** - Check for newer versions of @claudex/cli and prompt users for updates
- **updatedocs/** - Update index.md documentation based on git history changes
