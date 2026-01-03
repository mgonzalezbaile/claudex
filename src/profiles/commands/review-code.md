# Review Code Command

Perform comprehensive code review by orchestrating specialist agents. This command analyzes code quality, identifies bugs, suggests improvements, and provides actionable feedback through parallel delegation to language-specific experts.

## Step 1: Gather Context

Run git commands to identify what files need review:

```bash
# Check git status for staged and unstaged changes
git status --short

# Show staged changes with stats
git diff --cached --stat

# Show unstaged changes with stats
git diff --stat

# Show recent commits on current branch
git log --oneline -10

# Show changes in last commit
git diff HEAD~1 --stat
```

## Step 2: Clarify Scope

If the scope is ambiguous, ask the user to specify:

**Questions to ask:**
- Review staged changes?
- Review specific files or directories?
- Review recent commits (how many)?
- Review all uncommitted changes?
- Review specific pull request changes?

If the scope is clear from context (e.g., user said "review my staged changes"), proceed automatically.

## Step 3: Classify Files by Language

Group the files to review by their primary language/technology:

**File Extension Mapping:**
- `.ts`, `.tsx`, `.js`, `.jsx`, `.mjs`, `.cjs` → `principal-engineer-typescript`
- `.go` → `principal-engineer-go`
- `.py` → `principal-engineer-python`
- Prompt files (`.md` in `src/profiles/`, agent profiles, commands, tasks) → `prompt-engineer`
- Other extensions → `principal-engineer` (generic fallback)


## Step 4: Delegate to Specialist Agents

For each language group, spawn the appropriate specialist agent using the Task tool with this standardized review task:

```markdown
## Code Review Task

**Files to Review:**
- [list of file paths for this language]

**Review Criteria:**

Analyze the code changes and provide feedback in these categories:

1. **Critical Issues** 🔴 - Must fix before merging
   - Bugs, logic errors, edge case failures
   - Security vulnerabilities
   - Breaking changes or regressions

2. **Important Improvements** 🟡 - Should address
   - Code quality and readability issues
   - Performance concerns
   - Maintainability problems
   - Missing error handling

3. **Minor Suggestions** 🟢 - Optional improvements
   - Style and formatting
   - Documentation enhancements
   - Minor optimizations

4. **Positive Highlights** ✅ - What was done well
   - Good patterns and practices
   - Clever solutions
   - Excellent test coverage

**Output Format:**

Use this exact structure:

```markdown
# Code Review: [Language/Stack]

## Overall Assessment
[2-3 sentence summary of code quality]

## Critical Issues 🔴
[If none, state "None identified"]

- **[File:Line]** [Issue description]
  - Impact: [what could go wrong]
  - Recommendation: [how to fix]

## Important Improvements 🟡
[If none, state "None identified"]

- **[File:Line]** [Issue description]
  - Impact: [why this matters]
  - Recommendation: [suggested improvement]

## Minor Suggestions 🟢
[If none, state "None identified"]

- **[File:Line]** [Suggestion]
  - Benefit: [why this helps]

## Positive Highlights ✅
[If none, state "None identified"]

- [What was done well]

## Test Coverage Assessment
- [Assessment of test completeness]
- [Missing test scenarios if any]
```

**Important Notes:**
- Be specific with file:line references
- Provide actionable recommendations
- Balance criticism with recognition of good work
- Consider the project's conventions and context
```

**Delegation Strategy:**
- Spawn agents in PARALLEL when reviewing multiple language groups
- Each agent reviews only their language's files
- Wait for all agents to complete before proceeding

## Step 6: Aggregate Results

Collect the review reports from all specialist agents and combine them into a unified summary:

```markdown
# Code Review Summary

**Reviewed:** [file paths or commit range]
**Reviewers:** [list of specialist agents used]
**Date:** [current date]

## Overall Assessment

[Synthesize the overall code quality from all specialist reports]

## Critical Issues 🔴

[Aggregate all critical issues from all agents]
[If none across all reviews, state "None identified"]

## Important Improvements 🟡

[Aggregate all important improvements from all agents]
[If none across all reviews, state "None identified"]

## Minor Suggestions 🟢

[Aggregate all minor suggestions from all agents]
[If none across all reviews, state "None identified"]

## Positive Highlights ✅

[Aggregate all positive highlights from all agents]

## Test Coverage

[Synthesize test coverage assessments from all agents]

## Next Steps

1. [Prioritized action items based on severity]
2. [Suggested workflow - e.g., "Fix critical issues, then re-run tests"]

## Review Statistics

- Files reviewed: [N]
- Lines changed: [+X -Y]
- Critical issues: [N]
- Important improvements: [N]
- Minor suggestions: [N]
- Languages reviewed: [list]
```

## Step 7: Present to User

Display the aggregated review summary to the user and offer to:
- Deep dive into specific issues
- Review related code for consistency
- Provide refactoring examples
- Re-review after fixes are applied

## Important Notes

**Orchestration Principles:**
- Keep the command lean - delegate detailed analysis to specialists
- Spawn agents in parallel to maximize efficiency
- Each agent is the expert in their domain
- Aggregate results coherently for user-friendly output

**Edge Cases:**
- Single file reviews: Still delegate to appropriate specialist
- Mixed language changes: Spawn multiple specialists in parallel
- No changes found: Inform user clearly
- Prompt/agent file changes: Use `prompt-engineer` for specialized review

**When to Use This Command:**

- Before committing significant changes
- Before creating a pull request
- After implementing a new feature
- When refactoring existing code
- Before releases to review critical paths
- When learning a codebase (review recent quality work)

Regular code reviews catch issues early, improve quality, and maintain project health.
