package domain

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event that occurred in the system.
type DomainEvent interface {
	OccurredAt() time.Time
	AggregateID() uuid.UUID
	EventType() string
}

// BaseEvent provides common fields for all domain events.
type BaseEvent struct {
	aggregateID uuid.UUID
	occurredAt  time.Time
	eventType   string
}

// NewBaseEvent creates a new base event.
func NewBaseEvent(aggregateID uuid.UUID, eventType string) BaseEvent {
	return BaseEvent{
		aggregateID: aggregateID,
		occurredAt:  time.Now(),
		eventType:   eventType,
	}
}

// OccurredAt returns when the event occurred.
func (e BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// AggregateID returns the ID of the aggregate that generated this event.
func (e BaseEvent) AggregateID() uuid.UUID {
	return e.aggregateID
}

// EventType returns the type of this event.
func (e BaseEvent) EventType() string {
	return e.eventType
}
