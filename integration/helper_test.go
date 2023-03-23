package integration

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/services"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/require"
)

type testServices struct {
	*services.Services

	githubServer *github.TestServer
}

func setup(t *testing.T, repo string) *testServices {
	t.Helper()

	db := sql.NewTestDB(t)
	cfg := services.NewDefaultConfig()

	// Use stub github server
	githubServer, githubCfg := github.NewTestServer(t, github.WithRepo(repo))
	cfg.Github.Config = githubCfg

	svcs, _, err := services.New(logr.Discard(), db, cfg)
	require.NoError(t, err)
	return &testServices{
		Services:     svcs,
		githubServer: githubServer,
	}
}

func (s *testServices) createOrganization(t *testing.T, ctx context.Context) *organization.Organization {
	t.Helper()

	org, err := s.CreateOrganization(ctx, organization.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

func (s *testServices) createWorkspace(t *testing.T, ctx context.Context, org *organization.Organization) *workspace.Workspace {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	ws, err := s.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:         otf.String(uuid.NewString()),
		Organization: &org.Name,
	})
	require.NoError(t, err)
	return ws
}

func (s *testServices) createVCSProvider(t *testing.T, ctx context.Context, org *organization.Organization) *vcsprovider.VCSProvider {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	provider, err := s.CreateVCSProvider(ctx, vcsprovider.CreateOptions{
		Organization: org.Name,
		// tests require a legitimate cloud name to avoid invalid foreign
		// key error upon insert/update
		Cloud: "github",
		Name:  uuid.NewString(),
		Token: uuid.NewString(),
	})
	require.NoError(t, err)
	return provider
}

func (s *testServices) createModule(t *testing.T, ctx context.Context, org *organization.Organization) *module.Module {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	module, err := s.CreateModule(ctx, module.CreateOptions{
		Name:         uuid.NewString(),
		Provider:     uuid.NewString(),
		Organization: org.Name,
	})
	require.NoError(t, err)
	return module
}

func (s *testServices) createUser(t *testing.T, ctx context.Context, opts ...auth.NewUserOption) *auth.User {
	t.Helper()

	user, err := s.CreateUser(ctx, uuid.NewString(), opts...)
	require.NoError(t, err)
	return user
}

func (s *testServices) createTeam(t *testing.T, ctx context.Context, org *organization.Organization) *auth.Team {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	team, err := s.CreateTeam(ctx, auth.NewTeamOptions{
		Name:         uuid.NewString(),
		Organization: org.Name,
	})
	require.NoError(t, err)
	return team
}

func (s *testServices) createConfigurationVersion(t *testing.T, ctx context.Context, ws *workspace.Workspace) *configversion.ConfigurationVersion {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}

	cv, err := s.CreateConfigurationVersion(ctx, ws.ID, configversion.ConfigurationVersionCreateOptions{})
	require.NoError(t, err)
	return cv
}

func (s *testServices) createRun(t *testing.T, ctx context.Context, ws *workspace.Workspace, cv *configversion.ConfigurationVersion) *run.Run {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}
	if cv == nil {
		cv = s.createConfigurationVersion(t, ctx, ws)
	}

	run, err := s.CreateRun(ctx, ws.ID, run.RunCreateOptions{
		ConfigurationVersionID: otf.String(cv.ID),
	})
	require.NoError(t, err)
	return run
}

func (s *testServices) createVariable(t *testing.T, ctx context.Context, ws *workspace.Workspace) *variable.Variable {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}

	v, err := s.CreateVariable(ctx, ws.ID, variable.CreateVariableOptions{
		Key:      otf.String(uuid.NewString()),
		Value:    otf.String(uuid.NewString()),
		Category: variable.VariableCategoryPtr(variable.CategoryTerraform),
	})
	require.NoError(t, err)
	return v
}

func (s *testServices) createStateVersion(t *testing.T, ctx context.Context, ws *workspace.Workspace) *state.Version {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}

	file, err := os.ReadFile("./testdata/terraform.tfstate")
	require.NoError(t, err)

	sv, err := s.CreateStateVersion(ctx, state.CreateStateVersionOptions{
		State:       file,
		WorkspaceID: otf.String(ws.ID),
	})
	require.NoError(t, err)
	return sv
}

func (s *testServices) createRegistrySession(t *testing.T, ctx context.Context, org *organization.Organization, expiry *time.Time) *auth.RegistrySession {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	rs, err := s.CreateRegistrySession(ctx, auth.CreateRegistrySessionOptions{
		Organization: &org.Name,
		Expiry:       expiry,
	})
	require.NoError(t, err)
	return rs
}

func (s *testServices) createSession(t *testing.T, ctx context.Context, user *auth.User, expiry *time.Time) *auth.Session {
	t.Helper()

	if user == nil {
		user = s.createUser(t, ctx)
	}

	rs, err := s.CreateSession(ctx, auth.CreateSessionOptions{
		Request: httptest.NewRequest("", "/", nil),
		UserID:  &user.ID,
		Expiry:  expiry,
	})
	require.NoError(t, err)
	return rs
}

func (s *testServices) createToken(t *testing.T, ctx context.Context, user *auth.User) *auth.Token {
	t.Helper()

	// If user is provided then add it to context. Otherwise the context is
	// expected to contain a user if authz is to succeed.
	if user != nil {
		ctx = otf.AddSubjectToContext(ctx, user)
	}

	token, err := s.CreateToken(ctx, &auth.TokenCreateOptions{
		Description: "lorem ipsum...",
	})
	require.NoError(t, err)
	return token
}
