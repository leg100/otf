package integration

import (
	"bytes"
	"context"
	"io"
	"net/url"
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
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/releases"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

type (
	// daemon for integration test
	testDaemon struct {
		*daemon.Daemon
		// stub github server for test to use.
		*github.TestServer
		// event subscription for test to use.
		sub <-chan pubsub.Event
		// releases service to allow tests to download terraform
		releases.ReleasesService
	}

	// configures the daemon for integration tests
	config struct {
		daemon.Config
		// skip creation of default organization
		skipDefaultOrganization bool
		// customise path in which terraform bins are saved
		terraformBinDir string
	}
)

// setup an integration test with a daemon, organization, and a user context.
func setup(t *testing.T, cfg *config, gopts ...github.TestServerOption) (*testDaemon, *organization.Organization, context.Context) {
	t.Helper()

	if cfg == nil {
		cfg = &config{}
	}
	// Setup database if not specified
	if cfg.Database == "" {
		cfg.Database = sql.NewTestDB(t)
	}
	// Use shared secret if one is not specified
	if cfg.Secret == nil {
		cfg.Secret = sharedSecret
	}
	// Unless test has specified otherwise, disable checking for latest
	// terraform version
	if cfg.DisableLatestChecker == nil || !*cfg.DisableLatestChecker {
		cfg.DisableLatestChecker = internal.Bool(true)
	}
	// Skip TLS verification for tests because they'll be standing up various
	// stub TLS servers with self-certified certs.
	cfg.SkipTLSVerification = true

	daemon.ApplyDefaults(&cfg.Config)
	cfg.SSL = true
	cfg.CertFile = "./fixtures/cert.pem"
	cfg.KeyFile = "./fixtures/key.pem"

	// Start stub github server, unless test has set its own github stub
	var githubServer *github.TestServer
	if cfg.GithubHostname == "" {
		var githubURL *url.URL
		githubServer, githubURL = github.NewTestServer(t, gopts...)
		cfg.GithubHostname = githubURL.Host
	}

	// Configure logger; discard logs by default
	var logger logr.Logger
	if _, ok := os.LookupEnv("OTF_INTEGRATION_TEST_ENABLE_LOGGER"); ok {
		var err error
		logger, err = logr.New(&logr.Config{Verbosity: 9, Format: "default"})
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

	sub, err := d.Broker.Subscribe(ctx, "")
	require.NoError(t, err)

	releasesService := releases.NewService(releases.Options{
		Logger:          logger,
		DB:              d.DB,
		TerraformBinDir: cfg.terraformBinDir,
	})

	daemon := &testDaemon{
		Daemon:          d,
		TestServer:      githubServer,
		ReleasesService: releasesService,
		sub:             sub,
	}

	// create a dedicated user account and context for test to use.
	testUser, testUserCtx := daemon.createUserCtx(t)

	var org *organization.Organization
	if !cfg.skipDefaultOrganization {
		// create organization for test to use. Consume the created event too so
		// that tests that consume events don't receive this event.
		org = daemon.createOrganization(t, testUserCtx)
		// re-fetch user so that its ownership of the above org is included
		testUser = daemon.getUser(t, adminCtx, testUser.Username)
		// and re-add to context
		testUserCtx = internal.AddSubjectToContext(ctx, testUser)
	}

	return daemon, org, testUserCtx
}

func (s *testDaemon) createOrganization(t *testing.T, ctx context.Context) *organization.Organization {
	t.Helper()

	org, err := s.CreateOrganization(ctx, organization.CreateOptions{
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
		Kind:  vcs.KindPtr(vcs.GithubKind),
		Token: internal.String(uuid.NewString()),
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

// createUser is always invoked with the site admin context because only they
// are authorized to create users.
func (s *testDaemon) createUser(t *testing.T, opts ...auth.NewUserOption) *auth.User {
	t.Helper()

	user, err := s.CreateUser(adminCtx, "user-"+internal.GenerateRandomString(4), opts...)
	require.NoError(t, err)
	return user
}

// createUserCtx is always invoked with the site admin context because only they
// are authorized to create users.
func (s *testDaemon) createUserCtx(t *testing.T, opts ...auth.NewUserOption) (*auth.User, context.Context) {
	t.Helper()

	user := s.createUser(t, opts...)
	return user, internal.AddSubjectToContext(context.Background(), user)
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

	team, err := s.CreateTeam(ctx, org.Name, auth.CreateTeamOptions{
		Name: internal.String("team-" + internal.GenerateRandomString(4)),
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
	err = s.UploadConfig(ctx, cv.ID, tarball)
	require.NoError(t, err)
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

	run, err := s.CreateRun(ctx, ws.ID, run.CreateOptions{
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

	v, err := s.CreateWorkspaceVariable(ctx, ws.ID, variable.CreateVariableOptions{
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

	// If user is provided then add them to context. Otherwise the context is
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
	cfg.APIConfig.Token = string(token)
	cfg.APIConfig.Address = s.Hostname()

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

	tfpath := s.downloadTerraform(t, ctx, nil)

	// Create user token expressly for the terraform cli
	user := userFromContext(t, ctx)
	_, token := s.createToken(t, ctx, user)

	cmdargs := []string{command}
	cmdargs = append(cmdargs, args...)

	cmd := exec.Command(tfpath, cmdargs...)
	cmd.Dir = configPath

	cmd.Env = internal.SafeAppend(sharedEnvs, internal.CredentialEnv(s.Hostname(), token))

	// Run tf and collect stdout/stderr
	out, err := cmd.CombinedOutput()

	// strip ANSI escape codes because tests are not expecting them.
	return internal.StripAnsi(string(out)), err
}

func (s *testDaemon) otfcli(t *testing.T, ctx context.Context, args ...string) string {
	t.Helper()

	// Create user token expressly for the otf cli
	user := userFromContext(t, ctx)
	_, token := s.createToken(t, ctx, user)

	cmdargs := []string{"--address", s.Hostname(), "--token", string(token)}
	cmdargs = append(cmdargs, args...)

	var buf bytes.Buffer
	err := cli.NewCLI().Run(ctx, cmdargs, &buf)
	require.NoError(t, err)

	require.NoError(t, err, "otf cli failed: %s", buf.String())
	return buf.String()
}

func (s *testDaemon) downloadTerraform(t *testing.T, ctx context.Context, version *string) string {
	t.Helper()

	if version == nil {
		version = internal.String(releases.DefaultTerraformVersion)
	}
	tfpath, err := s.Download(ctx, *version, io.Discard)
	require.NoError(t, err)
	return tfpath
}
