package scheduler

import (
	"context"

	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
)

type fakeQueueFactory struct {
	q *fakeQueue
}

func (f *fakeQueueFactory) newQueue(queueOptions) eventHandler {
	f.q = &fakeQueue{}
	return f.q
}

type fakeQueue struct {
	gotWorkspace *workspace.Workspace
	gotRun       *run.Run
}

func (q *fakeQueue) handleWorkspace(ctx context.Context, ws *workspace.Workspace) error {
	q.gotWorkspace = ws
	return nil
}

func (q *fakeQueue) handleRun(ctx context.Context, run *run.Run) error {
	q.gotRun = run
	return nil
}
