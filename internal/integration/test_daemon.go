package integration

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/agent"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/cli"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/orgcreator"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

type (
	// daemon for integration tests
	testDaemon struct {
		*daemon.Daemon

		*github.TestServer
	}

	// configures the daemon for integration tests
	config struct {
		daemon.Config
	}
)

// setup configures and starts an otfd daemon for an integration test
func setup(t *testing.T, cfg *config, gopts ...github.TestServerOption) *testDaemon {
	t.Helper()

	if cfg == nil {
		cfg = &config{}
	}

	// Setup database if not specified
	if cfg.Database == "" {
		cfg.Database = sql.NewTestDB(t)
	}
	// Setup secret if not specified
	if cfg.Secret == nil {
		cfg.Secret = testutils.NewSecret(t)
	}
	daemon.ApplyDefaults(&cfg.Config)
	cfg.SSL = true
	cfg.CertFile = "./fixtures/cert.pem"
	cfg.KeyFile = "./fixtures/key.pem"

	// Start stub github server
	githubServer, githubCfg := github.NewTestServer(t, gopts...)
	cfg.Github.Config = githubCfg

	// Configure logger; discard logs by default
	var logger logr.Logger
	if _, ok := os.LookupEnv("OTF_INTEGRATION_TEST_ENABLE_LOGGER"); ok {
		var err error
		logger, err = logr.New(&logr.Config{Verbosity: 1, Format: "default"})
		require.NoError(t, err)
	} else {
		logger = logr.Discard()
	}

	// Confer superuser privileges on all calls to service endpoints
	ctx := internal.AddSubjectToContext(context.Background(), &internal.Superuser{Username: "app-user"})

	d, err := daemon.New(ctx, logger, cfg.Config)
	require.NoError(t, err)

	// start daemon and upon test completion check that it exited cleanly
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan error)
	started := make(chan struct{})
	go func() {
		err := d.Start(ctx, started)
		// if context was canceled don't report any error
		if ctx.Err() != nil {
			done <- nil
			return
		}
		require.NoError(t, err, "daemon exited with an error")
		done <- err
	}()
	// don't proceed until daemon has started.
	<-started

	t.Cleanup(func() {
		cancel() // terminates daemon
		<-done   // don't exit test until daemon is fully terminated
	})

	return &testDaemon{
		Daemon:     d,
		TestServer: githubServer,
	}
}

func (s *testDaemon) createOrganization(t *testing.T, ctx context.Context) *organization.Organization {
	t.Helper()

	org, err := s.CreateOrganization(ctx, orgcreator.OrganizationCreateOptions{
		Name: internal.String(internal.GenerateRandomString(4) + "-corp"),
	})
	require.NoError(t, err)
	return org
}

func (s *testDaemon) createWorkspace(t *testing.T, ctx context.Context, org *organization.Organization) *workspace.Workspace {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	ws, err := s.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:         internal.String("workspace-" + internal.GenerateRandomString(6)),
		Organization: &org.Name,
	})
	require.NoError(t, err)
	return ws
}

func (s *testDaemon) getWorkspace(t *testing.T, ctx context.Context, workspaceID string) *workspace.Workspace {
	t.Helper()

	ws, err := s.GetWorkspace(ctx, workspaceID)
	require.NoError(t, err)
	return ws
}

