package otf

const (
	OrganizationCreated        EventType = "organization_created"
	OrganizationDeleted        EventType = "organization_deleted"
	WorkspaceCreated           EventType = "workspace_created"
	WorkspaceDeleted           EventType = "workspace_deleted"
	RunCreated                 EventType = "run_created"
	RunCompleted               EventType = "run_completed"
	EventRunCanceled           EventType = "run_canceled"
	EventRunApplied            EventType = "run_applied"
	EventRunPlanned            EventType = "run_planned"
	EventRunPlannedAndFinished EventType = "run_planned_and_finished"
	EventPlanQueued            EventType = "plan_queued"
	EventApplyQueued           EventType = "apply_queued"
	EventError                 EventType = "error"
)

type EventType string

type Event struct {
	Type    EventType
	Payload interface{}
}

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
