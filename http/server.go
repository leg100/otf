package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/r3labs/sse/v2"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/hooks"
	"github.com/leg100/otf/triggerer"
	"github.com/leg100/surl"
)

const (
	// shutdownTimeout is the time given for outstanding requests to finish
	// before shutdown.
	shutdownTimeout = 1 * time.Second

	jsonApplication = "application/json"
)

type WebRoute string

// ServerConfig is the http server config
type ServerConfig struct {
	SSL                  bool
	CertFile, KeyFile    string
	EnableRequestLogging bool
	SiteToken            string // site admin token
	Secret               string // Secret for signing
	MaxConfigSize        int64  // Maximum permitted config upload size in bytes
}

func (cfg *ServerConfig) Validate() error {
	if cfg.SSL {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return fmt.Errorf("must provide both --cert-file and --key-file")
		}
	}
	if cfg.Secret == "" {
		return fmt.Errorf("--secret cannot be empty")
	}
	return nil
}

// Server provides an HTTP/S server
type Server struct {
	logr.Logger
	otf.Application // provides access to otf services
	ServerConfig
	*Router      // http router, exported so that other pkgs can add routes
	*surl.Signer // sign and validate signed URLs

	eventsServer     *sse.Server
	server           *http.Server
	vcsEventsHandler *triggerer.Triggerer
}

