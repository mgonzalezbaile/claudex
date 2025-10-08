<!-- Powered by BMAD™ Core -->

# Prompt Quality Checklist

## Overview

This checklist validates prompts against modern best practices from the OpenAI Prompt Engineering Playbook. Use this as a pre-flight check before deploying prompts to production.

**How to use:**
- Run this checklist via `*execute-checklist prompt-quality-checklist`
- Work through sections one by one (interactive mode) or all at once (YOLO mode)
- Each section validates specific quality dimensions

**Scoring:**
- ✅ PASS: Requirement met
- ❌ FAIL: Requirement not met (needs fixing)
- ⚠️ PARTIAL: Partially met (could be improved)
- N/A: Not applicable to this prompt

---

## Section 1: Structural Integrity

**LLM Instructions:**
Evaluate whether the prompt follows canonical structure and ordering. Check for presence and placement of key sections.

**Validation Items:**

### 1.1 Section Presence
- [ ] **Role/System section** present at top defining who the AI is
- [ ] **Global Rules section** present with session-wide constraints
- [ ] **Task/Goal section** present with clear objective
- [ ] **Input section** present (or placeholder if not applicable)
- [ ] **Output Contract section** present with format specification
- [ ] **Remember/Recency section** present (or critical rules repeated at bottom)

### 1.2 Section Ordering
- [ ] Role/System appears FIRST (high primacy)
- [ ] Global Rules appear near top (before task details)
- [ ] Task definition comes after rules
- [ ] Input fencing appears after task (or integrated appropriately)
- [ ] Examples (if any) appear after input
- [ ] Output Contract near end (before Remember)
- [ ] Remember/Recency section at very end

### 1.3 Section Quality
- [ ] Role section is 1-2 sentences (concise, not verbose)
- [ ] Global Rules use bullet format (not paragraphs)
- [ ] Task is single sentence or short bullet steps (not essay)
- [ ] Output Contract is specific and parseable (not vague)

**Section Score:** [X/13 items passed]

---

## Section 2: Clarity & Explicitness

**LLM Instructions:**
Assess whether instructions are explicit, concrete, and unambiguous. Check for vague language and ensure constraints are machine-checkable.

**Validation Items:**

### 2.1 Instruction Clarity
- [ ] All instructions are concrete (no "try to", "ideally", "if possible")
- [ ] Task objective is unambiguous (single clear interpretation)
- [ ] Constraints are explicit (not "be brief" but "max 3 sentences")
- [ ] All requirements are stated positively (not negative framing)

