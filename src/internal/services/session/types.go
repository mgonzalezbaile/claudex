// Package session provides session management services for Claudex.
// It handles session metadata, storage operations, and naming utilities.
package session

import "time"

// SessionItem represents session metadata for UI display and operations.
// It is used by both UI components and session management functions.
type SessionItem struct {
	Title       string
	Description string
	Date        string
	Created     time.Time
	ItemType    string // "new", "ephemeral", "session"
}

// FilterValue implements the list.Item interface for Bubble Tea filtering
func (i SessionItem) FilterValue() string { return i.Title }
