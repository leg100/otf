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

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/hooks"
	"github.com/leg100/surl"
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
	SSL                  bool
	CertFile, KeyFile    string
	EnableRequestLogging bool
	SiteToken            string // site admin token
	Secret               string // Secret for signing
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
	logr.Logger
	otf.Application // provides access to otf services
	ServerConfig
	*Router      // http router, exported so that other pkgs can add routes
	*surl.Signer // sign and validate signed URLs

	server *http.Server
}

// NewServer is the constructor for Server
func NewServer(logger logr.Logger, cfg ServerConfig, app otf.Application, db otf.DB, stateService otf.StateVersionService, variableService otf.VariableService, registrySessionService otf.RegistrySessionService) (*Server, error) {
	s := &Server{
		server:       &http.Server{},
		Logger:       logger,
		ServerConfig: cfg,
		Application:  app,
	}

	// configure URL signer
	s.Signer = surl.New([]byte(cfg.Secret),
		surl.PrefixPath("/signed"),
		surl.WithPathFormatter(),
		surl.WithBase58Expiry(),
		surl.SkipQuery(),
	)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	r := NewRouter()
	s.Router = r

	// Catch panics and return 500s
	r.Use(handlers.RecoveryHandler(handlers.PrintRecoveryStack(true)))

	r.GET("/.well-known/terraform.json", s.WellKnown)
	r.GET("/metrics", promhttp.Handler().ServeHTTP)
	r.GET("/healthz", GetHealthz)

	//
	// VCS event handling
	//
	events := make(chan cloud.VCSEvent, 100)
	r.Handle("/webhooks/vcs/{webhook_id}", hooks.NewHandler(logger, events, s.Application))
	// s.vcsEventsHandler = triggerer.NewTriggerer(app, logger, events)

	// These are signed URLs that expire after a given time.
	r.PathPrefix("/signed/{signature.expiry}").Sub(func(signed *Router) {
		signed.Use((&SignatureVerifier{s.Signer}).Handler)

		signed.GET("/runs/{run_id}/logs/{phase}", s.getLogs)
		signed.GET("/modules/download/{module_version_id}.tar.gz", s.downloadModuleVersion)
	})

	authMiddleware := &authTokenMiddleware{
		UserService:            app,
		AgentTokenService:      app,
		RegistrySessionService: registrySessionService,
		siteToken:              cfg.SiteToken,
	}

	r.PathPrefix("/api/v2").Sub(func(api *Router) {
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

		api.GET("/ping", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		// Authenticated endpoints
		api.Sub(func(r *Router) {
			// Ensure request has valid API bearer token
			r.Use(authMiddleware.handler)

			// Variables routes
			variableService.AddHandlers(r.Router)

			// StateVersion routes
			stateService.AddHandlers(r.Router)

			// ConfigurationVersion routes
			configService.AddHandlers(r.Router)

			// Watch routes
			watchService.AddHandlers(r.Router)

			// Run routes
			runService.AddHandlers(r.Router)

			// User routes
			userService.AddHandlers(r.Router)

			// Agent token routes
			agentTokenService.AddHandlers(r.Router)

			// Registry session routes
			registrySessionService.AddHandlers(r.Router)
		})
	})

	// module registry
	r.PathPrefix("/api/registry/v1/modules").Sub(func(r *Router) {
		// Ensure request has valid API bearer token
		r.Use(authMiddleware.handler)

		moduleService.AddHandlers(r.Router)
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
	// start handling incoming VCS events
	go s.vcsEventsHandler.Start(ctx)

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
