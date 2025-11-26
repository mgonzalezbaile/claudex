#!/bin/bash
# notification.sh - macOS notification and voice synthesis library

# Send macOS notification
# Usage: send_notification "title" "message" "sound_name"
send_notification() {
    local title="$1"
    local message="$2"
    local sound="${3:-Ping}"

    # Check if osascript is available
    if ! command -v osascript &> /dev/null; then
        echo "Warning: osascript not found, skipping notification" >&2
        return 1
    fi

    # Background execution to avoid blocking
    (
        osascript -e "display notification \"$message\" with title \"$title\" sound name \"$sound\"" 2>/dev/null
    ) &
    disown
}

# Voice synthesis announcement
# Usage: speak_message "message" "voice"
speak_message() {
    local message="$1"
    local voice="${2:-Albert}"

    # Check if say is available
    if ! command -v say &> /dev/null; then
        echo "Warning: say command not found, skipping voice" >&2
        return 1
    fi

    # Background execution to avoid blocking
    (
        say -v "$voice" "$message" 2>/dev/null
    ) &
    disown
}

# Combined notification (visual + voice)
# Usage: notify_agent_complete "agent_id" "session_name"
notify_agent_complete() {
    local agent_id="$1"
    local session_name="$2"

    # Check environment variables
    local enable_notifications="${CLAUDEX_NOTIFICATIONS_ENABLED:-true}"
    local enable_voice="${CLAUDEX_VOICE_ENABLED:-false}"

    # Format session name: remove UUID, replace dashes with spaces, title case
    # Example: "macos-notification-alerts-f2e890aa-f586-4e2d-935e-2d8b962b734f" -> "Macos Notification Alerts"
    local formatted_name=$(echo "$session_name" | sed -E 's/-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$//' | tr '-' ' ' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) tolower(substr($i,2))}1')

    # Send visual notification
    if [ "$enable_notifications" = "true" ]; then
        send_notification "$formatted_name" "Agent complete" "Ping"
    fi

    # Send voice notification
    if [ "$enable_voice" = "true" ]; then
        speak_message "Agent complete" "Albert"
    fi
}
