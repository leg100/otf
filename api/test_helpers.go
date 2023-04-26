package api

import (
	"context"
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/run"
)

type fakeRunService struct {
	ch chan otf.Event

	run.RunService
}

func (f *fakeRunService) Watch(context.Context, run.WatchOptions) (<-chan otf.Event, error) {
	return f.ch, nil
}

type fakeMarshaler struct {
	run *jsonapi.Run

	marshaler
}

func (f *fakeMarshaler) toRun(*run.Run, *http.Request) (*jsonapi.Run, error) {
	return f.run, nil
}
