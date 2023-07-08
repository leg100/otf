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

	InsertDBAction = "INSERT"
	UpdateDBAction = "UPDATE"
	DeleteDBAction = "DELETE"
)

type (
	// EventType identifies the type of event
	EventType string

	// Event represents an event in the lifecycle of an otf resource
	Event struct {
		Type    EventType
		Payload any
		Local   bool // for local node only and not to be published to rest of cluster
	}
	Table string

	// DBAction is the action carried out on a database row
	DBAction string
)

func NewCreatedEvent(payload any) Event {
	return Event{Type: CreatedEvent, Payload: payload}
}

func NewUpdatedEvent(payload any) Event {
	return Event{Type: UpdatedEvent, Payload: payload}
}

func NewDeletedEvent(payload any) Event {
	return Event{Type: DeletedEvent, Payload: payload}
}
