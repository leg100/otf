package pubsub

import (
	"log/slog"
	"time"
)

const (
	CreatedEvent EventType = "created"
	UpdatedEvent EventType = "updated"
	DeletedEvent EventType = "deleted"
)

type (
	// EventType identifies the type of event
	EventType string

	// Event represents an event in the lifecycle of an otf resource
	Event[T any] struct {
		Type    EventType
		Payload T
		Time    time.Time
	}
)

func (e Event[T]) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("type", string(e.Type)),
		slog.Any("payload", e.Payload),
	}
	return slog.GroupValue(attrs...)
}
