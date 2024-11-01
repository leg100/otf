// Package runner contains the runner, the component responsible for carrying
// out runs by executing terraform processes, either as part of the server
// or remotely via agents.
package runner

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/releases"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

const DefaultMaxJobs = 5

var PluginCacheDir = filepath.Join(os.TempDir(), "plugin-cache")

type (
	Runner struct {
		*RunnerMeta

		Sandbox         bool   // isolate privileged ops within sandbox
		Debug           bool   // toggle debug mode
		PluginCache     bool   // toggle use of terraform's shared plugin cache
		TerraformBinDir string // destination directory for terraform binaries

		client     client
		spawner    operationSpawner
		registered chan *RunnerMeta

		logger logr.Logger // logger that logs messages regardless of whether runner is a pool runner or not.
		v      int         // logger verbosity
	}

	Config struct {
		Name    string // descriptive name given to runner
		MaxJobs int    // number of jobs the runner can execute at any one time

		Sandbox         bool   // isolate privileged ops within sandbox
		Debug           bool   // toggle debug mode
		PluginCache     bool   // toggle use of terraform's shared plugin cache
		TerraformBinDir string // destination directory for terraform binaries
	}
)

func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	opts := Config{}
	flags.IntVar(&opts.MaxJobs, "concurrency", DefaultMaxJobs, "Number of runs that can be processed concurrently")
	flags.BoolVar(&opts.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.BoolVar(&opts.Debug, "debug", false, "Enable runner debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&opts.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	flags.StringVar(&opts.TerraformBinDir, "terraform-bins-dir", releases.DefaultTerraformBinDir, "Destination directory for terraform binary downloads.")
	return &opts
}

// newRunner constructs a runner.
func newRunner(
	logger logr.Logger,
	client client,
	spawner operationSpawner,
	isAgent bool,
	cfg Config,
) (*Runner, error) {
	if cfg.MaxJobs == 0 {
		cfg.MaxJobs = DefaultMaxJobs
	}
	r := &Runner{
		RunnerMeta: &RunnerMeta{
			Name:    cfg.Name,
			MaxJobs: cfg.MaxJobs,
		},
		client:     client,
		registered: make(chan *RunnerMeta),
		logger:     logger,
		spawner:    spawner,
	}
	if !isAgent {
		// Set a higher threshold for logging on server runner where the runner is
		// but one of several components and many of the actions that are logged
		// here are also logged on the service endpoints, resulting in duplicate
		// logs.
		r.v = 1
		// Distinguish log messages on server runner from other components.
		r.logger = logger.WithValues("component", "runner")
	}
	if cfg.Debug {
		r.logger.V(r.v).Info("enabled debug mode")
	}
	if cfg.Sandbox {
		if _, err := exec.LookPath("bwrap"); errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("sandbox mode requires bubblewrap: %w", err)
		}
		r.logger.V(r.v).Info("enabled sandbox mode")
	}
	if cfg.PluginCache {
		if err := os.MkdirAll(PluginCacheDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating plugin cache directory: %w", err)
		}
		r.logger.V(r.v).Info("enabled plugin cache", "path", PluginCacheDir)
	}
	return r, nil
}

// Start the runner daemon.
func (r *Runner) Start(ctx context.Context) error {
	r.logger.V(r.v).Info("starting runner", "version", internal.Version)

	// initialize terminator
	terminator := &terminator{mapping: make(map[JobSpec]cancelable)}

	// register runner with server, which responds with an updated runner
	// registrationMetadata, including a unique ID.
	registrationMetadata, err := r.client.register(ctx, registerOptions{
		Name:        r.Name,
		Version:     internal.Version,
		Concurrency: r.MaxJobs,
	})
	if err != nil {
		return fmt.Errorf("registering runner: %w", err)
	}
	r.logger.V(r.v).Info("registered successfully", "runner", registrationMetadata)
	// send registered runner to channel, letting caller know runner has
	// registered.
	go func() {
		r.registered <- registrationMetadata
	}()

	// Update metadata with metadata from server, which includes unique ID.
	r.RunnerMeta = registrationMetadata
	// Add metadata to the context in all calls, which is needed to authorize a
	// server runner with service endpoints.
	ctx = internal.AddSubjectToContext(ctx, registrationMetadata)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		defer func() {
			// send final status update using a context that is still valid
			// for a further 10 seconds unless daemon is forcefully shutdown.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if updateErr := r.client.updateStatus(ctx, registrationMetadata.ID, RunnerExited); updateErr != nil {
				err = fmt.Errorf("sending final status update: %w", updateErr)
			} else {
				r.logger.V(r.v).Info("sent final status update", "status", "exited")
			}
		}()

		// every 10 seconds update the runner status
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				// send runner status update
				status := RunnerIdle
				if terminator.totalJobs() > 0 {
					status = RunnerBusy
				}
				if err := r.client.updateStatus(ctx, registrationMetadata.ID, status); err != nil {
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
						r.logger.Error(err, "sending runner status update", "status", status)
					}
				} else {
					r.logger.V(9).Info("sent runner status update", "status", status)
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
				r.logger.Info("gracefully canceling in-progress jobs", "total", terminator.totalJobs())
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
				r.logger.V(r.v).Info("waiting for next job")
				// block on waiting for jobs
				jobs, err := r.client.getJobs(ctx, registrationMetadata.ID)
				if err != nil {
					return err
				}
				for _, j := range jobs {
					if j.Status == JobAllocated {
						r.logger.V(r.v).Info("received job", "job", j)
						// start job and receive job token in return
						token, err := r.client.startJob(ctx, j.Spec)
						if err != nil {
							if ctx.Err() != nil {
								return nil
							}
							r.logger.Error(err, "starting job and retrieving job token")
							continue
						}
						op, err := r.spawner.newOperation(j, token)
						if err != nil {
							r.logger.Error(err, "spawning job operation")
							continue
						}
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
						r.logger.V(r.v).Info("received cancelation signal", "force", *j.Signaled, "job", j)
						terminator.cancel(j.Spec, *j.Signaled, true)
					}
				}
				return nil
			}
			policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
			_ = backoff.RetryNotify(processJobs, policy, func(err error, next time.Duration) {
				r.logger.Error(err, "waiting for next job", "backoff", next)
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
func (r *Runner) Registered() <-chan *RunnerMeta {
	return r.registered
}
