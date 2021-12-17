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

	// html web app
	webApp *html.Application

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
	}

	// Construct web app
	webApp, err := html.NewApplication(logger, cfg.ApplicationConfig)
	if err != nil {
		return nil, err
	}
	s.webApp = webApp

	http.Handle("/", s.routes(cfg))

	return s, nil
}

// NewRouter constructs an HTTP router
func (server *Server) routes(cfg ServerConfig) http.Handler {
	router := mux.NewRouter()

	// Catch panics and return 500s
	router.Use(handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))

	// Optionally enable HTTP request logging
	if cfg.EnableRequestLogging {
		router.Use(server.loggingMiddleware)
	}

	// Add web app routes.
	server.webApp.AddRoutes(router)

	router.HandleFunc("/.well-known/terraform.json", server.WellKnown)
	router.HandleFunc("/metrics/cache.json", server.CacheStats)

	router.HandleFunc("/state-versions/{id}/download", server.DownloadStateVersion).Methods("GET")
	router.HandleFunc("/configuration-versions/{id}/upload", server.UploadConfigurationVersion).Methods("PUT")
	router.HandleFunc("/plans/{id}/logs", server.GetPlanLogs).Methods("GET")
	router.HandleFunc("/plans/{id}/logs", server.UploadPlanLogs).Methods("PUT")
	router.HandleFunc("/applies/{id}/logs", server.GetApplyLogs).Methods("GET")
	router.HandleFunc("/applies/{id}/logs", server.UploadApplyLogs).Methods("PUT")
	router.HandleFunc("/runs/{id}/plan", server.UploadPlanFile).Methods("PUT")
	router.HandleFunc("/runs/{id}/plan", server.GetPlanFile).Methods("GET")

	router.HandleFunc("/healthz", GetHealthz).Methods("GET")

	router.HandleFunc("/app/{org}/{workspace}/runs/{id}", server.GetRunLogs).Methods("GET")

	// Websocket connections
	server.registerEventRoutes(router)

	// Filter json-api requests
	sub := router.Headers("Accept", jsonapi.MediaType).Subrouter()

	// Filter api v2 requests
	sub = sub.PathPrefix("/api/v2").Subrouter()

	sub.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Organization routes
	sub.HandleFunc("/organizations", server.ListOrganizations).Methods("GET")
	sub.HandleFunc("/organizations", server.CreateOrganization).Methods("POST")
	sub.HandleFunc("/organizations/{name}", server.GetOrganization).Methods("GET")
	sub.HandleFunc("/organizations/{name}", server.UpdateOrganization).Methods("PATCH")
	sub.HandleFunc("/organizations/{name}", server.DeleteOrganization).Methods("DELETE")
	sub.HandleFunc("/organizations/{name}/entitlement-set", server.GetEntitlements).Methods("GET")

	// Workspace routes
	sub.HandleFunc("/organizations/{org}/workspaces", server.ListWorkspaces).Methods("GET")
	sub.HandleFunc("/organizations/{org}/workspaces/{name}", server.GetWorkspace).Methods("GET")
	sub.HandleFunc("/organizations/{org}/workspaces", server.CreateWorkspace).Methods("POST")
	sub.HandleFunc("/organizations/{org}/workspaces/{name}", server.UpdateWorkspace).Methods("PATCH")
	sub.HandleFunc("/organizations/{org}/workspaces/{name}", server.DeleteWorkspace).Methods("DELETE")
	sub.HandleFunc("/workspaces/{id}", server.UpdateWorkspaceByID).Methods("PATCH")
	sub.HandleFunc("/workspaces/{id}", server.GetWorkspaceByID).Methods("GET")
	sub.HandleFunc("/workspaces/{id}", server.DeleteWorkspaceByID).Methods("DELETE")
	sub.HandleFunc("/workspaces/{id}/actions/lock", server.LockWorkspace).Methods("POST")
	sub.HandleFunc("/workspaces/{id}/actions/unlock", server.UnlockWorkspace).Methods("POST")

	// StateVersion routes
	sub.HandleFunc("/workspaces/{workspace_id}/state-versions", server.CreateStateVersion).Methods("POST")
	sub.HandleFunc("/workspaces/{workspace_id}/current-state-version", server.CurrentStateVersion).Methods("GET")
	sub.HandleFunc("/state-versions/{id}", server.GetStateVersion).Methods("GET")
	sub.HandleFunc("/state-versions", server.ListStateVersions).Methods("GET")

	// ConfigurationVersion routes
	sub.HandleFunc("/workspaces/{workspace_id}/configuration-versions", server.CreateConfigurationVersion).Methods("POST")
	sub.HandleFunc("/configuration-versions/{id}", server.GetConfigurationVersion).Methods("GET")
	sub.HandleFunc("/workspaces/{workspace_id}/configuration-versions", server.ListConfigurationVersions).Methods("GET")

	// Run routes
	sub.HandleFunc("/runs", server.CreateRun).Methods("POST")
	sub.HandleFunc("/runs/{id}/actions/apply", server.ApplyRun).Methods("POST")
	sub.HandleFunc("/workspaces/{workspace_id}/runs", server.ListRuns).Methods("GET")
	sub.HandleFunc("/runs/{id}", server.GetRun).Methods("GET")
	sub.HandleFunc("/runs/{id}/actions/discard", server.DiscardRun).Methods("POST")
	sub.HandleFunc("/runs/{id}/actions/cancel", server.CancelRun).Methods("POST")
	sub.HandleFunc("/runs/{id}/actions/force-cancel", server.ForceCancelRun).Methods("POST")
	sub.HandleFunc("/runs/{id}/plan/json-output", server.GetJSONPlanByRunID).Methods("GET")

	// Plan routes
	sub.HandleFunc("/plans/{id}", server.GetPlan).Methods("GET")
	sub.HandleFunc("/plans/{id}/json-output", server.GetPlanJSON).Methods("GET")

	// Apply routes
	sub.HandleFunc("/applies/{id}", server.GetApply).Methods("GET")

	return router
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
