package internal

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
	EventLogFinished         EventType = "log_finished"
	EventVCS                 EventType = "vcs_event"

	CreatedEvent EventType = "created"
	UpdatedEvent EventType = "updated"
	DeletedEvent EventType = "deleted"
)

// EventType identifies the type of event
type EventType string

// Event represents an event in the lifecycle of an otf resource
type Event struct {
	Type    EventType
	Payload interface{}
	Local   bool // for local node only and not to be published to rest of cluster
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

// Subscriber is capable of creating a subscription to events.
type Subscriber interface {
	// Subscribe subscribes the caller to OTF events. Name uniquely identifies the
	// caller.
	Subscribe(ctx context.Context, name string) (<-chan Event, error)
}
