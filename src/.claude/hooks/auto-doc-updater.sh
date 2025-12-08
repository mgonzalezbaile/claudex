#!/bin/bash

# Auto-Doc Updater Hook
# Updates session documentation asynchronously every N tool executions
# Triggered on PostToolUse

set -euo pipefail

# Configuration
UPDATE_FREQUENCY="${CLAUDEX_AUTODOC_FREQUENCY:-5}"

# Logging configuration
# Use CLAUDEX_LOG_FILE if set, otherwise fallback to local file
if [ -z "${CLAUDEX_LOG_FILE:-}" ]; then
    LOG_FILE="./auto-doc-updater.log"
else
    LOG_FILE="$CLAUDEX_LOG_FILE"
    # Create parent directory if it doesn't exist
    LOG_DIR=$(dirname "$LOG_FILE")
    mkdir -p "$LOG_DIR" 2>/dev/null || true
fi

# Logging function with source prefix
log_message() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "$timestamp | [hook_auto_doc_updater] $1" >> "$LOG_FILE"
}

# Output JSON response and exit
output_and_exit() {
    cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "permissionDecision": "allow"
  }
}
EOF
    exit 0
}

# Start logging for this invocation
echo "===========================================================" >> "$LOG_FILE"
log_message "Hook triggered"

# Recursion Guard: Prevent hook from triggering itself if we call Claude internally
if [ "${CLAUDE_HOOK_INTERNAL:-}" == "1" ]; then
    log_message "Recursion detected (CLAUDE_HOOK_INTERNAL=1). Exiting."
    output_and_exit
fi

# Feature toggle check
if [ "${CLAUDEX_AUTODOC_SESSION_PROGRESS:-true}" = "false" ]; then
    log_message "Auto-documentation disabled (CLAUDEX_AUTODOC_SESSION_PROGRESS=false)"
    output_and_exit
fi

# Read JSON input from stdin
INPUT_JSON=$(cat)

# Extract session info
SESSION_ID=$(echo "$INPUT_JSON" | jq -r '.session_id // ""')
TRANSCRIPT_PATH=$(echo "$INPUT_JSON" | jq -r '.transcript_path // ""')

if [ -z "$SESSION_ID" ]; then
    log_message "No session ID found. Exiting."
    output_and_exit
fi

# Determine Session Folder
# Priority 1: Use CLAUDEX_SESSION_PATH environment variable
if [ ! -z "${CLAUDEX_SESSION_PATH:-}" ] && [ -d "$CLAUDEX_SESSION_PATH" ]; then
    SESSION_FOLDER="$CLAUDEX_SESSION_PATH"
