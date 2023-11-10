package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/tokens"
	"github.com/spf13/pflag"
)

const defaultConcurrency = 5

var (
	PluginCacheDir = filepath.Join(os.TempDir(), "plugin-cache")
	DefaultEnvs    = []string{
		"TF_IN_AUTOMATION=true",
		"CHECKPOINT_DISABLE=true",
	}
)

type (
	// Config is configuration for an agent daemon
	Config struct {
		Name            *string // descriptive name for agent
		Concurrency     int     // number of workers
		Sandbox         bool    // isolate privileged ops within sandbox
		Debug           bool    // toggle debug mode
		PluginCache     bool    // toggle use of terraform's shared plugin cache
		TerraformBinDir string  // destination directory for terraform binaries

		server bool // otfd (true) or otf-agent (false)
	}
)

func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	cfg := Config{}
	flags.IntVar(&cfg.Concurrency, "concurrency", defaultConcurrency, "Number of runs that can be processed concurrently")
	flags.BoolVar(&cfg.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.BoolVar(&cfg.Debug, "debug", false, "Enable agent debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&cfg.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	return &cfg
}

// daemon implements the agent itself.
type daemon struct {
	logr.Logger
	client

	agentID string   // unique ID assigned by server
	envs    []string // terraform environment variables
}

// New constructs a new agent daemon.
func New(ctx context.Context, logger logr.Logger, app client, cfg Config) (*daemon, error) {
	opts := registerAgentOptions{
		Name:        cfg.Name,
		Concurrency: cfg.Concurrency,
	}
	if cfg.Concurrency == 0 {
		opts.Concurrency = defaultConcurrency
	}
	if cfg.Debug {
		logger.V(0).Info("enabled debug mode")
	}
	if cfg.Sandbox {
		if _, err := exec.LookPath("bwrap"); errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("sandbox mode requires bubblewrap: %w", err)
		}
		logger.V(0).Info("enabled sandbox mode")
	}
	envs := DefaultEnvs
	if cfg.PluginCache {
		if err := os.MkdirAll(PluginCacheDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating plugin cache directory: %w", err)
		}
		envs = internal.SafeAppend(envs, "TF_PLUGIN_CACHE_DIR="+PluginCacheDir)
		logger.V(0).Info("enabled plugin cache", "path", PluginCacheDir)
	}
	// handle otf-agent specific behaviour
	if !cfg.server {
		// confirm token validity, and get pool ID from token
		at, err := app.GetAgentToken(ctx, "")
		if err != nil {
			return nil, fmt.Errorf("attempted authentication: %w", err)
		}
		opts.AgentPoolID = &at.AgentPoolID
		logger.Info("successfully authenticated", "organization", at.Organization, "token_id", at.ID)
	}
	// register agent with server
	agent, err := app.registerAgent(ctx, opts)
	if err != nil {
		return nil, err
	}
	// create daemon along with unique ID assigned by server
	return &daemon{
		client:  app,
		Logger:  logger,
		envs:    DefaultEnvs,
		agentID: agent.ID,
	}, nil
}

func (d *daemon) Start(ctx context.Context) error {
	for {
		// block on waiting for jobs
		jobs, err := d.getAgentJobs(ctx, d.agentID)
		if err != nil {
			return err
		}
		token, err := d.CreateRunToken(ctx, tokens.CreateRunTokenOptions{
			Organization: &run.Organization,
			RunID:        &run.ID,
		})
	}
}

// worker does a Job
type worker struct {
	client
	logr.Logger

	job *Job
}

// do the Job
func (w *worker) do(ctx context.Context) error {
	run, err := w.GetRun(ctx, w.job.RunID)
	if err != nil {
		return err
	}
	return nil
}
