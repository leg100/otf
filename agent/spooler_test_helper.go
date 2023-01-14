package agent

import (
	"context"

	"github.com/leg100/otf"
)

type fakeSpoolerApp struct {
	runs   []*otf.Run
	events chan otf.Event

	otf.Application
}

func (a *fakeSpoolerApp) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{
		Items:      a.runs,
		Pagination: otf.NewPagination(otf.ListOptions{}, len(a.runs)),
	}, nil
}

func (a *fakeSpoolerApp) Watch(_ context.Context, _ otf.WatchOptions) (<-chan otf.Event, error) {
	return a.events, nil
}
