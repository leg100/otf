package integration

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/cli"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/module"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/runner/agent"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/testutils"
	otfuser "github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

type (
	// daemon for integration test
	testDaemon struct {
		*daemon.Daemon
		// stub github server for test to use.
		*github.TestServer
		// run subscription for tests to check on run events
		runEvents <-chan pubsub.Event[*run.Event]
	}
)

// setup an integration test with a daemon, organization, and a user context.
func setup(t *testing.T, opts ...configOption) (*testDaemon, *organization.Organization, context.Context) {
	t.Helper()

	cfg := &config{
		Config: daemon.NewConfig(),
	}

	// Skip TLS verification for tests because they'll be standing up various
	// stub TLS servers with self-certified certs.
	cfg.SkipTLSVerification = true

	// Enable SSL by default, because tests running terraform binary mandate it
	cfg.SSL = true
	cfg.CertFile = "./fixtures/cert.pem"
	cfg.KeyFile = "./fixtures/key.pem"

	for _, fn := range opts {
		fn(cfg)
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
	// engine version
	if cfg.DisableLatestChecker == nil || !*cfg.DisableLatestChecker {
		cfg.DisableLatestChecker = new(true)
	}
	// Start stub github server, unless test has set its own github stub
	var githubServer *github.TestServer
	if !cfg.skipGithubStub {
		var githubURL *url.URL
		githubServer, githubURL = github.NewTestServer(t, cfg.githubOptions...)
		cfg.GithubHostname = &internal.WebURL{URL: *githubURL}
	}

	// Configure logger; discard logs by default
	var logger logr.Logger
	if _, ok := os.LookupEnv("OTF_INTEGRATION_TEST_ENABLE_LOGGER"); ok {
		var err error
		logger, err = logr.New(logr.Config{Verbosity: 9, Format: "default"})
		require.NoError(t, err)
		cfg.EnableRequestLogging = true
	} else {
		logger = logr.Discard()
	}

	// Confer superuser privileges on all calls to service endpoints
	ctx := authz.AddSubjectToContext(context.Background(), &authz.Superuser{Username: "app-user"})

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

	// Subscribe to run events
	runEvents, unsub := d.Runs.Watch(ctx)
	t.Cleanup(unsub)

	t.Cleanup(func() {
		cancel() // terminates daemon
		<-done   // don't exit test until daemon is fully terminated
	})

	daemon := &testDaemon{
		Daemon:     d,
		TestServer: githubServer,
		runEvents:  runEvents,
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
		testUserCtx = authz.AddSubjectToContext(ctx, testUser)
	}

	return daemon, org, testUserCtx
}

func (s *testDaemon) createOrganization(t *testing.T, ctx context.Context) *organization.Organization {
	t.Helper()

	org, err := s.Organizations.Create(ctx, organization.CreateOptions{
		Name: new(internal.GenerateRandomString(4) + "-corp"),
	})
	require.NoError(t, err)
	return org
}

func (s *testDaemon) createWorkspace(t *testing.T, ctx context.Context, org *organization.Organization) *workspace.Workspace {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	ws, err := s.Workspaces.Create(ctx, workspace.CreateOptions{
		Name:         new("workspace-" + internal.GenerateRandomString(6)),
		Organization: &org.Name,
	})
	require.NoError(t, err)
	return ws
}

func (s *testDaemon) getWorkspace(t *testing.T, ctx context.Context, workspaceID resource.TfeID) *workspace.Workspace {
	t.Helper()

	ws, err := s.Workspaces.Get(ctx, workspaceID)
	require.NoError(t, err)
	return ws
}

func (s *testDaemon) getRun(t *testing.T, ctx context.Context, runID resource.TfeID) *run.Run {
	t.Helper()

	run, err := s.Runs.Get(ctx, runID)
	require.NoError(t, err)
	return run
}

func (s *testDaemon) waitRunStatus(t *testing.T, ctx context.Context, runID resource.TfeID, status runstatus.Status) *run.Run {
	t.Helper()

	for event := range s.runEvents {
		if event.Payload.ID == runID {
			if event.Payload.Status == status {
				return s.getRun(t, ctx, runID)
			}
			if runstatus.Done(event.Payload.Status) && event.Payload.Status != status {
				t.Fatalf("expected run status %s but run finished with status %s", status, event.Payload.Status)
			}
		}
	}
	return nil
}

func (s *testDaemon) createVCSProvider(t *testing.T, ctx context.Context, org *organization.Organization, createOptions *vcs.CreateOptions) *vcs.Provider {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	opts := vcs.CreateOptions{
		Organization: org.Name,
		KindID:       github.TokenKindID,
		Token:        new(uuid.NewString()),
	}
	if createOptions != nil {
		opts.Name = createOptions.Name
	}

	provider, err := s.VCSProviders.Create(ctx, opts)
	require.NoError(t, err)
	return provider
}

func (s *testDaemon) createModule(t *testing.T, ctx context.Context, org *organization.Organization) *module.Module {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	module, err := s.Modules.CreateModule(ctx, module.CreateOptions{
		Name:         uuid.NewString(),
		Provider:     uuid.NewString(),
		Organization: org.Name,
	})
	require.NoError(t, err)
	return module
}

// createUser is always invoked with the site admin context because only they
// are authorized to create users.
func (s *testDaemon) createUser(t *testing.T, opts ...otfuser.NewUserOption) *otfuser.User {
	t.Helper()

	user, err := s.Users.Create(adminCtx, "user-"+internal.GenerateRandomString(4), opts...)
	require.NoError(t, err)
	return user
}

// createUserCtx is always invoked with the site admin context because only they
// are authorized to create users.
func (s *testDaemon) createUserCtx(t *testing.T, opts ...otfuser.NewUserOption) (*otfuser.User, context.Context) {
	t.Helper()

	user := s.createUser(t, opts...)
	return user, authz.AddSubjectToContext(context.Background(), user)
}

func (s *testDaemon) getUser(t *testing.T, ctx context.Context, username otfuser.Username) *otfuser.User {
	t.Helper()

	user, err := s.Users.GetUser(ctx, otfuser.UserSpec{Username: &username})
	require.NoError(t, err)
	return user
}

func (s *testDaemon) getUserCtx(t *testing.T, ctx context.Context, username otfuser.Username) (*otfuser.User, context.Context) {
	t.Helper()

	user, err := s.Users.GetUser(ctx, otfuser.UserSpec{Username: &username})
	require.NoError(t, err)
	return user, authz.AddSubjectToContext(ctx, user)
}

func (s *testDaemon) createTeam(t *testing.T, ctx context.Context, org *organization.Organization) *team.Team {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	team, err := s.Teams.Create(ctx, org.Name, team.CreateTeamOptions{
		Name: new("team-" + internal.GenerateRandomString(4)),
	})
	require.NoError(t, err)
	return team
}

func (s *testDaemon) getTeam(t *testing.T, ctx context.Context, org organization.Name, name string) *team.Team {
	t.Helper()

	team, err := s.Teams.Get(ctx, org, name)
	require.NoError(t, err)
	return team
}

func (s *testDaemon) createConfigurationVersion(t *testing.T, ctx context.Context, ws *workspace.Workspace, opts *configversion.CreateOptions) *configversion.ConfigurationVersion {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}
	if opts == nil {
		opts = &configversion.CreateOptions{}
	}

	cv, err := s.Configs.Create(ctx, ws.ID, *opts)
	require.NoError(t, err)
	return cv
}

