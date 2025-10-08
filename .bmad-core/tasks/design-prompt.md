<!-- Powered by BMAD™ Core -->

# Design Prompt Task

## ⚠️ CRITICAL EXECUTION NOTICE ⚠️

**THIS IS AN EXECUTABLE WORKFLOW - NOT REFERENCE MATERIAL**

When this task is invoked:

1. **DISABLE ALL EFFICIENCY OPTIMIZATIONS** - This workflow requires full user interaction
2. **MANDATORY STEP-BY-STEP EXECUTION** - Each section must be processed sequentially with user feedback
3. **ELICITATION IS REQUIRED** - Use the 1-9 format and wait for user response
4. **NO SHORTCUTS ALLOWED** - Complete prompts cannot be created without following this workflow

**VIOLATION INDICATOR:** If you create a complete prompt without user interaction, you have violated this workflow.

## Critical: Interactive Design Process

This task creates new prompts from scratch using structured elicitation to ensure the prompt meets user needs.

## CRITICAL: Mandatory Elicitation Format

**This is a HARD STOP requiring user interaction:**

**YOU MUST:**

1. Present section content/draft
2. Provide detailed rationale (explain decisions, trade-offs, assumptions)
3. **STOP and present numbered options 1-9:**
   - **Option 1:** Always "Proceed to next section"
   - **Options 2-9:** Select 8 methods from data/elicitation-methods
   - End with: "Select 1-9 or just type your question/feedback:"
4. **WAIT FOR USER RESPONSE** - Do not proceed until user selects option or provides feedback

**NEVER ask yes/no questions or use any other format.**

## Instructions

### 1. Gather High-Level Requirements

Start by understanding the use case. Ask:

**Context Questions:**
1. What should this prompt accomplish? (the main goal)
2. Who/what is the audience for the output?
3. What format should the output be in? (JSON, Markdown, plain text, etc.)
4. Are there any specific constraints or requirements? (length, tone, safety, etc.)
5. Will this prompt use tools/functions? (yes/no, if yes, which ones?)

**Capture responses** - these drive the design.

### 2. Section-by-Section Design

Process each section using the elicitation workflow:

---

#### Section 1: Role Definition

**Draft the Role section** (1-2 sentences):
- Who is the AI?
- What is their core mission?
- What expertise do they bring?

**Present draft with rationale:**
- Explain why this role fits the use case
- Note any assumptions made
- Highlight trade-offs (e.g., specialist vs generalist)

**MANDATORY ELICITATION:**
Present 1-9 options:
1. Proceed to next section
2-9. [Select from elicitation-methods]

"Select 1-9 or just type your question/feedback:"

**Wait for user response.**

---

#### Section 2: Global Rules

