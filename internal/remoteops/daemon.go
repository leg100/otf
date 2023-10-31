package remoteops

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/releases"
	"golang.org/x/sync/errgroup"
)

const (
	DefaultID          = "agent-001"
	DefaultConcurrency = 5
)

var (
	PluginCacheDir = filepath.Join(os.TempDir(), "plugin-cache")
	DefaultEnvs    = []string{
		"TF_IN_AUTOMATION=true",
		"CHECKPOINT_DISABLE=true",
	}
)

// daemon performs remote operations
type daemon struct {
	Config
	logr.Logger
	releases.Downloader

	client
	spooler     // spools new run events
	*terminator // terminates runs

	envs []string // terraform environment variables
}

// NewDaemon constructs a remote operations daemon, which is either in-process,
// part of otfd, or an external process, otf-agent, communicating with the
// server via RPC.
func NewDaemon(logger logr.Logger, app client, cfg Config) (*daemon, error) {
	if cfg.Concurrency == 0 {
		cfg.Concurrency = DefaultConcurrency
	}
	if cfg.isAgent && cfg.Organization == nil {
		return nil, fmt.Errorf("an agent requires a specific organization to be specified")
	}
	if cfg.Sandbox {
		if _, err := exec.LookPath("bwrap"); errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("sandbox mode requires bubblewrap: %w", err)
		}
		logger.V(0).Info("enabled sandbox mode")
	}
	if cfg.Debug {
		logger.V(0).Info("enabled debug mode")
	}
	dmon := &daemon{
		client:     app,
		Config:     cfg,
		Logger:     logger,
		Downloader: releases.NewDownloader(cfg.TerraformBinDir),
		envs:       DefaultEnvs,
		spooler:    newSpooler(app, logger, cfg),
		terminator: newTerminator(),
	}
	if cfg.PluginCache {
		if err := os.MkdirAll(PluginCacheDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating plugin cache directory: %w", err)
		}
		dmon.envs = append(dmon.envs, "TF_PLUGIN_CACHE_DIR="+PluginCacheDir)
		logger.V(0).Info("enabled plugin cache", "path", PluginCacheDir)
	}
	return dmon, nil
}

// NewAgent constructs a remote operations daemon that communicates with the
// server via RPC. It is typically invoked as a separate process, `otf-agent`.
func NewAgent(ctx context.Context, logger logr.Logger, cfg AgentConfig) (*daemon, error) {
	// Sends unauthenticated ping to server
	app, err := newClient(cfg.APIConfig)
	if err != nil {
		return nil, err
	}

	// Confirm token validity
	at, err := app.GetAgentToken(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("attempted authentication: %w", err)
	}
	logger.Info("successfully authenticated", "organization", at.Organization, "token_id", at.ID)

	// Ensure agent only processes runs for this org
	cfg.Organization = internal.String(at.Organization)
	// This is an agent.
	cfg.isAgent = true

	return NewDaemon(logger, app, cfg.Config)
}

// Start starts the daemon and its workers
func (a *daemon) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		if err := a.spooler.start(ctx); err != nil {
			// only report error if context has not been canceled
			if ctx.Err() == nil {
				return fmt.Errorf("spooler terminated: %w", err)
			}
		}
		return nil
	})
	g.Go(func() error {
		for i := 0; i < a.Concurrency; i++ {
			w := &worker{a}
			go w.Start(ctx)
		}

		for {
			select {
			case cancelation := <-a.getCancelation():
				a.cancel(cancelation.Run.ID, cancelation.Forceful)
			case <-ctx.Done():
				return nil
			}
		}
	})
	return g.Wait()
}