func (s *testDaemon) createVCSProvider(t *testing.T, ctx context.Context, org *organization.Organization) *vcsprovider.VCSProvider {
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

func (s *testDaemon) createModule(t *testing.T, ctx context.Context, org *organization.Organization) *module.Module {
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

func (s *testDaemon) createUser(t *testing.T, ctx context.Context, opts ...auth.NewUserOption) *auth.User {
	t.Helper()

	user, err := s.CreateUser(ctx, "user-"+internal.GenerateRandomString(4), opts...)
	require.NoError(t, err)
	return user
}

func (s *testDaemon) createUserCtx(t *testing.T, ctx context.Context, opts ...auth.NewUserOption) (*auth.User, context.Context) {
	t.Helper()

	user := s.createUser(t, ctx, opts...)
	return user, internal.AddSubjectToContext(ctx, user)
}

func (s *testDaemon) getUser(t *testing.T, ctx context.Context, username string) *auth.User {
	t.Helper()

	user, err := s.GetUser(ctx, auth.UserSpec{Username: &username})
	require.NoError(t, err)
	return user
}

func (s *testDaemon) getUserCtx(t *testing.T, ctx context.Context, username string) (*auth.User, context.Context) {
	t.Helper()

	user, err := s.GetUser(ctx, auth.UserSpec{Username: &username})
	require.NoError(t, err)
	return user, internal.AddSubjectToContext(ctx, user)
}

func (s *testDaemon) createTeam(t *testing.T, ctx context.Context, org *organization.Organization) *auth.Team {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	team, err := s.CreateTeam(ctx, auth.CreateTeamOptions{
		Name:         "team-" + internal.GenerateRandomString(4),
		Organization: org.Name,
	})
	require.NoError(t, err)
	return team
}

func (s *testDaemon) getTeam(t *testing.T, ctx context.Context, org, name string) *auth.Team {
	t.Helper()

	team, err := s.GetTeam(ctx, org, name)
	require.NoError(t, err)
	return team
}

func (s *testDaemon) createConfigurationVersion(t *testing.T, ctx context.Context, ws *workspace.Workspace, opts *configversion.ConfigurationVersionCreateOptions) *configversion.ConfigurationVersion {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}
	if opts == nil {
		opts = &configversion.ConfigurationVersionCreateOptions{}
	}

	cv, err := s.CreateConfigurationVersion(ctx, ws.ID, *opts)
	require.NoError(t, err)
	return cv
}

func (s *testDaemon) createAndUploadConfigurationVersion(t *testing.T, ctx context.Context, ws *workspace.Workspace, opts *configversion.ConfigurationVersionCreateOptions) *configversion.ConfigurationVersion {
	cv := s.createConfigurationVersion(t, ctx, ws, opts)
	tarball, err := os.ReadFile("./testdata/root.tar.gz")
	require.NoError(t, err)
	s.UploadConfig(ctx, cv.ID, tarball)
	return cv
}

func (s *testDaemon) createRun(t *testing.T, ctx context.Context, ws *workspace.Workspace, cv *configversion.ConfigurationVersion) *run.Run {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}
	if cv == nil {
		cv = s.createConfigurationVersion(t, ctx, ws, nil)
	}

	run, err := s.CreateRun(ctx, ws.ID, run.RunCreateOptions{
		ConfigurationVersionID: internal.String(cv.ID),
	})
	require.NoError(t, err)
	return run
}

func (s *testDaemon) createVariable(t *testing.T, ctx context.Context, ws *workspace.Workspace) *variable.Variable {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}

	v, err := s.CreateVariable(ctx, ws.ID, variable.CreateVariableOptions{
		Key:      internal.String("key-" + internal.GenerateRandomString(4)),
		Value:    internal.String("val-" + internal.GenerateRandomString(4)),
		Category: variable.VariableCategoryPtr(variable.CategoryTerraform),
	})
	require.NoError(t, err)
	return v
}

func (s *testDaemon) createStateVersion(t *testing.T, ctx context.Context, ws *workspace.Workspace) *state.Version {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}

	file, err := os.ReadFile("./testdata/terraform.tfstate")
	require.NoError(t, err)

	sv, err := s.CreateStateVersion(ctx, state.CreateStateVersionOptions{
		State:       file,
		WorkspaceID: internal.String(ws.ID),
	})
	require.NoError(t, err)
	return sv
}

func (s *testDaemon) getCurrentState(t *testing.T, ctx context.Context, wsID string) *state.Version {
	t.Helper()

	sv, err := s.GetCurrentStateVersion(ctx, wsID)
	require.NoError(t, err)
	return sv
}

func (s *testDaemon) createToken(t *testing.T, ctx context.Context, user *auth.User) (*tokens.UserToken, []byte) {
	t.Helper()

	// If user is provided then add it to context. Otherwise the context is
	// expected to contain a user if authz is to succeed.
	if user != nil {
		ctx = internal.AddSubjectToContext(ctx, user)
	}

	ut, token, err := s.CreateUserToken(ctx, tokens.CreateUserTokenOptions{
		Description: "lorem ipsum...",
	})
	require.NoError(t, err)
	return ut, token
}

