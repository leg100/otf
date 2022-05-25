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
	"github.com/gorilla/mux"
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

	UploadConfigurationVersionRoute WebRoute = "/configuration-versions/%v/upload"
	GetPlanLogsRoute                WebRoute = "plans/%v/logs"
	GetApplyLogsRoute               WebRoute = "applies/%v/logs"
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

	router := mux.NewRouter()

	// Catch panics and return 500s
	router.Use(handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))

	// Optionally enable HTTP request logging
	if cfg.EnableRequestLogging {
		router.Use(s.loggingMiddleware)
	}

	// Add web app routes.
	if err := html.AddRoutes(logger, cfg.ApplicationConfig, app, db, router); err != nil {
		return nil, err
	}

	router.HandleFunc("/.well-known/terraform.json", s.WellKnown)
	router.HandleFunc("/metrics/cache.json", s.CacheStats)

	router.HandleFunc("/state-versions/{id}/download", s.DownloadStateVersion).Methods("GET")
	router.HandleFunc("/configuration-versions/{id}/upload", s.UploadConfigurationVersion).Methods("PUT")
	router.HandleFunc("/plans/{id}/logs", s.GetPlanLogs).Methods("GET")
	router.HandleFunc("/plans/{id}/logs", s.UploadPlanLogs).Methods("PUT")
	router.HandleFunc("/applies/{id}/logs", s.GetApplyLogs).Methods("GET")
	router.HandleFunc("/applies/{id}/logs", s.UploadApplyLogs).Methods("PUT")
	router.HandleFunc("/runs/{id}/plan", s.UploadPlanFile).Methods("PUT")
	router.HandleFunc("/runs/{id}/plan", s.GetPlanFile).Methods("GET")

	router.HandleFunc("/healthz", GetHealthz).Methods("GET")

	// Websocket connections
	s.registerEventRoutes(router)

	// Filter json-api requests
	sub := router.Headers("Accept", jsonapi.MediaType).Subrouter()

	// Filter api v2 requests
	sub = sub.PathPrefix("/api/v2").Subrouter()

	sub.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Organization routes
	sub.HandleFunc("/organizations", s.ListOrganizations).Methods("GET")
	sub.HandleFunc("/organizations", s.CreateOrganization).Methods("POST")
	sub.HandleFunc("/organizations/{name}", s.GetOrganization).Methods("GET")
	sub.HandleFunc("/organizations/{name}", s.UpdateOrganization).Methods("PATCH")
	sub.HandleFunc("/organizations/{name}", s.DeleteOrganization).Methods("DELETE")
	sub.HandleFunc("/organizations/{name}/entitlement-set", s.GetEntitlements).Methods("GET")

	// Workspace routes
	sub.HandleFunc("/organizations/{organization_name}/workspaces", s.ListWorkspaces).Methods("GET")
	sub.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", s.GetWorkspace).Methods("GET")
	sub.HandleFunc("/organizations/{organization_name}/workspaces", s.CreateWorkspace).Methods("POST")
	sub.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", s.UpdateWorkspace).Methods("PATCH")
	sub.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", s.DeleteWorkspace).Methods("DELETE")
	sub.HandleFunc("/workspaces/{id}", s.UpdateWorkspace).Methods("PATCH")
	sub.HandleFunc("/workspaces/{id}", s.GetWorkspace).Methods("GET")
	sub.HandleFunc("/workspaces/{id}", s.DeleteWorkspace).Methods("DELETE")
	sub.HandleFunc("/workspaces/{id}/actions/lock", s.LockWorkspace).Methods("POST")
	sub.HandleFunc("/workspaces/{id}/actions/unlock", s.UnlockWorkspace).Methods("POST")

	// StateVersion routes
	sub.HandleFunc("/workspaces/{workspace_id}/state-versions", s.CreateStateVersion).Methods("POST")
	sub.HandleFunc("/workspaces/{workspace_id}/current-state-version", s.CurrentStateVersion).Methods("GET")
	sub.HandleFunc("/state-versions/{id}", s.GetStateVersion).Methods("GET")
	sub.HandleFunc("/state-versions", s.ListStateVersions).Methods("GET")

	// ConfigurationVersion routes
	sub.HandleFunc("/workspaces/{workspace_id}/configuration-versions", s.CreateConfigurationVersion).Methods("POST")
	sub.HandleFunc("/configuration-versions/{id}", s.GetConfigurationVersion).Methods("GET")
	sub.HandleFunc("/workspaces/{workspace_id}/configuration-versions", s.ListConfigurationVersions).Methods("GET")

	// Run routes
	sub.HandleFunc("/runs", s.CreateRun).Methods("POST")
	sub.HandleFunc("/runs/{id}/actions/apply", s.ApplyRun).Methods("POST")
	sub.HandleFunc("/workspaces/{workspace_id}/runs", s.ListRuns).Methods("GET")
	sub.HandleFunc("/runs/{id}", s.GetRun).Methods("GET")
	sub.HandleFunc("/runs/{id}/actions/discard", s.DiscardRun).Methods("POST")
	sub.HandleFunc("/runs/{id}/actions/cancel", s.CancelRun).Methods("POST")
	sub.HandleFunc("/runs/{id}/actions/force-cancel", s.ForceCancelRun).Methods("POST")
	sub.HandleFunc("/runs/{id}/plan/json-output", s.GetJSONPlanByRunID).Methods("GET")

	// Plan routes
	sub.HandleFunc("/plans/{id}", s.GetPlan).Methods("GET")
	sub.HandleFunc("/plans/{id}/json-output", s.GetPlanJSON).Methods("GET")

	// Apply routes
	sub.HandleFunc("/applies/{id}", s.GetApply).Methods("GET")

	http.Handle("/", router)

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
