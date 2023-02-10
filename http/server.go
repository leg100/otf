package http

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/felixge/httpsnoop"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
)

const (
	// shutdownTimeout is the time given for outstanding requests to finish
	// before shutdown.
	shutdownTimeout = 1 * time.Second

	jsonApplication = "application/json"
)

var (
	// Files embedded within the go binary
	//
	//go:embed static
	embedded embed.FS

	// The same files but on the local disk
	localDisk = os.DirFS("http/html")
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

	r.Use(setOrganization)

	// Serve static assets (JS, CSS, etc) from within go binary. Dev mode
	// sources files from local disk instead.
	var fs http.FileSystem
	if cfg.DevMode {
		fs = &cacheBuster{localDisk}
	} else {
		fs = &cacheBuster{embedded}
	}
	r.PathPrefix("/static/").Handler(http.FileServer(fs)).Methods("GET")

	// Prometheus metrics
	r.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	// TODO: marshal at compile-time
	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		payload, err := json.Marshal(struct {
			Version string
			Commit  string
			Built   string
		}{
			Version: otf.Version,
			Commit:  otf.Commit,
			Built:   otf.Built,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Content-type", jsonApplication)
		w.Write(payload)
	})

	// TODO: marshal at compile-time
	r.HandleFunc("/.well-known/terraform.json", func(w http.ResponseWriter, r *http.Request) {
		payload, err := json.Marshal(struct {
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
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		w.Header().Set("Content-type", jsonApplication)
		w.Write(payload)
	})

	api := r.PathPrefix("/api/v2").Subrouter()

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

	api.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

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
