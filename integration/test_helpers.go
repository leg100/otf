package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/chromedp/chromedp"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/client"
	"github.com/leg100/otf/cmd"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/daemon"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/orgcreator"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/tokens"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/require"
)

type (
	// daemon for integration tests
	testDaemon struct {
		*daemon.Daemon

		vcsServer
	}

	// configures the daemon for integration tests
	config struct {
		daemon.Config

		repo string // create repo on stub github server
	}

	// some tests want to know whether a webhook has been created on the vcs
	// server
	vcsServer interface {
		HasWebhook() bool
	}
)

// setup configures and starts an otfd daemon for an integration test
func setup(t *testing.T, cfg *config) *testDaemon {
	t.Helper()

	// Options for the stub github server.
	var ghopts []github.TestServerOption

	if cfg == nil {
		cfg = &config{}
	}

	// Setup database if not specified
	if cfg.Database == "" {
		cfg.Database = sql.NewTestDB(t)
	}
	// Setup secret if not specified
	if cfg.Secret == "" {
		cfg.Secret = otf.GenerateRandomString(16)
	}
	// Add repo to stub github server if specified
	if cfg.repo != "" {
		ghopts = append(ghopts, github.WithRepo(cfg.repo))
	}
	daemon.ApplyDefaults(&cfg.Config)
	cfg.SSL = true
	cfg.CertFile = "./fixtures/cert.pem"
	cfg.KeyFile = "./fixtures/key.pem"

	// Start stub github server
	githubServer, githubCfg := github.NewTestServer(t, ghopts...)
	cfg.Github.Config = githubCfg

	// Configure logger; discard logs by default
	var logger logr.Logger
	if _, ok := os.LookupEnv("OTF_INTEGRATION_TEST_ENABLE_LOGGER"); ok {
		var err error
		logger, err = cmd.NewLogger(&cmd.LoggerConfig{Level: "debug", Color: "true"})
		require.NoError(t, err)
	} else {
		logger = logr.Discard()
	}

	// Confer superuser privileges on all calls to service endpoints
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{Username: "app-user"})

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
		Daemon:    d,
		vcsServer: githubServer,
	}
}

func (s *testDaemon) createOrganization(t *testing.T, ctx context.Context) *organization.Organization {
	t.Helper()

	org, err := s.CreateOrganization(ctx, orgcreator.OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
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
		Name:         otf.String(uuid.NewString()),
		Organization: &org.Name,
	})
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

	user, err := s.CreateUser(ctx, uuid.NewString(), opts...)
	require.NoError(t, err)
	return user
}

func (s *testDaemon) createUserCtx(t *testing.T, ctx context.Context, opts ...auth.NewUserOption) (*auth.User, context.Context) {
	t.Helper()

	user, err := s.CreateUser(ctx, uuid.NewString(), opts...)
	require.NoError(t, err)
	return user, otf.AddSubjectToContext(ctx, user)
}

func (s *testDaemon) createTeam(t *testing.T, ctx context.Context, org *organization.Organization) *auth.Team {
	t.Helper()

	if org == nil {
		org = s.createOrganization(t, ctx)
	}

	team, err := s.CreateTeam(ctx, auth.CreateTeamOptions{
		Name:         uuid.NewString(),
		Organization: org.Name,
	})
	require.NoError(t, err)
	return team
}

func (s *testDaemon) createConfigurationVersion(t *testing.T, ctx context.Context, ws *workspace.Workspace) *configversion.ConfigurationVersion {
	t.Helper()

	if ws == nil {
		ws = s.createWorkspace(t, ctx, nil)
	}

	cv, err := s.CreateConfigurationVersion(ctx, ws.ID, configversion.ConfigurationVersionCreateOptions{})
	require.NoError(t, err)
	return cv
}

