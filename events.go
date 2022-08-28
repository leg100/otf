package otf

const (
	EventOrganizationCreated EventType = "organization_created"
	EventOrganizationDeleted EventType = "organization_deleted"
	EventWorkspaceCreated    EventType = "workspace_created"
	EventWorkspaceDeleted    EventType = "workspace_deleted"
	// EventLatestRunUpdate is an update to the "latest" run for a workspace
	EventLatestRunUpdate EventType = "run_latest_update"
	EventRunCreated      EventType = "run_created"
	EventRunStatusUpdate EventType = "run_status_update"
	EventRunDeleted      EventType = "run_deleted"
	EventRunCancel       EventType = "run_cancel"
	EventRunForceCancel  EventType = "run_force_cancel"
	EventError           EventType = "error"
)

// EventType identifies the type of event
type EventType string

// Event represents an event in the lifecycle of an oTF resource
type Event struct {
	Type    EventType
	Payload interface{}
}

// EventService allows interacting with events
type EventService interface {
	Publish(Event)
	Subscribe(id string) (Subscription, error)
}

// Subscription represents a stream of events for a subscriber
type Subscription interface {
	// Event stream for all subscriber's event.
	C() <-chan Event

	// Closes the event stream channel and disconnects from the event service.
	Close() error
}
