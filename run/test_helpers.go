package run

import (
	"context"
	"net/http"

	"github.com/leg100/otf"
)

type fakeSubscriber struct {
	ch chan otf.Event

	otf.PubSubService
}

func (f *fakeSubscriber) Subscribe(context.Context, string) (<-chan otf.Event, error) {
	return f.ch, nil
}

type fakeService struct {
	ch chan otf.Event

	Service
}

func (f *fakeService) Watch(context.Context, WatchOptions) (<-chan otf.Event, error) {
	return f.ch, nil
}

type fakeJSONAPIMarshaler struct {
	marshaled []byte
	jsonapiMarshaler
}

func (f *fakeJSONAPIMarshaler) MarshalJSONAPI(*Run, *http.Request) ([]byte, error) {
	return f.marshaled, nil
}
