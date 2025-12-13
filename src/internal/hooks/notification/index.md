# hooks/notification

Notification hook for sending macOS notifications with optional voice synthesis.

## Key Files

- **notifier.go** - Handler for notification events with platform-specific notification dispatch

## Key Types

- `Handler` - Processes notification events and sends notifications via `notify.Notifier`

## Behavior

1. Validates notification input (message required)
2. Logs notification processing via `shared.Logger`
3. Gets notification configuration for type via `notify.GetNotificationConfig()`
4. Sends notification via `notify.Notifier.Send()` with title, message, sound
5. If CLAUDEX_VOICE_ENABLED=true, speaks message via `notify.Notifier.Speak()`
6. Returns error on notification failure (voice failure is logged but doesn't fail hook)

## Notification Types

Notification type maps to predefined configurations (title, sound):
- Agent completion notifications
- Session lifecycle notifications
- Custom notification types

## Configuration

Voice synthesis controlled by environment variable:
- `CLAUDEX_VOICE_ENABLED=true` - Enable voice synthesis
- `CLAUDEX_VOICE_ENABLED=false` - Disable voice synthesis (default)

## Platform Support

Notifications implemented via [../notify/](../../notify/index.md):
- macOS: osascript + say
- Other platforms: no-op notifier

## Usage

Hook is invoked by application code when notification events occur. Executable located at `.claude/hooks/Notification`.
