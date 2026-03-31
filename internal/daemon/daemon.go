// Package daemon configures and starts the otfd daemon and its subsystems.
package daemon

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/authenticator"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/disco"
	"github.com/leg100/otf/internal/dynamiccreds"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/forgejo"
	"github.com/leg100/otf/internal/github"
	githubui "github.com/leg100/otf/internal/github/ui"
	"github.com/leg100/otf/internal/gitlab"
	"github.com/leg100/otf/internal/healthz"
	"github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/iap"
	"github.com/leg100/otf/internal/loginserver"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/module"
	moduleui "github.com/leg100/otf/internal/module/ui"
	"github.com/leg100/otf/internal/notifications"
	"github.com/leg100/otf/internal/organization"
	orgui "github.com/leg100/otf/internal/organization/ui"
	"github.com/leg100/otf/internal/policy"
	"github.com/leg100/otf/internal/repohooks"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	runui "github.com/leg100/otf/internal/run/ui"
	"github.com/leg100/otf/internal/runner"
	runnerui "github.com/leg100/otf/internal/runner/ui"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sshkey"
	sshkeyui "github.com/leg100/otf/internal/sshkey/ui"
	"github.com/leg100/otf/internal/state"
	stateui "github.com/leg100/otf/internal/state/ui"
	"github.com/leg100/otf/internal/team"
	teamui "github.com/leg100/otf/internal/team/ui"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/ui"
	"github.com/leg100/otf/internal/ui/session"
	"github.com/leg100/otf/internal/user"
	userui "github.com/leg100/otf/internal/user/ui"
	"github.com/leg100/otf/internal/variable"
	variableui "github.com/leg100/otf/internal/variable/ui"
	"github.com/leg100/otf/internal/vcs"
	vcsui "github.com/leg100/otf/internal/vcs/ui"
	"github.com/leg100/otf/internal/workspace"
	workspaceui "github.com/leg100/otf/internal/workspace/ui"
	"golang.org/x/sync/errgroup"
)

type (
	Daemon struct {
		DB            *sql.DB
		Organizations *organization.Service
		Policies      *policy.Service
		Runs          *run.Service
		Workspaces    *workspace.Service
		Variables     *variable.Service
		Notifications *notifications.Service
		State         *state.Service
		Configs       *configversion.Service
		Modules       *module.Service
		VCSProviders  *vcs.Service
		Tokens        *tokens.Service
		Sessions      *session.Service
		Teams         *team.Service
		Users         *user.Service
		GithubApp     *github.Service
		RepoHooks     *repohooks.Service
		Runners       *runner.Service
		Connections   *connections.Service
		System        *internal.HostnameService
		SSHKeys       *sshkey.Service

		netListener net.Listener
		server      *http.Server
		subsystems  []*Subsystem
	}
)

type policyPlanAdapter struct {
	runs *run.Service
}

func (a policyPlanAdapter) GetRunPlanJSON(ctx context.Context, runID resource.TfeID) ([]byte, error) {
	return a.runs.GetRunPlanJSON(ctx, runID)
}

type policyRunAdapter struct {
	runs       *run.Service
	workspaces *workspace.Service
}

