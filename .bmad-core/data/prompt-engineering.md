# OpenAI Prompt Engineering Playbook (Concise)

> This guide is OpenAI-first, non‑RAG, and focuses on **how to structure prompts** for reliability, quality, and low overhead.

---

## 1) Modern Model Behavior

Modern OpenAI models follow **explicit, concrete instructions** more literally than prior generations. A single, unambiguous sentence (e.g., “Respond with JSON only, no extra text.”) reliably shapes behavior—provided it appears early and isn’t contradicted later. When instructions clash, models tend to favor the **latest, most specific** cue they read, but the safest approach is to avoid conflicts and **mirror must‑follow rules** at both the top and bottom of the prompt.

Defaults matter: if tone/length/audience aren’t specified, models choose reasonable—sometimes verbose—defaults. You get more predictable results by setting explicit **boundaries** (length caps, audience, style), expressing them **positively** (“Keep answers under 3 sentences.”) and making them **machine-checkable** (e.g., output contracts).

## 2) Prompt Anatomy & Optimal Ordering (core)

**Canonical order** (top → bottom) and why it works:

1. **System / Role** – Who the agent is and the non‑negotiable mission. This frames *all* behavior and leverages primacy.
2. **Global Rules** – Session‑wide invariants (tone, safety, format bans, concision). Keep these short and bullet‑pointed.
3. **Task / Goal** – The concrete thing to do *now*. Single sentence first; optional bullets for steps.
4. **Inputs / Context** – The minimum necessary data for this turn (paste, brief summary, or key facts). Fence it clearly.
5. **Examples (optional)** – Zero/one/few‑shot exemplars showing *style* and *format*. Keep them canonical and short.
6. **Output Contract** – Exact response shape (e.g., JSON schema, table columns). Add validation hints.
7. **Recency Nudges (optional)** – Re‑state 1–2 must‑follow constraints (e.g., “Remember: JSON only.”).

### Minimal anatomy template

```text
# ROLE
You are a <specialist persona>. Your job: <one‑line mission>.

# GLOBAL RULES
- Audience: <who>
- Style: <tone words>
- Brevity: <limit>
- Safety/Boundaries: <red lines>

# TASK
<one sentence goal>

# INPUT
<<<INPUT_START
{insert minimal context}
INPUT_END>>>

# EXAMPLES (optional)
- Input → Output pairs (brief).

# OUTPUT CONTRACT
Return JSON only with keys: {"field_a": string, "field_b": integer, "notes": string|null}. No extra text.

# REMEMBER (optional)
- JSON only. No prose.
```

### Practical notes

* **Keep “Role” and “Global Rules” short.** Long prologues dilute signal.
* **Fence context** (triple backticks, XML‑like tags, or sentinel markers) so the model knows where data ends and instructions resume.
* **One task per prompt.** Split multi‑objectives across turns unless they’re tightly coupled.
* **Repeat the non‑negotiables** (e.g., “JSON only”) at bottom to counter drift in long contexts.
* **Stop conditions**: if needed, state when to ask for missing info vs. proceed with best effort.

---

## 3) Formatting & Delimiters (Markdown · XML tags · JSON)

**Goal:** Use the *least* structure that guarantees reliability. Prefer formats the *consumer* (human or code) can parse with minimal ambiguity.

### When to use what

* **Markdown** → Human‑readable docs, short analyses, or emails. Use `#` headers for sections; lists for rules; fenced code blocks for examples/output.
* **JSON** → Machine‑consumed outputs or deterministic parsing. Require **one top‑level object** with fixed keys. Forbid extra prose.
* **XML‑style tags / Sentinels** → To *fence inputs* or *demarcate sections* inside the prompt where Markdown isn’t precise enough. Examples: `<INPUT>…</INPUT>`, `<<<INPUT_START` … `INPUT_END>>>`.
* **Tables (Markdown)** → Human scanning of structured results. Avoid if a downstream parser needs the data—use JSON instead.
* **YAML** → Only if you control both ends and indentation won’t be an issue. JSON is safer for strict parsing.

### Reliable fencing patterns

```text
# INPUT (fenced with sentinels)
<<<INPUT_START
<user_email>
Subject: …
Body: …
</user_email>
INPUT_END>>>
```

```text
# INPUT (XML‑style)
<INPUT>
  <customer_id>cst_123</customer_id>
  <notes>VIP tier</notes>
</INPUT>
```

````markdown
# INPUT (Markdown code fence)
```txt
paste raw text here
````

````

**Tips**
- Don’t nest code fences; switch languages (` ```json`, ` ```txt`).
- Use **block labels** (INPUT, POLICY, EXAMPLES) so the model understands purpose.

### JSON output contracts (strict)
Ask for JSON **and nothing else**. Provide a shape and allowed values.
```text
Return **only** this JSON object in a single code block, no prose before/after.
Schema (informal):
{
  "status": "ok" | "needs_info" | "refuse",
  "summary": string,
  "actions": string[],
  "confidence": 0..1
}
If information is missing, set status="needs_info" and include missing fields in actions.
````

