package run

import (
	"context"

	"github.com/leg100/otf"
)

type fakeSubscriber struct {
	ch chan otf.Event

	otf.PubSubService
}

func (f *fakeSubscriber) Subscribe(context.Context, string) (<-chan otf.Event, error) {
	return f.ch, nil
}
