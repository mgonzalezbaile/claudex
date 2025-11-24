#!/bin/bash

# SessionEnd hook for debugging/testing
# This hook is triggered when a Claude Code session ends

# Use CLAUDEX_LOG_FILE if set, otherwise fallback to local file
if [ -z "${CLAUDEX_LOG_FILE:-}" ]; then
    LOG_FILE="./session-end.log"
else
    LOG_FILE="$CLAUDEX_LOG_FILE"
    # Create parent directory if it doesn't exist
    LOG_DIR=$(dirname "$LOG_FILE")
    mkdir -p "$LOG_DIR" 2>/dev/null || true
fi

# Prevent infinite loop - check if we're already in a hook
if [ "$IN_SESSION_END_HOOK" = "1" ]; then
    echo "$(date '+%Y-%m-%d %H:%M:%S') | [hook_session_end] Skipping hook (already in hook context)" >> "$LOG_FILE"
    exit 0
fi

# Set flag to prevent recursive hook calls
export IN_SESSION_END_HOOK=1

# Read JSON input from stdin
INPUT_JSON=$(cat)

# Extract session_id using basic text processing (jq-free approach)
SESSION_ID=$(echo "$INPUT_JSON" | grep -o '"session_id":"[^"]*"' | cut -d'"' -f4)
REASON=$(echo "$INPUT_JSON" | grep -o '"reason":"[^"]*"' | cut -d'"' -f4)

# Run the slow operations in a completely detached background process
# Using nohup and redirecting stdin from /dev/null to fully detach
# Pass the CLAUDEX_LOG_FILE to the background process
nohup bash -c '
    LOG_FILE="'"$LOG_FILE"'"
    SESSION_ID="'"$SESSION_ID"'"
    REASON="'"$REASON"'"

    # Create log entry with timestamp and source prefix
    echo "===========================================================" >> "$LOG_FILE"
    echo "$(date "+%Y-%m-%d %H:%M:%S") | [hook_session_end] SessionEnd hook triggered" >> "$LOG_FILE"
    echo "$(date "+%Y-%m-%d %H:%M:%S") | [hook_session_end] Session ID: $SESSION_ID" >> "$LOG_FILE"
    echo "$(date "+%Y-%m-%d %H:%M:%S") | [hook_session_end] Reason: $REASON" >> "$LOG_FILE"
    echo "$(date "+%Y-%m-%d %H:%M:%S") | [hook_session_end] Current directory: $(pwd)" >> "$LOG_FILE"
    echo "$(date "+%Y-%m-%d %H:%M:%S") | [hook_session_end] User: $USER" >> "$LOG_FILE"
    echo "$(date "+%Y-%m-%d %H:%M:%S") | [hook_session_end] Session ended successfully" >> "$LOG_FILE"
    echo "" >> "$LOG_FILE"

    # Generate a random sentence using Claude (using full path)
    echo "$(date "+%Y-%m-%d %H:%M:%S") | [hook_session_end] Generating random sentence from Claude..." >> "$LOG_FILE"
    export IN_SESSION_END_HOOK=1
    RANDOM_SENTENCE=$(/opt/homebrew/bin/claude -p "Generate a single random creative sentence. Only output the sentence, nothing else." 2>&1)

    # Add the random sentence to the log
    echo "$(date "+%Y-%m-%d %H:%M:%S") | [hook_session_end] Random sentence: $RANDOM_SENTENCE" >> "$LOG_FILE"
    echo "===========================================================" >> "$LOG_FILE"
    echo "" >> "$LOG_FILE"
' </dev/null >/dev/null 2>&1 &

# Disown the background process so it's not tied to this shell
disown

# Exit immediately
exit 0
