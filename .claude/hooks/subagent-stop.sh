#!/bin/bash

# SubagentStop hook to update session documentation
# Triggered when a subagent (Task tool) completes

# Determine Project Root (assuming script is in .claude/hooks/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
LOG_FILE="$PROJECT_ROOT/.claude/hooks/subagent-stop.log"

# Logging function
log_message() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "$timestamp | $1" >> "$LOG_FILE"
}

echo "===========================================================" >> "$LOG_FILE"
log_message "Hook triggered"

# Recursion Guard: Prevent hook from triggering itself
if [ "$CLAUDE_HOOK_INTERNAL" == "1" ]; then
    log_message "Recursion detected (CLAUDE_HOOK_INTERNAL=1). Exiting."
    exit 0
fi

# Read JSON input from stdin
INPUT_JSON=$(cat)

# Extract info using jq
SESSION_ID=$(echo "$INPUT_JSON" | jq -r '.session_id // ""')
TRANSCRIPT_PATH=$(echo "$INPUT_JSON" | jq -r '.agent_transcript_path // ""')
AGENT_ID=$(echo "$INPUT_JSON" | jq -r '.agent_id // ""')

log_message "Session ID: $SESSION_ID"
log_message "Agent ID: $AGENT_ID"
log_message "Transcript: $TRANSCRIPT_PATH"
log_message "CLAUDEX_SESSION_PATH: $CLAUDEX_SESSION_PATH"

if [ -z "$SESSION_ID" ] || [ -z "$TRANSCRIPT_PATH" ]; then
    log_message "Missing session_id or transcript_path. Exiting."
    exit 0
fi

# Find session folder
# Priority 1: Use CLAUDEX_SESSION_PATH environment variable (most reliable)
if [ ! -z "$CLAUDEX_SESSION_PATH" ] && [ -d "$CLAUDEX_SESSION_PATH" ]; then
    SESSION_FOLDER="$CLAUDEX_SESSION_PATH"
    log_message "Found session folder via env var: $SESSION_FOLDER"