func (s *testDaemon) createNotificationConfig(t *testing.T, ctx context.Context, ws *workspace.Workspace) *notifications.Config {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}

	nc, err := s.CreateNotificationConfiguration(ctx, ws.ID, notifications.CreateConfigOptions{
		DestinationType: notifications.DestinationGeneric,
		Enabled:         internal.Bool(true),
		Name:            internal.String(uuid.NewString()),
		URL:             internal.String("http://example.com"),
	})
	require.NoError(t, err)
	return nc
}

func (s *testDaemon) createAgentToken(t *testing.T, ctx context.Context, organization string) []byte {
	t.Helper()

	token, err := s.CreateAgentToken(ctx, tokens.CreateAgentTokenOptions{
		Organization: organization,
		Description:  "lorem ipsum...",
	})
	require.NoError(t, err)
	return token
}

func (s *testDaemon) createSubscriber(t *testing.T, ctx context.Context) <-chan pubsub.Event {
	t.Helper()

	sub, err := s.Subscribe(ctx, "")
	require.NoError(t, err)
	return sub
}

// startAgent starts an external agent, configuring it with the given
// organization and configuring it to connect to the daemon.
func (s *testDaemon) startAgent(t *testing.T, ctx context.Context, organization string, cfg agent.ExternalConfig) {
	t.Helper()

	// Configure logger; discard logs by default
	var logger logr.Logger
	if _, ok := os.LookupEnv("OTF_INTEGRATION_TEST_ENABLE_LOGGER"); ok {
		var err error
		logger, err = logr.New(&logr.Config{Verbosity: 1, Format: "default"})
		require.NoError(t, err)
	} else {
		logger = logr.Discard()
	}

	token := s.createAgentToken(t, ctx, organization)
	cfg.HTTPConfig = http.NewConfig()
	cfg.HTTPConfig.Token = string(token)
	cfg.HTTPConfig.Address = s.Hostname()
	cfg.HTTPConfig.Insecure = true // daemon uses self-signed cert

	agent, err := agent.NewExternalAgent(ctx, logger, cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		err := agent.Start(ctx)
		close(done)
		require.NoError(t, err)
	}()

	t.Cleanup(func() {
		cancel() // terminate agent
		<-done   // don't exit test until agent fully terminated
	})
}

func (s *testDaemon) tfcli(t *testing.T, ctx context.Context, command, configPath string, args ...string) string {
	t.Helper()

	out, err := s.tfcliWithError(t, ctx, command, configPath, args...)
	require.NoError(t, err, "tf cli failed: %s", out)
	return out
}

func (s *testDaemon) tfcliWithError(t *testing.T, ctx context.Context, command, configPath string, args ...string) (string, error) {
	t.Helper()

	// Create user token expressly for the terraform cli
	user, err := auth.UserFromContext(ctx)
	require.NoError(t, err)
	_, token := s.createToken(t, ctx, user)

	cmdargs := []string{command, "-no-color"}
	cmdargs = append(cmdargs, args...)

	cmd := exec.Command("terraform", cmdargs...)
	cmd.Dir = configPath
	cmd.Env = append(envs,
		internal.CredentialEnv(s.Hostname(), token),
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (s *testDaemon) otfcli(t *testing.T, ctx context.Context, args ...string) string {
	t.Helper()

	// Create user token expressly for the otf cli
	user, err := auth.UserFromContext(ctx)
	require.NoError(t, err)
	_, token := s.createToken(t, ctx, user)

	cmdargs := []string{"--address", s.Hostname(), "--token", string(token)}
	cmdargs = append(cmdargs, args...)

	var buf bytes.Buffer
	err = (&cli.CLI{}).Run(ctx, cmdargs, &buf)
	require.NoError(t, err)

	require.NoError(t, err, "otf cli failed: %s", buf.String())
	return buf.String()
}
