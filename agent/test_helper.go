package agent

import "github.com/leg100/otf"

type testRunLister struct {
	runs []*otf.Run
}

func (l *testRunLister) List(opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{Items: l.runs}, nil
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
