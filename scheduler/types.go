package scheduler

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

// interfaces purely for faking purposes
type queueFactory interface {
	newQueue(app otf.Application, logger logr.Logger, ws *otf.Workspace) eventHandler
}

type eventHandler interface {
	handleEvent(context.Context, otf.Event) error
}
