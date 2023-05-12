package internal

import (
	"context"
)

type (
	Broker interface {
		PubSubService

		// Register an unmarshaler for unmarshaling database table events into OTF
		// events.
		Register(table string, unmarshaler EventUnmarshaler)
	}

	// EventUnmarshaler unmarshals a database event payload into an OTF event payload.
	EventUnmarshaler interface {
		UnmarshalEvent(ctx context.Context, payload []byte, op EventType) (any, error)
	}
)
