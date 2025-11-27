In Progress:
  - Add engineer with flexible skills (typescript, python, etc)
    - Goal: Refactor monolithic engineer agent definitions into composable parts to support various stacks (TypeScript, Go, PHP) and specializations (Frontend, Backend, Fullstack).
    - Context: Currently, agents in `claudex-go/profiles/agents` (like `principal-typescript-engineer`) are monolithic. We want to decompose them using the `profiles/roles` and `profiles/skills` structure.
    - Implementation Plan:
      - Extract the base "Engineer" role into `profiles/roles/engineer.md`. This should contain:
        - Core workflow (Activation, Pre-Implementation, Execution, Testing, Story Updates).
        - Common orchestration interface (Reporting, Escalation).
        - Standard output formats and commands.
      - Extract specific technical skills into `profiles/skills/`.
        - Ensure `typescript.md` is complete.
        - Add `python.md`, `go.md`, `php.md`, `frontend.md`, `backend.md` as needed.
      - Create composable agent definitions in `profiles/agents/` that combine the Engineer role with specific skills.
      - Ensure the composition mechanism works (e.g., an agent file that imports/references the role and multiple skills).

To Do: 
  - Adjust Agents' models:
    - Documenter hook: haiku
    - Architect: opus
    - Researcher: haiku/sonnet
    - Engineer: sonnet/opus
    - Prompt Engineer: sonnet/opus
  - Analyze all hooks via loggers: /hooks (look into Stop, SubagentStart)
    - Documenter:
      - Analyze if documenter could be better if triggered on PreCompact or Stop
      - Idea: With autocompact disabled, when context is full you get an error. Use gemini to read the transcript and produce or update the overview document. Remove from the transcript tools executions and similar non-relevant things (like I did with claudex already). It's important that gemini produce several docs, being the overview a small one with pointers to be loaded only on demand.
  - Find a solution to allow the user to define where the relevant documentation is located (product, architecture, standards, etc)
  - Review all agents to adjust their output format to make sure they provide enough context to the caller (team lead) but avoid verbose responses
  - Create QA that is responsible for defining the cases to be covered by the test and evals suite. They will receive the definition of the feature as input and come up with the test and evals suite definition at a product/business level. The QA should execute in parallel of the Architect.
  - Refactor resume session feature:
    - Resume should: 1. ask the user if they want to start with fresh memory or continue with the previous one; 2.1 if continue is chosen the execution is like the one we have currently, 2.2 if fresh memory is chosen then a new session-id is generated, a new folder is created as a clone of the previous one and  the previous one is removed (this is what makes it different from "fork", which keeps both folders). In case of "fork", we need to refactor to allow the user to enter a description like with new sessions, with the difference that the current session folder is cloned and kept.


Done:
  - Enable macOS notifications for Claude Code subagents and Claude notifications. Add voice + visual alerts when subagent tasks complete using the SubagentStop hook and when Claude sends a Notification via notification hook.
  - Refactor resume session feature:
    - Resume should: 1. ask the user if they want to start with fresh memory or continue with the previous one; 2.1 if continue is chosen the execution is like the one we have currently, 2.2 if fresh memory is chosen then a new session-id is generated, a new folder is created as a clone of the previous one and  the previous one is removed (this is what makes it different from "fork", which keeps both folders). In case of "fork", we need to refactor to allow the user to enter a description like with new sessions, with the difference that the current session folder is cloned and kept.