func (s *testDaemon) createAndUploadConfigurationVersion(t *testing.T, ctx context.Context, ws *workspace.Workspace, opts *configversion.CreateOptions) *configversion.ConfigurationVersion {
	cv := s.createConfigurationVersion(t, ctx, ws, opts)
	tarball := testutils.ReadFile(t, "./testdata/root.tar.gz")
	err := s.Configs.UploadConfig(ctx, cv.ID, tarball)
	require.NoError(t, err)
	return cv
}

func (s *testDaemon) createRun(t *testing.T, ctx context.Context, ws *workspace.Workspace, cv *configversion.ConfigurationVersion, opts *run.CreateOptions) *run.Run {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}
	if cv == nil {
		cv = s.createConfigurationVersion(t, ctx, ws, nil)
	}
	if opts == nil {
		opts = &run.CreateOptions{}
	}
	opts.ConfigurationVersionID = &cv.ID

	run, err := s.Runs.Create(ctx, ws.ID, *opts)
	require.NoError(t, err)
	return run
}

func (s *testDaemon) createVariable(t *testing.T, ctx context.Context, ws *workspace.Workspace, opts *variable.CreateVariableOptions) *variable.Variable {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}

	if opts == nil {
		opts = &variable.CreateVariableOptions{
			Key:      new("key-" + internal.GenerateRandomString(4)),
			Value:    new("val-" + internal.GenerateRandomString(4)),
			Category: internal.Ptr(variable.CategoryTerraform),
		}
	}
	v, err := s.Variables.CreateWorkspaceVariable(ctx, ws.ID, *opts)
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

	sv, err := s.State.Create(ctx, state.CreateStateVersionOptions{
		State:       file,
		WorkspaceID: ws.ID,
		// serial matches that in ./testdata/terraform.tfstate
		Serial: new(int64(9)),
	})
	require.NoError(t, err)
	return sv
}

func (s *testDaemon) getCurrentState(t *testing.T, ctx context.Context, wsID resource.TfeID) *state.Version {
	t.Helper()

	sv, err := s.State.GetCurrent(ctx, wsID)
	require.NoError(t, err)
	return sv
}

