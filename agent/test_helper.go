package agent

import (
	"context"

	"github.com/leg100/otf"
)

type testRunService struct {
	runs []*otf.Run

	otf.RunService
}

func (l *testRunService) ListRun(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{
		Items:      l.runs,
		Pagination: otf.NewPagination(otf.ListOptions{}, 1),
	}, nil
}

type testSubscriber struct {
	sub testSubscription
}

func (s *testSubscriber) Subscribe(id string) (otf.Subscription, error) {
	return &s.sub, nil
}

type testSubscription struct {
	c chan otf.Event
}

func (s *testSubscription) C() <-chan otf.Event { return s.c }

func (s *testSubscription) Close() error { return nil }
