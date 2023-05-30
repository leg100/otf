/*
Package agent provides a daemon capable of running remote operations on behalf of a user.
*/
package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/client"
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

// agent processes runs.
type agent struct {
	Config
	client.Client
	logr.Logger

	spooler     // spools new run events
	*terminator // terminates runs
	Downloader  // terraform cli downloader

	envs []string // terraform environment variables
}

// NewAgent is the constructor for an agent
func NewAgent(logger logr.Logger, app client.Client, cfg Config) (*agent, error) {
	if cfg.Concurrency == 0 {
		cfg.Concurrency = DefaultConcurrency
	}
	if cfg.External && cfg.Organization == nil {
		return nil, fmt.Errorf("external agent requires organization to be specified")
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

	agent := &agent{
		Client:     app,
		Config:     cfg,
		Logger:     logger,
		envs:       DefaultEnvs,
		spooler:    newSpooler(app, logger, cfg),
		terminator: newTerminator(),
		Downloader: newTerraformDownloader(),
	}

	if cfg.PluginCache {
		if err := os.MkdirAll(PluginCacheDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating plugin cache directory: %w", err)
		}
		agent.envs = append(agent.envs, "TF_PLUGIN_CACHE_DIR="+PluginCacheDir)
		logger.V(0).Info("enabled plugin cache", "path", PluginCacheDir)
	}

	return agent, nil
}

// NewExternalAgent constructs an external agent, which communicates with otfd
// via http
func NewExternalAgent(ctx context.Context, logger logr.Logger, cfg ExternalConfig) (*agent, error) {
	// Sends unauthenticated ping to server
	app, err := client.New(cfg.HTTPConfig)
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
	// Mark agent as external.
	cfg.External = true

	return NewAgent(logger, app, cfg.Config)
}

// Start starts the agent daemon and its workers
func (a *agent) Start(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := a.spooler.start(ctx); err != nil {
			return fmt.Errorf("spooler terminated: %w", err)
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
