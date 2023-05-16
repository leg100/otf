package internal

import (
	"context"
	"reflect"
)

type (
	Broker interface {
		PubSubService

		// Register a type with the broker. Registration allows the broker to
		// relay events via postgres, giving the broker the means to marshal or
		// unmarshal event payloads, and to emit events when updates and deletes
		// cascade on the table associated with the type.
		//
		// The getter provides a fallback mechanism: when the event payload
		// exceeds postgres' maximum size then only the ID of the type is
		// relayed and the receiver can fetch the event using the getter.
		Register(t reflect.Type, table string, getter Getter)
	}

	// Getter retrieves an event payload using its ID.
	Getter interface {
		GetByID(context.Context, string) (any, error)
	}
)
