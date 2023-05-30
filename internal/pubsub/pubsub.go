// Package pubsub provides cluster-wide publishing and subscribing of events
package pubsub

import "context"

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
