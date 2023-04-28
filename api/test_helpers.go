package api

import (
	"context"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf"
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
	run *Run

	marshaler
}

func (f *fakeMarshaler) toRun(*run.Run, *http.Request) (*Run, []jsonapi.MarshalOption, error) {
	return f.run, nil, nil
}
