<!-- Powered by BMAD™ Core -->

# Create Output Contract Task

## ⚠️ CRITICAL EXECUTION NOTICE ⚠️

**THIS IS AN EXECUTABLE WORKFLOW - NOT REFERENCE MATERIAL**

This task designs strict, parseable output contracts (schemas) for LLM responses.

## Instructions

### 1. Gather Requirements

Ask user about the use case:

**Required Information:**
1. **What data needs to be extracted/generated?** (list the fields/information)
2. **Who/what consumes this output?** (human, code, database, API, etc.)
3. **Preferred format?** (JSON, Markdown, CSV, table, or let me recommend)

**Optional Information:**
4. **Edge cases?** (missing data, empty results, errors - how to handle?)
5. **Validation needs?** (required fields, value constraints, relationships)
6. **Existing schema?** (do you have a partial schema or examples?)

### 2. Determine Optimal Format

Based on consumer, recommend format:

**JSON** → Best for:
- Machine consumption (APIs, databases)
- Strict parsing requirements
- Complex nested structures
- Type safety needs
- Automated validation

**Markdown** → Best for:
- Human consumption
- Documentation/reports
- Mixed prose and structure
- Readability priority

**CSV/Table** → Best for:
- Tabular data
- Spreadsheet import
- Simple flat structures
- Bulk data export

**Hybrid** → Best for:
- Human AND machine consumption
- Summary text + structured data
- Reports with embedded data

Present recommendation with rationale and confirm with user.

### 3. Design Schema

Based on format choice, design contract:

---

#### For JSON Schemas

**Structure the schema with:**

1. **Top-level object** (always use one root object, not loose fields)
2. **Field definitions** with:
   - Type: `string | integer | boolean | array | object | null`
   - Constraints: enums, ranges, patterns
   - Required vs optional
   - Default/fallback values

**Example Schema Design Process:**

**User Need**: "Extract product info: name, price, categories, availability"

**Schema Draft**:
```json
{
  "name": string,                    // required, never null
  "price": number,                   // required, positive, 2 decimals
  "currency": string,                // required, 3-letter code (USD, EUR)
  "categories": string[],            // required, empty array [] if none
  "in_stock": boolean,               // required, true/false
  "quantity": integer | null,        // optional, null if unknown
  "notes": string | null             // optional, null if none
}
```

**Validation Rules to Add**:
```
- name: non-empty string, max 200 chars
- price: positive number, exactly 2 decimal places
- currency: must be valid ISO 4217 code
- categories: array of 1-20 strings, each max 50 chars
- quantity: if present, must be >= 0
```

**Edge Case Handling**:
```
- Missing required field → Set to sensible default or return error status
- Invalid format → Explain in separate error field
- Empty result → Return empty array or null with reason
```

---

#### For Markdown Schemas

**Define structure with:**

