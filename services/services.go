// Package services builds and configures a dependency tree of otf services. For
// use by the otfd binary and for integration testing.
package services

import (
	"context"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/agent"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/client"
	"github.com/leg100/otf/cloud"
	cmdutil "github.com/leg100/otf/cmd"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/github"
	"github.com/leg100/otf/gitlab"
	"github.com/leg100/otf/http"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/logs"
	"github.com/leg100/otf/module"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/pubsub"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/scheduler"
	"github.com/leg100/otf/state"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/vcsprovider"
	"github.com/leg100/otf/workspace"
)

type (
	Services struct {
		Config
		Handlers []otf.Handlers

		organization.OrganizationService
		auth.AuthService
		variable.VariableService
		vcsprovider.VCSProviderService
		state.StateService
		workspace.WorkspaceService
		module.ModuleService
		otf.HostnameService
		configversion.ConfigurationVersionService
		run.RunService
		repo.RepoService
		logs.LogsService
		pubsub.Broker
	}

	Config struct {
		AgentConfig          *agent.Config
		LoggerConfig         *cmdutil.LoggerConfig
		CacheConfig          *inmem.CacheConfig
		Github               cloud.CloudOAuthConfig
		Gitlab               cloud.CloudOAuthConfig
		Secret               string // secret for signing URLs
		SiteToken            string
		Host                 string
		Address, Database    string
		MaxConfigSize        int64
		SSL                  bool
		CertFile, KeyFile    string
		EnableRequestLogging bool
		DevMode              bool
	}

	daemon interface {
		Start(context.Context) error
	}
)

func NewDefaultConfig() Config {
	return Config{
		AgentConfig: &agent.Config{
			Concurrency: agent.DefaultConcurrency,
		},
		CacheConfig: &inmem.CacheConfig{},
		Github: cloud.CloudOAuthConfig{
			Config:      github.Defaults(),
			OAuthConfig: github.OAuthDefaults(),
		},
		Gitlab: cloud.CloudOAuthConfig{
			Config:      gitlab.Defaults(),
			OAuthConfig: gitlab.OAuthDefaults(),
		},
	}
}