**Prompt snippet**

````markdown
# OUTPUT CONTRACT
Respond with **one** code block containing strict JSON.
```json
{
  "status": "ok",
  "summary": "…",
  "actions": [],
  "confidence": 0.0
}
````

No markdown or text outside the JSON code block.

````

**Common failure guards**
- Add a **validator hint**: “If you can’t populate a field, use `null` or an empty array.”
- Provide **enums** and ranges (e.g., `0..1` for confidence) to reduce creative drift.
- For multi‑item outputs, require **JSONL** or a top‑level array only if the consumer expects it.

### Hybrid patterns (text + machine data)
If you truly need both human text and JSON, separate them clearly and label them:
```text
# HUMAN SUMMARY
<one short paragraph>

# MACHINE OUTPUT (JSON ONLY)
```json
{ … }
````

````
Or enforce order with sentinels:
```text
BEGIN_SUMMARY
<one short paragraph>
END_SUMMARY
BEGIN_JSON
{ … }
END_JSON
````

### Do / Don’t

* **Do**: Keep headings and rule lists short; prefer bullets over paragraphs.
* **Do**: Use explicit **“No extra text.”** when you need pure JSON.
* **Don’t**: Mix analysis prose inside a JSON code block.
* **Don’t**: Overuse decorative formatting; it dilutes signal.

### Micro‑checklist

* [ ] Zero/one/few chosen for task complexity.
* [ ] Examples short, canonical, diverse.
* [ ] Fenced and labeled (INPUT/OUTPUT).
* [ ] Outputs exactly match the contract.
* [ ] Optional: one concise counter‑example to block a failure mode.

---

## 7) Long‑Context Discipline

**Goal:** Keep only what the model needs *right now*. Placement beats volume.

### Keep the window tight

* **Rolling window**: include only the last N turns needed for the current task.
* **Summarize history**: compress older turns into 3–6 bullet facts the task depends on.
* **Pin critical rules**: repeat 1–2 non‑negotiables at both **top** and **bottom**.

### Quarantine and label

* Fence large pastes with sentinels and labels (`INPUT`, `POLICY`, `NOTES`).
* Use short **headnotes** above long inputs: “This is a product manual. Extract specs only.”

### Prevent conflict & poisoning

* **Order of authority**: System/Role > Global Rules > Task > Inputs > Examples.
* If Inputs contradict rules, tell the model to **ignore** the conflicting part and proceed or ask.
* Add a simple **conflict check**: “If instructions conflict, prefer System/Global Rules and state the conflict briefly.”

### Summarize with intent

Provide a target‑aware summary format:

```text
# CONTEXT SUMMARY (for task)
- Product: …
- Constraints: …
- Decision criteria: …
```

### Micro‑checklist

* [ ] Only last N relevant turns included.
* [ ] Older content summarized as task facts.
* [ ] Conflicts resolved by explicit authority order.
* [ ] Inputs fenced; headnote added.
* [ ] Non‑negotiables mirrored at bottom.

---

## 8) Tool‑Use Prompting (for API/tool‑calling setups)

**Goal:** Make tool calls predictable, minimal, and correct.

### Define tools clearly

```text
# TOOLS
- search(query: string) → results[]  // use for web lookup when facts are missing
- calc(expr: string) → number        // use for arithmetic; prefer exact math to estimation
- send_email(to, subject, body) → id // only after explicit user approval
```

### Usage rules

* **When to call**: “Call a tool **only** if required to complete the task or verify a critical fact.”
* **Ask‑before‑acting**: “If an action affects external systems (email, file writes), ask for confirmation.”
* **Minimal viable toolset**: include only tools you actually use this turn.
* **Parameter discipline**: echo exact param names and types; avoid extra keys.
* **Failure handling**: if a tool fails, retry **once** with a corrected argument; otherwise report the failure in the output channel.
* **Cost/latency guard**: prefer fewer, richer calls over many tiny calls.

### Selection logic (in prompt)

```text
# TOOL SELECTION
- If the task is answerable from given INPUT, do not call tools.
- If a numeric calculation is needed, call calc().
- If up‑to‑date facts are needed and not provided, call search().
- If any required parameter is missing, ask 1 clarifying question instead of calling tools.
```

### Output shape for tool calls

If your platform expects a structured decision, ask for it:

```json
{
  "should_call_tool": true | false,
  "tool_name": "search" | "calc" | null,
  "arguments": {"query": "…"},
  "reason": "one short sentence"
}
```

### Micro‑checklist

* [ ] Tools named, described, and typed.
* [ ] Clear rules for when to call vs. ask.
* [ ] Parameter names exact; no extras.
* [ ] Single retry policy documented.
* [ ] Structured tool decision optional but available.

---

## 9) Reasoning Scaffolds (lightweight)

**Goal:** Improve reliability with *minimal* planning. Avoid verbose chain‑of‑thought; prefer short, outcome‑focused scaffolds.

### Patterns

* **Plan‑then‑answer (concise):**

