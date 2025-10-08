<!-- Powered by BMAD™ Core -->

# Analyze Prompt Task

## ⚠️ CRITICAL EXECUTION NOTICE ⚠️

**THIS IS AN EXECUTABLE WORKFLOW - NOT REFERENCE MATERIAL**

This task analyzes existing prompts for structure, clarity, conflicts, and optimization opportunities.

## Instructions

### 1. Gather Input

Ask the user to provide the prompt to analyze. Accept:
- Direct paste of prompt text
- File path to read
- URL to fetch (if web access available)

If prompt is very long (>2000 tokens estimated), confirm with user before proceeding.

### 2. Parse Prompt Structure

Analyze the prompt and identify these sections (mark as MISSING if not found):

**Canonical Structure Elements:**
- [ ] **Role/System**: Who the AI is (persona, expertise, mission)
- [ ] **Global Rules**: Session-wide constraints (tone, audience, brevity, boundaries)
- [ ] **Task/Goal**: What to do (should be concrete and singular)
- [ ] **Input/Context**: Data provided for the task (should be fenced)
- [ ] **Examples**: Demonstrations of desired behavior (optional)
- [ ] **Output Contract**: Exact response format/schema
- [ ] **Recency Nudges**: Repeated critical constraints at end

**Note the actual order** of sections as they appear in the prompt.

### 3. Structural Analysis

Evaluate structure against canonical ordering:

#### Ordering Issues
- Are sections in recommended order (Role→Rules→Task→Input→Examples→Contract→Nudges)?
- Are critical elements at the top (high primacy)?
- Are must-follow rules repeated at bottom (recency effect)?

#### Section Quality
- **Role**: Clear and specific? Or vague/missing?
- **Global Rules**: Concise bullets? Or verbose paragraphs?
- **Task**: Single clear objective? Or multiple conflicting goals?
- **Input**: Properly fenced? Or mixed with instructions?
- **Examples**: Short and canonical? Or too many/verbose?
- **Output Contract**: Strict and parseable? Or ambiguous?

### 4. Clarity & Conflict Analysis

#### Instruction Clarity
- Are instructions explicit and concrete?
- Are boundaries expressed positively? ("Keep under 3 sentences" vs "Don't be verbose")
- Are constraints machine-checkable?

#### Conflicts Detection
Scan for contradictory instructions:
- Do different sections give opposing guidance?
- Are there implicit vs explicit conflicts?
- Does user-provided context contradict rules?

**Example conflicts:**
- "Be concise" in rules, then "Explain in detail" in task
- "JSON only" in contract, then "Provide reasoning" in task
- "Professional tone" vs "Casual style"

### 5. Missing Critical Elements

Identify what's missing that could cause issues:

**Safety & Boundaries:**
- [ ] No safety constraints defined
- [ ] No refusal pattern provided
- [ ] No handling for ambiguous/missing info

**Output Specification:**
- [ ] No format specified (JSON, Markdown, etc.)
- [ ] No length constraints
- [ ] No validation hints
- [ ] No handling for edge cases

**Context Management:**
- [ ] Inputs not fenced (risk of instruction injection)
- [ ] No labels on data sections
- [ ] Context too large for task

### 6. Token Efficiency Review

Assess prompt economy:
- Redundant instructions that could be consolidated?
- Verbose sections that could be bullet-pointed?
- Unnecessary examples or explanations?
- Over-specification where simpler would work?

### 7. Generate Analysis Report

Present findings in this structure:

```markdown
# Prompt Analysis Report

## Overall Assessment
[2-3 sentence summary of prompt quality and main issues]

**Estimated Complexity**: [Simple | Moderate | Complex]
**Primary Issues**: [List top 3 issues]

---

## Structure Analysis

### Canonical Sections Present
✅ Role/System: [Present/Missing - brief note]
✅ Global Rules: [Present/Missing - brief note]
✅ Task/Goal: [Present/Missing - brief note]
✅ Input/Context: [Present/Missing - brief note]
⚠️ Examples: [Present/Missing - brief note]
❌ Output Contract: [Present/Missing - brief note]
⚠️ Recency Nudges: [Present/Missing - brief note]

### Section Order
**Current Order**: [List actual order]
**Recommended Order**: Role→Rules→Task→Input→Examples→Contract→Nudges
**Issues**: [Describe ordering problems if any]

---

## Clarity Issues

### Instruction Clarity
[Bullet list of unclear or ambiguous instructions with line references]

### Boundary Expression
[Note any negative boundaries that should be positive]

### Machine-Checkability
[List constraints that aren't verifiable]

---

## Conflicts Detected

[For each conflict found:]
**Conflict #N**: [Description]
- **Location**: [Where in prompt]
- **Impact**: [How this affects behavior]
- **Resolution**: [Suggested fix]

---

## Missing Critical Elements

### Safety & Boundaries
- [ ] [List missing safety elements]

### Output Specification
- [ ] [List missing output specs]

### Context Management
- [ ] [List context issues]

---

## Token Efficiency Opportunities

1. [Specific redundancy or verbosity with example]
2. [Consolidation opportunity]
3. [Simplification suggestion]

**Estimated Token Savings**: ~[X]% with optimization

---

## Scoring

Rate the prompt on these dimensions (1-10):

- **Structure**: [Score]/10 - [Rationale]
- **Clarity**: [Score]/10 - [Rationale]
- **Completeness**: [Score]/10 - [Rationale]
- **Efficiency**: [Score]/10 - [Rationale]
- **Safety**: [Score]/10 - [Rationale]

**Overall Quality**: [Score]/10

---

## Recommendations

### High Priority (Address First)
1. [Most critical issue with specific fix]
2. [Second critical issue]
3. [Third critical issue]

### Medium Priority (Improve Quality)
1. [Enhancement suggestion]
2. [Enhancement suggestion]

### Low Priority (Nice to Have)
1. [Optional improvement]

---

## Next Steps

Would you like me to:
1. **Optimize this prompt** - Apply fixes and generate improved version (*optimize)
2. **Fix specific issue** - Address one particular problem you're concerned about
3. **Run quality checklist** - Validate against full quality criteria (*execute-checklist)
4. **Design from scratch** - Start fresh with better structure (*design)
```

### 8. User Interaction

After presenting the report:
- Ask if user wants deeper analysis on any section
- Offer to run *optimize command to generate improved version
- Suggest *execute-checklist for comprehensive validation

## Best Practices

- **Be Specific**: Reference actual prompt text when citing issues
- **Be Constructive**: Frame problems as improvement opportunities
- **Be Educational**: Explain WHY issues matter, not just WHAT is wrong
- **Be Practical**: Prioritize fixes by impact
- **Be Encouraging**: Acknowledge what works well in the prompt

## Example Analysis Flow

```
User: "Analyze this prompt: [paste]"
→ Parse structure
→ Identify canonical sections
→ Check ordering
→ Scan for conflicts
→ Note missing elements
→ Calculate efficiency
→ Generate report
→ Present findings
→ Offer next steps
```
