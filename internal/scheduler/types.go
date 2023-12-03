package scheduler

import (
	"context"

	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

// interfaces purely for faking purposes
type queueFactory interface {
	newQueue(opts queueOptions) eventHandler
}

type eventHandler interface {
	handleRun(context.Context, *run.Run) error
	handleWorkspace(context.Context, *workspace.Workspace) error
}