func (a policyRunAdapter) GetRunInfo(ctx context.Context, runID resource.TfeID) (*policy.RunInfo, error) {
	r, err := a.runs.GetRun(ctx, runID)
	if err != nil {
		return nil, err
	}
	ws, err := a.workspaces.GetWorkspace(ctx, r.WorkspaceID)
	if err != nil {
		return nil, err
	}
	info := &policy.RunInfo{
		ID:                     r.ID.String(),
		CreatedAt:              r.CreatedAt.Format(time.RFC3339),
		Message:                r.Message,
		IsDestroy:              r.IsDestroy,
		Refresh:                r.Refresh,
		RefreshOnly:            r.RefreshOnly,
		ReplaceAddrs:           r.ReplaceAddrs,
		Speculative:            r.PlanOnly,
		TargetAddrs:            r.TargetAddrs,
		ConfigurationVersionID: r.ConfigurationVersionID,
		Variables:              map[string]policy.RunVariableInfo{},
	}
	if r.CreatedBy != nil {
		info.CreatedBy = r.CreatedBy.String()
	}
	info.Organization.Name = r.Organization.String()
	info.Project.ID = ""
	info.Project.Name = ""
	info.Workspace.ID = ws.ID.String()
	info.Workspace.Name = ws.Name
	info.Workspace.CreatedAt = ws.CreatedAt.Format(time.RFC3339)
	info.Workspace.Description = ws.Description
	info.Workspace.ExecutionMode = string(ws.ExecutionMode)
	info.Workspace.AutoApply = ws.AutoApply
	info.Workspace.WorkingDirectory = ws.WorkingDirectory
	info.Workspace.Tags = append([]string(nil), ws.Tags...)
	info.Workspace.TagBindings = make([]policy.TagBinding, 0, len(ws.Tags))
	for _, tag := range ws.Tags {
		info.Workspace.TagBindings = append(info.Workspace.TagBindings, policy.TagBinding{
			Key:       tag,
			Value:     nil,
			Inherited: false,
		})
	}
	if ws.Connection != nil {
		info.Workspace.VCSRepo = map[string]any{
			"identifier":         ws.Connection.Repo.String(),
			"display_identifier": ws.Connection.Repo.String(),
			"branch":             ws.Connection.Branch,
			"ingress_submodules": false,
		}
	}
	if r.IngressAttributes != nil {
		info.CommitSHA = r.IngressAttributes.CommitSHA
	}
	for _, v := range r.Variables {
		info.Variables[v.Key] = policy.RunVariableInfo{
			Category:  "terraform",
			Sensitive: false,
		}
	}
	return info, nil
}

type policyVariableAdapter struct {
	variables *variable.Service
}

func (a policyVariableAdapter) ListRunVariables(ctx context.Context, runID resource.TfeID) ([]policy.Variable, error) {
	vars, err := a.variables.ListEffectiveVariables(ctx, runID)
	if err != nil {
		return nil, err
	}
	items := make([]policy.Variable, len(vars))
	for i, v := range vars {
		items[i] = policy.Variable{
			Key:       v.Key,
			Category:  string(v.Category),
			Sensitive: v.Sensitive,
			HCL:       v.HCL,
		}
		if !v.Sensitive {
			items[i].Value = v.Value
		}
	}
	return items, nil
}

