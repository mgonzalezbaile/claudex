# Claudex Development Notes

## Task Status

### 🚀 In Progress
- [ ] **Dynamic Profiles ("Custom")**:
  - [ ] Generate agent definition dynamically based on user description

### 📋 ToDo
- **Profile Composition Engine**:


---

## Overview
This document tracks feature ideas, architectural decisions, and implementation notes for the Claudex framework.

## Architecture Considerations

### Session Directory Execution
Execute `claude` directly in the session directory to give it focused access to accumulated context. Benefits:
- Direct access to all session-specific context
- Avoids overloading with irrelevant project files
- Subagents inherit the same focused context
- Reduces context pollution from the broader project

## Feature Roadmap

### Session Management (Core)
- [x] **Dashboard / TUI**: Terminal UI for managing the lifecycle.
- [x] **Refine Team Lead Agent Context Passing**: Ensure Team Lead agent always passes complete context to delegated agents.
- [x] **Enforce Session-Scoped Documentation**: Update all profiles and agents to ensure documentation output is always created within the session's folder.
- [x] **Automatic Documentation Updates**: Implemented via `SubagentStop` hook to capture agent execution summaries in `session-history.md`.
- [x] **Create New Session**:
  - [x] Input description
  - [x] Auto-generate session slug (via Claude)
  - [x] Create persistent folder in `sessions/`
- [x] **Resume Session**:
  - [x] List existing sessions (sorted by recent use)
  - [x] Re-attach to Claude Session ID
- [x] **Fork Session**:
  - [x] Clone session folder and artifacts
  - [x] Create new independent session
  - [ ] Allow user to enter a new description for the forked session
- [x] **Ephemeral Session**:
  - [x] Run Claude without persistent storage/history

### Profile & Agent System
*Philosophy: Keep system prompts minimal for maximum reliability. Logic should be split into composable blocks, and deep context should be loaded on-demand, not pre-loaded.*

**IMPORTANT NOTE**: All documentation files must be created in the session's folder. All agents should look into the session's folder to gather context before exploring the codebase.

- [x] **Profile Composition Engine**:
  - [x] **Base Template**: Define shared building blocks (Tone, Format, Rules) common to all agents.
  - [x] **Role Definitions**: Distinct files for each persona (e.g., `roles/architect.md`).
  - [x] **Skill Mixins**: Reusable blocks for specific tech stacks (e.g., `skills/typescript.md`, `skills/python.md`).
  - [x] **Assembly Logic**: Update `claudex-go` to dynamically assemble (Template + Role + Skills) at startup.

- [ ] **Context Injection & Documentation**:
  - [ ] **Context Map**: Define a standard (e.g., `.claudex/context.md`) where users list key project docs (Standards, Features).
  - [ ] **On-Demand Loading Patterns**: Instructions for agents to *search/read* specific docs only when relevant task arises (Lazy Loading).
  - [ ] **User Overrides**: Allow users to inject custom "knowledge files" into the session scope at startup.

- [ ] **Standard Agent Library (Built-in)**:
  - [ ] **Team Lead**: Orchestrator. Focus on delegation and **Aggressive Parallelization**.
    - [ ] **Context Passing**: Always pass complete context to delegated agents (files, directories, research findings)
    - [ ] **Avoid Redundant Search**: Agents should not need to re-search codebase/internet for information Team Lead already has
  - [ ] **Researcher**: Current `architect-assistant`. Analyst. Focus on deep research (code/docs) and producing **Research Documents**.
  - [ ] **Architect**: Planner. Focus on creating **Execution Plans** based on Researcher output.
    - Review .bmad-core/templates/execution-plan-tmpl.yaml to make it more concise
    - Create plan optimized for agents parallelization
  - [ ] **Principal Software Engineer**: Builder. Focus on implementation. Supports **Skill Mixins** (Python, TS, etc.).
  - [ ] **Principal AI/Prompt Engineer**: Expert in Evals, Prompting, and LLM Systems.
  - [ ] **QA Engineer**: Validator. Focus on test coverage and quality assurance.
  - [ ] **Context Curator** (Background): Agent responsible for automatic documentation updates via hooks.

