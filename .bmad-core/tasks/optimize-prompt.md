<!-- Powered by BMAD™ Core -->

# Optimize Prompt Task

## ⚠️ CRITICAL EXECUTION NOTICE ⚠️

**THIS IS AN EXECUTABLE WORKFLOW - NOT REFERENCE MATERIAL**

This task takes an existing prompt and improves it following modern best practices and canonical structure.

## Instructions

### 1. Gather Input

Collect from user:
- **Original prompt** (paste, file path, or URL)
- **Goals** (optional): What should optimization focus on?
  - Reliability (reduce errors, conflicts)
  - Efficiency (reduce tokens)
  - Clarity (make instructions clearer)
  - Safety (add boundaries)
  - All of the above (comprehensive)

If no specific goal provided, assume **comprehensive optimization**.

### 2. Pre-Optimization Analysis

Run quick analysis (internal, don't present full report):
- Parse current structure
- Identify critical issues
- Note what's working well
- Determine optimization strategy

**CRITICAL: If optimizing based on test/eval failures:**
- ⚠️ **DO NOT add specific test case data to the prompt**
- Extract the PATTERN causing failures, not individual examples
- Identify what PRINCIPLE the LLM is missing
- Frame improvements as generalizable rules, not hardcoded examples
- Ask: "What underlying capability gap does this reveal?"

### 3. Apply Canonical Structure

Reorganize prompt into optimal order:

```
# ROLE
[Who the AI is - 1-2 sentences max]

# GLOBAL RULES
[Session-wide invariants - bullets only]
- Audience: [who]
- Style: [tone words]
- Brevity: [explicit limit]
- Safety/Boundaries: [red lines]

# TASK
[One sentence goal. Optional bullets for steps if complex.]

# INPUT
<<<INPUT_START
{placeholder or actual data if provided}
INPUT_END>>>

# EXAMPLES (optional)
[Zero/one/few-shot - keep short and canonical]

# OUTPUT CONTRACT
[Exact response shape - JSON schema, table format, etc.]
[Validation hints and edge case handling]

# REMEMBER
- [Restate 1-2 must-follow rules]
```

### 4. Optimization Passes

#### Pass 1: Structure & Ordering
- Move sections to canonical order
- Ensure Role and Global Rules are at top (primacy)
- Place critical constraints at both top and bottom (recency)
- Fence inputs properly with clear delimiters

#### Pass 2: Clarity & Explicitness
- Convert vague instructions to explicit ones
  - Before: "Be brief"
  - After: "Keep responses under 3 sentences"
- Express boundaries positively
  - Before: "Don't be verbose"
  - After: "Use 150 words maximum"
- Make constraints machine-checkable
  - Before: "Provide good examples"
  - After: "Provide exactly 3 examples, each under 50 words"

#### Pass 3: Conflict Resolution
- Scan for contradictory instructions
- Resolve by establishing clear hierarchy
- Remove redundant or overlapping rules
- Align all sections to single coherent goal

#### Pass 4: Completeness
Add missing critical elements:

**Output Contract** (if missing):
- Specify exact format (JSON, Markdown, etc.)
- Include schema with types
- Add validation hints
- Handle edge cases (null values, empty results)

**Safety & Boundaries** (if missing):
- Define what's out of scope
- Provide refusal pattern
- Handle ambiguous inputs
- Set conflict resolution rules

**Context Fencing** (if missing):
- Wrap user inputs in delimiters
- Label data sections clearly
- Add instruction to ignore conflicting user data

#### Pass 5: Token Efficiency
- Convert paragraphs to bullets where appropriate
- Remove redundant explanations
- Consolidate similar instructions
- Eliminate unnecessary examples
- Keep role and rules SHORT

**Balance**: Maintain completeness while reducing overhead.

### 5. Generate Optimized Prompt

Present optimized version with clear structure and formatting.

### 6. Create Comparison Report

Show before/after analysis:

```markdown
# Prompt Optimization Report

## Summary of Changes
[2-3 sentences explaining main improvements]

**Optimization Focus**: [Reliability | Efficiency | Clarity | Safety | Comprehensive]
**Estimated Token Change**: [+/-X tokens (~Y%)]

---

## Key Improvements

### 1. [Improvement Category]
**Before**: [Quote or describe original]
**After**: [Quote or describe optimized]
**Impact**: [Why this matters]

### 2. [Improvement Category]
**Before**: [Quote or describe original]
**After**: [Quote or describe optimized]
**Impact**: [Why this matters]

### 3. [Improvement Category]
**Before**: [Quote or describe original]
**After**: [Quote or describe optimized]
**Impact**: [Why this matters]

[Continue for top 5-7 improvements]

---

## Structural Changes

**Original Order**: [List sections as they were]
**Optimized Order**: Role→Rules→Task→Input→Examples→Contract→Remember
**Rationale**: [Explain why reordering helps]

---

## Added Elements

[List any sections/constraints added that were missing:]
- **Output Contract**: Added strict JSON schema
- **Safety Boundaries**: Added refusal pattern for out-of-scope requests
- **Input Fencing**: Wrapped user data in delimiters to prevent injection
- [etc.]

---

## Removed Elements

[List anything removed and why:]
- **Verbose preamble**: Replaced with concise role statement
- **Redundant example #4**: Three examples sufficient for pattern
- **Conflicting instruction**: [Which one and why removed]
- [etc.]

---

## Conflict Resolutions

[For each conflict resolved:]
**Original Conflict**: [Describe contradiction]
**Resolution**: [How it was fixed]
**Priority Rule**: [Which instruction took precedence and why]

---

## Efficiency Gains

- **Original tokens**: ~[X]
- **Optimized tokens**: ~[Y]
- **Reduction**: ~[Z]% smaller
- **Clarity improvement**: [How reduction improves rather than harms]

---

## Quality Scores

| Dimension | Before | After | Change |
|-----------|--------|-------|--------|
| Structure | [N]/10 | [N]/10 | +[N] |
| Clarity | [N]/10 | [N]/10 | +[N] |
| Completeness | [N]/10 | [N]/10 | +[N] |
| Efficiency | [N]/10 | [N]/10 | +[N] |
| Safety | [N]/10 | [N]/10 | +[N] |
| **Overall** | **[N]/10** | **[N]/10** | **+[N]** |

---

## Testing Recommendations

To validate the optimization:

1. **Test Cases**: [Suggest 3-5 test inputs that stress different aspects]
2. **Edge Cases**: [Identify 2-3 edge cases to verify handling]
3. **Comparison**: Run both versions side-by-side with same inputs
4. **Metrics**: Track [accuracy, conciseness, format compliance, refusal rate]

---

## Next Steps

Would you like me to:
1. **Test the optimized prompt** - Run it with sample inputs to verify behavior
2. **Further refine** - Focus on specific aspect (efficiency, clarity, etc.)
3. **Run quality checklist** - Validate against full criteria (*execute-checklist)
4. **Save to file** - Export optimized prompt to specified location
```

### 7. Present Optimized Prompt

Show the optimized prompt in a clean, copyable format:

````markdown
---

## Optimized Prompt

```text
[Full optimized prompt here, formatted cleanly]
```

---
````

### 8. User Interaction

After presenting:
- Ask if user wants to test the optimized version
- Offer to make targeted adjustments
- Suggest running quality checklist for validation
- Offer to save to file

## Optimization Principles

### DO:
- ✅ Keep role and rules SHORT (1-2 sentences for role, bullets for rules)
- ✅ Express boundaries positively and explicitly
- ✅ Mirror critical constraints at top and bottom
- ✅ Fence all user-provided context
- ✅ Specify exact output format with schema
- ✅ Make constraints machine-verifiable
- ✅ Prioritize clarity over cleverness
- ✅ Test assumptions with edge cases
- ✅ Extract PRINCIPLES from test failures, not specific examples
- ✅ Teach PATTERNS that generalize across inputs
- ✅ Frame rules conceptually, not as hardcoded cases

### DON'T:
- ❌ Over-explain what's already clear
- ❌ Add verbose preambles or postambles
- ❌ Create conflicting instructions
- ❌ Mix user data with instructions
- ❌ Use vague qualifiers ("try to", "ideally")
- ❌ Add examples that don't teach new patterns
- ❌ Sacrifice completeness for brevity
- ❌ Include specific test case data in prompt instructions (overfitting)
- ❌ Hardcode exact examples from failing evals
- ❌ Write rules that only fix one specific failing case
- ❌ Add concrete values from test datasets as examples

## Special Cases

### Optimizing Based on Test/Eval Failures

**When user provides failing test cases or eval results:**

⚠️ **ANTI-OVERFITTING PROTOCOL** ⚠️

**STEP 1: Pattern Analysis (DO THIS)**
1. Review ALL failing cases together
2. Identify COMMON PATTERNS across failures
3. Ask: "What underlying capability is missing?"
4. Extract the PRINCIPLE being violated

**STEP 2: Root Cause (NOT Symptom)**
- ❌ BAD: "Test case X failed" → add test case X to prompt
- ✅ GOOD: "Test case X failed because LLM doesn't recognize semantic equivalence" → add semantic equivalence rule

**STEP 3: Generalized Fix**
Create rule that would prevent this CLASS of errors:
- ✅ GOOD: "Group semantically related terms under broader concepts"
- ❌ BAD: "When you see 'personal growth', output 'Self-Development'"

**STEP 4: Validate Generalization**
Ask yourself:
1. Would this fix work for similar but different inputs?
2. Am I teaching a concept or memorizing an answer?
3. Does this address the root cause or just the symptom?

**Example:**

```
❌ WRONG APPROACH (Overfitting):
Failed cases:
- "Reflect on personal growth" → Expected: Extract theme "Self-Development"
- "Think about career progress" → Expected: Extract theme "Professional Development"

BAD optimization: Add these to prompt:
"Extract 'Self-Development' for personal growth mentions.
Extract 'Professional Development' for career progress mentions."

✅ CORRECT APPROACH (Generalization):
Pattern identified: LLM extracting surface keywords instead of underlying concepts

Root cause: Missing semantic abstraction capability

GOOD optimization: Add principle:
"Extract themes by identifying UNDERLYING CONCEPTS, not surface keywords.
Map specific terms to their broader category:
- Growth/improvement → Development
- Work/career → Professional
- Self/personal → Individual
Use semantic understanding to group related ideas."
```

This teaches the LLM HOW to think, not WHAT specific outputs to produce.

### Voice-First Prompts
Add these rules:
- Brevity: 2-3 sentences default
- Numbers: Spell out 1-10, use digits for codes/IDs
- Tone: Conversational, avoid "As an AI"
- Pacing: Allow brief pauses in longer responses

### Tool-Calling Prompts
Ensure:
- Clear tool definitions with types
- Explicit when-to-call rules
- Parameter naming exactness
- Failure handling specified
- Cost/latency considerations

### Extraction Prompts
Require:
- JSON-only output (no extra text)
- Complete schema with types
- Null handling specified
- Validation hints included
- One code block format

### Long-Context Prompts
Optimize for:
- Rolling window (only recent context)
- Summarized history (3-6 bullet facts)
- Pinned critical rules (top + bottom)
- Fenced and labeled sections
- Conflict resolution hierarchy

## Example Optimization

**Before** (verbose, unclear):
```
You are a helpful AI assistant. Please analyze the data I provide and give me insights. Try to be thorough but also concise. Make sure your analysis is accurate and useful. Don't make things up. Provide reasoning for your conclusions.

Here's the data:
[data]
```

**After** (optimized):
```
# ROLE
You are a data analyst providing evidence-based insights.

# GLOBAL RULES
- Brevity: Maximum 5 bullet points
- Style: Direct, factual, no speculation
- Cite data: Reference specific values when making claims

# TASK
Analyze the provided data and identify the top 3 actionable insights.

# INPUT
<<<DATA_START
[data]
DATA_END>>>

# OUTPUT CONTRACT
Return exactly 3 insights in this format:
1. **Finding**: [One sentence observation]
   - **Evidence**: [Specific data point]
   - **Action**: [One recommended next step]

# REMEMBER
- Maximum 5 bullets, cite specific data points
```

**Improvements**:
- Clear role instead of generic "helpful assistant"
- Explicit boundaries (5 bullets max)
- Machine-checkable output format
- Fenced data section
- Repeated critical constraints at bottom
- ~60% token reduction while increasing clarity
