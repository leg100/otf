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

func (a *fakeSpoolerApp) Watch(_ context.Context, _ run.WatchOptions) (<-chan otf.Event, error) {
	return a.events, nil
}
