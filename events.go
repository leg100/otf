package otf

import (
	"context"
)

const (
	EventOrganizationCreated EventType = "organization_created"
	EventOrganizationDeleted EventType = "organization_deleted"
	EventWorkspaceCreated    EventType = "workspace_created"
	EventWorkspaceRenamed    EventType = "workspace_renamed"
	EventWorkspaceDeleted    EventType = "workspace_deleted"
	EventRunCreated          EventType = "run_created"
	EventRunStatusUpdate     EventType = "run_status_update"
	EventRunDeleted          EventType = "run_deleted"
	EventRunCancel           EventType = "run_cancel"
	EventRunForceCancel      EventType = "run_force_cancel"
	EventError               EventType = "error"
	EventInfo                EventType = "info"
	EventLogChunk            EventType = "log_update"
	EventVCS                 EventType = "vcs_event"
)

// EventType identifies the type of event
type EventType string

// Event represents an event in the lifecycle of an otf resource
type Event struct {
	Type    EventType
	Payload interface{}
}

// PubSubService provides low-level access to pub-sub behaviours. Access is
// unauthenticated.
type PubSubService interface {
	Publisher
	Subscriber
}

type Publisher interface {
	// Publish an event
	Publish(Event)
}

// Subscriber creates a subscription to OTF events.
type Subscriber interface {
	// Subscribe subscribes the caller to OTF events. Name uniquely identifies the
	// caller.
	Subscribe(ctx context.Context, name string) (<-chan Event, error)
}

// EventService allows interacting with events. Access is authenticated.
type EventService interface {
	// WatchLogs provides access to a stream of phase logs. The WatchLogsOptions filters
	// events. Context must be cancelled to close stream.
	//
	// TODO(@leg100): it would be clearer to the caller if the stream is closed by
	// returning a stream object with a Close() method. The calling code would
	// call WatchLogs(), and then defer a Close(), which is more readable IMO.
	WatchLogs(context.Context, WatchLogsOptions) (<-chan Chunk, error)
}

// WatchLogsOptions filters logs returned by the WatchLogs endpoint.
type WatchLogsOptions WatchOptions
