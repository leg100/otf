package api

import (
	"context"
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
)

type (
	fakeRunService struct {
		ch chan pubsub.Event

		run.RunService
	}

	fakeConfigService struct {
		configversion.ConfigurationVersionService
	}
)

func (f *fakeRunService) Watch(context.Context, run.WatchOptions) (<-chan pubsub.Event, error) {
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