else
    # Priority 2: Try to find it relative to CWD
    CWD_SESSION_PATTERN="$(pwd)/sessions/*-${SESSION_ID}"
    SESSION_FOLDERS=($(ls -d $CWD_SESSION_PATTERN 2>/dev/null || true))
    
    if [ ${#SESSION_FOLDERS[@]} -gt 0 ]; then
        SESSION_FOLDER="${SESSION_FOLDERS[0]}"
    else
        log_message "Could not locate session folder for ID: $SESSION_ID"
        output_and_exit
    fi
fi

# Frequency Control
COUNTER_FILE="$SESSION_FOLDER/.doc-update-counter"
LAST_PROCESSED_FILE="$SESSION_FOLDER/.last-processed-line-overview"

# Initialize counter if not exists
if [ ! -f "$COUNTER_FILE" ]; then
    echo "0" > "$COUNTER_FILE"
fi

# Read current count
CURRENT_COUNT=$(cat "$COUNTER_FILE")
NEW_COUNT=$((CURRENT_COUNT + 1))

log_message "Counter: $NEW_COUNT / $UPDATE_FREQUENCY"

if [ "$NEW_COUNT" -lt "$UPDATE_FREQUENCY" ]; then
    # Update counter and exit
    echo "$NEW_COUNT" > "$COUNTER_FILE"
    log_message "Threshold not reached. Exiting."
    output_and_exit
fi

# Reset counter
echo "0" > "$COUNTER_FILE"
log_message "Threshold reached. Starting documentation update..."

if [ ! -f "$TRANSCRIPT_PATH" ]; then
    log_message "Transcript file not found: $TRANSCRIPT_PATH"
    output_and_exit
fi

# Determine where to start reading the transcript
START_LINE=1
if [ -f "$LAST_PROCESSED_FILE" ]; then
    START_LINE=$(cat "$LAST_PROCESSED_FILE")
    START_LINE=$((START_LINE + 1))
fi

# Calculate total lines
TOTAL_LINES=$(wc -l < "$TRANSCRIPT_PATH")

if [ "$START_LINE" -gt "$TOTAL_LINES" ]; then
    log_message "No new lines in transcript. Exiting."
    output_and_exit
fi

log_message "Processing transcript lines $START_LINE to $TOTAL_LINES"

# Extract the increment
# We use tail +n to start from line n
TRANSCRIPT_INCREMENT=$(tail -n "+$START_LINE" "$TRANSCRIPT_PATH")

# Update the last processed line marker immediately to avoid double processing
echo "$TOTAL_LINES" > "$LAST_PROCESSED_FILE"

# Run the heavy processing in background
(
    # Filter relevant content to reduce token usage
    # We want:
    # 1. Assistant messages (what Claude said)
    # 2. Tool results from Task tools (completed sub-agents)
    # We ignore: User messages (usually short), other tool uses/results
    
    log_message "Background process started for analysis..."
    
    # Extract relevant parts using jq
    # This filters for:
    # - Type: "assistant" (Claude's thoughts/replies) - extract only the text content
    # - Type: "user" with completed Task tool (Sub-agent results) - extract only the agent's response
    RELEVANT_CONTENT=$(echo "$TRANSCRIPT_INCREMENT" | jq -c '
        if .type == "assistant" and .message.content then
            {
                type: "assistant_message",
                timestamp: .timestamp,
                content: [.message.content[] | select(.type == "text") | .text]
            }
        elif (.type == "user" and .toolUseResult.status == "completed" and .toolUseResult.agentId != null and .toolUseResult.agentId != "") then
            {
                type: "agent_result",
                timestamp: .timestamp,
                agentId: .toolUseResult.agentId,
                content: [.toolUseResult.content[] | select(.type == "text") | .text]
            }
        else
            empty
        end
    ')
    
    CONTENT_LENGTH=$(echo "$RELEVANT_CONTENT" | wc -c)
    log_message "Extracted relevant content ($CONTENT_LENGTH bytes)"
    
    if [ -z "$RELEVANT_CONTENT" ]; then
        log_message "No relevant content found in increment. Exiting background process."
        exit 0
    fi
    
    # List existing documentation files
    DOC_FILES=$(ls -1 "$SESSION_FOLDER"/*.md 2>/dev/null | grep -v "session-history.md" || echo "")
    
    DOC_CONTEXT=""
    if [ ! -z "$DOC_FILES" ]; then
        DOC_CONTEXT="Existing documentation files:\n"
        for f in $DOC_FILES; do
            DOC_CONTEXT+="- $(basename "$f")\n"
        done
    else
        DOC_CONTEXT="No existing documentation files."
    fi
    
    # Load shared prompt template
    HOOKS_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROMPT_TEMPLATE=$(cat "$HOOKS_DIR/prompts/session-overview-documenter.md")

    # Substitute variables
    PROMPT=$(eval "echo \"$PROMPT_TEMPLATE\"")

    log_message "Calling Claude to update documentation..."
    
    # Call Claude with recursion guard
    # We use -p to provide the prompt
    export CLAUDE_HOOK_INTERNAL=1
    
    # We capture output but primarily rely on Claude using tools to write files
    OUTPUT=$(claude -p "$PROMPT" --model haiku 2>&1)
    EXIT_CODE=$?
    
    log_message "Claude finished with exit code $EXIT_CODE"
    log_message "Output summary: ${OUTPUT:0:200}..."

) >/dev/null 2>&1 &

# Disown to detach
disown

# Always allow the original tool use to proceed
cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "permissionDecision": "allow"
  }
}
EOF

exit 0

