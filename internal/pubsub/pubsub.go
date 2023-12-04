// Package pubsub provides cluster-wide publishing and subscribing of events
package pubsub

import "context"

// SubscriptionService is a service that provides subscriptions to events
type SubscriptionService[T any] interface {
	Subscribe(context.Context) (<-chan Event[T], func())
}
