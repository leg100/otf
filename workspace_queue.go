package otf

// WorkspaceQueue is a queue of runs for a workspace
type WorkspaceQueue interface {
	// Update updates a queue in response to a run event and emits a new event
	// if there has been a change to the queue order
	Update(*Event) *Event
	// Get retrieves the queue of runs, active first.
	Get() []*Run
}
