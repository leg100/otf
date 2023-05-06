package api

import (
	"context"
	"net/http"

	"github.com/DataDog/jsonapi"
	internal "github.com/leg100/otf"
	"github.com/leg100/otf/api/types"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/run"
)

type (
	fakeRunService struct {
		ch chan internal.Event

		run.RunService
	}

	fakeConfigService struct {
		configversion.ConfigurationVersionService
	}
)

func (f *fakeRunService) Watch(context.Context, run.WatchOptions) (<-chan internal.Event, error) {
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