func (s *testDaemon) createToken(t *testing.T, ctx context.Context, user *otfuser.User) (*otfuser.UserToken, []byte) {
	t.Helper()

	// If user is provided then add them to context. Otherwise the context is
	// expected to contain a user if authz is to succeed.
	if user != nil {
		ctx = authz.AddSubjectToContext(ctx, user)
	}

	ut, token, err := s.Users.CreateToken(ctx, otfuser.CreateUserTokenOptions{
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

	nc, err := s.Notifications.Create(ctx, ws.ID, notifications.CreateConfigOptions{
		DestinationType: notifications.DestinationGeneric,
		Enabled:         new(true),
		Name:            new(uuid.NewString()),
		URL:             new("http://example.com"),
	})
	require.NoError(t, err)
	return nc
}

// startAgent starts a pool agent, configuring it with the given organization
// and configuring it to connect to the daemon. The corresponding agent type is
// returned once registered, along with a function to shutdown the agent down.
func (s *testDaemon) startAgent(t *testing.T, ctx context.Context, org organization.Name, poolID *resource.TfeID, token string, opts ...runnerConfigOption) (*runner.RunnerMeta, func()) {
	t.Helper()

	// Configure logger; discard logs by default
	var logger logr.Logger
	if _, ok := os.LookupEnv("OTF_INTEGRATION_TEST_ENABLE_LOGGER"); ok {
		var err error
		logger, err = logr.New(logr.Config{Verbosity: 1, Format: "default"})
		require.NoError(t, err)
	} else {
		logger = logr.Discard()
	}

	if token == "" {
		if poolID == nil {
			pool, err := s.Runners.CreateAgentPool(ctx, runner.CreateAgentPoolOptions{
				Name:         uuid.NewString(),
				Organization: org,
			})
			require.NoError(t, err)
			poolID = &pool.ID
		}
		_, tokenBytes, err := s.Runners.CreateAgentToken(ctx, *poolID, runner.CreateAgentTokenOptions{
			Description: "lorem ipsum...",
		})
		require.NoError(t, err)
		token = string(tokenBytes)
	}

	cfg := runner.NewDefaultConfig()
	for _, fn := range opts {
		fn(cfg)
	}

	// Set a routeable URL for the agent to locate the server. We can't reliably use the
	// server hostname because in some tests that can be set to something
	// unrouteable, e.g. the dynamic provider credential test sets it to
	// something arbitrary.
	routeableURL := url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("localhost:%d", s.ListenAddress.Port),
	}

	runner, err := agent.New(logger, routeableURL.String(), token, cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		err := runner.Start(ctx)
		close(done)
		require.NoError(t, err)
	}()

	t.Cleanup(func() {
		cancel() // terminate agent
		<-done   // don't exit test until agent fully terminated
	})
	// Wait for agent to register itself
	<-runner.Started()
	return runner.RunnerMeta, cancel
}

func (s *testDaemon) engineCLI(t *testing.T, ctx context.Context, engine string, command, configPath string, args ...string) string {
	t.Helper()

	out, err := s.engineCLIWithError(t, ctx, engine, command, configPath, args...)
	require.NoError(t, err, "engine cli failed: %s", out)
	return out
}

func (s *testDaemon) engineCLIWithError(t *testing.T, ctx context.Context, engine string, command, configPath string, args ...string) (string, error) {
	t.Helper()

	if engine == "" {
		engine = terraformPath
	}

	// Create user token expressly for the terraform cli
	user := userFromContext(t, ctx)
	_, token := s.createToken(t, ctx, user)

	cmdargs := []string{command}
	cmdargs = append(cmdargs, args...)

	cmd := exec.Command(engine, cmdargs...)
	cmd.Dir = configPath

	cmd.Env = internal.SafeAppend(sharedEnvs, internal.CredentialEnv(s.System.Hostname(), token))

	// Run tf and collect stdout/stderr
	out, err := cmd.CombinedOutput()

	// strip ANSI escape codes because tests are not expecting them.
	return internal.StripAnsi(string(out)), err
}

func (s *testDaemon) otfCLI(t *testing.T, ctx context.Context, args ...string) string {
	t.Helper()

	// Create user token expressly for the otf cli
	user := userFromContext(t, ctx)
	_, token := s.createToken(t, ctx, user)

	cmdargs := []string{"--url", s.System.URL("/"), "--token", string(token)}
	cmdargs = append(cmdargs, args...)

	var buf bytes.Buffer
	err := cli.NewCLI().Run(ctx, cmdargs, &buf)
	require.NoError(t, err)

	require.NoError(t, err, "otf cli failed: %s", buf.String())
	return buf.String()
}

// getLocalURL retrieves a response from the URL of the daemon under test.
//
// NOTE: it takes care to use the local listening address rather than the
// hostname that might have been assigned to the daemon, which might skew the
// test.
func (s *testDaemon) getLocalURL(t *testing.T, path string) *http.Response {
	localURL := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("localhost:%d", s.ListenAddress.Port),
		Path:   path,
	}
	resp, err := http.Get(localURL.String())
	require.NoError(t, err)
	return resp
}
