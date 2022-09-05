package agent

import (
	"context"

	"github.com/leg100/otf"
)

type testRunService struct {
	runs []*otf.Run

	otf.RunService
}

func (l *testRunService) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{
		Items:      l.runs,
		Pagination: otf.NewPagination(otf.ListOptions{}, 1),
	}, nil
}

type testSubscriber struct {
	ch chan otf.Event
}

func (s *testSubscriber) Subscribe(_ context.Context) <-chan otf.Event {
	return s.ch
}
