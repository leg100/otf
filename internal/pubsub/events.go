package pubsub

import "log/slog"

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
	}
)

func NewCreatedEvent[T any](payload T) Event[T] {
	return Event[T]{Type: CreatedEvent, Payload: payload}
}

func NewUpdatedEvent[T any](payload T) Event[T] {
	return Event[T]{Type: UpdatedEvent, Payload: payload}
}

func NewDeletedEvent[T any](payload T) Event[T] {
	return Event[T]{Type: DeletedEvent, Payload: payload}
}

func (e Event[T]) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("type", string(e.Type)),
		slog.Any("payload", e.Payload),
	}
	return slog.GroupValue(attrs...)
}
