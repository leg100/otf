package html

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/stretchr/testify/require"
)

type fakeAppOption func(*Application)

func withFakeSiteToken(token string) fakeAppOption {
	return func(app *Application) {
		app.siteToken = token
	}
}

func withAuthenticators(authenticators []*Authenticator) fakeAppOption {
	return func(app *Application) {
		app.authenticators = authenticators
	}
}

func newFakeWebApp(t *testing.T, app otf.Application, opts ...fakeAppOption) *Application {
	views, err := NewViewEngine(false)
	require.NoError(t, err)

	a := &Application{
		Application: app,
		Logger:      logr.Discard(),
		viewEngine:  views,
	}

	for _, o := range opts {
		o(a)
	}

	return a
}

type fakeApp struct {
	fakeUser         *otf.User
	fakeAgentToken   *otf.AgentToken
	fakeOrganization *otf.Organization
	fakeWorkspace    *otf.Workspace
	fakeRun          *otf.Run

	otf.Application
}

func (u *fakeApp) GetUser(context.Context, otf.UserSpec) (*otf.User, error) {
	return u.fakeUser, nil
}

func (u *fakeApp) GetSessionByToken(context.Context, string) (*otf.Session, error) {
	return otf.NewSession("user-fake", "127.0.0.1")
}

func (u *fakeApp) ListSessions(context.Context, string) ([]*otf.Session, error) {
	return nil, nil
}

func (u *fakeApp) CreateToken(ctx context.Context, userID string, opts *otf.TokenCreateOptions) (*otf.Token, error) {
	return otf.NewToken(userID, opts.Description)
}

func (u *fakeApp) ListTokens(ctx context.Context, userID string) ([]*otf.Token, error) {
	return nil, nil
}

func (u *fakeApp) DeleteToken(context.Context, string, string) error { return nil }

func (u *fakeApp) ListTeams(ctx context.Context, organizationName string) ([]*otf.Team, error) {
	return nil, nil
}

func (u *fakeApp) GetOrganization(ctx context.Context, name string) (*otf.Organization, error) {
	return u.fakeOrganization, nil
}

func (u *fakeApp) GetWorkspace(context.Context, string) (*otf.Workspace, error) {
	return u.fakeWorkspace, nil
}

func (u *fakeApp) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      []*otf.Workspace{u.fakeWorkspace},
		Pagination: otf.NewPagination(opts.ListOptions, 1),
	}, nil
}

func (u *fakeApp) CreateWorkspace(ctx context.Context, opts otf.CreateWorkspaceOptions) (*otf.Workspace, error) {
	return u.fakeWorkspace, nil
}

func (u *fakeApp) ListWorkspacePermissions(context.Context, string) ([]*otf.WorkspacePermission, error) {
	return nil, nil
}

func (u *fakeApp) SetWorkspacePermission(context.Context, string, string, rbac.Role) error {
	return nil
}

func (u *fakeApp) UnsetWorkspacePermission(context.Context, string, string) error {
	return nil
}

func (u *fakeApp) GetRun(context.Context, string) (*otf.Run, error) {
	return u.fakeRun, nil
}

func (u *fakeApp) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{
		Items:      []*otf.Run{u.fakeRun},
		Pagination: otf.NewPagination(opts.ListOptions, 1),
	}, nil
}

func (u *fakeApp) GetChunk(context.Context, otf.GetChunkOptions) (otf.Chunk, error) {
	return otf.Chunk{Data: []byte("fake-logs")}, nil
}

func (u *fakeApp) CreateAgentToken(ctx context.Context, opts otf.CreateAgentTokenOptions) (*otf.AgentToken, error) {
	return otf.NewAgentToken(opts)
}

func (u *fakeApp) GetAgentToken(ctx context.Context, id string) (*otf.AgentToken, error) {
	return u.fakeAgentToken, nil
}

func (u *fakeApp) ListAgentTokens(ctx context.Context, _ string) ([]*otf.AgentToken, error) {
	return []*otf.AgentToken{u.fakeAgentToken}, nil
}

func (u *fakeApp) DeleteAgentToken(context.Context, string) (*otf.AgentToken, error) {
	return u.fakeAgentToken, nil
}
