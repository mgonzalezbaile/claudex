<!-- Powered by BMAD™ Core -->

# Create Execution Plan Document

## ⚠️ CRITICAL EXECUTION NOTICE ⚠️

**THIS IS AN EXECUTABLE WORKFLOW - NOT REFERENCE MATERIAL**

When this task is invoked:

1. **CLARIFY BEFORE DOCUMENTING** - Resolve ALL questions with the user before producing document sections
2. **USE MCP TOOLS** - Query documentation via context7 MCP, use sequential-thinking for complex analysis
3. **FINAL DECISIONS ONLY** - Documents must contain ONLY final decisions, not alternatives or rationale discussions
4. **MANDATORY STEP-BY-STEP EXECUTION** - Each section must be processed sequentially with user feedback
5. **ELICITATION IS REQUIRED** - When `elicit: true`, you MUST use the 1-9 format and wait for user response

**VIOLATION INDICATOR:** If you create document sections with alternatives, options, or rationale before clarifying with user, you have violated this workflow.

## Critical: Pre-Document Investigation Phase

**BEFORE starting document creation:**

1. **Gather Context**
   - Read any provided PRD, architecture documents, user stories, or requirements
   - Identify all libraries, SDKs, frameworks, third-party services mentioned

2. **Query Documentation** (MANDATORY)
   - Use `mcp__context7__resolve-library-id` to identify libraries
   - Use `mcp__context7__get-library-docs` to fetch up-to-date documentation
   - Examples: Firebase Functions, OpenAI SDK, TypeScript, React, etc.

3. **Analyze Complexity** (when applicable)
   - Use `mcp__sequential-thinking__sequentialthinking` for:
     - Complex architectural decisions
     - Multiple alternative approaches
     - Interconnected system changes
     - Trade-off analysis

4. **Clarify Requirements**
   - Ask user about ANY unclear technical decisions
   - Confirm implementation approaches
   - Verify assumptions
   - Resolve ambiguities
   - **WAIT for user confirmation before proceeding**

5. **Lock in Final Decisions**
   - Once all questions are answered, document ONLY the final decisions
   - Do NOT include alternative options in the document
   - Do NOT include rationale discussions in the document
   - Keep document focused on what will be implemented

## MCP Tool Usage Examples

### Using context7 for Documentation

```
# Step 1: Resolve library ID
Use: mcp__context7__resolve-library-id
Input: "firebase-functions"
Output: Library ID for Firebase Functions

# Step 2: Get documentation
Use: mcp__context7__get-library-docs
Input: {library_id: "...", query: "how to create http callable functions"}
Output: Up-to-date documentation about callable functions
```

### Using sequential-thinking for Complex Analysis

```
Use: mcp__sequential-thinking__sequentialthinking
Input: {
  "task": "Analyze the best approach for implementing user preference caching with multiple storage options (Redis, in-memory, database)",
  "context": "Firebase Cloud Functions with cold start concerns, budget constraints, need for consistency"
}
Output: Structured thinking process with step-by-step analysis
```

## CRITICAL: Mandatory Elicitation Format

**When `elicit: true`, this is a HARD STOP requiring user interaction:**

**YOU MUST:**

1. Present section content
2. Provide detailed rationale (explain trade-offs, assumptions, decisions made)
3. **STOP and present numbered options 1-9:**
   - **Option 1:** Always "Proceed to next section"
   - **Options 2-9:** Select 8 methods from data/elicitation-methods
   - End with: "Select 1-9 or just type your question/feedback:"
4. **WAIT FOR USER RESPONSE** - Do not proceed until user selects option or provides feedback

**WORKFLOW VIOLATION:** Creating content for elicit=true sections without user interaction violates this task.

**NEVER ask yes/no questions or use any other format.**

## Processing Flow

1. **Pre-Investigation Phase** (described above)
   - Gather context
   - Query documentation via context7
   - Use sequential-thinking for complex decisions
   - Clarify ALL questions with user
   - Lock in final decisions

2. **Load Template** - Use execution-plan-tmpl.yaml

3. **Set Preferences** - Show current mode (Interactive), confirm output file

4. **Process Each Section:**
   - Skip if condition unmet
   - Use context7 MCP when referencing library/framework documentation
   - Use sequential-thinking MCP for complex implementation decisions
   - Draft content using section instruction
   - Present content + detailed rationale
   - **IF elicit: true** → MANDATORY 1-9 options format
   - Save to file if possible

5. **Continue Until Complete**

## Detailed Rationale Requirements

When presenting section content, ALWAYS include rationale that explains:

- Trade-offs and choices made (what was chosen over alternatives and why)
- Key assumptions made during drafting
- Interesting or questionable decisions that need user attention
- Areas that might need validation

## Elicitation Results Flow

After user selects elicitation method (2-9):

1. Execute method from data/elicitation-methods
2. Present results with insights
3. Offer options:
   - **1. Apply changes and update section**
   - **2. Return to elicitation menu**
   - **3. Ask any questions or engage further with this elicitation**

## Test Execution Command Format

**CRITICAL:** Always use this exact format for test commands:

```bash
cd /Users/maikel/Workspace/Pelago/voiced/pelago/apps/voiced/functions && env FIRESTORE_EMULATOR_HOST=localhost:8080 FIREBASE_AUTH_EMULATOR_HOST=localhost:9099 MOCK_OPENAI=true NODE_OPTIONS='--experimental-vm-modules' yarn jest --testPathPattern=<file_path> --testNamePattern=<name_pattern>
```

Where:
- `<file_path>` contains the test file path to be executed
- `<name_pattern>` allows you to execute a subset of tests

## YOLO Mode

User can type `#yolo` to toggle to YOLO mode (process all sections at once).

## CRITICAL REMINDERS

**❌ NEVER:**

- Create document sections before clarifying all questions with user
- Include alternatives, options, or rationale discussions in final document
- Skip using context7 MCP when documentation queries are needed
- Ask yes/no questions for elicitation
- Use any format other than 1-9 numbered options
- Create new elicitation methods

**✅ ALWAYS:**

- Use context7 MCP to query up-to-date documentation
- Use sequential-thinking MCP for complex analysis
- Clarify ALL questions before producing document sections
- Document ONLY final decisions in the execution plan
- Use exact 1-9 format when elicit: true
- Select options 2-9 from data/elicitation-methods only
- Provide detailed rationale explaining decisions
- End with "Select 1-9 or just type your question/feedback:"
- Use exact test command format as specified

## Success Criteria

An execution plan is complete and correct when:

1. All questions have been clarified with the user
2. Document contains ONLY final decisions (no alternatives or rationale)
3. Executive Summary clearly states what, why, and how
4. Implementation Overview includes high-level flow and code changes
5. Test suite is fully defined with exact test commands
6. File-by-file implementation provides clear guidance
7. Code quality checks are specified
8. Implementation checklist breaks work into actionable tasks
9. All relevant documentation was queried via context7 MCP
10. Complex decisions were analyzed via sequential-thinking MCP
