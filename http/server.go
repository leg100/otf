package http

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/hashicorp/jsonapi"
	"github.com/leg100/ots"
	"github.com/urfave/negroni"
)

const (
	// ShutdownTimeout is the time given for outstanding requests to finish
	// before shutdown.
	ShutdownTimeout = 1 * time.Second

	jsonApplication = "application/json"
)

// Query schema decoder, caches structs, and safe for sharing
var decoder = schema.NewDecoder()

// The HTTP/S server
type Server struct {
	server *http.Server
	router *mux.Router
	ln     net.Listener
	err    chan error

	SSL               bool
	CertFile, KeyFile string

	// Listening Address in the form <ip>:<port>
	Addr string

	OrganizationService         ots.OrganizationService
	WorkspaceService            ots.WorkspaceService
	StateVersionService         ots.StateVersionService
	ConfigurationVersionService ots.ConfigurationVersionService
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

	// Add default set of negroni middleware to routes: (i) Logging (ii)
	// Recovery (iii) Static File serving (we don't use this one...)
	n := negroni.Classic()
	n.UseHandler(router)

	return n
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

// Port returns the TCP port for the running server.
// This is useful in tests where we allocate a random port by using ":0".
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
