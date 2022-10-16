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
)

// EventType identifies the type of event
type EventType string

// Event represents an event in the lifecycle of an oTF resource
type Event struct {
	Type    EventType
	Payload interface{}
}

// PubSubService provides low-level access to pub-sub behaviours. Access is
// unauthenticated.
type PubSubService interface {
	// Publish an event
	Publish(Event)
	// Subscribe creates a subscription to a stream of errors. Name is a
	// unique identifier describing the subscriber.
	Subscribe(ctx context.Context, name string) (<-chan Event, error)
}

// EventService allows interacting with events. Access is authenticated.
type EventService interface {
	// Watch provides access to a stream of events. The WatchOptions filters
	// events.
	Watch(context.Context, WatchOptions) (<-chan Event, error)
	// WatchLogs provides access to a stream of phase logs. The WatchLogsOptions filters
	// events.
	WatchLogs(context.Context, WatchLogsOptions) (<-chan Chunk, error)
}

// WatchOptions filters events returned by the Watch endpoint.
type WatchOptions struct {
	// Name to uniquely describe the watcher. If not provided then a
	// name will be auto generated.
	Name *string
	// Filter by organization name
	OrganizationName *string `schema:"organization_name"`
	// Filter by workspace name. Must be specified in tandem with
	// OrganizationName.
	WorkspaceName *string `schema:"workspace_name"`
}

// WatchLogsOptions filters logs returned by the WatchLogs endpoint.
type WatchLogsOptions WatchOptions