// New builds and configures otf services and their http handlers.
func New(logger logr.Logger, db otf.DB, cfg Config) (*Services, error) {
	hostnameService := otf.NewHostnameService(cfg.Host)

	renderer, err := html.NewViewEngine(cfg.DevMode)
	if err != nil {
		return nil, fmt.Errorf("setting up web page renderer: %w", err)
	}
	cloudService, err := inmem.NewCloudService(cfg.Github.Config, cfg.Gitlab.Config)
	if err != nil {
		return nil, err
	}
	cache, err := inmem.NewCache(*cfg.CacheConfig)
	if err != nil {
		return nil, err
	}
	logger.Info("started cache", "max_size", cfg.CacheConfig.Size, "ttl", cfg.CacheConfig.TTL)

	broker := pubsub.NewBroker(logger, db)

	// Setup url signer
	signer := otf.NewSigner(cfg.Secret)

	orgService := organization.NewService(organization.Options{
		Logger:   logger,
		DB:       db,
		Renderer: renderer,
		Broker:   broker,
	})
	authService, err := auth.NewService(auth.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		Configs:             []cloud.CloudOAuthConfig{cfg.Github, cfg.Gitlab},
		SiteToken:           cfg.SiteToken,
		HostnameService:     hostnameService,
		OrganizationService: orgService,
	})
	if err != nil {
		return nil, fmt.Errorf("setting up auth service: %w", err)
	}
	vcsProviderService := vcsprovider.NewService(vcsprovider.Options{
		Logger:       logger,
		DB:           db,
		Renderer:     renderer,
		CloudService: cloudService,
	})
	repoService := repo.NewService(repo.Options{
		Logger:             logger,
		DB:                 db,
		CloudService:       cloudService,
		HostnameService:    hostnameService,
		Publisher:          broker,
		VCSProviderService: vcsProviderService,
	})

	workspaceService := workspace.NewService(workspace.Options{
		Logger:              logger,
		DB:                  db,
		Broker:              broker,
		Renderer:            renderer,
		RepoService:         repoService,
		TeamService:         authService,
		OrganizationService: orgService,
		VCSProviderService:  vcsProviderService,
	})
	configService := configversion.NewService(configversion.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		Cache:               cache,
		Signer:              signer,
		MaxUploadSize:       cfg.MaxConfigSize,
	})
	runService := run.NewService(run.Options{
		Logger:                      logger,
		DB:                          db,
		Renderer:                    renderer,
		WorkspaceAuthorizer:         workspaceService,
		WorkspaceService:            workspaceService,
		ConfigurationVersionService: configService,
		VCSProviderService:          vcsProviderService,
		Broker:                      broker,
		Cache:                       cache,
		Signer:                      signer,
	})
	logsService := logs.NewService(logs.Options{
		Logger:        logger,
		DB:            db,
		RunAuthorizer: runService,
		Cache:         cache,
		Broker:        broker,
		Verifier:      signer,
	})
	moduleService := module.NewService(module.Options{
		Logger:             logger,
		DB:                 db,
		Renderer:           renderer,
		HostnameService:    hostnameService,
		VCSProviderService: vcsProviderService,
		Signer:             signer,
		RepoService:        repoService,
	})
	stateService := state.NewService(state.Options{
		Logger:              logger,
		DB:                  db,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
		Cache:               cache,
	})
	variableService := variable.NewService(variable.Options{
		Logger:              logger,
		DB:                  db,
		Renderer:            renderer,
		WorkspaceAuthorizer: workspaceService,
		WorkspaceService:    workspaceService,
	})

	handlers := []otf.Handlers{
		authService,
		workspaceService,
		orgService,
		variableService,
		vcsProviderService,
		stateService,
		moduleService,
		configService,
		runService,
		logsService,
		repoService,
	}

	return &Services{
		Config:                      cfg,
		Handlers:                    handlers,
		AuthService:                 authService,
		WorkspaceService:            workspaceService,
		OrganizationService:         orgService,
		VariableService:             variableService,
		VCSProviderService:          vcsProviderService,
		StateService:                stateService,
		ModuleService:               moduleService,
		HostnameService:             hostnameService,
		ConfigurationVersionService: configService,
		RunService:                  runService,
		LogsService:                 logsService,
		RepoService:                 repoService,
		Broker:                      broker,
	}, nil
}

func (s *Services) NewServer(logger logr.Logger) (*http.Server, net.Listener, error) {
	server, err := http.NewServer(logger, http.ServerConfig{
		SSL:                  s.SSL,
		CertFile:             s.CertFile,
		KeyFile:              s.KeyFile,
		EnableRequestLogging: s.EnableRequestLogging,
		DevMode:              s.DevMode,
		Middleware: []mux.MiddlewareFunc{
			auth.AuthenticateToken(s.AuthService, s.SiteToken),
			auth.AuthenticateSession(s.AuthService),
		},
		Handlers: s.Handlers,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("setting up http server: %w", err)
	}
	ln, err := net.Listen("tcp", s.Address)
	if err != nil {
		return nil, nil, err
	}
	// Unless user has set a hostname, set the hostname to the listening address
	// of the http server.
	if s.Config.Host == "" {
		listenAddress := ln.Addr().(*net.TCPAddr)
		s.SetHostname(otf.NormalizeAddress(listenAddress))
	}
	return server, ln, nil
}

// If --hostname not set then use http server's listening address as

func (s *Services) NewAgent(logger logr.Logger) (daemon, error) {
	return agent.NewAgent(
		logger.WithValues("component", "agent"),
		client.LocalClient{
			AgentTokenService:           s.AuthService,
			WorkspaceService:            s.WorkspaceService,
			OrganizationService:         s.OrganizationService,
			VariableService:             s.VariableService,
			StateService:                s.StateService,
			HostnameService:             s.HostnameService,
			ConfigurationVersionService: s.ConfigurationVersionService,
			RegistrySessionService:      s.AuthService,
			RunService:                  s.RunService,
			LogsService:                 s.LogsService,
		},
		*s.AgentConfig,
	)
}

func (s *Services) StartScheduler(ctx context.Context, logger logr.Logger, db otf.DB) error {
	return scheduler.Start(ctx, scheduler.Options{
		Logger:           logger,
		WorkspaceService: s.WorkspaceService,
		RunService:       s.RunService,
		DB:               db,
		Subscriber:       s.Broker,
	})
}
