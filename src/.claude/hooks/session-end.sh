#!/bin/bash

# SessionEnd hook to update session documentation
# This hook is triggered when a Claude Code session ends

# Determine Project Root (assuming script is in .claude/hooks/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Use CLAUDEX_LOG_FILE if set, otherwise fallback to local file
if [ -z "${CLAUDEX_LOG_FILE:-}" ]; then
    LOG_FILE="$PROJECT_ROOT/.claude/hooks/session-end.log"
else
    LOG_FILE="$CLAUDEX_LOG_FILE"
    # Create parent directory if it doesn't exist
    LOG_DIR=$(dirname "$LOG_FILE")
    mkdir -p "$LOG_DIR" 2>/dev/null || true
fi

# Logging function with source prefix
log_message() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "$timestamp | [hook_session_end] $1" >> "$LOG_FILE"
}

# Output JSON response and exit
output_and_exit() {
    cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionEnd",
    "permissionDecision": "allow"
  }
}
EOF
    exit 0
}

echo "===========================================================" >> "$LOG_FILE"
log_message "Hook triggered (SessionEnd)"

# Recursion Guard: Prevent hook from triggering itself
if [ "$CLAUDE_HOOK_INTERNAL" == "1" ]; then
    log_message "Recursion detected (CLAUDE_HOOK_INTERNAL=1). Exiting."
    output_and_exit
fi

# Feature toggle check
if [ "${CLAUDEX_AUTODOC_SESSION_END:-true}" = "false" ]; then
    log_message "Session end autodoc disabled (CLAUDEX_AUTODOC_SESSION_END=false)"
    output_and_exit
fi

# Read JSON input from stdin
INPUT_JSON=$(cat)

# Extract info using jq
SESSION_ID=$(echo "$INPUT_JSON" | jq -r '.session_id // ""')
TRANSCRIPT_PATH=$(echo "$INPUT_JSON" | jq -r '.transcript_path // ""')
REASON=$(echo "$INPUT_JSON" | jq -r '.reason // ""')

log_message "Session ID: $SESSION_ID"
log_message "Transcript: $TRANSCRIPT_PATH"
log_message "Reason: $REASON"
log_message "CLAUDEX_SESSION_PATH: $CLAUDEX_SESSION_PATH"

if [ -z "$SESSION_ID" ] || [ -z "$TRANSCRIPT_PATH" ]; then
    log_message "Missing session_id or transcript_path. Exiting."
    output_and_exit
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
            output_and_exit
        fi
    fi
fi

log_message "Session Folder: $SESSION_FOLDER"

if [ ! -f "$TRANSCRIPT_PATH" ]; then
    log_message "Transcript file not found: $TRANSCRIPT_PATH"
    output_and_exit
fi

# ---------------------------------------------------------
# Update Session Overview Documentation on Session End
# ---------------------------------------------------------
log_message "Starting session overview documentation update for session end"

(
    # Shared state file for incremental processing
    LAST_PROCESSED_FILE="$SESSION_FOLDER/.last-processed-line-overview"
    START_LINE=1

    if [ -f "$LAST_PROCESSED_FILE" ]; then
        START_LINE=$(cat "$LAST_PROCESSED_FILE")
        START_LINE=$((START_LINE + 1))
    fi

    TOTAL_LINES=$(wc -l < "$TRANSCRIPT_PATH" 2>/dev/null || echo "0")

    if [ "$TOTAL_LINES" -eq 0 ] || [ "$START_LINE" -gt "$TOTAL_LINES" ]; then
        log_message "No new transcript lines to process (start: $START_LINE, total: $TOTAL_LINES)"
        exit 0
    fi

    log_message "Processing transcript lines $START_LINE to $TOTAL_LINES"

    # Extract increment
    TRANSCRIPT_INCREMENT=$(tail -n "+$START_LINE" "$TRANSCRIPT_PATH")

    # Update marker immediately to prevent double processing
    echo "$TOTAL_LINES" > "$LAST_PROCESSED_FILE"

    # Filter relevant content (assistant messages and agent results)
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

    if [ -z "$RELEVANT_CONTENT" ] || [ "$CONTENT_LENGTH" -lt 10 ]; then
        log_message "No relevant content found in transcript increment"
        exit 0
    fi

    # List existing documentation
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
    PROMPT_TEMPLATE=$(cat "$SCRIPT_DIR/prompts/session-overview-documenter.md")

    # Substitute variables
    PROMPT=$(eval "echo \"$PROMPT_TEMPLATE\"")

    log_message "Calling Claude to update session-overview.md (triggered by session end)"

    # Call Claude with recursion guard
    export CLAUDE_HOOK_INTERNAL=1
    OUTPUT=$(claude -p "$PROMPT" --model haiku 2>&1)
    EXIT_CODE=$?

    log_message "Claude finished with exit code $EXIT_CODE"
    log_message "Output summary: ${OUTPUT:0:200}..."

    # Reset PostToolUse counter to prevent redundant update
    COUNTER_FILE="$SESSION_FOLDER/.doc-update-counter"
    if [ -f "$COUNTER_FILE" ]; then
        echo "0" > "$COUNTER_FILE"
        log_message "Reset PostToolUse counter to 0"
    fi

) >/dev/null 2>&1 &

# Disown to detach from parent shell
disown

log_message "Session overview documentation update dispatched in background"
log_message "Main script exiting"

# Always allow the original session end to proceed
cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionEnd",
    "permissionDecision": "allow"
  }
}
EOF

exit 0
