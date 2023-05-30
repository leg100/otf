package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/felixge/httpsnoop"
	gorillaHandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/json"
)

const (
	ModuleV1Prefix = "/v1/modules"
	APIPrefixV2    = "/api/v2"

	// shutdownTimeout is the time given for outstanding requests to finish
	// before shutdown.
	shutdownTimeout = 1 * time.Second
)

var (
	healthzPayload = json.MustMarshal(struct {
		Version string
		Commit  string
		Built   string
	}{
		Version: internal.Version,
		Commit:  internal.Commit,
		Built:   internal.Built,
	})

	// endpoints with these prefixes require authentication
	AuthenticatedPrefixes = []string{
		APIPrefixV2,
		ModuleV1Prefix,
		paths.UIPrefix,
	}
)

type (
	// ServerConfig is the http server config
	ServerConfig struct {
		SSL                  bool
		CertFile, KeyFile    string
		EnableRequestLogging bool
		DevMode              bool

		Handlers   []internal.Handlers
		Middleware []mux.MiddlewareFunc
	}

	// Server is the http server for OTF
	Server struct {
		logr.Logger
		ServerConfig

		server *http.Server
	}
)

// NewServer constructs the http server for OTF
func NewServer(logger logr.Logger, cfg ServerConfig) (*Server, error) {
	if cfg.SSL {
		if cfg.CertFile == "" || cfg.KeyFile == "" {
			return nil, fmt.Errorf("must provide both --cert-file and --key-file")
		}
	}

	r := mux.NewRouter()

	// Catch panics and return 500s
	r.Use(gorillaHandlers.RecoveryHandler(gorillaHandlers.PrintRecoveryStack(true)))

	// Redirect paths with a trailing slash to path without, e.g. /runs/ ->
	// /runs. Uses an HTTP301.
	r.StrictSlash(true)

	r.Handle("/", http.RedirectHandler("/app/organizations", http.StatusFound))

	// Serve static files
	if err := html.AddStaticHandler(logger, r, cfg.DevMode); err != nil {
		return nil, err
	}

	// Prometheus metrics
	r.HandleFunc("/metrics", promhttp.Handler().ServeHTTP)

	r.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.Write(healthzPayload)
	})

	r.HandleFunc(path.Join(APIPrefixV2, "ping"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Subrouter for service routes
	svcRouter := r.NewRoute().Subrouter()

	// Subject service routes to provided middleware, verifying tokens,
	// sessions.
	svcRouter.Use(cfg.Middleware...)

	// Add handlers for each service
	for _, h := range cfg.Handlers {
		h.AddHandlers(svcRouter)
	}

	// Add tfp api version header to every api response
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, APIPrefixV2) {
				// Version 2.5 is the minimum version terraform requires for the
				// newer 'cloud' configuration block:
				// https://developer.hashicorp.com/terraform/cli/cloud/settings#the-cloud-block
				w.Header().Set("TFP-API-Version", "2.5")
			}
			next.ServeHTTP(w, r)
		})
	})

	// Optionally log every request
	if cfg.EnableRequestLogging {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				m := httpsnoop.CaptureMetrics(next, w, r)
				logger.Info("request",
					"duration", fmt.Sprintf("%dms", m.Duration.Milliseconds()),
					"status", m.Code,
					"method", r.Method,
					"path", fmt.Sprintf("%s?%s", r.URL.Path, r.URL.RawQuery))
			})
		})
	}

	return &Server{
		Logger:       logger,
		ServerConfig: cfg,
		server:       &http.Server{Handler: r},
	}, nil
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

// APIRouter wraps the given router with a router suitable for API routes.
func APIRouter(r *mux.Router) *mux.Router {
	return r.PathPrefix(APIPrefixV2).Subrouter()
}
