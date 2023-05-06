package scheduler

import (
	"context"

	internal "github.com/leg100/otf"
)

// interfaces purely for faking purposes
type queueFactory interface {
	newQueue(opts queueOptions) eventHandler
}

type eventHandler interface {
	handleEvent(context.Context, internal.Event) error
}
