// Package pubsub provides cluster-wide publishing and subscribing of events
package pubsub

import "context"

// SubscriptionService is a service that provides subscriptions to events
type SubscriptionService[T any] interface {
	Subscribe() (<-chan Event[T], func())
	SubscribeWithContext(ctx context.Context) <-chan Event[T]
}
