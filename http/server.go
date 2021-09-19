package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/urfave/negroni"
)

const (
	// ShutdownTimeout is the time given for outstanding requests to finish
	// before shutdown.
	ShutdownTimeout = 1 * time.Second

	jsonApplication = "application/json"

	UploadConfigurationVersionRoute WebRoute = "/configuration-versions/%v/upload"
	GetPlanLogsRoute                WebRoute = "/plans/%v/logs"
	GetApplyLogsRoute               WebRoute = "/applies/%v/logs"
)

type WebRoute string

// Server provides an HTTP/S server
type Server struct {
	server *http.Server
	ln     net.Listener
	err    chan error

	logr.Logger

	SSL               bool
	CertFile, KeyFile string

	// Listening Address in the form <ip>:<port>
	Addr string

	// Hostname, used within absolute URL links, defaults to localhost
	Hostname string

	OrganizationService         otf.OrganizationService
	WorkspaceService            otf.WorkspaceService
	StateVersionService         otf.StateVersionService
	ConfigurationVersionService otf.ConfigurationVersionService
	RunService                  otf.RunService
	PlanService                 otf.PlanService
	ApplyService                otf.ApplyService
}

// NewServer is the contructor for Server
func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		err:    make(chan error),
	}

	http.Handle("/", NewRouter(s))

	return s
}

// NewRouter constructs a negroni-wrapped HTTP router
func NewRouter(server *Server) *negroni.Negroni {
	router := mux.NewRouter()

	router.HandleFunc("/.well-known/terraform.json", server.WellKnown)

	router.HandleFunc("/state-versions/{id}/download", server.DownloadStateVersion).Methods("GET")
	router.HandleFunc("/configuration-versions/{id}/upload", server.UploadConfigurationVersion).Methods("PUT")
	router.HandleFunc("/plans/{id}/logs", server.GetPlanLogs).Methods("GET")
	router.HandleFunc("/applies/{id}/logs", server.GetApplyLogs).Methods("GET")
	router.HandleFunc("/runs/{id}/logs", server.UploadLogs).Methods("POST")

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
	sub.HandleFunc("/runs/{id}/plan/json-output", server.GetRunPlanJSON).Methods("GET")

	// Plan routes
	sub.HandleFunc("/plans/{id}", server.GetPlan).Methods("GET")
	sub.HandleFunc("/plans/{id}/json-output", server.GetPlanJSON).Methods("GET")

	// Apply routes
	sub.HandleFunc("/applies/{id}", server.GetApply).Methods("GET")

	// Setup negroni and middleware
	n := negroni.New()
	// Catch panics and return 500s
	n.Use(negroni.NewRecovery())

	// Log requests
	n.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		start := time.Now()

		next(rw, r)

		res := rw.(negroni.ResponseWriter)
		server.Info("request",
			"duration", time.Since(start).Milliseconds(),
			"status", res.Status(),
			"method", r.Method,
			"path", fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery))
	})

	n.UseHandler(router)

	return n
}

// GetURL returns an absolute URL corresponding to the given route and params
func (s *Server) GetURL(route WebRoute, param ...interface{}) string {
	url := &url.URL{
		Scheme: "https",
		Host:   s.Hostname,
		Path:   fmt.Sprintf(string(route), param...),
	}

	return url.String()
}

// Open validates the server options and begins listening on the bind address.
func (s *Server) Open() (err error) {
	if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
		return err
	}

	// Begin serving requests on the listener. We use Serve() instead of
	// ListenAndServe() because it allows us to check for listen errors (such as
	// trying to use an already open port) synchronously.
	go func() {
		if s.SSL {
			s.err <- s.server.ServeTLS(s.ln, s.CertFile, s.KeyFile)
		} else {
			s.err <- s.server.Serve(s.ln)
		}
	}()

	return nil
}

// Port returns the TCP port for the running server.  This is useful in tests
// where we allocate a random port by using ":0".
func (s *Server) Port() int {
	if s.ln == nil {
		return 0
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

// Wait blocks until server stops listening or context is cancelled.
func (s *Server) Wait(ctx context.Context) error {
	select {
	case err := <-s.err:
		return err
	case <-ctx.Done():
		return s.server.Close()
	}
}

// Close gracefully shuts down the server.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}
