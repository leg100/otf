package agent

import (
	"context"

	"github.com/leg100/otf/internal/client"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
)

type fakeSpoolerApp struct {
	runs   []*run.Run
	events chan pubsub.Event

	client.Client
}

func (a *fakeSpoolerApp) ListRuns(ctx context.Context, opts run.RunListOptions) (*run.RunList, error) {
	return &run.RunList{
		Items:      a.runs,
		Pagination: resource.NewPagination(resource.PageOptions{}, int64(len(a.runs))),
	}, nil
}

func (a *fakeSpoolerApp) Watch(ctx context.Context, opts run.WatchOptions) (<-chan pubsub.Event, error) {
	// the non-fake watch takes care of closing channel when context is
	// terminated.
	go func() {
		<-ctx.Done()
		close(a.events)
	}()
	return a.events, nil
}
