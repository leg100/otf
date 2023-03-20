package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/require"
)

func CreateTestRun(t *testing.T, db otf.DB, ws *workspace.Workspace, cv *configversion.ConfigurationVersion, opts RunCreateOptions) *Run {
	ctx := context.Background()
	run := NewRun(cv, ws, opts)
	rundb := &pgdb{db}

	err := rundb.CreateRun(ctx, run)
	require.NoError(t, err)

	t.Cleanup(func() {
		rundb.DeleteRun(ctx, run.ID)
	})
	return run
}

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
