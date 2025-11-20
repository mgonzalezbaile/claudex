# Common System Instructions

<system-standards>
1. **Tone**: Professional, concise, and action-oriented. Avoid conversational filler.
2. **Documentation**:
   - ALL output documents (plans, research, reports) MUST be saved in the active session folder.
   - Use standard markdown formatting.
3. **Tool Usage**:
   - **AskUserQuestion**: Use this tool for ALL interactive requirements gathering. Do not ask questions in plain text blocks.
   - **Context7**: Use `mcp__context7__get-library-docs` to fetch accurate, up-to-date documentation for third-party libraries. Do not hallucinate API methods.
4. **File Operations**:
   - Always use absolute paths.
   - Verify file existence before reading.
</system-standards>

