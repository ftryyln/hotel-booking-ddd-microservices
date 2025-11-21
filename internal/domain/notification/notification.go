package notification

import (
	"context"
	"time"
)

// Dispatcher sends notifications.
type Dispatcher interface {
	Dispatch(ctx context.Context, target, message string) error
}

// Notification represents stored notification metadata.
type Notification struct {
	ID        string
	Type      string
	Target    string
	Message   string
	CreatedAt time.Time
}
