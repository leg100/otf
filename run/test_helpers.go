package run

import (
	"context"
	"net/http"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/require"
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

type (
	fakeWebServices struct {
		runs []*Run
		ws   *workspace.Workspace

		RunService
		WorkspaceService
	}

	fakeWebServiceOption func(*fakeWebServices)
)

func withWorkspace(workspace *workspace.Workspace) fakeWebServiceOption {
	return func(svc *fakeWebServices) {
		svc.ws = workspace
	}
}

func withRuns(runs ...*Run) fakeWebServiceOption {
	return func(svc *fakeWebServices) {
		svc.runs = runs
	}
}

func newTestWebHandlers(t *testing.T, opts ...fakeWebServiceOption) *webHandlers {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)

	var svc fakeWebServices
	for _, fn := range opts {
		fn(&svc)
	}

	return &webHandlers{
		Renderer:         renderer,
		WorkspaceService: &svc,
		logsdb:           &svc,
		starter:          &svc,
		svc:              &svc,
	}
}

func (f *fakeWebServices) GetWorkspaceByName(context.Context, string, string) (*workspace.Workspace, error) {
	return f.ws, nil
}

func (f *fakeWebServices) GetWorkspace(context.Context, string) (*workspace.Workspace, error) {
	return f.ws, nil
}

func (f *fakeWebServices) ListRuns(ctx context.Context, opts RunListOptions) (*RunList, error) {
	return &RunList{
		Items:      f.runs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.runs)),
	}, nil
}

func (f *fakeWebServices) GetLogs(context.Context, string, otf.PhaseType) ([]byte, error) {
	return nil, nil
}

func (f *fakeWebServices) Cancel(ctx context.Context, runID string) (*Run, error) { return nil, nil }

func (f *fakeWebServices) get(ctx context.Context, runID string) (*Run, error) {
	return f.runs[0], nil
}

func (f *fakeWebServices) startRun(ctx context.Context, workspaceID string, strategy runStrategy) (*Run, error) {
	return f.runs[0], nil
}
