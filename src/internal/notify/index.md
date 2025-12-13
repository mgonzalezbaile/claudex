# Notification System

Platform-aware notification and voice synthesis system for Claudex hooks.

## Key Files

- **notifier.go** - Notifier interface, factory, and configuration
- **macos.go** - macOS implementation using osascript and say
- **noop.go** - No-op fallback for non-macOS platforms

## Key Types

- `Notifier` - Interface for sending notifications and voice synthesis
- `Config` - Configuration for notifications and voice (enabled flags, sounds, voices)
- `Dependencies` - Dependency injection interface for Commander
- `macOSNotifier` - macOS implementation using AppleScript and say command
- `noopNotifier` - Silent no-op implementation

## Usage

Factory function `New(cfg Config, deps Dependencies)` returns platform-specific implementation. On macOS, uses `osascript` for notifications and `say` for voice synthesis. On other platforms, returns no-op notifier. Includes predefined notification types with default titles and sounds.

See [../services/hooks/](../services/hooks/) for usage in idle timeout and permission prompts.
