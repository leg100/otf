package run

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

type fakeSubService struct {
	ch chan pubsub.Event[*Run]
}

func (f *fakeSubService) Subscribe(context.Context) (<-chan pubsub.Event[*Run], func()) {
	return f.ch, nil
}

type (
	fakeWebServices struct {
		runs []*Run
		ws   *workspace.Workspace

		// fakeWebServices does not implement all of webRunClient
		webRunClient
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
		Renderer: renderer,
		workspaces: &workspace.FakeService{
			Workspaces: []*workspace.Workspace{svc.ws},
		},
		runs: &svc,
	}
}

func (f *fakeWebServices) Create(ctx context.Context, workspaceID string, opts CreateOptions) (*Run, error) {
	return f.runs[0], nil
}

func (f *fakeWebServices) GetPolicy(context.Context, string) (internal.WorkspacePolicy, error) {
	return internal.WorkspacePolicy{}, nil
}

func (f *fakeWebServices) List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error) {
	return resource.NewPage(f.runs, opts.PageOptions, nil), nil
}

func (f *fakeWebServices) getLogs(context.Context, string, internal.PhaseType) ([]byte, error) {
	return nil, nil
}

func (f *fakeWebServices) Cancel(context.Context, string) error { return nil }

func (f *fakeWebServices) Get(ctx context.Context, runID string) (*Run, error) {
	return f.runs[0], nil
}

func (f *fakeWebServices) Apply(ctx context.Context, runID string) error {
	return nil
}
