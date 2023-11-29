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
	"github.com/leg100/otf/internal/releases"
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
		Concurrency     int     // number of jobs the agent can execute at any one time
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
	client

	envs       []string // terraform environment variables
	config     Config
	downloader releases.Downloader
	logger     logr.Logger // logger that logs messages regardless of whether agent is a pool agent or not.
	poolLogger logr.Logger // logger that only logs messages if the agent is a pool agent.
}

// New constructs an agent daemon.
func New(logger logr.Logger, app client, cfg Config) (*daemon, error) {
	if _, ok := app.(*rpcClient); !ok {
		// agent is deemed a server agent if it is not using an RPC client.
		cfg.server = true
	}
	poolLogger := logger
	if cfg.server {
		// disable logging for server agents otherwise the server logs are
		// likely to contain duplicate logs from both the agent daemon and the
		// agent service, but still make logger available to server agent when
		// it does need to log something.
		poolLogger = logr.NewNoopLogger()
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
	d := &daemon{
		client:     app,
		envs:       DefaultEnvs,
		downloader: releases.NewDownloader(cfg.TerraformBinDir),
		config:     cfg,
		poolLogger: poolLogger,
		logger:     logger,
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
	app, err := NewRPCClient(apiConfig, nil)
	if err != nil {
		return nil, err
	}
	return New(logger, app, cfg)
}

// Start the agent daemon.
func (d *daemon) Start(ctx context.Context) error {
	d.poolLogger.Info("starting agent", "version", internal.Version)

	// initialize terminator
	terminator := &terminator{mapping: make(map[JobSpec]cancelable)}

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
		Version:     internal.Version,
		Concurrency: d.config.Concurrency,
	})
	if err != nil {
		return err
	}
	registeredKeyValues := []any{"agent_id", agent.ID}
	if agent.AgentPoolID != nil {
		registeredKeyValues = append(registeredKeyValues, "agent_pool_id", *agent.AgentPoolID)
	}
	d.logger.Info("registered successfully", registeredKeyValues...)

	if d.config.server {
		// server agents should identify themselves as a serverAgent
		// (pool agents identify themselves as a poolAgent, but the
		// bearer token middleware takes care of that server-side).
		ctx = internal.AddSubjectToContext(ctx, &serverAgent{Agent: agent})
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		defer func() {
			// send final status update using a context that is still valid
			// for a further 10 seconds unless daemon is forcefully shutdown.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if updateErr := d.updateAgentStatus(ctx, agent.ID, AgentExited); updateErr != nil {
				err = fmt.Errorf("sending final status update: %w", updateErr)
			} else {
				d.logger.Info("sent final status update", "status", "exited")
			}
		}()

		// every 10 seconds update the agent status
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				// send agent status update
				status := AgentIdle
				if terminator.totalJobs() > 0 {
					status = AgentBusy
				}
				if err := d.updateAgentStatus(ctx, agent.ID, status); err != nil {
					if ctx.Err() != nil {
						// context canceled
						return nil
					}
					if errors.Is(err, internal.ErrConflict) {
						// exit, compelling agent to re-register - this may
						// happen when the server has de-registered the agent,
						// which it may do when it hasn't heard from the agent
						// in a while and the agent only belatedly succeeds in
						// sending an update.
						return errors.New("agent status update failed due to conflict; agent needs to re-register")
					} else {
						d.poolLogger.Error(err, "sending agent status update", "status", status)
					}
				} else {
					d.poolLogger.V(9).Info("sent agent status update", "status", status)
				}
			case <-ctx.Done():
				// context canceled
				return nil
			}
		}
	})

	g.Go(func() (err error) {
		defer func() {
			if terminator.totalJobs() > 0 {
				d.logger.Info("gracefully canceling in-progress jobs", "total", terminator.totalJobs())
				// The interrupt sent to the main process is also sent to the
				// forked terraform processes, so there is no need to send the
				// latter another interrupt but merely set the cancel semaphore
				// on each operation.
				terminator.stopAll()
			}
		}()

		// fetch jobs allocated to this agent and launch workers to do jobs; also
		// handle cancelation signals for jobs
		for {
			// block on waiting for jobs
			var jobs []*Job
			getJobs := func() (err error) {
				d.poolLogger.Info("waiting for next job")
				jobs, err = d.getAgentJobs(ctx, agent.ID)
				return err
			}
			policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
			_ = backoff.RetryNotify(getJobs, policy, func(err error, next time.Duration) {
				d.poolLogger.Error(err, "waiting for next job", "backoff", next)
			})
			// only stop retrying if context is canceled
			if ctx.Err() != nil {
				return nil
			}
			for _, j := range jobs {
				if j.Status == JobAllocated {
					d.poolLogger.Info("received job", "job", j)
					// start job and receive job token in return
					token, err := d.startJob(ctx, j.Spec)
					if err != nil {
						if ctx.Err() != nil {
							return nil
						}
						d.poolLogger.Error(err, "starting job")
						continue
					}
					d.poolLogger.V(0).Info("started job")
					op := newOperation(newOperationOptions{
						logger:     d.poolLogger.WithValues("job", j),
						client:     d.client,
						job:        j,
						downloader: d.downloader,
						envs:       d.envs,
						token:      token,
					})
					// check operation in with the terminator, so that if a cancelation signal
					// arrives it can be handled accordingly for the duration of the operation.
					terminator.checkIn(j.Spec, op)
					op.V(0).Info("started job")
					g.Go(func() error {
						op.doAndFinish()
						terminator.checkOut(op.job.Spec)
						return nil
					})
				} else if j.Signaled != nil {
					d.poolLogger.Info("received cancelation signal", "force", *j.Signaled, "job", j)
					terminator.cancel(j.Spec, *j.Signaled, true)
				}
			}
		}
	})
	return g.Wait()
}
