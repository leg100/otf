package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/leg100/otf/internal"
	otfapi "github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/releases"
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
	// Config is configuration for an runner daemon
	Config struct {
		Name            string // descriptive name for agent
		Concurrency     int    // number of jobs the runner can execute at any one time
		Sandbox         bool   // isolate privileged ops within sandbox
		Debug           bool   // toggle debug mode
		PluginCache     bool   // toggle use of terraform's shared plugin cache
		TerraformBinDir string // destination directory for terraform binaries
	}
)

// runner carries out jobs.
type runner struct {
	*client

	envs       []string // terraform environment variables
	config     Config
	downloader downloader
	registered chan *runner
	logger     logr.Logger // logger that logs messages regardless of whether runner is a pool runner or not.
	poolLogger logr.Logger // logger that only logs messages if the runner is a pool runner.

	isAgent bool
}

type Options struct {
	Logger logr.Logger
	Config Config
	client *client

	// whether runner is an agent (true) or a server (false).
	isAgent bool
}

// downloader downloads terraform versions
type downloader interface {
	Download(ctx context.Context, version string, w io.Writer) (string, error)
}

// newRunner constructs a runner.
func newRunner(opts Options) (*runner, error) {
	poolLogger := opts.Logger
	if !opts.isAgent {
		// disable logging for server runners otherwise the server logs are
		// likely to contain duplicate logs from both the runner daemon and the
		// runner service, but still make logger available to server runner when
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
	d := &runner{
		client:     opts.client,
		envs:       DefaultEnvs,
		downloader: releases.NewDownloader(opts.Config.TerraformBinDir),
		registered: make(chan *runner),
		config:     opts.Config,
		poolLogger: poolLogger,
		logger:     opts.Logger,
		isAgent:    opts.isAgent,
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

// NewPoolDaemon constructs a pool runner daemon that communicates with the otfd server via RPC.
func NewPoolDaemon(logger logr.Logger, cfg Config, apiConfig otfapi.Config) (*runner, error) {
	rpcClient, err := newRPCClient(apiConfig, nil)
	if err != nil {
		return nil, err
	}
	return newRunner(Options{
		Logger:  logger,
		Config:  cfg,
		client:  rpcClient,
		isAgent: true,
	})
}

// Start the runner daemon.
func (d *runner) Start(ctx context.Context) error {
	d.poolLogger.Info("starting runner", "version", internal.Version)

	// initialize terminator
	terminator := &terminator{mapping: make(map[JobSpec]cancelable)}

	if !d.isAgent {
		// prior to registration, the server runner identifies itself as an
		// unregisteredServerrunner (the pool runner identifies itself as an
		// unregisteredPoolrunner but the server-side token middleware handles
		// that).
		ctx = internal.AddSubjectToContext(ctx, &unregisteredServerrunner{})
	}

	// register runner with server
	runner, err := d.runners.registerrunner(ctx, registerrunnerOptions{
		Name:        d.config.Name,
		Version:     internal.Version,
		Concurrency: d.config.Concurrency,
	})
	if err != nil {
		return fmt.Errorf("registering runner: %w", err)
	}
	d.poolLogger.Info("registered successfully", "runner", runner)
	// send registered runner to channel, letting caller know runner has
	// registered.
	go func() {
		d.registered <- runner
	}()

	if !d.isAgent {
		// server runners should identify themselves as a serverrunner
		// (pool runners identify themselves as a poolrunner, but the
		// bearer token middleware takes care of that server-side).
		ctx = internal.AddSubjectToContext(ctx, &serverrunner{runner: runner})
	}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		defer func() {
			// send final status update using a context that is still valid
			// for a further 10 seconds unless daemon is forcefully shutdown.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if updateErr := d.runners.updaterunnerStatus(ctx, runner.ID, runnerExited); updateErr != nil {
				err = fmt.Errorf("sending final status update: %w", updateErr)
			} else {
				d.logger.Info("sent final status update", "status", "exited")
			}
		}()

		// every 10 seconds update the runner status
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				// send runner status update
				status := runnerIdle
				if terminator.totalJobs() > 0 {
					status = runnerBusy
				}
				if err := d.runners.updaterunnerStatus(ctx, runner.ID, status); err != nil {
					if ctx.Err() != nil {
						// context canceled
						return nil
					}
					if errors.Is(err, internal.ErrConflict) {
						// exit, compelling runner to re-register - this may
						// happen when the server has de-registered the runner,
						// which it may do when it hasn't heard from the runner
						// in a while and the runner only belatedly succeeds in
						// sending an update.
						return errors.New("runner status update failed due to conflict; runner needs to re-register")
					} else {
						d.poolLogger.Error(err, "sending runner status update", "status", status)
					}
				} else {
					d.poolLogger.V(9).Info("sent runner status update", "status", status)
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

		// fetch jobs allocated to this runner and launch workers to do jobs; also
		// handle cancelation signals for jobs
		for {
			processJobs := func() (err error) {
				d.poolLogger.Info("waiting for next job")
				// block on waiting for jobs
				jobs, err := d.runners.getrunnerJobs(ctx, runner.ID)
				if err != nil {
					return err
				}
				for _, j := range jobs {
					if j.Status == JobAllocated {
						d.poolLogger.Info("received job", "job", j)
						// start job and receive job token in return
						token, err := d.runners.startJob(ctx, j.Spec)
						if err != nil {
							if ctx.Err() != nil {
								return nil
							}
							d.poolLogger.Error(err, "starting job")
							continue
						}
						d.poolLogger.V(0).Info("started job")
						op := newOperation(operationOptions{
							logger:       d.poolLogger.WithValues("job", j),
							client:       d.client,
							config:       d.config,
							runnerID:     runner.ID,
							job:          j,
							downloader:   d.downloader,
							envs:         d.envs,
							jobToken:     token,
							isPoolrunner: d.isAgent,
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

// Registered returns the daemon's corresponding runner on a channel once it has
// successfully registered.
func (d *runner) Registered() <-chan *runner {
	return d.registered
}