// New builds a new daemon, establishes a connection to the database and
// migrates it to the latest schema, constructs services and subsystems and
// returns the daemon.
//
// Start() is then expected to be called to start the daemon.
//
// Close() must be called to release resources.
func New(ctx context.Context, logger logr.Logger, cfg Config) (*Daemon, error) {
	if internal.DevMode {
		logger.Info("enabled developer mode")
	}
	if err := cfg.Valid(); err != nil {
		return nil, err
	}
	logger.V(1).Info("set default engine", "engine", cfg.DefaultEngine)

	db, err := sql.New(ctx, logger, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("creating database pool: %w", err)
	}

	// sqlListener listens to database events.
	sqlListener := sql.NewListener(logger, db)

	// netListener opens a TCP port for listening on.
	netListener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, err
	}

	hostnameService := internal.NewHostnameService(
		logger,
		cfg.Host,
		cfg.WebhookHost,
		netListener.Addr().(*net.TCPAddr),
	)

	// responder responds to TFE API requests
	responder := tfeapi.NewResponder(logger)

	// Setup url signer
	signer := tfeapi.NewSigner(cfg.Secret)

	tokensService, err := tokens.NewService(tokens.Options{
		Logger: logger,
		Secret: cfg.Secret,
	})
	if err != nil {
		return nil, fmt.Errorf("setting up authentication middleware: %w", err)
	}

	authorizer := authz.NewAuthorizer(logger)

	orgService := organization.NewService(organization.Options{
		Logger:                       logger,
		Authorizer:                   authorizer,
		DB:                           db,
		RestrictOrganizationCreation: cfg.RestrictOrganizationCreation,
		TokensService:                tokensService,
	})

	teamService := team.NewService(team.Options{
		Logger:              logger,
		Authorizer:          authorizer,
		DB:                  db,
		OrganizationService: orgService,
		TokensService:       tokensService,
	})
	userService := user.NewService(user.Options{
		Logger:        logger,
		Authorizer:    authorizer,
		DB:            db,
		TokensService: tokensService,
		SiteToken:     cfg.SiteToken,
		TeamService:   teamService,
	})
	// promote nominated users to site admin
	if err := userService.SetSiteAdmins(ctx, cfg.SiteAdmins...); err != nil {
		return nil, err
	}

	sessionService := session.NewService(logger, tokensService)

	// Authenticate API/UI requests. An authenticated request is one that
	// has a particular header containing a credential. The token service
	// middleware checks the request against a series of 'authenticators': each
	// authenticator checks if the request contains a particular credential; if
	// there is a successful match then it returns the corresponding subject,
	// e.g. a user or organization API token, etc.
	tokensService.Middleware.Authenticators = []tokens.Authenticator{
		// Authenticate UI session cookies.
		&session.Authenticator{
			Client: tokensService,
		},
		// Authenticate requests from Google IAP.
		iap.NewAuthenticator(
			cfg.GoogleIAPAudience,
			userService,
		),
		// Authenticate API requests from the site admin using their special
		// non-JWT token. It's important that this authenticator comes *before*
		// the JWT authenticator otherwise the JWT authenticator would try to
		// parse the site token as a JWT and return an error.
		&user.SiteAdminAuthenticator{
			SiteToken: cfg.SiteToken,
		},
		// Authenticate API requests with an authorization header containing a
		// JWT token.
		&tokens.JWTAuthenticator{
			Client: tokensService,
		},
	}

	configService := configversion.NewService(configversion.Options{
		Logger:     logger,
		Authorizer: authorizer,
		DB:         db,
	})

	vcsService := vcs.NewService(vcs.Options{
		Logger:              logger,
		Authorizer:          authorizer,
		DB:                  db,
		SourceIconRegistrar: configService,
		SkipTLSVerification: cfg.SkipTLSVerification,
	})

	vcsEventBroker := &vcs.Broker{}

	githubAppService := github.NewService(github.Options{
		Logger:              logger,
		Authorizer:          authorizer,
		DB:                  db,
		VCSService:          vcsService,
		GithubAPIURL:        cfg.GithubHostname,
		SkipTLSVerification: cfg.SkipTLSVerification,
		VCSEventBroker:      vcsEventBroker,
	})

	repoService := repohooks.NewService(ctx, repohooks.Options{
		Logger: logger,
		DB:     db,
		URLs:   hostnameService,
		Client: struct {
			*organization.OrganizationService
			*vcs.VCSService
		}{
			OrganizationService: orgService,
			VCSService:          vcsService,
		},
		VCSEventBroker: vcsEventBroker,
	})

	connectionService := connections.NewService(ctx, connections.Options{
		Logger:             logger,
		DB:                 db,
		VCSProviderService: vcsService,
		RepoHooksService:   repoService,
	})
	engineService := engine.NewService(engine.Options{
		Logger: logger,
		DB:     db,
	})
	if cfg.DisableLatestChecker == nil || !*cfg.DisableLatestChecker {
		engineService.StartLatestChecker(ctx)
	}
	workspaceService := workspace.NewService(workspace.Options{
		Logger:            logger,
		Authorizer:        authorizer,
		DB:                db,
		Listener:          sqlListener,
		ConnectionService: connectionService,
		DefaultEngine:     cfg.DefaultEngine,
		EngineService:     engineService,
	})

	runService := run.NewService(run.Options{
		Logger:     logger,
		Authorizer: authorizer,
		DB:         db,
		Listener:   sqlListener,
		Client: struct {
			*organization.OrganizationService
			*workspace.WorkspaceService
			*configversion.ConfigService
			*vcs.VCSService
			*engine.EngineService
			*tokens.TokensService
		}{
			OrganizationService: orgService,
			WorkspaceService:    workspaceService,
			ConfigService:       configService,
			VCSService:          vcsService,
			EngineService:       engineService,
			TokensService:       tokensService,
		},
		VCSEventSubscriber: vcsEventBroker,
		DaemonCtx:          ctx,
	})
	moduleService := module.NewService(module.Options{
		Logger:             logger,
		Authorizer:         authorizer,
		DB:                 db,
		VCSProviderService: vcsService,
		ConnectionsService: connectionService,
		RepohookService:    repoService,
		VCSEventSubscriber: vcsEventBroker,
	})
	stateService := state.NewService(state.Options{
		Logger:     logger,
		Authorizer: authorizer,
		DB:         db,
	})
	variableService := variable.NewService(variable.Options{
		Logger:           logger,
		Authorizer:       authorizer,
		DB:               db,
		WorkspaceService: workspaceService,
		RunClient:        runService,
	})
	policyService := policy.NewService(policy.Options{
		Logger:          logger,
		Authorizer:      authorizer,
		DB:              db,
		Runs:            policyRunAdapter{runs: runService, workspaces: workspaceService},
		Workspaces:      workspaceService,
		Configs:         configService,
		VCSProviders:    vcsService,
		RepoHooks:       repoService,
		States:          stateService,
		Variables:       policyVariableAdapter{variables: variableService},
		Plans:           policyPlanAdapter{runs: runService},
		VCSEventSub:     vcsEventBroker,
		SentinelPath:    cfg.RunnerConfig.SentinelPath,
		SentinelWorkDir: cfg.RunnerConfig.SentinelWorkDir,
	})
	runService.SetPolicyService(policyService)
	dynamiccredsService, err := dynamiccreds.NewService(dynamiccreds.Options{
		HostnameService: hostnameService,
		PublicKeyPath:   cfg.PublicKeyPath,
		PrivateKeyPath:  cfg.PrivateKeyPath,
		Logger:          logger,
	})
	if err != nil {
		return nil, err
	}
	runnerService := runner.NewService(runner.ServiceOptions{
		Logger:                    logger,
		Authorizer:                authorizer,
		DB:                        db,
		RunService:                runService,
		WorkspaceService:          workspaceService,
		TokensService:             tokensService,
		Listener:                  sqlListener,
		DynamicCredentialsService: dynamiccredsService,
		HostnameService:           hostnameService,
	})
	authenticatorService, err := authenticator.NewAuthenticatorService(ctx, authenticator.Options{
		Logger:               logger,
		URLClient:            hostnameService,
		SessionClient:        sessionService,
		UserService:          userService,
		IDTokenHandlerConfig: cfg.OIDC,
		SkipTLSVerification:  cfg.SkipTLSVerification,
	})
	if err != nil {
		return nil, err
	}

	// Forgejo registrations
	forgejo.RegisterVCSKind(
		vcsService,
		cfg.ForgejoHostname,
		cfg.SkipTLSVerification,
	)

	// Gitlab registrations
	gitlab.RegisterVCSKind(
		vcsService,
		cfg.GitlabHostname,
		cfg.SkipTLSVerification,
	)
	err = gitlab.RegisterOAuthHandler(
		authenticatorService,
		cfg.GitlabHostname,
		cfg.GitlabClientID,
		cfg.GitlabClientSecret,
		cfg.SkipTLSVerification,
	)
	if err != nil {
		return nil, fmt.Errorf("registering gitlab oauth client: %w", err)
	}

	// Github registrations
	err = github.RegisterOAuthHandler(
		authenticatorService,
		cfg.GithubHostname,
		cfg.GithubClientID,
		cfg.GithubClientSecret,
		cfg.SkipTLSVerification,
	)
	if err != nil {
		return nil, fmt.Errorf("registering github oauth client: %w", err)
	}

	sshkeyService := sshkey.NewService(sshkey.Options{
		Logger:     logger,
		Authorizer: authorizer,
		DB:         db,
	})

	serverRunner, err := runner.New(
		logger,
		runnerService,
		func(_ string) runner.OperationClient {
			return struct {
				*organization.OrganizationService
				*workspace.WorkspaceService
				*run.RunService
				*configversion.ConfigService
				*variable.VariableService
				*state.StateService
				*internal.HostnameService
				*sshkey.SSHKeyService
				*runner.RunnerService
				*engine.EngineService
			}{
				OrganizationService: orgService,
				WorkspaceService:    workspaceService,
				RunService:          runService,
				ConfigService:       configService,
				VariableService:     variableService,
				StateService:        stateService,
				HostnameService:     hostnameService,
				SSHKeyService:       sshkeyService,
				RunnerService:       runnerService,
				EngineService:       engineService,
			}
		},
		cfg.RunnerConfig,
	)
	if err != nil {
		return nil, err
	}

	notificationService := notifications.NewService(notifications.Options{
		Logger:     logger,
		Authorizer: authorizer,
		DB:         db,
		Listener:   sqlListener,
	})

	// Handlers for the TFE API
	stateTFEAPI := state.NewTFEAPI(
		struct {
			*state.StateService
			*workspace.WorkspaceService
		}{
			StateService:     stateService,
			WorkspaceService: workspaceService,
		},
		responder,
		signer,
	)
	tfeapiHandlers := &tfeapi.Handlers{
		Verifier: signer,
		Handlers: []internal.Handlers{
			&variable.TFEAPI{
				Client: struct {
					*variable.VariableService
					*workspace.WorkspaceService
				}{
					VariableService:  variableService,
					WorkspaceService: workspaceService,
				},
				Responder: responder,
			},
			&notifications.TFEAPI{
				Client:    notificationService,
				Responder: responder,
			},
			&vcs.TFEAPI{
				Client:    vcsService,
				Responder: responder,
			},
			&runner.TFEAPI{
				Client:    runnerService,
				Responder: responder,
			},
			&team.TFEAPI{
				Client:    teamService,
				Responder: responder,
			},
			&sshkey.TFEAPI{
				Client:    sshkeyService,
				Responder: responder,
			},
			configversion.NewTFEAPI(
				logger,
				configService,
				responder,
				signer,
				cfg.MaxConfigSize,
			),
			organization.NewTFEAPI(
				orgService,
				responder,
			),
			user.NewTFEAPI(
				userService,
				responder,
			),
			user.NewTFEAPI(
				userService,
				responder,
			),
			run.NewTFEAPI(
				struct {
					*run.RunService
					*workspace.WorkspaceService
				}{
					RunService:       runService,
					WorkspaceService: workspaceService,
				},
				authorizer,
				signer,
				responder,
			),
			workspace.NewTFEAPI(
				workspaceService,
				responder,
				authorizer,
			),
			stateTFEAPI,
		},
	}

	apihandlers := &api.Handlers{
		Handlers: []internal.Handlers{
			&organization.API{
				Responder: responder,
				Client:    orgService,
			},
			&workspace.API{
				Responder: responder,
				Client:    workspaceService,
			},
			&run.API{
				Responder: responder,
				Client:    runService,
			},
			&configversion.API{
				Responder: responder,
				Client:    configService,
			},
			&user.API{
				Responder: responder,
				Client:    userService,
			},
			&team.API{
				Responder: responder,
				Client:    teamService,
			},
			&state.API{
				Responder: responder,
				Client:    stateService,
				TFEAPI:    stateTFEAPI,
			},
			&sshkey.API{
				Client: sshkeyService,
			},
			&variable.API{
				Client:    variableService,
				Responder: responder,
			},
			&runner.API{
				Client:    runnerService,
				Responder: responder,
			},
		},
	}

	// Handlers for the UI web app
	runuiHandlers := runui.NewHandlers(
		logger,
		struct {
			*run.RunService
			*user.UserService
			*workspace.WorkspaceService
			*configversion.ConfigService
		}{
			RunService:       runService,
			UserService:      userService,
			WorkspaceService: workspaceService,
			ConfigService:    configService,
		},
		authorizer,
	)
	uiHandlers := &ui.Handlers{
		Authorizer:   authorizer,
		Policies:     policyService,
		VCSProviders: vcsService,
		Workspaces:   workspaceService,
		Handlers: []internal.Handlers{
			&teamui.Handlers{
				Client: struct {
					*user.UserService
					*team.TeamService
				}{
					UserService: userService,
					TeamService: teamService,
				},
				Authorizer: authorizer,
			},
			stateui.NewHandlers(
				struct {
					*state.StateService
					*workspace.WorkspaceService
				}{
					WorkspaceService: workspaceService,
					StateService:     stateService,
				},
				authorizer,
			),
			orgui.NewHandlers(orgService, cfg.RestrictOrganizationCreation),
			vcsui.NewHandlers(vcsService),
			variableui.NewHandlers(
				struct {
					*variable.VariableService
					*workspace.WorkspaceService
				}{
					VariableService:  variableService,
					WorkspaceService: workspaceService,
				},
				authorizer,
			),
			runuiHandlers,
			&runnerui.Handlers{
				Client: struct {
					*runner.RunnerService
					*workspace.WorkspaceService
				}{
					RunnerService:    runnerService,
					WorkspaceService: workspaceService,
				},
				Authorizer: authorizer,
			},
			&sshkeyui.Handlers{
				Client: sshkeyService,
			},
			&userui.Handlers{
				Client: userService,
			},
			&workspaceui.Handlers{
				Client: struct {
					*run.RunService
					*user.UserService
					*workspace.WorkspaceService
					*configversion.ConfigService
					*sshkey.SSHKeyService
					*engine.EngineService
					*vcs.VCSService
					*team.TeamService
				}{
					RunService:       runService,
					UserService:      userService,
					WorkspaceService: workspaceService,
					ConfigService:    configService,
					SSHKeyService:    sshkeyService,
					EngineService:    engineService,
					VCSService:       vcsService,
					TeamService:      teamService,
				},
				Authorizer:     authorizer,
				SingleRunTable: runuiHandlers.SingleRunTable,
			},
			githubui.NewHandlers(
				githubAppService,
				hostnameService,
				cfg.GithubHostname,
				cfg.SkipTLSVerification,
				authorizer,
			),
			moduleui.NewHandlers(
				struct {
					*module.ModuleService
					*vcs.VCSService
					*internal.HostnameService
				}{
					ModuleService:   moduleService,
					VCSService:      vcsService,
					HostnameService: hostnameService,
				},
				authorizer,
			),
		},
	}

	// Compile list of all handlers
	handlers := []internal.Handlers{
		&healthz.Check{Client: db},
		apihandlers,
		tfeapiHandlers,
		uiHandlers,
		&userui.LoginHandlers{
			Client: struct {
				*authenticator.AuthenticatorService
				*session.Service
			}{
				AuthenticatorService: authenticatorService,
				Service:              sessionService,
			},
			SiteToken: cfg.SiteToken,
		},
		authenticatorService,
		loginserver.NewServer(loginserver.Options{
			Secret:      cfg.Secret,
			UserService: userService,
		}),
		disco.Service{},
		&github.AppEventHandler{
			Logger:     logger,
			Publisher:  vcsEventBroker,
			GithubApps: githubAppService,
			VCSService: vcsService,
		},
		repoService,
		dynamiccredsService,
		&module.Registry{
			Client: moduleService,
			Signer: signer,
		},
	}

	// Construct subsystems; ordered by start order
	subsystems := []*Subsystem{
		// The listener is started first because it is responsible for listening
		// for database events, and other subsystems rely on it to be listening
		// before they start, such as the runner (which generates a new runner
		// event), and the allocator (which receives the new runner event).
		{
			Name:   "listener",
			Logger: logger,
			System: sqlListener,
		},
		{
			Name:   "reporter",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.ReporterLockID),
			System: run.NewReporter(
				logger,
				vcsService,
				workspaceService,
				runService,
				configService,
				hostnameService,
			),
		},
		{
			Name:   "run_metrics",
			Logger: logger,
			System: runService.MetricsCollector,
		},
		{
			Name:   "timeout",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.TimeoutLockID),
			System: &run.Timeout{
				Logger:                logger.WithValues("component", "timeout"),
				OverrideCheckInterval: cfg.OverrideTimeoutCheckInterval,
				PlanningTimeout:       cfg.PlanningTimeout,
				ApplyingTimeout:       cfg.ApplyingTimeout,
				Runs:                  runService,
			},
		},
		{
			Name:   "run-deleter",
			Logger: logger,
			System: &resource.Deleter[*run.Run]{
				Logger:                logger.WithValues("component", "run-deleter"),
				OverrideCheckInterval: cfg.OverrideDeleterInterval,
				Client:                &runDeleterAdapter{svc: runService},
				AgeThreshold:          cfg.DeleteRunsAfter,
			},
		},
		{
			Name:   "config-deleter",
			Logger: logger,
			System: &resource.Deleter[*configversion.ConfigurationVersion]{
				Logger:                logger.WithValues("component", "config-deleter"),
				OverrideCheckInterval: cfg.OverrideDeleterInterval,
				Client:                &configVersionDeleterAdapter{svc: configService},
				AgeThreshold:          cfg.DeleteConfigsAfter,
			},
		},
		{
			Name:   "notifier",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.NotifierLockID),
			System: notifications.NewNotifier(notifications.NotifierOptions{
				Logger:             logger,
				HostnamesClient:    hostnameService,
				WorkspaceClient:    workspaceService,
				RunClient:          runService,
				NotificationClient: notificationService,
				DB:                 db,
			}),
		},
		{
			Name:   "job-allocator",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.AllocatorLockID),
			System: runnerService.NewAllocator(logger),
		},
		{
			Name:   "runner-manager",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.RunnerManagerLockID),
			System: runnerService.NewManager(),
		},
		{
			Name:   "job-signaler",
			Logger: logger,
			DB:     db,
			System: runnerService.Signaler,
		},
	}
	if !cfg.DisableRunner {
		subsystems = append(subsystems, &Subsystem{
			Name:   "runner-daemon",
			Logger: logger,
			DB:     db,
			System: serverRunner,
		})
	}
	if !cfg.DisableScheduler {
		subsystems = append(subsystems, &Subsystem{
			Name:   "scheduler",
			Logger: logger,
			DB:     db,
			LockID: internal.Ptr(sql.SchedulerLockID),
			System: run.NewScheduler(run.SchedulerOptions{
				Logger:          logger,
				WorkspaceClient: workspaceService,
				RunClient:       runService,
			}),
		})
	}

	// Construct web server and start listening on port
	server, err := http.NewServer(logger, http.ServerConfig{
		SSL:                  cfg.SSL,
		CertFile:             cfg.CertFile,
		KeyFile:              cfg.KeyFile,
		EnableRequestLogging: cfg.EnableRequestLogging,
		Middleware:           []mux.MiddlewareFunc{tokensService.Middleware.Authenticate},
		Handlers:             handlers,
	})
	if err != nil {
		return nil, fmt.Errorf("setting up http server: %w", err)
	}

	return &Daemon{
		Organizations: orgService,
		Policies:      policyService,
		System:        hostnameService,
		Runs:          runService,
		Workspaces:    workspaceService,
		Variables:     variableService,
		Notifications: notificationService,
		State:         stateService,
		Configs:       configService,
		Modules:       moduleService,
		VCSProviders:  vcsService,
		Tokens:        tokensService,
		Sessions:      sessionService,
		Teams:         teamService,
		Users:         userService,
		RepoHooks:     repoService,
		GithubApp:     githubAppService,
		Connections:   connectionService,
		Runners:       runnerService,
		SSHKeys:       sshkeyService,
		DB:            db,
		netListener:   netListener,
		server:        server,
		subsystems:    subsystems,
	}, nil
}