else
    # Priority 2: Current working directory
    CWD_SESSION_PATTERN="$(pwd)/sessions/*-${SESSION_ID}"
    SESSION_FOLDERS=($(ls -d $CWD_SESSION_PATTERN 2>/dev/null || true))
    
    if [ ${#SESSION_FOLDERS[@]} -gt 0 ]; then
        SESSION_FOLDER="${SESSION_FOLDERS[0]}"
        log_message "Found session folder via CWD: $SESSION_FOLDER"
    else
        # Priority 3: Script location (fallback)
        SCRIPT_SESSION_PATTERN="$PROJECT_ROOT/sessions/*-${SESSION_ID}"
        SESSION_FOLDERS=($(ls -d $SCRIPT_SESSION_PATTERN 2>/dev/null || true))
        
        if [ ${#SESSION_FOLDERS[@]} -gt 0 ]; then
            SESSION_FOLDER="${SESSION_FOLDERS[0]}"
            log_message "Found session folder via script path: $SESSION_FOLDER"
        else
            log_message "No session folder found. Exiting."
            exit 0
        fi
    fi
fi

HISTORY_FILE="$SESSION_FOLDER/session-history.md"

log_message "Session Folder: $SESSION_FOLDER"
log_message "History File: $HISTORY_FILE"

if [ ! -f "$TRANSCRIPT_PATH" ]; then
    log_message "Transcript file not found: $TRANSCRIPT_PATH"
    exit 0
fi

# Ensure history file exists
if [ ! -f "$HISTORY_FILE" ]; then
    echo "# Session History" > "$HISTORY_FILE"
    echo "" >> "$HISTORY_FILE"
fi

# Read existing history (last 200 lines to give context but avoid huge prompts)
EXISTING_HISTORY=$(tail -n 200 "$HISTORY_FILE")

# Read transcript (last 500 lines to capture recent activity)
TRANSCRIPT_CONTENT=$(tail -n 500 "$TRANSCRIPT_PATH")

# Start background processing to unblock parent process
(
    log_message "Background process started"
    # ---------------------------------------------------------
    # Update other documentation files in the session folder
    # ---------------------------------------------------------
log_message "Checking for other documentation files to update..."

# Find all .md files in session folder, excluding session-history.md
# We use a while loop to process each file found
find "$SESSION_FOLDER" -maxdepth 1 -name "*.md" | while read -r DOC_FILE; do
    BASENAME=$(basename "$DOC_FILE")
    
    # Skip session-history.md as it is handled separately
    if [ "$BASENAME" == "session-history.md" ]; then
        continue
    fi

    log_message "Found documentation: $BASENAME"
    
    # Read document content
    DOC_CONTENT=$(cat "$DOC_FILE")
    
    # Skip empty files or files that are too large (arbitrary limit ~100KB to be safe with context window)
    DOC_SIZE=$(wc -c < "$DOC_FILE")
    if [ "$DOC_SIZE" -gt 100000 ]; then
        log_message "Skipping $BASENAME (too large: $DOC_SIZE bytes)"
        continue
    fi

    # Construct Prompt for updating documentation
    DOC_PROMPT="You are an AI assistant maintaining documentation for a coding session.
A subagent has just completed a task.
Here is the transcript of the subagent's execution:
---
$TRANSCRIPT_CONTENT
---

Here is the current content of the document '$BASENAME':
---
$DOC_CONTENT
---

Instructions:
1. Update the document '$BASENAME' to reflect the changes and progress made by the subagent.
2. IMPORTANT: Do not remove existing information unless it is clearly obsolete or incorrect.
3. If the new information is distinct (e.g., a new feature, topic, or research result), APPEND it to the document or create a new section.
4. If the new information updates existing content, MERGE it intelligently.
5. Maintain the existing format and style of the document.
6. If the document tracks status (e.g., roadmap, todo list), mark items as complete or in progress.
7. Output ONLY the full updated content of the document. Do not include any conversational text or markdown code fences around the whole output unless they are part of the document itself."

    log_message "Updating $BASENAME with Claude..."
    
    # Call Claude to get updated content
    # Using -p as per user preference for prompt
    # Set CLAUDE_HOOK_INTERNAL=1 to prevent recursion
    UPDATED_CONTENT=$(CLAUDE_HOOK_INTERNAL=1 claude -p "$DOC_PROMPT")
    
    if [ $? -eq 0 ] && [ ! -z "$UPDATED_CONTENT" ]; then
        # Overwrite the file with new content
        echo "$UPDATED_CONTENT" > "$DOC_FILE"
        log_message "Successfully updated $BASENAME"
    else
        log_message "Failed to update $BASENAME (Claude call failed or empty output)"
    fi
done

# ---------------------------------------------------------
# Update Session History
# ---------------------------------------------------------

# Construct Prompt for Claude
PROMPT="You are an AI assistant helping to document a coding session.
A subagent (Agent ID: $AGENT_ID) has just completed a task.
Your goal is to append a concise summary of what the subagent did to the session history.

Here is the recent content of the session history (for context):
---
$EXISTING_HISTORY
---

Here is the transcript of the subagent's execution (JSONL format):
---
$TRANSCRIPT_CONTENT
---

Instructions:
1. Analyze the transcript to understand what the subagent accomplished.
2. Look for tool uses (file edits, commands run) and their outcomes.
3. Identify if the task was completed successfully or if there were errors.
4. Generate a markdown entry to append to the history.
5. The entry should start with a header: '## 🤖 Subagent Execution'
6. Include the Agent ID and a Status (Completed/Failed).
7. Provide a bulleted summary of actions taken.
8. Be concise. Do not repeat the entire history.
9. Output ONLY the markdown to be appended. Do not include any conversational text."

# Call Claude
log_message "Calling Claude to analyze transcript for history..."
SUMMARY=$(CLAUDE_HOOK_INTERNAL=1 claude -p "$PROMPT")

if [ $? -eq 0 ] && [ ! -z "$SUMMARY" ]; then
    echo "" >> "$HISTORY_FILE"
    echo "$SUMMARY" >> "$HISTORY_FILE"
    echo "" >> "$HISTORY_FILE"
    echo "---" >> "$HISTORY_FILE"
    log_message "Updated session history with Claude's analysis."
else
    log_message "Claude call failed or returned empty. Fallback to simple logging."
    
    # Fallback logic (original simple append)
    TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
    echo "" >> "$HISTORY_FILE"
    echo "## 🤖 Agent Execution ($TIMESTAMP)" >> "$HISTORY_FILE"
    echo "**Agent ID**: \`$AGENT_ID\`" >> "$HISTORY_FILE"
    echo "**Status**: Completed (Analysis Failed)" >> "$HISTORY_FILE"
    echo "" >> "$HISTORY_FILE"
    echo "*(Claude analysis failed, check logs)*" >> "$HISTORY_FILE"
    echo "" >> "$HISTORY_FILE"
    echo "---" >> "$HISTORY_FILE"
fi

) >/dev/null 2>&1 &
# Disown to ensure it keeps running after script exits
disown

log_message "Main script exiting"
exit 0
