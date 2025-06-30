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
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
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
		registered chan struct{}

		logger logr.Logger // logger that logs messages regardless of whether runner is a pool runner or not.
		v      int         // logger verbosity
	}
)

// newRunner constructs a runner.
func newRunner(
	logger logr.Logger,
	client client,
	spawner operationSpawner,
	isAgent bool,
	cfg Config,
) (*Runner, error) {
	r := &Runner{
		RunnerMeta: &RunnerMeta{
			Name:    cfg.Name,
			MaxJobs: cfg.MaxJobs,
		},
		client:     client,
		registered: make(chan struct{}),
		logger:     logger,
		spawner:    spawner,
	}
	if !isAgent {
		// Set a higher threshold for logging on server runner where the runner is
		// but one of several components and many of the actions that are logged
		// here are also logged on the service endpoints, resulting in duplicate
		// logs.
		r.v = 1
		// Distinguish log messages in server runner component from other
		// components.
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

	// Initialize tracker, to keep track of active jobs.
	tracker := &jobTracker{mapping: make(map[resource.TfeID]cancelable)}

	// Authenticate as unregistered runner with the registration endpoint. This
	// is only necessary for the server runner; the agent runner relies on
	// middleware to authenticate as an unregistered runner on the server.
	ctx = authz.AddSubjectToContext(ctx, &unregistered{})

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

	// Update metadata with metadata from server, which includes unique ID.
	r.RunnerMeta = registrationMetadata
	// Add metadata to the context in all calls, which is needed to authorize a
	// server runner with service endpoints.
	ctx = authz.AddSubjectToContext(ctx, registrationMetadata)

	// send notification to channel, letting caller know runner has registered.
	go func() {
		r.registered <- struct{}{}
	}()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		defer func() {
			// send final status update using a context that is still valid
			// for a further 10 seconds unless daemon is forcefully shutdown.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			r.logger.V(r.v).Info("sending final status update before shutting down")

			if updateErr := r.client.updateStatus(ctx, registrationMetadata.ID, RunnerExited); updateErr != nil {
				err = fmt.Errorf("sending final status update: %w", updateErr)
			}
			r.logger.V(r.v).Info("sent final status update")
		}()

		// every 10 seconds update the runner status
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				// send runner status update
				status := RunnerIdle
				if tracker.totalJobs() > 0 {
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
			if tracker.totalJobs() > 0 {
				r.logger.Info("gracefully canceling in-progress jobs", "total", tracker.totalJobs())
				// NOTE: The interrupt sent to the main process is also sent to
				// any forked terraform processes, which is what we want, but it
				// is also necessary to instruct operations to stop.
				tracker.stopAll()
			}
		}()

		// fetch jobs allocated to this runner and spawn operations to do jobs; also
		// handle cancelation signals for jobs
		for {
			processJobs := func() error {
				r.logger.V(r.v).Info("waiting for next job")
				// block on waiting for jobs
				jobs, err := r.client.awaitAllocatedJobs(ctx, registrationMetadata.ID)
				if err != nil {
					return err
				}
				for _, j := range jobs {
					if j.Status != JobAllocated {
						// Skip jobs in a state other than allocated.
						continue
					}
					r.logger.V(r.v).Info("received job", "job", j)
					// start job and receive job token in return
					token, err := r.client.startJob(ctx, j.ID)
					if err != nil {
						if ctx.Err() != nil {
							// context cancelled means process is shutting
							// down.
							return nil
						}
						return fmt.Errorf("starting job and retrieving job token: %w", err)
					}
					op, err := r.spawner.newOperation(j, token)
					if err != nil {
						return fmt.Errorf("spawning job operation: %w", err)
					}
					// check job operation in with the tracker, so that if the
					// runner shuts down then the operation is shutdown too.
					tracker.checkIn(j.ID, op)
					op.V(0).Info("started job")
					g.Go(func() error {
						op.doAndFinish()
						tracker.checkOut(op.job.ID)
						return nil
					})
				}
				return nil
			}
			policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
			_ = backoff.RetryNotify(processJobs, policy, func(err error, next time.Duration) {
				r.logger.Error(err, "processing jobs", "backoff", next)
			})
			// only stop retrying if context is canceled
			if ctx.Err() != nil {
				return nil
			}
		}
	})
	return g.Wait()
}

func (r *Runner) Started() <-chan struct{} {
	return r.registered
}
