#!/bin/bash

# SubagentStop hook to update session documentation
# Triggered when a subagent (Task tool) completes

# Determine Project Root (assuming script is in .claude/hooks/)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

# Use CLAUDEX_LOG_FILE if set, otherwise fallback to local file
if [ -z "${CLAUDEX_LOG_FILE:-}" ]; then
    LOG_FILE="$PROJECT_ROOT/.claude/hooks/subagent-stop.log"
else
    LOG_FILE="$CLAUDEX_LOG_FILE"
    # Create parent directory if it doesn't exist
    LOG_DIR=$(dirname "$LOG_FILE")
    mkdir -p "$LOG_DIR" 2>/dev/null || true
fi

# Logging function with source prefix
log_message() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "$timestamp | [hook_subagent_stop] $1" >> "$LOG_FILE"
}

echo "===========================================================" >> "$LOG_FILE"
log_message "Hook triggered (SubagentStop)"

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

log_message "✅ AGENT FINISHED - Agent ID: $AGENT_ID"
log_message "Session ID: $SESSION_ID"
log_message "Transcript: $TRANSCRIPT_PATH"
log_message "CLAUDEX_SESSION_PATH: $CLAUDEX_SESSION_PATH"

# === NOTIFICATION SYSTEM ===
# Source notification library
NOTIFICATION_LIB="$(dirname "$0")/lib/notification.sh"
if [ -f "$NOTIFICATION_LIB" ]; then
    source "$NOTIFICATION_LIB"

    # Extract session name from path
    SESSION_NAME=$(basename "$CLAUDEX_SESSION_PATH" 2>/dev/null || echo "unknown")

    # Send notification
    notify_agent_complete "$AGENT_ID" "$SESSION_NAME"

    log_message "📢 Notification sent for agent: $AGENT_ID"
else
    log_message "⚠️  Notification library not found: $NOTIFICATION_LIB"
fi
# === END NOTIFICATION SYSTEM ===

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

# === NOTIFICATION SYSTEM ===
# Source notification library
NOTIFICATION_LIB="$(dirname "$0")/lib/notification.sh"
if [ -f "$NOTIFICATION_LIB" ]; then
    source "$NOTIFICATION_LIB"

    # Extract session name from path
    SESSION_NAME=$(basename "$SESSION_FOLDER" 2>/dev/null || echo "unknown")

    # Send notification
    notify_agent_complete "$AGENT_ID" "$SESSION_NAME"

    log_message "📢 Notification sent for agent: $AGENT_ID"
else
    log_message "⚠️  Notification library not found: $NOTIFICATION_LIB"
fi
# === END NOTIFICATION SYSTEM ===

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
# Update Session History (REMOVED)
# ---------------------------------------------------------
# The session history update logic has been removed as requested.
# This file is now handled by the auto-doc-updater hook (session-overview.md) or other mechanisms.

) >/dev/null 2>&1 &
# Disown to ensure it keeps running after script exits
disown

# ---------------------------------------------------------
# Update Session Overview Documentation on Agent Stop
# ---------------------------------------------------------
log_message "Starting session overview documentation update for agent: $AGENT_ID"

(
    # Determine transcript to process
    # Priority: Use main session transcript if available
    # Extract base path from agent transcript and construct main transcript path
    AGENT_TRANSCRIPT_DIR=$(dirname "$TRANSCRIPT_PATH")
    MAIN_TRANSCRIPT="${AGENT_TRANSCRIPT_DIR}/${SESSION_ID}.jsonl"

    if [ -f "$MAIN_TRANSCRIPT" ]; then
        TRANSCRIPT_TO_PROCESS="$MAIN_TRANSCRIPT"
        log_message "Using main session transcript: $MAIN_TRANSCRIPT"
    else
        # Fallback to agent transcript
        TRANSCRIPT_TO_PROCESS="$TRANSCRIPT_PATH"
        log_message "Main transcript not found, using agent transcript: $TRANSCRIPT_PATH"
    fi

    # Shared state file for incremental processing
    LAST_PROCESSED_FILE="$SESSION_FOLDER/.last-processed-line-overview"
    START_LINE=1

    if [ -f "$LAST_PROCESSED_FILE" ]; then
        START_LINE=$(cat "$LAST_PROCESSED_FILE")
        START_LINE=$((START_LINE + 1))
    fi

    TOTAL_LINES=$(wc -l < "$TRANSCRIPT_TO_PROCESS" 2>/dev/null || echo "0")

    if [ "$TOTAL_LINES" -eq 0 ] || [ "$START_LINE" -gt "$TOTAL_LINES" ]; then
        log_message "No new transcript lines to process (start: $START_LINE, total: $TOTAL_LINES)"
        exit 0
    fi

    log_message "Processing transcript lines $START_LINE to $TOTAL_LINES"

    # Extract increment
    TRANSCRIPT_INCREMENT=$(tail -n "+$START_LINE" "$TRANSCRIPT_TO_PROCESS")

    # Update marker immediately to prevent double processing
    echo "$TOTAL_LINES" > "$LAST_PROCESSED_FILE"

    # Filter relevant content (same logic as auto-doc-updater)
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

    log_message "Calling Claude to update session-overview.md (triggered by agent stop)"

    # Call Claude with recursion guard
    export CLAUDE_HOOK_INTERNAL=1
    OUTPUT=$(claude -p "$PROMPT" --model haiku 2>&1)
    EXIT_CODE=$?

    log_message "Claude finished with exit code $EXIT_CODE"
    log_message "Output summary: ${OUTPUT:0:200}..."

    # Reset PostToolUse counter to prevent redundant update shortly after
    COUNTER_FILE="$SESSION_FOLDER/.doc-update-counter"
    if [ -f "$COUNTER_FILE" ]; then
        echo "0" > "$COUNTER_FILE"
        log_message "Reset PostToolUse counter to 0 to prevent duplicate documentation update"
    fi

) >/dev/null 2>&1 &

# Disown to detach from parent shell
disown

log_message "Session overview documentation update dispatched in background"

log_message "Main script exiting"
exit 0
