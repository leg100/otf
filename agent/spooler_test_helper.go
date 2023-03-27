package agent

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/client"
	"github.com/leg100/otf/run"
)

type fakeSpoolerApp struct {
	runs   []*run.Run
	events chan otf.Event

	client.Client
}

func (a *fakeSpoolerApp) ListRuns(ctx context.Context, opts run.RunListOptions) (*run.RunList, error) {
	return &run.RunList{
		Items:      a.runs,
		Pagination: otf.NewPagination(otf.ListOptions{}, len(a.runs)),
	}, nil
}

func (a *fakeSpoolerApp) Watch(ctx context.Context, opts run.WatchOptions) (<-chan otf.Event, error) {
	// the non-fake watch takes care of closing channel when context is
	// terminated.
	go func() {
		<-ctx.Done()
		close(a.events)
	}()
	return a.events, nil
}
