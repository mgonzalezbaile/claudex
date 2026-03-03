# Product Definition: Claudex

## 1. Overview
**Claudex** (Cloud Experience) is an advanced wrapper and enhancement layer built on top of **Claude Code** (by Anthropic). While Claude Code provides a powerful CLI interface for AI-assisted development, Claudex aims to elevate this experience by introducing persistent session management, persona-based workflows, and dynamic context maintenance through background hooks.

The core philosophy of Claudex is to treat an AI interaction not just as a transient CLI stream, but as a stateful, manageable project asset.

### 1.1. Prerequisites
Before using Claudex, the following dependencies are required:
*   **Claude Code CLI:** The native `claude` command-line tool must be installed and authenticated (`claude auth login`).
*   **Go Runtime:** Required to build and run the `claudex-go` session manager (Go 1.21+ recommended).
*   **Unix-like Environment:** The installer and shell integration currently support macOS and Linux (Bash/Zsh).

## 2. Core Value Proposition
Claudex addresses several limitations of the native Claude Code CLI:
*   **Session Persistence:** Raw Claude sessions are often transient. Claudex persists sessions as folders on the filesystem, allowing users to stop, resume, and inspect context at any time.
*   **Context Management:** By storing session data in files, users can clean the immediate context window while retaining the "memory" of the session in the file system.
*   **Persona/Role Injection:** Users can instantiate Claude with specific roles (e.g., "Team Lead", "Architect", "Lawyer") without manually prompting for them every time.
*   **Forking Capabilities:** A unique ability to "branch" a conversation—taking an existing context and diverging into a new direction without losing the original state.

## 3. Architecture

Claudex consists of two main components:

### 3.1. The Installer (`claudex-install`)
The entry point for integrating Claudex into a user's workspace.
*   **Function:** Installs Claudex into a specific repository or folder.
*   **Mechanism:** Generates and copies (or symlinks) configuration files into a `.claude` directory within the target workspace.
*   **Goal:** Ensures that when Claudex runs, it has access to workspace-specific settings and context files.

### 3.2. The Runner & Session Manager (`claudex-go`)
The main executable and user interface (TUI).
*   **Function:** Acts as the gateway to the Claude CLI.
*   **Features:**
    *   **Dashboard:** A TUI (Terminal User Interface) for selecting actions.
    *   **Lifecycle Management:** Handles creating New sessions, Resuming existing ones, and Forking.
    *   **Ephemeral Mode:** Offers a "quick start" mode for transient tasks where no persistent session folder or history is required.
    *   **Dynamic Naming:** Uses Claude itself to generate descriptive slugs for sessions based on user descriptions.
    *   **Profile Selection:** Allows users to inject specific system prompts (Profiles) at startup.

### 3.3. Context Synchronization (Native Hooks)
Instead of using a custom proxy, Claudex leverages the built-in **hooks** functionality of Claude Code.
*   **Function:** Automatically triggers background processes during the conversation lifecycle.
*   **Mechanism:** Claudex configures native Claude hooks to execute maintenance scripts when specific events occur (e.g., after tool use or periodically).
*   **Goal:** Keeps the file-based session context (roadmap, architecture docs, todo lists) synchronized with the active conversation, ensuring the "Session Folder" remains a truthful artifact of the work done.

## 4. User Workflows

### 4.1. Installation
1.  User navigates to a project folder (e.g., `~/my-project`).
2.  User executes the Claudex installer.
3.  Claudex creates/configures the `.claude` folder, establishing the "environment" for the AI.

### 4.2. Session Initiation
When the user runs `claudex`, they are presented with the Session Manager:
*   **New Session:**
    1.  User provides a description (e.g., "Refactor the login API").
    2.  Claudex generates a unique session slug (e.g., `login-refactor-v1`).
    3.  User selects a **Profile** (e.g., "Senior Go Engineer").
    4.  Claudex launches Claude Code with the profile loaded as a system prompt.
*   **Resume Session:**
    1.  User selects an existing session from the list.
    2.  Claudex re-attaches to the specific Claude Session ID associated with that folder.
    3.  Context and history are preserved.
*   **Fork Session:**
    1.  User selects an existing session.
    2.  Claudex clones the session folder (including all context files).
    3.  A new, independent session starts, inheriting the history of the parent but allowing for new divergence.

### 4.3. Execution
During the session:
1.  **Profile Enforcement:** The selected profile (Agent) guides the behavior of Claude throughout the session.
2.  **Automated Context Sync:** As the user interacts with Claude, the configured native hooks execute in the background. These hooks monitor the conversation flow and automatically update the session's context files (e.g., updating `roadmap.md` when a task is completed), ensuring the file system always reflects the current state of the session.

## 5. Technical Concepts

### Session as a Folder
A Session is defined by a directory in `.claudex/sessions/`.
*   Contains metadata: `.description`, `.created`, `.last_used`.
*   Contains context artifacts: Any files generated or modified during the specific session.
*   Contains state: Links to the upstream Claude API Session ID (UUID).

### Profiles
Profiles are template definitions stored in `.profiles/`. They primarily consist of system prompts that define:
*   **Tone:** (e.g., Strict, Educational, Concise).
*   **Domain Knowledge:** (e.g., Legal constraints, Golang best practices).
*   **Format Requirements:** (e.g., "Always output code in blocks").

## 6. Standard Profile Library
Claudex ships with a set of "Standard Agents" designed to work together:

### 6.1. Team Lead (The Orchestrator)
*   **Role:** Manager and Delegator.
*   **Behavior:** Does **not** write code or execute implementation tasks.
*   **Responsibilities:**
    *   Analyzes requirements.
    *   Delegates tasks to the best-skilled specialized agent.
    *   **Parallelization:** Maximizes efficiency by spawning multiple agents (e.g., Frontend + Backend) simultaneously.

### 6.2. Principal Software Engineer
*   **Role:** The Builder.
*   **Responsibilities:**
    *   Executes the plan provided by the Plan agent.
    *   **Skill Injection:** Can be loaded with specific "Skill Mixins" (e.g., Python Expert, TypeScript Expert, Go Expert, PHP Expert) to match the project stack.

### 6.3. Principal AI/Prompt Engineer
*   **Role:** AI & Evals Expert.
*   **Responsibilities:**
    *   Specializes in prompt engineering and LLM evaluations ("evals").
    *   Optimizes agent interactions and system prompts.

### 6.4. QA Engineer
*   **Role:** Quality Assurance.
*   **Responsibilities:**
    *   Validates implementation against requirements.
    *   Writes test cases and ensures coverage.

### 6.5. Context Curator (Background Agent)
*   **Role:** Documentation & Memory.
*   **Behavior:** Triggered automatically via hooks (not manually invoked).
*   **Responsibilities:**
    *   Gathers output from executed agents.
    *   Updates session documentation (roadmaps, decision logs) to keep the context fresh.
