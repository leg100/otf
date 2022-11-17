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

	"github.com/allegro/bigcache"
	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/signer"
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
	// Listening Address in the form <ip>:<port>
	Addr                 string
	SSL                  bool
	CertFile, KeyFile    string
	EnableRequestLogging bool
	// site admin token
	SiteToken string
	// Secret for signing
	Secret string
	// Maximum permitted config upload size in bytes
	MaxConfigSize int64
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
	ServerConfig

	server *http.Server
	ln     net.Listener

	logr.Logger

	// provides access to otf services
	otf.Application

	CacheService *bigcache.BigCache

	// server-side-events server
	eventsServer *sse.Server
	// the http router, exported so that other pkgs can add routes
	*Router
	*signer.Signer
}

// NewServer is the constructor for Server
func NewServer(logger logr.Logger, cfg ServerConfig, app otf.Application, db otf.DB, cache *bigcache.BigCache) (*Server, error) {
	s := &Server{
		server:       &http.Server{},
		Logger:       logger,
		ServerConfig: cfg,
		Application:  app,
		eventsServer: newSSEServer(),
		Signer:       signer.New([]byte(cfg.Secret), signer.SkipQuery()),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	r := NewRouter()
	s.Router = r

	// Catch panics and return 500s
	r.Use(handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))

	// Optionally enable HTTP request logging
	if cfg.EnableRequestLogging {
		r.Use(s.loggingMiddleware)
	}

	r.GET("/.well-known/terraform.json", s.WellKnown)
	r.GET("/metrics", promhttp.Handler().ServeHTTP)
	r.GET("/healthz", GetHealthz)

	// These are signed URLs that expire after a given time. They don't use
	// bearer authentication.
	r.PathPrefix("/signed/{signature.expiry}").Sub(func(signed *Router) {
		signed.Use((&signatureVerifier{s.Signer}).handler)

		signed.GET("/runs/{run_id}/logs/{phase}", s.getLogs)
		signed.PUT("/configuration-versions/{id}/upload", s.UploadConfigurationVersion())
	})

	r.PathPrefix("/api/v2").Sub(func(api *Router) {
		api.GET("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		// Authenticated endpoints
		api.Sub(func(r *Router) {
			// Ensure request has valid API token
			r.Use((&authTokenMiddleware{
				UserService:       app,
				AgentTokenService: app,
				siteToken:         cfg.SiteToken,
			}).handler)

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
			r.GET("/state-versions/{id}/download", s.DownloadStateVersion)

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
		})
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

	s.Info("started server", "address", s.ln.Addr().String(), "ssl", s.SSL)

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
	s.Info("gracefully shutting down server...")

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
