package otf

const (
	EventOrganizationCreated   EventType = "organization_created"
	EventOrganizationDeleted   EventType = "organization_deleted"
	EventWorkspaceCreated      EventType = "workspace_created"
	EventWorkspaceDeleted      EventType = "workspace_deleted"
	EventRunCreated            EventType = "run_created"
	EventRunCompleted          EventType = "run_completed"
	EventRunCanceled           EventType = "run_canceled"
	EventRunApplied            EventType = "run_applied"
	EventRunPlanned            EventType = "run_planned"
	EventRunPlannedAndFinished EventType = "run_planned_and_finished"
	EventRunErrored            EventType = "run_errored"
	EventPlanQueueable         EventType = "plan_queueable"
	EventPlanQueued            EventType = "plan_queued"
	EventApplyQueued           EventType = "apply_queued"
	EventError                 EventType = "error"
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
