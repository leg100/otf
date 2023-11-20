package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/logr"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

const DefaultConcurrency = 5

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
	flags.IntVar(&cfg.Concurrency, "concurrency", DefaultConcurrency, "Number of runs that can be processed concurrently")
	flags.BoolVar(&cfg.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.BoolVar(&cfg.Debug, "debug", false, "Enable agent debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&cfg.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	return &cfg
}

// daemon implements the agent itself.
type daemon struct {
	logr.Logger
	client
	*terminator

	envs   []string // terraform environment variables
	config Config
}

// New constructs an agent daemon.
func New(logger logr.Logger, app client, cfg Config) (*daemon, error) {
	if _, ok := app.(*rpcClient); !ok {
		// agent is deemed a server agent if it is not using an RPC client.
		cfg.server = true
	}
	if cfg.Concurrency == 0 {
		cfg.Concurrency = DefaultConcurrency
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
	if cfg.server {
		// disable logging for server agents otherwise the server logs are
		// likely to contain duplicate logs from both the agent daemon and the
		// agent service.
		logger = logr.NewNoopLogger()
	}
	d := &daemon{
		Logger:     logger,
		client:     app,
		envs:       DefaultEnvs,
		terminator: &terminator{mapping: make(map[JobSpec]cancelable)},
		config:     cfg,
	}
	if cfg.PluginCache {
		if err := os.MkdirAll(PluginCacheDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating plugin cache directory: %w", err)
		}
		d.envs = append(d.envs, "TF_PLUGIN_CACHE_DIR="+PluginCacheDir)
		logger.V(0).Info("enabled plugin cache", "path", PluginCacheDir)
	}
	return d, nil
}

// NewRPC constructs a agent daemon that communicates with the server via RPC.
func NewRPC(logger logr.Logger, cfg Config, apiConfig otfapi.Config) (*daemon, error) {
	app, err := NewRPCClient(apiConfig)
	if err != nil {
		return nil, err
	}
	return New(logger, app, cfg)
}

// Start the agent daemon.
func (d *daemon) Start(ctx context.Context) error {
	d.Info("starting agent", "version", internal.Version)

	if d.config.server {
		// prior to registration, the server agent identifies itself as an
		// unregisteredServerAgent (the non-server agent identifies itself as an
		// unregisteredPoolAgent but the server-side token middleware handles
		// that).
		ctx = internal.AddSubjectToContext(ctx, &unregisteredServerAgent{})
	}

	// register agent with server
	agent, err := d.registerAgent(ctx, registerAgentOptions{
		Name:        d.config.Name,
		Concurrency: d.config.Concurrency,
	})
	if err != nil {
		return err
	}
	registeredKeyValues := []any{"agent_id", agent.ID}
	if agent.AgentPoolID != nil {
		registeredKeyValues = append(registeredKeyValues, "agent_pool_id", *agent.AgentPoolID)
	}
	d.Info("registered successfully", registeredKeyValues...)

	if d.config.server {
		// server agents should identify themselves as a serverAgent
		// (non-server agents identify themselves as a poolAgent, but the
		// bearer token middleware takes care of that server-side).
		ctx = internal.AddSubjectToContext(ctx, &serverAgent{Agent: agent})
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		// every 10 seconds update the agent status
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				// send agent status update
				status := AgentIdle
				if d.totalJobs() > 0 {
					status = AgentBusy
				}
				if err := d.updateAgentStatus(ctx, agent.ID, status); err != nil {
					if ctx.Err() != nil {
						goto finalupdate
					}
					d.Error(err, "sending agent status update", "status", status)
				}
				d.V(9).Info("sent agent status update", "status", status)
			case <-ctx.Done():
				goto finalupdate
			}
		}
	finalupdate:
		// send final status update using a context that is still valid
		// for a further 10 seconds.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		d.Info("sending final status update", "status", "exited")
		if err := d.updateAgentStatus(ctx, agent.ID, AgentExited); err != nil {
			return fmt.Errorf("sending final status update: %w", err)
		}
		return nil
	})

	// fetch jobs allocated to this agent and launch workers to do jobs; also
	// handle cancelation signals for jobs
	for {
		// block on waiting for jobs
		var jobs []*Job
		getJobs := func() (err error) {
			d.Info("waiting for next job")
			jobs, err = d.getAgentJobs(ctx, agent.ID)
			return err
		}
		policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
		err := backoff.RetryNotify(getJobs, policy, func(err error, next time.Duration) {
			d.Error(err, "waiting for next job", "backoff", next)
		})
		if err != nil {
			// ctx canceled
			break
		}
		for _, j := range jobs {
			if j.Status == JobAllocated {
				d.Info("received job", "job", j)
				w := &worker{
					Logger:     d.WithValues("job", j),
					client:     d.client,
					job:        j,
					envs:       d.envs,
					terminator: d.terminator,
				}
				g.Go(func() error {
					w.doAndHandleError(ctx)
					return nil
				})
			} else if j.signal != nil {
				d.Info("received signal", "signal", *j.signal, "job", j)
				switch *j.signal {
				case cancelSignal:
					d.cancel(j.JobSpec, false)
				case forceCancelSignal:
					d.cancel(j.JobSpec, true)
				default:
					d.Error(nil, "invalid signal received", "job", j, "signal", *j.signal)
				}
			}
		}
	}
	// TODO: exit cleanly when ctrl-c'd
	return g.Wait()
}
