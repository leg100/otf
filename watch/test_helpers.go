package watch

import (
	"context"
	"net/http"

	"github.com/leg100/otf"
	"github.com/r3labs/sse/v2"
)

type fakeSubscriber struct {
	ch chan otf.Event
}

func (f *fakeSubscriber) Subscribe(context.Context, string) (<-chan otf.Event, error) {
	return f.ch, nil
}

type fakeApp struct {
	ch chan otf.Event
}

func (f *fakeApp) Watch(context.Context, otf.WatchOptions) (<-chan otf.Event, error) {
	return f.ch, nil
}

type fakeEventsServer struct {
	published chan *sse.Event
}

func (f *fakeEventsServer) CreateStream(string) *sse.Stream              { return nil }
func (f *fakeEventsServer) RemoveStream(string)                          {}
func (f *fakeEventsServer) ServeHTTP(http.ResponseWriter, *http.Request) {}
func (f *fakeEventsServer) Publish(id string, e *sse.Event) {
	f.published <- e
}