func (s *testDaemon) createRun(t *testing.T, ctx context.Context, ws *workspace.Workspace, cv *configversion.ConfigurationVersion) *run.Run {
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

func (s *testDaemon) createVariable(t *testing.T, ctx context.Context, ws *workspace.Workspace) *variable.Variable {
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

func (s *testDaemon) createStateVersion(t *testing.T, ctx context.Context, ws *workspace.Workspace) *state.Version {
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

func (s *testDaemon) createToken(t *testing.T, ctx context.Context, user *auth.User) (*tokens.Token, []byte) {
	t.Helper()

	// If user is provided then add it to context. Otherwise the context is
	// expected to contain a user if authz is to succeed.
	if user != nil {
		ctx = otf.AddSubjectToContext(ctx, user)
	}

	ut, token, err := s.CreateToken(ctx, tokens.CreateTokenOptions{
		Description: "lorem ipsum...",
	})
	require.NoError(t, err)
	return ut, token
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
func (s *testDaemon) startAgent(t *testing.T, ctx context.Context, organization string, cfg agent.Config) {
	t.Helper()

	token := s.createAgentToken(t, ctx, organization)

	clientCfg := http.NewConfig()
	clientCfg.Address = s.Hostname()
	clientCfg.Insecure = true // daemon uses self-signed cert
	clientCfg.Token = string(token)
	app, err := client.New(*clientCfg)
	require.NoError(t, err)

	cfg.External = true
	cfg.Organization = &organization
	//
	// Configure logger; discard logs by default
	var logger logr.Logger
	if _, ok := os.LookupEnv("OTF_INTEGRATION_TEST_ENABLE_LOGGER"); ok {
		var err error
		logger, err = cmd.NewLogger(&cmd.LoggerConfig{Level: "debug", Color: "true"})
		require.NoError(t, err)
	} else {
		logger = logr.Discard()
	}

	agent, err := agent.NewAgent(logger, app, cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan error)
	go func() {
		done <- agent.Start(ctx)
	}()

	t.Cleanup(func() {
		cancel() // terminate agent
		<-done   // don't exit test until agent fully terminated
	})
}

func (s *testDaemon) tfcli(t *testing.T, ctx context.Context, command, configPath string, args ...string) string {
	t.Helper()

	user, err := auth.UserFromContext(ctx)
	require.NoError(t, err)
	_, token := s.createToken(t, ctx, user)

	cmdargs := []string{command, "-no-color"}
	cmdargs = append(cmdargs, args...)

	cmd := exec.Command("terraform", cmdargs...)
	cmd.Dir = configPath
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"SSL_CERT_FILE=" + os.Getenv("SSL_CERT_FILE"),
		otf.CredentialEnv(s.Hostname(), token),
	}
	if proxy, ok := os.LookupEnv("HTTPS_PROXY"); ok {
		cmd.Env = append(cmd.Env, "HTTPS_PROXY="+proxy)
	}
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "terraform %s failed: %s", command, out)
	return string(out)
}

func (s *testDaemon) otfcli(t *testing.T, ctx context.Context, args ...string) string {
	t.Helper()

	user, err := auth.UserFromContext(ctx)
	require.NoError(t, err)
	_, token := s.createToken(t, ctx, user)

	cmdargs := []string{"--address", s.Hostname()}
	cmdargs = append(cmdargs, args...)

	cmd := exec.Command("../_build/otf", cmdargs...)
	cmd.Env = []string{
		"HOME=" + os.Getenv("HOME"),
		"PATH=" + os.Getenv("PATH"),
		"SSL_CERT_FILE=" + os.Getenv("SSL_CERT_FILE"),
		otf.CredentialEnv(s.Hostname(), token),
		"OTF_ADDRESS=" + s.Hostname(),
	}
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "otf cli failed: %s", out)
	return string(out)
}

type cmdOption func(*exec.Cmd)

func createBrowserCtx(t *testing.T) context.Context {
	t.Helper()

	headless := false
	if v, ok := os.LookupEnv("OTF_E2E_HEADLESS"); ok {
		var err error
		headless, err = strconv.ParseBool(v)
		if err != nil {
			panic("cannot parse OTF_E2E_HEADLESS")
		}
	}

	// must create an allocator before creating the browser
	allocator, cancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", headless),
			chromedp.Flag("hide-scrollbars", true),
			chromedp.Flag("mute-audio", true),
			chromedp.Flag("ignore-certificate-errors", true),
			chromedp.Flag("disable-gpu", true),
		)...)
	t.Cleanup(cancel)

	// now create the browser
	ctx, cancel := chromedp.NewContext(allocator)
	t.Cleanup(cancel)

	// Ensure ~/.terraform.d exists - 'terraform login' has a bug whereby it tries to
	// persist the API token it receives to a temporary file in ~/.terraform.d but
	// fails if ~/.terraform.d doesn't exist yet. This only happens when
	// CHECKPOINT_DISABLE is set, because the checkpoint would otherwise handle
	// creating that directory first...
	os.MkdirAll(path.Join(os.Getenv("HOME"), ".terraform.d"), 0o755)

	return ctx
}

func workspacePath(hostname, org, name string) string {
	return "https://" + hostname + "/app/organizations/" + org + "/workspaces/" + name
}

func organizationPath(hostname, org string) string {
	return "https://" + hostname + "/app/organizations/" + org
}

// newRootModule creates a terraform root module, returning its directory path
func newRootModule(t *testing.T, hostname, organization, workspace string) string {
	t.Helper()

	config := []byte(fmt.Sprintf(`
terraform {
  backend "remote" {
	hostname = "%s"
	organization = "%s"

	workspaces {
	  name = "%s"
	}
  }
}
resource "null_resource" "e2e" {}
`, hostname, organization, workspace))

	root := t.TempDir()
	err := os.WriteFile(filepath.Join(root, "main.tf"), config, 0o600)
	require.NoError(t, err)

	return root
}
