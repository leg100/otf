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
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/releases"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
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
		Name            string // descriptive name for agent
		Concurrency     int    // number of jobs the agent can execute at any one time
		Sandbox         bool   // isolate privileged ops within sandbox
		Debug           bool   // toggle debug mode
		PluginCache     bool   // toggle use of terraform's shared plugin cache
		TerraformBinDir string // destination directory for terraform binaries
	}
)

func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	cfg := Config{}
	flags.IntVar(&cfg.Concurrency, "concurrency", DefaultConcurrency, "Number of runs that can be processed concurrently")
	flags.BoolVar(&cfg.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.BoolVar(&cfg.Debug, "debug", false, "Enable agent debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&cfg.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	flags.StringVar(&cfg.Name, "name", "", "Give agent a descriptive name. Optional.")
	return &cfg
}

// daemon implements the agent itself.
type daemon struct {
	*daemonClient

	envs       []string // terraform environment variables
	config     Config
	downloader releases.Downloader
	registered chan *Agent
	logger     logr.Logger // logger that logs messages regardless of whether agent is a pool agent or not.
	poolLogger logr.Logger // logger that only logs messages if the agent is a pool agent.

	isPoolAgent bool
}

type DaemonOptions struct {
	Logger logr.Logger
	Config Config
	client *daemonClient

	// whether daemon is for a pool agent (true) or for a server agent (false).
	isPoolAgent bool
}

// newDaemon constructs an agent daemon.
func newDaemon(opts DaemonOptions) (*daemon, error) {
	poolLogger := opts.Logger
	if !opts.isPoolAgent {
		// disable logging for server agents otherwise the server logs are
		// likely to contain duplicate logs from both the agent daemon and the
		// agent service, but still make logger available to server agent when
		// it does need to log something.
		poolLogger = logr.NewNoopLogger()
	}
	if opts.Config.Concurrency == 0 {
		opts.Config.Concurrency = DefaultConcurrency
	}
	if opts.Config.Debug {
		opts.Logger.V(0).Info("enabled debug mode")
	}
	if opts.Config.Sandbox {
		if _, err := exec.LookPath("bwrap"); errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("sandbox mode requires bubblewrap: %w", err)
		}
		opts.Logger.V(0).Info("enabled sandbox mode")
	}
	d := &daemon{
		envs:        DefaultEnvs,
		downloader:  releases.NewDownloader(opts.Config.TerraformBinDir),
		registered:  make(chan *Agent),
		config:      opts.Config,
		poolLogger:  poolLogger,
		logger:      opts.Logger,
		isPoolAgent: opts.isPoolAgent,
	}
	if opts.Config.PluginCache {
		if err := os.MkdirAll(PluginCacheDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating plugin cache directory: %w", err)
		}
		d.envs = append(d.envs, "TF_PLUGIN_CACHE_DIR="+PluginCacheDir)
		opts.Logger.V(0).Info("enabled plugin cache", "path", PluginCacheDir)
	}
	return d, nil
}

type ServerDaemonOptions struct {
	Logger                      logr.Logger
	Config                      Config
	RunService                  *run.Service
	WorkspaceService            *workspace.Service
	VariableService             *variable.Service
	ConfigurationVersionService configversion.Service
	StateService                *state.Service
	LogsService                 *logs.Service
	AgentService                Service
	HostnameService             internal.HostnameService
}

// NewServerDaemon constructs a server agent daemon that is part of the otfd
// server.
func NewServerDaemon(logger logr.Logger, cfg Config, opts ServerDaemonOptions) (*daemon, error) {
	return newDaemon(DaemonOptions{
		Logger: logger,
		Config: cfg,
		client: &daemonClient{
			runs:       opts.RunService,
			workspaces: opts.WorkspaceService,
			state:      opts.StateService,
			variables:  opts.VariableService,
			configs:    opts.ConfigurationVersionService,
			logs:       opts.LogsService,
			agents:     opts.AgentService,
			server:     opts.HostnameService,
		},
	})
}

// NewPoolDaemon constructs a pool agent daemon that communicates with the otfd server via RPC.
func NewPoolDaemon(logger logr.Logger, cfg Config, apiConfig otfapi.Config) (*daemon, error) {
	rpcClient, err := newRPCDaemonClient(apiConfig, nil)
	if err != nil {
		return nil, err
	}
	return newDaemon(DaemonOptions{
		Logger:      logger,
		Config:      cfg,
		client:      rpcClient,
		isPoolAgent: true,
	})
}

// Start the agent daemon.
func (d *daemon) Start(ctx context.Context) error {
	d.poolLogger.Info("starting agent", "version", internal.Version)

	// initialize terminator
	terminator := &terminator{mapping: make(map[JobSpec]cancelable)}

	if !d.isPoolAgent {
		// prior to registration, the server agent identifies itself as an
		// unregisteredServerAgent (the pool agent identifies itself as an
		// unregisteredPoolAgent but the server-side token middleware handles
		// that).
		ctx = internal.AddSubjectToContext(ctx, &unregisteredServerAgent{})
	}

	// register agent with server
	agent, err := d.agents.registerAgent(ctx, registerAgentOptions{
		Name:        d.config.Name,
		Version:     internal.Version,
		Concurrency: d.config.Concurrency,
	})
	if err != nil {
		return err
	}
	d.poolLogger.Info("registered successfully", "agent", agent)
	// send registered agent to channel, letting caller know agent has
	// registered.
	go func() {
		d.registered <- agent
	}()

	if !d.isPoolAgent {
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

			if updateErr := d.agents.updateAgentStatus(ctx, agent.ID, AgentExited); updateErr != nil {
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
				if err := d.agents.updateAgentStatus(ctx, agent.ID, status); err != nil {
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
			processJobs := func() (err error) {
				d.poolLogger.Info("waiting for next job")
				// block on waiting for jobs
				jobs, err := d.agents.getAgentJobs(ctx, agent.ID)
				if err != nil {
					return err
				}
				for _, j := range jobs {
					if j.Status == JobAllocated {
						d.poolLogger.Info("received job", "job", j)
						// start job and receive job token in return
						token, err := d.agents.startJob(ctx, j.Spec)
						if err != nil {
							if ctx.Err() != nil {
								return nil
							}
							d.poolLogger.Error(err, "starting job")
							continue
						}
						d.poolLogger.V(0).Info("started job")
						op := newOperation(newOperationOptions{
							logger:      d.poolLogger.WithValues("job", j),
							client:      d.daemonClient,
							config:      d.config,
							agentID:     agent.ID,
							job:         j,
							downloader:  d.downloader,
							envs:        d.envs,
							token:       token,
							isPoolAgent: d.isPoolAgent,
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
				return nil
			}
			policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
			_ = backoff.RetryNotify(processJobs, policy, func(err error, next time.Duration) {
				d.poolLogger.Error(err, "waiting for next job", "backoff", next)
			})
			// only stop retrying if context is canceled
			if ctx.Err() != nil {
				return nil
			}
		}
	})
	return g.Wait()
}

// Registered returns the daemon's corresponding agent on a channel once it has
// successfully registered.
func (d *daemon) Registered() <-chan *Agent {
	return d.registered
}
