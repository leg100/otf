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

type fakeSubscriber struct {
	ch chan pubsub.Event

	pubsub.PubSubService
}

func (f *fakeSubscriber) Subscribe(context.Context, string) (<-chan pubsub.Event, error) {
	return f.ch, nil
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

func (f *fakeWebServices) CreateRun(ctx context.Context, workspaceID string, opts RunCreateOptions) (*Run, error) {
	return f.runs[0], nil
}

func (f *fakeWebServices) ListRuns(ctx context.Context, opts RunListOptions) (*resource.Page[*Run], error) {
	return resource.NewPage(f.runs, opts.PageOptions, nil), nil
}

func (f *fakeWebServices) GetLogs(context.Context, string, internal.PhaseType) ([]byte, error) {
	return nil, nil
}

func (f *fakeWebServices) Cancel(ctx context.Context, runID string) (*Run, error) { return nil, nil }

func (f *fakeWebServices) GetRun(ctx context.Context, runID string) (*Run, error) {
	return f.runs[0], nil
}

func (f *fakeWebServices) startRun(context.Context, string, Operation) (*Run, error) {
	return f.runs[0], nil
}
