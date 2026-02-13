package workspace

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/require"
)

func NewTestWorkspace(t *testing.T, opts *CreateOptions) *Workspace {
	if opts == nil {
		opts = &CreateOptions{}
	}
	if opts.Organization == nil {
		name := organization.NewTestName(t)
		opts.Organization = &name
	}
	if opts.Name == nil {
		opts.Name = new(uuid.NewString())
	}

	factory := &factory{
		defaultEngine: engine.Default,
		engines: &fakeReleasesService{
			latestVersion: "1.9.0",
		},
	}
	ws, err := factory.NewWorkspace(t.Context(), *opts)
	require.NoError(t, err)
	return ws
}

type FakeService struct {
	Workspaces []*Workspace
	Policy     Policy
}

func (f *FakeService) ListConnectedWorkspaces(ctx context.Context, vcsProviderID resource.TfeID, repoPath vcs.Repo) ([]*Workspace, error) {
	return f.Workspaces, nil
}

func (f *FakeService) Create(context.Context, CreateOptions) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Update(_ context.Context, _ resource.TfeID, opts UpdateOptions) (*Workspace, error) {
	f.Workspaces[0].Update(opts)
	return f.Workspaces[0], nil
}

func (f *FakeService) List(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error) {
	return resource.NewPage(f.Workspaces, opts.PageOptions, nil), nil
}

func (f *FakeService) Get(context.Context, resource.TfeID) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) GetByName(context.Context, organization.Name, string) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Watch(context.Context) (<-chan pubsub.Event[*Event], func()) {
	return nil, nil
}

func (f *FakeService) Delete(context.Context, resource.TfeID) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Lock(context.Context, resource.TfeID, *resource.TfeID) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) Unlock(context.Context, resource.TfeID, *resource.TfeID, bool) (*Workspace, error) {
	return f.Workspaces[0], nil
}

func (f *FakeService) ListTags(context.Context, organization.Name, ListTagsOptions) (*resource.Page[*Tag], error) {
	return nil, nil
}

func (f *FakeService) GetWorkspacePolicy(context.Context, resource.TfeID) (Policy, error) {
	return f.Policy, nil
}

func (f *FakeService) AddTags(ctx context.Context, workspaceID resource.TfeID, tags []TagSpec) error {
	return nil
}

func (f *FakeService) RemoveTags(ctx context.Context, workspaceID resource.TfeID, tags []TagSpec) error {
	return nil
}

func (f *FakeService) SetPermission(ctx context.Context, workspaceID, teamID resource.TfeID, role authz.Role) error {
	return nil
}

func (f *FakeService) UnsetPermission(ctx context.Context, workspaceID, teamID resource.TfeID) error {
	return nil
}

type fakeReleasesService struct {
	latestVersion string
}

func (f *fakeReleasesService) GetLatest(ctx context.Context, engine *engine.Engine) (string, time.Time, error) {
	return f.latestVersion, time.Time{}, nil
}
