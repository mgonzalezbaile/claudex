# Setup Usecase

Initializes .claude directory structure with hooks, agents, commands, and project-specific configuration.

## Key Files

- **setup.go** - Main setup workflow orchestration
- **agents.go** - Agent profile assembly from role and skill templates

## Key Types

- `SetupUseCase` - Orchestrates .claude directory setup workflow

## Usage

The `Execute` method creates the complete .claude directory structure:
1. Creates hooks/, agents/, and commands/agents/ directories
2. Copies hooks from ~/.config/claudex/hooks/
3. Copies agent profiles to both agents/ and commands/agents/
4. Detects project stacks and generates principal-engineer-{stack} agents
5. Creates settings.local.json with hooks configuration
