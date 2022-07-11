package http

import (
	"context"
	"fmt"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/handlers"

	"net"
	"net/http"
	"time"

	"github.com/allegro/bigcache"
	"github.com/go-logr/logr"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	httputil "github.com/leg100/otf/http/util"
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
	ApplicationConfig html.Config

	// Listening Address in the form <ip>:<port>
	Addr string

	SSL               bool
	CertFile, KeyFile string

	EnableRequestLogging bool

	// site authentication token
	SiteToken string
}

// Server provides an HTTP/S server
type Server struct {
	ServerConfig

	server *http.Server
	ln     net.Listener

	logr.Logger

	// provides access to otf services
	otf.Application

	CacheService *bigcache.BigCache
}

// NewServer is the constructor for Server
func NewServer(logger logr.Logger, cfg ServerConfig, app otf.Application, db otf.DB, cache *bigcache.BigCache) (*Server, error) {
	s := &Server{
		server:       &http.Server{},
		Logger:       logger,
		ServerConfig: cfg,
		Application:  app,
	}

	// Validate SSL params
	if cfg.SSL {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, fmt.Errorf("must provide both --cert-file and --key-file")
		}

		// Tell http utilities that we're using SSL.
		httputil.SSL = true
	}

	r := html.NewRouter()

	// Catch panics and return 500s
	r.Use(handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))

	// Optionally enable HTTP request logging
	if cfg.EnableRequestLogging {
		r.Use(s.loggingMiddleware)
	}

	// Add web app routes.
	if err := html.AddRoutes(logger, cfg.ApplicationConfig, app, r); err != nil {
		return nil, err
	}

	r.GET("/.well-known/terraform.json", s.WellKnown)
	r.GET("/metrics/cache.json", s.CacheStats)

	r.GET("/state-versions/{id}/download", s.DownloadStateVersion)
	r.PUT("/configuration-versions/{id}/upload", s.UploadConfigurationVersion)

	r.GET("/runs/{run_id}/logs/{phase}", s.getLogs)
	r.PUT("/runs/{run_id}/logs/{phase}", s.putLogs)

	r.GET("/healthz", GetHealthz)

	// Websocket connections
	s.registerEventRoutes(r)

	// JSON-API API endpoints
	japi := r.Headers("Accept", jsonapi.MediaType).PathPrefix("/api/v2")
	japi.Sub(func(r *html.Router) {
		// Ensure request has valid API token
		r.Use((&authTokenMiddleware{
			svc:       app.UserService(),
			siteToken: cfg.SiteToken,
		}).handler)

		r.GET("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		// Organization routes
		r.GET("/organizations", s.ListOrganizations)
		r.PST("/organizations", s.CreateOrganization)
		r.GET("/organizations/{name}", s.GetOrganization)
		r.PTC("/organizations/{name}", s.UpdateOrganization)
		r.DEL("/organizations/{name}", s.DeleteOrganization)
		r.GET("/organizations/{name}/entitlement-set", s.GetEntitlements)

		// Workspace routes
		r.GET("/organizations/{organization_name}/workspaces", s.ListWorkspaces)
		r.GET("/organizations/{organization_name}/workspaces/{workspace_name}", s.GetWorkspace)
		r.PST("/organizations/{organization_name}/workspaces", s.CreateWorkspace)
		r.PTC("/organizations/{organization_name}/workspaces/{workspace_name}", s.UpdateWorkspace)
		r.DEL("/organizations/{organization_name}/workspaces/{workspace_name}", s.DeleteWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/actions/lock", s.LockWorkspace)
		r.PST("/organizations/{organization_name}/workspaces/{workspace_name}/actions/unlock", s.UnlockWorkspace)
		r.PTC("/workspaces/{id}", s.UpdateWorkspace)
		r.GET("/workspaces/{id}", s.GetWorkspace)
		r.DEL("/workspaces/{id}", s.DeleteWorkspace)
		r.PST("/workspaces/{id}/actions/lock", s.LockWorkspace)
		r.PST("/workspaces/{id}/actions/unlock", s.UnlockWorkspace)

		// StateVersion routes
		r.PST("/workspaces/{workspace_id}/state-versions", s.CreateStateVersion)
		r.GET("/workspaces/{workspace_id}/current-state-version", s.CurrentStateVersion)
		r.GET("/state-versions/{id}", s.GetStateVersion)
		r.GET("/state-versions", s.ListStateVersions)

		// ConfigurationVersion routes
		r.PST("/workspaces/{workspace_id}/configuration-versions", s.CreateConfigurationVersion)
		r.GET("/configuration-versions/{id}", s.GetConfigurationVersion)
		r.GET("/workspaces/{workspace_id}/configuration-versions", s.ListConfigurationVersions)

		// Run routes
		r.PST("/runs", s.CreateRun)
		r.PST("/runs/{id}/actions/apply", s.ApplyRun)
		r.GET("/workspaces/{workspace_id}/runs", s.ListRuns)
		r.GET("/runs/{id}", s.GetRun)
		r.PST("/runs/{id}/actions/discard", s.DiscardRun)
		r.PST("/runs/{id}/actions/cancel", s.CancelRun)
		r.PST("/runs/{id}/actions/force-cancel", s.ForceCancelRun)
		r.GET("/organizations/{organization_name}/runs/queue", s.GetRunsQueue)

		// Plan routes
		r.GET("/plans/{plan_id}", s.GetPlan)
		r.GET("/plans/{plan_id}/json-output", s.GetPlanJSON)

		// Apply routes
		r.GET("/applies/{apply_id}", s.GetApply)

		// User routes
		r.GET("/account/details", s.GetCurrentUser)
	})

	http.Handle("/", r)

	return s, nil
}

// Open begins listening on the bind address and waits until server exits due to
// error or the context is cancelled.
func (s *Server) Open(ctx context.Context) (err error) {
	if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
		return err
	}

	errch := make(chan error)

	// Begin serving requests on the listener. We use Serve() instead of
	// ListenAndServe() because it allows us to check for listen errors (such as
	// trying to use an already open port) synchronously.
	go func() {
		if s.SSL {
			errch <- s.server.ServeTLS(s.ln, s.CertFile, s.KeyFile)
		} else {
			errch <- s.server.Serve(s.ln)
		}
	}()

	s.Logger.Info("started server", "address", s.Addr, "ssl", s.SSL)

	// Block until server stops listening or context is cancelled.
	select {
	case err := <-errch:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	case <-ctx.Done():
		return s.shutdown()
	}
}

// shutdown attempts to gracefully shuts down the server before a timeout
// expires at which point it forcefully closes the server.
func (s *Server) shutdown() error {
	s.Logger.Info("gracefully shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		return s.server.Close()
	}

	return nil
}

// newLoggingMiddleware returns middleware that logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m := httpsnoop.CaptureMetrics(next, w, r)

		s.Logger.Info("request",
			"duration", fmt.Sprintf("%dms", m.Duration.Milliseconds()),
			"status", m.Code,
			"method", r.Method,
			"path", fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery))
	})
}