1. **Heading hierarchy** (required # ## ### levels)
2. **Section requirements** (what must be in each section)
3. **List formats** (bullet, numbered, checklist)
4. **Table schema** (if tables used)
5. **Content constraints** (length, tone, style)

**Example Markdown Schema**:

```markdown
# [Product Name] - required, max 100 chars

## Overview - required section
[2-3 sentence description of product]

## Specifications - required section
- **Price**: [format: $X.XX USD]
- **Availability**: [In Stock | Out of Stock | Pre-Order]
- **Categories**: [comma-separated list]

## Additional Notes - optional section
[Any extra information, max 200 words]
```

---

#### For Table Schemas

**Define columns with:**

1. **Column names** (exact header text)
2. **Data types** (string, number, boolean, date)
3. **Constraints** (max length, format, required/optional)
4. **Sort order** (if applicable)

**Example Table Schema**:

```
| Column        | Type    | Required | Constraints              |
|---------------|---------|----------|--------------------------|
| Product Name  | string  | yes      | max 100 chars            |
| Price         | number  | yes      | positive, 2 decimals     |
| Currency      | string  | yes      | 3-letter ISO code        |
| In Stock      | boolean | yes      | true/false only          |
| Categories    | string  | yes      | comma-separated, max 200 |
| Notes         | string  | no       | max 500 chars            |
```

---

#### For Hybrid Schemas

**Separate concerns clearly:**

```markdown
# HUMAN SUMMARY
[2-3 sentence plain English summary]

# MACHINE DATA (JSON)
```json
{
  "structured_data": "here"
}
````

Or use sentinels:
```
BEGIN_SUMMARY
[prose here]
END_SUMMARY

BEGIN_DATA
{json here}
END_DATA
```

---

### 4. Add Strict Contract Instructions

Generate the prompt instructions for the output contract:

#### For JSON:

```text
# OUTPUT CONTRACT

Return **ONLY** a single JSON object in one code block.
No text before or after the JSON.

Schema:
```json
{
  "field_name": "type (constraints)",
  "another_field": "type (constraints)"
}
````

Validation rules:
- Required fields must never be null or omitted
- Optional fields use null if data missing
- Arrays use [] if empty, never null
- Numbers must match specified precision
- Enums must be exact values (case-sensitive)

Error handling:
- If critical data missing, set "status": "incomplete"
- Include "errors": [] array with issue descriptions
- Never return partial/malformed JSON

# REMEMBER
- JSON only, one code block, no extra text
- Follow schema exactly
```

#### For Markdown:

```text
# OUTPUT CONTRACT

Return content in the following Markdown structure:

# [Required H1 Heading]

## [Required H2 Section]
[Content requirements...]

## [Optional H2 Section]
[Content requirements...]

Format rules:
- Use exact heading levels specified
- Required sections must never be empty
- Lists must use specified format (- or 1.)
- Tables must include all specified columns
- Maximum [X] words per section

# REMEMBER
- Follow exact Markdown structure
- All required sections must be present
```

#### For Tables:

```text
# OUTPUT CONTRACT

Return a Markdown table with exactly these columns:

| Column1 | Column2 | Column3 |
|---------|---------|---------|
| value   | value   | value   |

Rules:
- All rows must have all columns
- Use exact column names (case-sensitive)
- Follow data type constraints
- Sort by [column] ascending/descending
- Include header row

# REMEMBER
- Exact columns, all rows complete
```

### 5. Generate Complete Contract

Present the output contract with all components:

```markdown
# Output Contract Specification

## Format
**Type**: [JSON | Markdown | Table | Hybrid]
**Consumer**: [Who/what uses this output]

## Schema

[Full schema with types, constraints, examples]

## Validation Rules

[List all validation requirements]

## Edge Case Handling

**Missing Data**:
[How to handle missing required vs optional fields]

**Empty Results**:
[What to return when no data available]

**Errors**:
[How to communicate parsing/validation errors]

**Invalid Input**:
[How to handle malformed or unexpected input]

## Prompt Instructions

[Ready-to-use OUTPUT CONTRACT section for the prompt]

```text
# OUTPUT CONTRACT
[Complete contract text to paste into prompt]
```

## Example Output

**Valid Example**:
```
[Show what correct output looks like]
```

**Invalid Example** (DON'T):
```
[Show common mistakes to avoid]
```

## Testing Checklist

To validate this contract:

- [ ] Test with complete valid data
- [ ] Test with missing optional fields
- [ ] Test with missing required fields
- [ ] Test with empty result set
- [ ] Test with invalid data types
- [ ] Test with edge case values (very long, very short, special chars)
- [ ] Test with multiple items (if array)
- [ ] Verify machine parseability (if JSON/structured)

## Next Steps

Would you like me to:
1. **Integrate into prompt** - Add this contract to existing prompt (*optimize)
2. **Test the contract** - Generate test cases to validate
3. **Refine schema** - Adjust specific fields or constraints
4. **Create template** - Convert to reusable template (*template)
```

### 6. User Interaction

After presenting contract:
- Ask if schema meets their needs
- Offer to adjust specific fields or constraints
- Suggest test cases to validate
- Offer to integrate into their prompt

## Best Practices

### DO:
- ✅ Make required vs optional explicit
- ✅ Specify exact types and constraints
- ✅ Handle edge cases (null, empty, error)
- ✅ Provide validation hints
- ✅ Include concrete examples
- ✅ Use one top-level object for JSON (not loose fields)
- ✅ Test parseability with code/systems that will consume it

### DON'T:
- ❌ Leave ambiguity in field requirements
- ❌ Use vague types ("value" instead of "string | integer")
- ❌ Forget edge case handling
- ❌ Over-specify when loose schema works
- ❌ Mix formats without clear separation
- ❌ Assume LLM will infer constraints

## Common Patterns

### Status + Data Pattern

Good for operations that can fail:

```json
{
  "status": "success" | "incomplete" | "error",
  "data": {
    // actual result fields
  },
  "errors": string[] | null,
  "metadata": {
    "processed_at": "ISO-8601 timestamp",
    "confidence": 0.0-1.0
  }
}
```

### List + Summary Pattern

Good for multi-item results:

```json
{
  "summary": {
    "total": integer,
    "filtered": integer,
    "categories": string[]
  },
  "items": [
    // array of result objects
  ]
}
```

### Confidence + Reasoning Pattern

Good for analytical tasks:

```json
{
  "answer": string,
  "confidence": 0.0-1.0,
  "reasoning": string,
  "sources": string[]
}
```

### Partial Success Pattern

Good for bulk operations:

```json
{
  "total_processed": integer,
  "successful": integer,
  "failed": integer,
  "results": [
    {
      "id": string,
      "status": "success" | "failed",
      "data": object | null,
      "error": string | null
    }
  ]
}
```

## Type Specification Best Practices

### Strings
```json
"field": string           // any string
"field": "enum1" | "enum2"  // only these exact values
"field": string // max 200 chars  // with constraint comment
```

### Numbers
```json
"field": number           // any number
"field": integer          // whole numbers only
"field": number // 0.0-1.0  // with range constraint
"field": number // 2 decimals  // with precision constraint
```

### Booleans
```json
"field": boolean          // true or false only
```

### Arrays
```json
"field": string[]         // array of strings
"field": object[]         // array of objects
"field": [] // empty if none  // never null
```

### Optional Fields
```json
"field": string | null    // optional, use null if missing
"field": string // optional, omit if not available  // can be omitted
```

### Nested Objects
```json
"field": {
  "subfield1": type,
  "subfield2": type
}
```

## Format Selection Guide

| Consumer Type | Recommended Format | Rationale |
|---------------|-------------------|-----------|
| REST API | JSON | Standard API format, strict parsing |
| Database Insert | JSON | Direct ORM mapping, type safety |
| Frontend UI | JSON | Easy JavaScript parsing |
| Report/Email | Markdown | Human-readable, formatted |
| Spreadsheet | Table/CSV | Direct import, tabular structure |
| Log/Audit | Markdown | Readable, searchable |
| Both Human & Code | Hybrid | Summary for humans, data for code |
| Voice Assistant | Plain Text | No structure needed, conversational |
