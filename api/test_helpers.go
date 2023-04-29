package api

import (
	"context"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/api/types"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/run"
)

type (
	fakeRunService struct {
		ch chan otf.Event

		run.RunService
	}

	fakeConfigService struct {
		configversion.ConfigurationVersionService
	}
)

func (f *fakeRunService) Watch(context.Context, run.WatchOptions) (<-chan otf.Event, error) {
	return f.ch, nil
}

type fakeMarshaler struct {
	run *types.Run

	marshaler
}

func (f *fakeMarshaler) toRun(*run.Run, *http.Request) (*types.Run, []jsonapi.MarshalOption, error) {
	return f.run, nil, nil
}

func (f *fakeConfigService) UploadConfig(context.Context, string, []byte) error {
	return nil
}
