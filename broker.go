package internal

import (
	"context"
	"reflect"
)

type (
	Broker interface {
		PubSubService
		Register(t reflect.Type, getter Getter)
	}

	// Getter retrieves an event payload using its ID.
	Getter interface {
		GetByID(context.Context, string) (any, error)
	}
)
