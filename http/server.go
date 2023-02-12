package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
)

const (
	// shutdownTimeout is the time given for outstanding requests to finish
	// before shutdown.
	shutdownTimeout = 1 * time.Second
)

var (
	healthzPayload = mustMarshal(struct {
		Version string
		Commit  string
		Built   string
	}{
		Version: otf.Version,
		Commit:  otf.Commit,
		Built:   otf.Built,
	})
	discoveryPayload = mustMarshal(struct {
		ModulesV1  string `json:"modules.v1"`
		MotdV1     string `json:"motd.v1"`
		StateV2    string `json:"state.v2"`
		TfeV2      string `json:"tfe.v2"`
		TfeV21     string `json:"tfe.v2.1"`
		TfeV22     string `json:"tfe.v2.2"`
		VersionsV1 string `json:"versions.v1"`
	}{
		ModulesV1:  "/api/v2/",
		MotdV1:     "/api/terraform/motd",
		StateV2:    "/api/v2/",
		TfeV2:      "/api/v2/",
		TfeV21:     "/api/v2/",
		TfeV22:     "/api/v2/",
		VersionsV1: "https://checkpoint-api.hashicorp.com/v1/versions/",
	})
)

// ServerConfig is the http server config
type ServerConfig struct {
	SSL                  bool
	CertFile, KeyFile    string
	EnableRequestLogging bool
	DevMode              bool
}

// Server provides an HTTP/S server
type Server struct {
	logr.Logger
	ServerConfig

	server *http.Server
}

// NewServer is the constructor for a http server
func NewServer(logger logr.Logger, cfg ServerConfig, apis ...otf.HTTPAPI) (*Server, error) {
	s := &Server{
		server:       &http.Server{},
		Logger:       logger,
		ServerConfig: cfg,
	}

	if cfg.SSL {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, fmt.Errorf("must provide both --cert-file and --key-file")
		}
	}

	r := mux.NewRouter()

	// Catch panics and return 500s
	r.Use(handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))

	// Redirect paths with a trailing slash to path without, e.g. /runs/ ->
	// /runs. Uses an HTTP301.
	r.StrictSlash(true)

	r.Handle("/", http.RedirectHandler("/organizations", http.StatusFound))

	// Serve static files
	html.AddStaticHandler(r, cfg.DevMode)

	// Prometheus metrics
	r.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.Write(healthzPayload)
	})

	r.HandleFunc("/.well-known/terraform.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.Write(discoveryPayload)
	})

	// Add tfp api version header to every response
	//
	// TODO: only set this on api routes
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Version 2.5 is the minimum version terraform requires for the
			// newer 'cloud' configuration block:
			// https://developer.hashicorp.com/terraform/cli/cloud/settings#the-cloud-block
			w.Header().Set("TFP-API-Version", "2.5")
			next.ServeHTTP(w, r)
		})
	})

	// Get/set session organization
	//
	// TODO: only set this on web app routes
	r.Use(html.SetOrganization)

	r.HandleFunc("/api/v2/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	for _, api := range apis {
		api.AddHandlers(r)
	}

	// Toggle logging HTTP requests
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

// mustMarshal marshals a value into json and panics upon error
func mustMarshal(v any) []byte {
	encoded, err := json.Marshal(v)
	if err != nil {
		panic(err.Error())
	}
	return encoded
}