// Start the otfd daemon and block until ctx is cancelled or an error is
// returned. The started channel is closed once the daemon has started.
func (d *Daemon) Start(ctx context.Context, started chan struct{}) error {
	// Cancel context the first time a func started with g.Go() fails
	g, ctx := errgroup.WithContext(ctx)

	for _, ss := range d.subsystems {
		if err := ss.Start(ctx, g); err != nil {
			return err
		}
		// Wait for subsystem to finish starting up if it exposes the ability to
		// do so.
		wait, ok := ss.System.(interface{ Started() <-chan struct{} })
		if ok {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second * 10):
				return fmt.Errorf("timed out waiting for subsystem to start: %s", ss.Name)
			case <-wait.Started():
			}
		}
	}

	// Run HTTP/JSON-API server and web app
	g.Go(func() error {
		if err := d.server.Start(ctx, d.netListener); err != nil {
			return fmt.Errorf("http server terminated: %w", err)
		}
		return nil
	})

	// Inform the caller the daemon has started
	close(started)

	// Block until error or Ctrl-C received.
	return g.Wait()
}

// Close releases daemon resources.
func (d *Daemon) Close() {
	// close active connections to http server
	d.netListener.Close()
	// close all db connections upon exit
	d.DB.Close()
}

// configVersionDeleterAdapter adapts configversion.Service to the
// resource.deleterClient interface.
type configVersionDeleterAdapter struct {
	svc *configversion.Service
}

func (a *configVersionDeleterAdapter) ListOlderThan(ctx context.Context, age time.Time) ([]*configversion.ConfigurationVersion, error) {
	return a.svc.ListConfigVersionsOlderThan(ctx, age)
}

func (a *configVersionDeleterAdapter) Delete(ctx context.Context, id resource.TfeID) error {
	return a.svc.DeleteConfigVersion(ctx, id)
}

// runDeleterAdapter adapts run.Service to the resource.deleterClient interface.
type runDeleterAdapter struct {
	svc *run.Service
}

func (a *runDeleterAdapter) ListOlderThan(ctx context.Context, age time.Time) ([]*run.Run, error) {
	return a.svc.ListRunsOlderThan(ctx, age)
}

func (a *runDeleterAdapter) Delete(ctx context.Context, id resource.TfeID) error {
	return a.svc.DeleteRun(ctx, id)
}