```text
# PLAN (brief)
- Step 1: …
- Step 2: …
- Step 3: …
# ANSWER
<final answer only>
```

* **Checklists:** Provide a 3–6 item checklist the model must tick before answering.
* **Deliberate verify:** “Do a silent self‑check for math/logic; if an issue is found, correct before answering.”
* **Rubric‑guided output:** Give criteria (e.g., accuracy, brevity, relevance) and ask the model to meet them without narrating thought.

### What to avoid

* Long reasoning dumps; they increase latency and risk overfitting to the scaffold.
* Exposing hidden scratchpads when you only need the result.

### Micro‑checklist

* [ ] Brief plan or checklist only when needed.
* [ ] No verbose reasoning in final output.
* [ ] Include a one‑line verify step for math/logic.

---

## 10) Safety, Boundaries & Refusal Patterns

**Goal:** Make refusals consistent and useful while preventing oversharing.

### Structure

Place safety right after **Global Rules** (high authority, high visibility).

```text
# SAFETY & BOUNDARIES
- Do not provide: medical, legal, or financial advice.
- No disallowed content (violence, sexual content involving minors, etc.).
- If asked to break rules, refuse briefly and offer a safe alternative.
- If instructions conflict, follow System/Global Rules.
```

### Refusal pattern (concise)

```text
I can’t help with that. Here’s a safer alternative you can try: <one‑line suggestion>.
```

### Guardrails

* **Scope first**: restate what you *can* do.
* **No meta‑preambles**: refuse in one sentence; then offer one helpful alternative.
* **Ambiguity handling**: if a request might cross a line, ask a clarifying question before refusing.

### Micro‑checklist

* [ ] Safety block near the top.
* [ ] Explicit “can do / can’t do” bullets.
* [ ] Short, helpful refusal template.
* [ ] Conflict priority stated.

---

## 11) Voice‑First Prompting Addendum (prompt‑only rules)

**Goal:** Tweak prompts so spoken output sounds natural without extra systems work.

### Defaults

* **Brevity:** 2–3 sentences by default; allow expansion on request.
* **Numbers:** Spell out small numbers; prefer “five” over “5” unless codes/IDs.
* **Tone:** Friendly, clear, conversational; avoid formal flourishes.
* **Pacing:** Allow one short pause (", …") in longer answers.

### Snippet

```text
# VOICE RULES
- Keep replies to 2–3 sentences.
- Prefer everyday words and contractions.
- Spell out numbers up to ten; preserve codes/IDs in digits.
- Avoid phrases like “As an AI,” “Certainly,” or long disclaimers.
```

### Micro‑checklist

* [ ] Brevity and tone set for speech.
* [ ] Number handling specified.
* [ ] Banned robotic phrases listed.

---

## 12) Templates & Checklists

### A) General Task Template

```text
# ROLE
You are a <specialist>. Your job: <one‑line mission>.

# GLOBAL RULES
- Audience: <who>
- Tone: <2–3 words>
- Brevity: ≤ 3 sentences
- Bans: no preambles, no “As an AI,”

# SAFETY & BOUNDARIES
- <short list>

# TASK
<one sentence goal>

# INPUT
<<<INPUT_START
{minimal context}
INPUT_END>>>

# EXAMPLES (optional)
…

# OUTPUT CONTRACT
… (strict JSON or specified format)

# REMEMBER
- <restate 1–2 non‑negotiables>
```

### B) Extract‑to‑JSON Template

````text
# ROLE
You extract structured data from text.

# GLOBAL RULES
- Output JSON only, one code block, no prose.
- If a value is unknown, use null.

# TASK
Extract fields from INPUT.

# INPUT
```txt
{paste}
````

# OUTPUT CONTRACT

```json
{"id":"","name":"","amount":0,"currency":"USD","notes":null}
```

````

### C) Tool‑Calling Agent Template
```text
# ROLE
You decide whether to call tools to complete the task.

# TOOLS
- search(query: string)
- calc(expr: string)

# RULES
- Call a tool only if required.
- If parameters are missing, ask 1 question.
- External actions require user confirmation.

# TASK
<goal>

# INPUT
<<<INPUT_START
{context}
INPUT_END>>>

# OUTPUT CONTRACT
```json
{"should_call_tool":false,"tool_name":null,"arguments":{},"reason":""}
````

````

### D) Analysis Brief Template
```text
# ROLE
You produce tight, decision‑ready briefs.

# GLOBAL RULES
- Answer‑first; then one reason or example.
- ≤ 120 words unless asked to expand.

# TASK
<analysis goal>

# INPUT
```txt
{notes}
````

# OUTPUT FORMAT

* Heading (≤ 10 words)
* 3 bullets: finding, implication, action

```

### Pre‑flight Checklist
- [ ] Role and Global Rules are short and at the top.
- [ ] Task is a single, clear sentence.
- [ ] Inputs are fenced and minimal.
- [ ] Examples (if any) are canonical and short.
- [ ] Output contract is strict and parseable.
- [ ] Non‑negotiables mirrored at the bottom.
- [ ] Safety block present near top.

```
