package ots

const (
	OrganizationCreated   EventType = "organization_created"
	OrganizationDeleted   EventType = "organization_deleted"
	WorkspaceCreated      EventType = "workspace_created"
	WorkspaceDeleted      EventType = "workspace_deleted"
	RunCreated            EventType = "run_created"
	RunCompleted          EventType = "run_completed"
	RunCanceled           EventType = "run_canceled"
	RunApplied            EventType = "run_applied"
	RunPlanned            EventType = "run_planned"
	RunPlannedAndFinished EventType = "run_planned_and_finished"
	PlanQueued            EventType = "plan_queued"
	ApplyQueued           EventType = "apply_queued"
)

type EventType string

type Event struct {
	Type    EventType
	Payload interface{}
}

type EventService interface {
	Publish(Event)
	Subscribe(id string) Subscription
}

// Subscription represents a stream of events for a subscriber
type Subscription interface {
	// Event stream for all subscriber's event.
	C() <-chan Event

	// Closes the event stream channel and disconnects from the event service.
	Close() error
}
