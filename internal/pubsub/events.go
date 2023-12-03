package pubsub

const (
	EventError       EventType = "error"
	EventInfo        EventType = "info"
	EventLogChunk    EventType = "log_update"
	EventLogFinished EventType = "log_finished"
	EventVCS         EventType = "vcs_event"

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