// NewServer is the constructor for Server
func NewServer(logger logr.Logger, cfg ServerConfig, app otf.Application, db otf.DB, stateService otf.StateVersionService, variableService otf.VariableService, registrySessionService otf.RegistrySessionService) (*Server, error) {
	s := &Server{
		server:       &http.Server{},
		Logger:       logger,
		ServerConfig: cfg,
		Application:  app,
		eventsServer: newSSEServer(),
	}

	// configure URL signer
	s.Signer = surl.New([]byte(cfg.Secret),
		surl.PrefixPath("/signed"),
		surl.WithPathFormatter(),
		surl.WithBase58Expiry(),
		surl.SkipQuery(),
	)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	r := NewRouter()
	s.Router = r

	// Catch panics and return 500s
	r.Use(handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))

	r.GET("/.well-known/terraform.json", s.WellKnown)
	r.GET("/metrics", promhttp.Handler().ServeHTTP)
	r.GET("/healthz", GetHealthz)

	//
	// VCS event handling
	//
	events := make(chan cloud.VCSEvent, 100)
	r.Handle("/webhooks/vcs/{webhook_id}", hooks.NewHandler(logger, events, s.Application))
	s.vcsEventsHandler = triggerer.NewTriggerer(app, logger, events)

	// These are signed URLs that expire after a given time.
	r.PathPrefix("/signed/{signature.expiry}").Sub(func(signed *Router) {
		// TODO: use mux.Middleware interface
		signed.Use((&signatureVerifier{s.Signer}).handler)

		signed.GET("/runs/{run_id}/logs/{phase}", s.getLogs)
		signed.PUT("/configuration-versions/{id}/upload", s.UploadConfigurationVersion())
		signed.GET("/modules/download/{module_version_id}.tar.gz", s.downloadModuleVersion)
	})

	authMiddleware := &authTokenMiddleware{
		UserService:            app,
		AgentTokenService:      app,
		RegistrySessionService: registrySessionService,
		siteToken:              cfg.SiteToken,
	}

	r.PathPrefix("/api/v2").Sub(func(api *Router) {
		// Add tfp api version header to every response
		api.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Version 2.5 is the minimum version terraform requires for the
				// newer 'cloud' configuration block:
				// https://developer.hashicorp.com/terraform/cli/cloud/settings#the-cloud-block
				w.Header().Set("TFP-API-Version", "2.5")
				next.ServeHTTP(w, r)
			})
		})

		api.GET("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		// Authenticated endpoints
		api.Sub(func(r *Router) {
			// Ensure request has valid API bearer token
			r.Use(authMiddleware.handler)

			// Organization routes
			r.GET("/organizations", s.ListOrganizations)
			r.PST("/organizations", s.CreateOrganization)
			r.GET("/organizations/{name}", s.GetOrganization)
			r.PTC("/organizations/{name}", s.UpdateOrganization)
			r.DEL("/organizations/{name}", s.DeleteOrganization)
			r.GET("/organizations/{name}/entitlement-set", s.GetEntitlements)

			// Workspace routes
			r.GET("/organizations/{organization_name}/workspaces", s.ListWorkspaces)
			r.PST("/organizations/{organization_name}/workspaces", s.CreateWorkspace)
			r.GET("/organizations/{organization_name}/workspaces/{workspace_name}", s.GetWorkspaceByName)
			r.PTC("/organizations/{organization_name}/workspaces/{workspace_name}", s.UpdateWorkspaceByName)
			r.DEL("/organizations/{organization_name}/workspaces/{workspace_name}", s.DeleteWorkspaceByName)

			r.PTC("/workspaces/{workspace_id}", s.UpdateWorkspace)
			r.GET("/workspaces/{workspace_id}", s.GetWorkspace)
			r.DEL("/workspaces/{workspace_id}", s.DeleteWorkspace)
			r.PST("/workspaces/{workspace_id}/actions/lock", s.LockWorkspace)
			r.PST("/workspaces/{workspace_id}/actions/unlock", s.UnlockWorkspace)

			// Variables routes
			variableService.AddHandlers(r.Router)

			// StateVersion routes
			stateService.AddHandlers(r.Router)

			// ConfigurationVersion routes
			r.PST("/workspaces/{workspace_id}/configuration-versions", s.CreateConfigurationVersion)
			r.GET("/configuration-versions/{id}", s.GetConfigurationVersion)
			r.GET("/workspaces/{workspace_id}/configuration-versions", s.ListConfigurationVersions)
			r.GET("/configuration-versions/{id}/download", s.DownloadConfigurationVersion)

			// Run routes
			r.PST("/runs", s.CreateRun)
			r.PST("/runs/{id}/actions/apply", s.ApplyRun)
			r.GET("/runs", s.ListRuns)
			r.GET("/workspaces/{workspace_id}/runs", s.ListRuns)
			r.GET("/runs/{id}", s.GetRun)
			r.PST("/runs/{id}/actions/discard", s.DiscardRun)
			r.PST("/runs/{id}/actions/cancel", s.CancelRun)
			r.PST("/runs/{id}/actions/force-cancel", s.ForceCancelRun)
			r.GET("/organizations/{organization_name}/runs/queue", s.GetRunsQueue)

			// Run routes for exclusive use by remote agents
			r.PST("/runs/{id}/actions/start/{phase}", s.startPhase)
			r.PST("/runs/{id}/actions/finish/{phase}", s.finishPhase)
			r.PUT("/runs/{run_id}/logs/{phase}", s.putLogs)
			r.GET("/runs/{run_id}/planfile", s.getPlanFile)
			r.PUT("/runs/{run_id}/planfile", s.uploadPlanFile)
			r.GET("/runs/{run_id}/lockfile", s.getLockFile)
			r.PUT("/runs/{run_id}/lockfile", s.uploadLockFile)

			// Event routes
			r.GET("/watch", s.watch)

			// Plan routes
			r.GET("/plans/{plan_id}", s.getPlan)
			r.GET("/plans/{plan_id}/json-output", s.getPlanJSON)

			// Apply routes
			r.GET("/applies/{apply_id}", s.GetApply)

			// User routes
			r.GET("/account/details", s.GetCurrentUser)

			// Agent token routes
			r.GET("/agent/details", s.GetCurrentAgent)
			r.PST("/agent/create", s.CreateAgentToken)

			// Registry session routes
			registrySessionService.AddHandlers(r.Router)
		})
	})

	// module registry
	r.PathPrefix("/api/registry/v1/modules").Sub(func(r *Router) {
		// Ensure request has valid API bearer token
		r.Use(authMiddleware.handler)

		r.GET("/{organization}/{name}/{provider}/versions", s.listModuleVersions)
		r.GET("/{organization}/{name}/{provider}/{version}/download", s.getModuleVersionDownloadLink)
	})

	// Optionally log all HTTP requests
	if cfg.EnableRequestLogging {
		http.Handle("/", s.loggingMiddleware(r))
	} else {
		http.Handle("/", r)
	}

	return s, nil
}

// Start starts serving http traffic on the given listener and waits until the server exits due to
// error or the context is cancelled.
func (s *Server) Start(ctx context.Context, ln net.Listener) (err error) {
	// start handling incoming VCS events
	go s.vcsEventsHandler.Start(ctx)

	errch := make(chan error)

	go func() {
		if s.SSL {
			errch <- s.server.ServeTLS(ln, s.CertFile, s.KeyFile)
		} else {
			errch <- s.server.Serve(ln)
		}
	}()

	s.Info("started server", "address", ln.Addr().String(), "ssl", s.SSL)

	// Block until server stops listening or context is cancelled.
	select {
	case err := <-errch:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		s.Info("gracefully shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			return s.server.Close()
		}

		return nil
	}
}

// newLoggingMiddleware returns middleware that logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(next, w, r)

		s.Info("request",
			"duration", fmt.Sprintf("%dms", m.Duration.Milliseconds()),
			"status", m.Code,
			"method", r.Method,
			"path", fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery))
	})
}

func newSSEServer() *sse.Server {
	srv := sse.New()
	// we don't use last-event-item functionality so turn it off
	srv.AutoReplay = false
	// encode payloads into base64 otherwise the JSON string payloads corrupt
	// the SSE protocol
	srv.EncodeBase64 = true
	return srv
}
