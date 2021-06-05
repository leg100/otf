package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"github.com/leg100/ots"
)

// ShutdownTimeout is the time given for outstanding requests to finish before shutdown.
const ShutdownTimeout = 1 * time.Second

type Server struct {
	server *http.Server
	router *mux.Router
	ln     net.Listener
	err    chan error

	CertFile, KeyFile string
	Port              int

	OrganizationService ots.OrganizationService
}

func NewServer() *Server {
	s := &Server{
		server: &http.Server{},
		router: mux.NewRouter(),
		err:    make(chan error),
	}

	// Filter json-api requests
	r := s.router.Headers("Accept", jsonapi.MediaType).Subrouter()

	r.HandleFunc("/organizations", s.ListOrganizations).Methods("GET")
	r.HandleFunc("/organizations/{name}", s.GetOrganization).Methods("GET")
	r.HandleFunc("/organizations/{name}", s.CreateOrganization).Methods("POST")
	r.HandleFunc("/organizations/{name}", s.UpdateOrganization).Methods("PATCH")
	r.HandleFunc("/organizations/{name}", s.DeleteOrganization).Methods("DELETE")
	r.HandleFunc("/organizations/{name}/entitlement-set", s.GetEntitlements).Methods("GET")

	http.Handle("/", r)

	return s
}

// Open validates the server options and begins listening on the bind address.
func (s *Server) Open() (err error) {
	if s.CertFile == "" || s.KeyFile == "" {
		return errors.New("path to certificate and/or key cannot be empty")
	}

	if s.ln, err = net.Listen("tcp", s.Addr()); err != nil {
		return err
	}

	// Begin serving requests on the listener. We use ServeTLS() instead of
	// ListenAndServeTLS() because it allows us to check for listen errors (such
	// as trying to use an already open port) synchronously.
	go func() {
		s.err <- s.server.ServeTLS(s.ln, s.CertFile, s.KeyFile)
	}()

	return nil
}

func (s *Server) Addr() string {
	return fmt.Sprintf(":%d", s.Port)
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
