package watch

import (
	"context"
	"net/http"

	"github.com/leg100/otf"
	otfrun "github.com/leg100/otf/run"
	"github.com/r3labs/sse/v2"
)

type fakePubSubService struct {
	ch chan otf.Event

	otf.PubSubService
}

func (f *fakePubSubService) Subscribe(context.Context, string) (<-chan otf.Event, error) {
	return f.ch, nil
}

type fakeApp struct {
	ch chan otf.Event
}

func (f *fakeApp) Watch(context.Context, otf.WatchOptions) (<-chan otf.Event, error) {
	return f.ch, nil
}

type fakeRunJSONAPIConverter struct {
	want []byte
}

func (f *fakeRunJSONAPIConverter) MarshalJSONAPI(*otfrun.Run, *http.Request) ([]byte, error) {
	return f.want, nil
}

type fakeEventsServer struct {
	published chan *sse.Event
	eventsServer
}

func (f *fakeEventsServer) CreateStream(string) *sse.Stream              { return nil }
func (f *fakeEventsServer) RemoveStream(string)                          {}
func (f *fakeEventsServer) ServeHTTP(http.ResponseWriter, *http.Request) {}
func (f *fakeEventsServer) Publish(id string, e *sse.Event) {
	f.published <- e
}