### 2.2 Boundary Expression
- [ ] Length limits are numeric ("≤120 words" not "concise")
- [ ] Tone/style specified with adjectives (not left to default)
- [ ] Audience explicitly stated (not assumed)
- [ ] Scope boundaries defined (what's in/out of scope)

### 2.3 Machine-Checkable Constraints
- [ ] Output format verifiable (JSON schema, table columns, etc.)
- [ ] Numeric constraints have exact values (not ranges like "about 100")
- [ ] Enums are listed explicitly (not "appropriate values")
- [ ] Boolean conditions are unambiguous (not "mostly" or "usually")

**Section Score:** [X/12 items passed]

---

## Section 3: Conflict Prevention

**LLM Instructions:**
Scan for contradictory instructions, ambiguous priorities, and potential conflicts between sections.

**Validation Items:**

### 3.1 Internal Consistency
- [ ] No direct contradictions (e.g., "be brief" + "explain in detail")
- [ ] Tone directives align (no "professional" + "casual")
- [ ] Length constraints don't conflict (no "3 sentences" + "thorough analysis")
- [ ] Format requirements align (no "JSON only" + "provide reasoning")

### 3.2 Priority Hierarchy
- [ ] Conflict resolution rule stated (if input vs rules conflict, which wins?)
- [ ] Tool-calling priority clear (when to call vs answer directly)
- [ ] Multiple objectives ranked or separated (not competing goals)
- [ ] User input authority specified (can it override rules?)

### 3.3 Recency Reinforcement
- [ ] Critical constraints repeated at bottom (Remember section)
- [ ] Most important rules appear both top and bottom
- [ ] No new conflicting rules introduced at end

**Section Score:** [X/9 items passed]

---

## Section 4: Completeness

**LLM Instructions:**
Identify missing elements that could cause failures or ambiguous behavior.

**Validation Items:**

### 4.1 Output Specification
- [ ] Format explicitly stated (JSON, Markdown, plain text, etc.)
- [ ] Schema/structure provided (not just "return results")
- [ ] Edge case handling defined (empty results, errors, missing data)
- [ ] Validation hints included (type constraints, required fields)

### 4.2 Safety & Boundaries
- [ ] Out-of-scope requests handled (refusal pattern or scope statement)
- [ ] Safety constraints stated if needed (no harmful content, etc.)
- [ ] Ambiguous input handling defined (ask vs proceed vs refuse)
- [ ] Conflict resolution rule present

### 4.3 Context Management
- [ ] Input properly fenced (delimiters, labels, clear separation)
- [ ] Input sections labeled (INPUT, CONTEXT, DATA, etc.)
- [ ] Instruction to ignore conflicting user input (if applicable)
- [ ] Handling for long context if applicable

**Section Score:** [X/12 items passed]

---

## Section 5: Token Efficiency

**LLM Instructions:**
Assess whether the prompt is as concise as possible without sacrificing clarity or completeness.

**Validation Items:**

### 5.1 Unnecessary Verbosity
- [ ] Role section ≤2 sentences (not full paragraph)
- [ ] Global Rules are bullets (not prose)
- [ ] No redundant explanations (saying same thing multiple ways)
- [ ] Examples are minimal (not excessive demonstrations)

### 5.2 Strategic Consolidation
- [ ] Similar rules consolidated (not scattered)
- [ ] Instructions grouped logically (not repeated)
- [ ] Examples serve distinct purposes (not redundant patterns)
- [ ] No preambles or postambles ("As an AI...", "In conclusion...")

### 5.3 Balance Check
- [ ] Concision doesn't harm clarity (still understandable)
- [ ] Critical details not removed for brevity
- [ ] Efficiency gains don't introduce ambiguity

**Section Score:** [X/11 items passed]

---

## Section 6: Format Discipline

**LLM Instructions:**
Verify appropriate format choice and correct usage of formatting conventions.

**Validation Items:**

### 6.1 Format Choice
- [ ] JSON used for machine consumption (not human narrative)
- [ ] Markdown used for human-readable docs (not strict parsing)
- [ ] Tables used for tabular data (not nested structures)
- [ ] Hybrid formats separated clearly (if both human + machine)

### 6.2 Fencing & Delimiters
- [ ] Input data fenced with clear delimiters (```, <<<>>>, XML tags)
- [ ] Fencing style appropriate for content (code fence for code, etc.)
- [ ] Labels used to identify sections (INPUT, POLICY, EXAMPLES)
- [ ] Nested fencing avoided (or handled with different markers)

### 6.3 JSON Contracts (if applicable)
- [ ] "JSON only, no extra text" explicitly stated
- [ ] Complete schema provided with types
- [ ] Required vs optional fields marked
- [ ] Null handling specified
- [ ] Edge case instructions included (empty arrays, missing data)

### 6.4 Markdown Contracts (if applicable)
- [ ] Heading structure specified (# ## ###)
- [ ] Section requirements defined
- [ ] List formats stated (bullets, numbered, checklist)
- [ ] Content constraints included (length, style)

**Section Score:** [X/14 items passed]

---

## Section 7: Special Patterns (if applicable)

**LLM Instructions:**
For specialized prompt types, validate pattern-specific requirements.

**Validation Items:**

### 7.1 Tool-Calling Prompts
- [ ] All tools defined with types (param: type → return_type)
- [ ] When-to-call rules explicit (not just tool descriptions)
- [ ] When-NOT-to-call rules present (avoid unnecessary calls)
- [ ] Parameter exactness emphasized (use exact names)
- [ ] Failure handling defined (retry policy, error reporting)
- [ ] External action confirmation required (email, writes, etc.)

### 7.2 Extraction Prompts
- [ ] "JSON only, no extra text" stated prominently
- [ ] Null handling explicit (use null vs omit vs default)
- [ ] "Never fabricate" rule present
- [ ] Confidence threshold stated (if applicable)
- [ ] Required vs optional fields clear

### 7.3 Analysis Prompts
- [ ] Answer-first structure specified
- [ ] Strict word/finding limit set
- [ ] Finding → Implication → Action pattern present
- [ ] Evidence citation required
- [ ] Prioritization criteria defined (top N, by impact, etc.)

### 7.4 Voice-First Prompts
- [ ] Brevity constraint for spoken output (2-3 sentences)
- [ ] Number formatting specified (spell out vs digits)
- [ ] Conversational tone required
- [ ] Banned robotic phrases listed ("As an AI", etc.)

### 7.5 Long-Context Prompts
- [ ] Context window management addressed
- [ ] Rolling window or summarization strategy
- [ ] Critical rules pinned at top and bottom
- [ ] Section labels and fencing for large inputs
- [ ] Conflict resolution hierarchy stated

**Section Score:** [X/applicable items passed]
**Note:** Score only applicable pattern(s); mark others N/A

---

## Section 8: Testing Readiness

**LLM Instructions:**
Determine if the prompt is ready for validation testing.

**Validation Items:**

### 8.1 Test Case Identification
- [ ] Happy path test case identifiable (normal valid input)
- [ ] Edge case test case identifiable (boundary conditions)
- [ ] Error case test case identifiable (malformed input)
- [ ] Ambiguous case test case identifiable (unclear input)

### 8.2 Measurement Criteria
- [ ] Output format is verifiable (can check JSON validity, etc.)
- [ ] Success criteria clear (what makes a good response?)
- [ ] Failure modes defined (what would be wrong output?)
- [ ] Performance expectations set (latency, tokens if applicable)

**Section Score:** [X/8 items passed]

---

## Final Report

**LLM Instructions:**
After completing all sections, generate summary report:

### Overall Assessment

**Total Score:** [X/applicable items] ([Y%])

**Grade:**
- A (90-100%): Production-ready, excellent quality
- B (80-89%): Good quality, minor improvements suggested
- C (70-79%): Acceptable, several improvements needed
- D (60-69%): Needs significant revision
- F (<60%): Major issues, requires redesign

### Critical Issues (❌ FAIL items)

[List all failed items with section references]

1. [Issue description - Section X.Y]
2. [Issue description - Section X.Y]
...

### Improvement Opportunities (⚠️ PARTIAL items)

[List all partial items with suggestions]

1. [Item description - Section X.Y] → [Suggestion]
2. [Item description - Section X.Y] → [Suggestion]
...

### Strengths (✅ PASS items worth noting)

[Highlight particularly well-executed aspects]

1. [Strength - Section X.Y]
2. [Strength - Section X.Y]
...

### Priority Recommendations

**Must Fix (High Priority):**
1. [Most critical issue with specific fix]
2. [Second critical issue]
3. [Third critical issue]

**Should Improve (Medium Priority):**
1. [Enhancement suggestion]
2. [Enhancement suggestion]

**Nice to Have (Low Priority):**
1. [Optional improvement]

### Next Steps

- [ ] Address all critical issues (❌ FAIL items)
- [ ] Review partial items (⚠️ PARTIAL)
- [ ] Run test cases (Section 8)
- [ ] Consider optimization suggestions
- [ ] Re-run checklist after fixes

---

## Appendix: Common Issues & Fixes

### Issue: "Vague length constraint"
**Example:** "Be concise"
**Fix:** "Maximum 3 sentences (60 words)"

### Issue: "Missing output format"
**Example:** "Provide the results"
**Fix:** "Return JSON with keys: {name: string, value: number}"

### Issue: "Conflicting tone"
**Example:** "Professional but casual"
**Fix:** Choose one: "Professional, direct tone" OR "Conversational, friendly tone"

### Issue: "Unfenced input"
**Example:** "Here's the data: {data}"
**Fix:** "<<<INPUT_START\n{data}\nINPUT_END>>>"

### Issue: "Negative boundary"
**Example:** "Don't be verbose"
**Fix:** "Use 150 words maximum"

### Issue: "Ambiguous task"
**Example:** "Analyze the document"
**Fix:** "Extract top 3 insights from document focusing on cost savings"
