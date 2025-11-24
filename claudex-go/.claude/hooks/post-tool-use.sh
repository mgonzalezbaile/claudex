#!/bin/bash

# PostToolUse hook to log tool usage completion
# Triggered after any tool use completes

set -euo pipefail

# Logging configuration
# Use CLAUDEX_LOG_FILE if set, otherwise fallback to local file
if [ -z "${CLAUDEX_LOG_FILE:-}" ]; then
    LOG_FILE="./post-tool-use.log"
else
    LOG_FILE="$CLAUDEX_LOG_FILE"
    # Create parent directory if it doesn't exist
    LOG_DIR=$(dirname "$LOG_FILE")
    mkdir -p "$LOG_DIR" 2>/dev/null || true
fi

# Logging function with source prefix
log_message() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "$timestamp | [hook_post_tool_use] $1" >> "$LOG_FILE"
}

# Start logging for this invocation
echo "===========================================================" >> "$LOG_FILE"
log_message "Hook triggered (PostToolUse)"

# Read JSON input from stdin
INPUT_JSON=$(cat)

# Extract relevant information
TOOL_NAME=$(echo "$INPUT_JSON" | jq -r '.tool_name // ""')
AGENT_ID=$(echo "$INPUT_JSON" | jq -r '.agent_id // ""')
SESSION_ID=$(echo "$INPUT_JSON" | jq -r '.session_id // ""')
TOOL_USE_ID=$(echo "$INPUT_JSON" | jq -r '.tool_use_id // ""')
TRANSCRIPT_PATH=$(echo "$INPUT_JSON" | jq -r '.transcript_path // ""')
STATUS=$(echo "$INPUT_JSON" | jq -r '.status // ""')

# Log tool completion
log_message "Tool: $TOOL_NAME"
log_message "Tool Use ID: $TOOL_USE_ID"
log_message "Transcript: $TRANSCRIPT_PATH"
log_message "Status: $STATUS"

# Log agent info if it's a Task tool
if [ "$TOOL_NAME" = "Task" ]; then
    log_message "Agent ID: $AGENT_ID"
    log_message "Session ID: $SESSION_ID"
fi

# Log the full input JSON for debugging (optional, can be commented out if too verbose)
log_message "========== FULL POST TOOL USE DATA =========="
log_message "$INPUT_JSON"
log_message "========== END POST TOOL USE DATA =========="

# Always allow - this is just for logging
cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "permissionDecision": "allow"
  }
}
EOF

exit 0

