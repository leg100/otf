package run

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
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

	fakeWebLogsService struct{}
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

func newTestWebHandlers(_ *testing.T, opts ...fakeWebServiceOption) *webHandlers {
	var svc fakeWebServices
	for _, fn := range opts {
		fn(&svc)
	}

	return &webHandlers{
		authorizer: authz.NewAllowAllAuthorizer(),
		workspaces: &workspace.FakeService{
			Workspaces: []*workspace.Workspace{svc.ws},
		},
		runs: &svc,
		logs: &fakeWebLogsService{},
	}
}

func (f *fakeWebServices) Create(ctx context.Context, workspaceID resource.TfeID, opts CreateOptions) (*Run, error) {
	return f.runs[0], nil
}

func (f *fakeWebServices) GetPolicy(context.Context, string) (authz.WorkspacePolicy, error) {
	return authz.WorkspacePolicy{}, nil
}

func (f *fakeWebServices) List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error) {
	return resource.NewPage(f.runs, opts.PageOptions, nil), nil
}

func (f *fakeWebServices) Cancel(context.Context, resource.TfeID) error { return nil }

func (f *fakeWebServices) Get(ctx context.Context, runID resource.TfeID) (*Run, error) {
	return f.runs[0], nil
}

func (f *fakeWebServices) Apply(ctx context.Context, runID resource.TfeID) error {
	return nil
}

func (f *fakeWebLogsService) GetAllLogs(context.Context, resource.TfeID, internal.PhaseType) ([]byte, error) {
	return nil, nil
}
