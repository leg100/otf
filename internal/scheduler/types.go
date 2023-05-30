package scheduler

import (
	"context"

	"github.com/leg100/otf/internal/pubsub"
)

// interfaces purely for faking purposes
type queueFactory interface {
	newQueue(opts queueOptions) eventHandler
}

type eventHandler interface {
	handleEvent(context.Context, pubsub.Event) error
}