**Draft Global Rules** (bullet list):
- Audience: [who will consume this?]
- Style: [tone, voice, formality]
- Brevity: [explicit length constraint]
- Safety/Boundaries: [what's out of scope?]
- [Other session-wide rules based on requirements]

**Present draft with rationale:**
- Explain each rule's purpose
- Note any assumptions about audience or context
- Highlight boundary decisions

**MANDATORY ELICITATION:**
Present 1-9 options:
1. Proceed to next section
2-9. [Select from elicitation-methods]

"Select 1-9 or just type your question/feedback:"

**Wait for user response.**

---

#### Section 3: Task Definition

**Draft Task section** (one clear sentence):
- State the concrete objective
- Keep it singular and focused
- Optionally add bullet steps if multi-step

**Present draft with rationale:**
- Explain how task aligns with goal
- Note any simplifications made
- Highlight scope decisions (what's included/excluded)

**MANDATORY ELICITATION:**
Present 1-9 options:
1. Proceed to next section
2-9. [Select from elicitation-methods]

"Select 1-9 or just type your question/feedback:"

**Wait for user response.**

---

#### Section 4: Input Specification

**Draft Input section:**
- Choose fencing style (sentinels, XML tags, code fence)
- Add labels for data sections
- Include placeholder or example data

**Present draft with rationale:**
- Explain fencing choice (why this delimiter?)
- Note any assumptions about input structure
- Highlight safety considerations (injection prevention)

**MANDATORY ELICITATION:**
Present 1-9 options:
1. Proceed to next section
2-9. [Select from elicitation-methods]

"Select 1-9 or just type your question/feedback:"

**Wait for user response.**

---

#### Section 5: Examples (Optional)

**Ask:** Does this prompt need examples to demonstrate desired behavior?

If YES:
- Determine how many (zero-shot, one-shot, few-shot)
- Draft examples (keep short and canonical)
- Ensure examples show format and style

If NO:
- Skip this section
- Note: Many modern models work well zero-shot

**If drafting examples, present with rationale:**
- Explain why these examples were chosen
- Note what each example teaches
- Highlight diversity (do they cover different cases?)

**MANDATORY ELICITATION (if examples included):**
Present 1-9 options:
1. Proceed to next section
2-9. [Select from elicitation-methods]

"Select 1-9 or just type your question/feedback:"

**Wait for user response.**

---

#### Section 6: Output Contract

**Draft Output Contract:**
- Specify exact format (JSON, Markdown, table, etc.)
- Include schema with field types
- Add validation hints
- Handle edge cases (null, empty, errors)

**For JSON contracts:**
```json
{
  "field_name": "type (string | integer | boolean | array | object)",
  "another_field": "allowed_values (enum if applicable)"
}
```

**For Markdown contracts:**
```markdown
# Expected Output Format
- Heading structure
- Bullet requirements
- Table schema if applicable
```

**Present draft with rationale:**
- Explain format choice (why JSON vs Markdown?)
- Note schema design decisions
- Highlight validation strategy

**MANDATORY ELICITATION:**
Present 1-9 options:
1. Proceed to next section
2-9. [Select from elicitation-methods]

"Select 1-9 or just type your question/feedback:"

**Wait for user response.**

---

#### Section 7: Recency Nudges

**Draft Remember section:**
- Restate 1-2 most critical constraints
- Mirror must-follow rules from earlier sections
- Keep very brief (2-3 bullets max)

**Example:**
```
# REMEMBER
- JSON only, no extra text
- Maximum 3 items in response
```

**Present draft with rationale:**
- Explain which rules were chosen for repetition
- Note why these are most critical

**MANDATORY ELICITATION:**
Present 1-9 options:
1. Proceed to final assembly
2-9. [Select from elicitation-methods]

"Select 1-9 or just type your question/feedback:"

**Wait for user response.**

---

### 3. Assemble Final Prompt

Combine all sections in canonical order:

```text
# ROLE
[Section 1 content]

# GLOBAL RULES
[Section 2 content]

# TASK
[Section 3 content]

# INPUT
[Section 4 content]

# EXAMPLES (optional)
[Section 5 content if included]

# OUTPUT CONTRACT
[Section 6 content]

# REMEMBER
[Section 7 content]
```

### 4. Present Final Prompt

Show the complete prompt in clean, copyable format:

````markdown
---

## Designed Prompt

```text
[Full prompt here]
```

---
````

### 5. Design Summary

Provide a summary report:

```markdown
# Prompt Design Summary

## Design Goals
- **Primary Objective**: [Main goal from requirements]
- **Audience**: [Who consumes output]
- **Format**: [Output format]
- **Constraints**: [Key limitations or boundaries]

## Design Decisions

### 1. [Decision Category]
**Choice**: [What was decided]
**Rationale**: [Why this choice]
**Trade-off**: [What was sacrificed for this choice]

### 2. [Decision Category]
**Choice**: [What was decided]
**Rationale**: [Why this choice]
**Trade-off**: [What was sacrificed for this choice]

[Continue for 4-6 key decisions]

## Prompt Characteristics

- **Estimated Length**: ~[X] tokens
- **Complexity**: [Simple | Moderate | Complex]
- **Interaction Style**: [One-shot | Conversational | Tool-calling]
- **Safety Level**: [Basic | Standard | High]

## Testing Recommendations

To validate this design:

1. **Happy Path**: [Test case for normal usage]
2. **Edge Case**: [Test case for boundary condition]
3. **Error Case**: [Test case for malformed input]
4. **Ambiguous Case**: [Test case for unclear input]

## Next Steps

Would you like me to:
1. **Test the prompt** - Run with sample inputs to verify behavior
2. **Optimize further** - Reduce tokens or improve clarity (*optimize)
3. **Run quality checklist** - Validate against full criteria (*execute-checklist)
4. **Save to file** - Export prompt to specified location
5. **Create template** - Convert to reusable template with placeholders (*template)
```

### 6. User Interaction

After presenting:
- Ask if user wants to test the prompt
- Offer to make adjustments to specific sections
- Suggest running quality checklist
- Offer to create reusable template version

## Elicitation Results Flow

After user selects elicitation method (2-9):

1. Execute method from data/elicitation-methods
2. Present results with insights
3. Offer options:
   - **1. Apply changes and proceed to next section**
   - **2. Return to elicitation menu for this section**
   - **3. Ask any questions or engage further with this elicitation**

## YOLO Mode

User can type `#yolo` to toggle to YOLO mode:
- Process all sections at once
- Draft complete prompt based on initial requirements
- Present for feedback

**Note:** YOLO mode is faster but may miss important nuances. Interactive mode is recommended for critical prompts.

## Best Practices

### DO:
- ✅ Ask clarifying questions when requirements are vague
- ✅ Explain trade-offs and assumptions in rationale
- ✅ Keep role and rules sections SHORT
- ✅ Use positive boundaries ("Keep under 3 sentences")
- ✅ Make output contracts strict and parseable
- ✅ Test assumptions with edge cases

### DON'T:
- ❌ Skip elicitation steps to save time
- ❌ Create complete prompts without user feedback
- ❌ Assume you know user's intent without asking
- ❌ Add unnecessary complexity or verbosity
- ❌ Use vague qualifiers ("try to", "ideally")
- ❌ Mix multiple objectives in one task

## Special Case Handling

### Tool-Calling Prompts
If user indicates tools will be used:
- Add TOOLS section after GLOBAL RULES
- Define each tool with types: `tool_name(param: type) → return_type`
- Add tool selection rules
- Include failure handling
- Specify when to ask vs execute

### Voice-First Prompts
If output will be spoken:
- Add voice-specific rules:
  - Brevity: 2-3 sentences default
  - Numbers: Spell out 1-10
  - Tone: Conversational
  - Avoid: "As an AI", "Certainly"

### Long-Context Prompts
If dealing with large inputs:
- Add context management rules
- Specify summarization strategy
- Define conflict resolution hierarchy
- Pin critical rules at top and bottom

### Multi-Turn Conversation Prompts
If part of conversation system:
- Define conversation state handling
- Specify context retention rules
- Add persona consistency constraints
- Handle conversation closure

## Example Elicitation Flow

```
Harper: "Let's design your prompt. What should this prompt accomplish?"
User: "Extract product info from descriptions"

Harper: "Got it. Who will use this output? (humans or machines?)"
User: "Our product database system"

Harper: "Perfect. What format? (JSON, Markdown, CSV?)"
User: "JSON"

Harper: [Drafts Role section]
"# ROLE
You extract structured product data from text descriptions."

Rationale: Focused specialist role matches extraction task.
Assumption: No need for explanatory output, pure extraction.
Trade-off: Won't provide reasoning, just data.

Select 1-9 or just type your question/feedback:
1. Proceed to next section
2. [Elicitation method]
...

User: "1"

Harper: [Continues to Global Rules section...]
```