- [x] **Profile Selection**: Choose from built-in profiles at startup.
- [x] **Profile Loading**: Inject profile content as system prompt.
- [ ] **Dynamic Profiles ("Custom")**:
  - [ ] Generate agent definition dynamically based on user description
  - [ ] **Agent Templates**: Define blocks that can be filled based on user description
- [ ] **"None" Profile**: Option to start without any system prompt.
- [ ] **Multi-tool support**: Execute prompts with alternative AI tools (Gemini, etc.)

### Context Management & Hooks (Native)
*Leveraging Claude Code's built-in hooks functionality.*

- [ ] **Native Hooks Configuration**:
  - [ ] `PreToolUse` / `PostToolUse` hooks
  - [ ] `SessionStart` / `SessionEnd` hooks
- [ ] **Prompt Interception & Enhancement**:
  - [ ] **`PreToolUse` for Task tool**: Intercept subagent spawning to modify prompts
  - [ ] **Dynamic Context Injection**: Auto-inject relevant docs/context based on subagent_type
  - [ ] **Prompt Templates**: Apply agent-specific prompt enhancements before execution
  - [ ] **Validation Layer**: Enforce prompt structure/requirements before spawning
- [ ] **Context Refresh**:
  - [ ] **`/reload-context`**: Refresh session context without losing state.
  - [ ] **`/exit` hook**: Summarize session upon exit.
- [ ] **Automatic Documentation Updates**:
  - [ ] **`PostToolUse` hook**: Background process triggered after every tool execution
  - [ ] Captures tool execution output to keep session documentation current
  - [ ] Ensures documentation stays synchronized with session progress
- [ ] **Automatic Execution Plan Tracking** *(Needs Discovery)*:
  - [ ] Background process to keep execution plan documents up-to-date
  - [ ] Track multi-phase plans (Phase 1, Phase 2, etc.) with task checklists
  - [ ] Automatically update task status as agents make progress
  - [ ] Discovery needed: trigger conditions, update frequency, integration points

### Infrastructure & Installation
- [x] **Installer Script**:
  - [x] Link `.claude` configuration to target workspace
  - [x] Setup global profiles directory
- [ ] **Global vs Local Profiles**: Implement "Cascading Configuration" (Local overrides Global).
- [ ] **Enable MCPs**: Configure default MCPs during installation.

## Detailed Ideas

### Session Lifecycle Hooks
- **`/exit` hook**: Capture session end
  - Summarize session and create resumption file
  - Run in background to avoid blocking user
- **SubagentStop hook**: Capture agent execution results
  - Update session context when `message.stop_reason == end_turn`
  - **Smart documentation**: Only create/update docs when truly valuable

### Hook System Ideas
- **preCompact hook**: Intercept before context compaction to save state.
- **Command hooks**: Pass session path to all commands.
- **Custom command hooks**: Integrate external tools (Gemini, Codex).

## Directory Structure
```
.claudex/
├── profiles/       # Global Agent/Profile definitions
├── context/        # (Proposed) Global context maps
├── sessions/       # (In project) Active and archived sessions
└── hooks/          # (Proposed) Hook scripts
```

## Technical Notes

### Session Management Commands
```bash
# Create and activate a session
session_id=$(claude --system-prompt "prompt" "activate" --output-format json | jq -r '.session_id')

# Resume a session
claude --resume $session_id
```

### Prompt Interception via PreToolUse Hook
The `PreToolUse` hook for the `Task` tool receives the complete subagent invocation:

```json
{
  "hook_event_name": "PreToolUse",
  "tool_name": "Task",
  "session_id": "...",
  "tool_input": {
    "description": "Task description",
    "subagent_type": "architect",
    "prompt": "Original prompt text..."
  }
}
```

**Implementation possibilities**:
- Modify `tool_input.prompt` to inject session-specific context
- Add standard instructions based on `tool_input.subagent_type`
- Inject relevant documentation references from session folder
- Enforce prompt structure/templates for consistency
- Return modified JSON to Claude Code with enhanced prompt

**Use cases**:
- Auto-inject "Read session docs first" instructions
- Add agent-specific guidelines (e.g., "Focus on parallelization" for team-lead)
- Append coding standards or architectural constraints
- Include references to recent session artifacts

## Current Issues & Feedback
- **File organization**: Claude currently generates documents anywhere; should be constrained to session folder.
- **Context Sync**: Needs the native hooks implementation to be fully automatic.
