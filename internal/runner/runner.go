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

	"github.com/cenkalti/backoff"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/releases"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
)

const DefaultConcurrency = 5

var PluginCacheDir = filepath.Join(os.TempDir(), "plugin-cache")

type (
	Runner struct {
		*runnerMeta

		Logger          logr.Logger
		Client          client
		Concurrency     int    // number of jobs the runner can execute at any one time
		Sandbox         bool   // isolate privileged ops within sandbox
		Debug           bool   // toggle debug mode
		PluginCache     bool   // toggle use of terraform's shared plugin cache
		TerraformBinDir string // destination directory for terraform binaries

		logger       logr.Logger // logger that logs messages regardless of whether runner is a pool runner or not.
		remoteLogger logr.Logger // logger that only logs messages if the runner is a pool runner.
		spawner      operationSpawner
		isRemote     bool
		registered   chan *Runner
	}

	// downloader downloads terraform versions
	downloader interface {
		Download(ctx context.Context, version string, w io.Writer) (string, error)
	}

	Options struct {
		Logger logr.Logger
		Client client

		Concurrency     int    // number of jobs the runner can execute at any one time
		Sandbox         bool   // isolate privileged ops within sandbox
		Debug           bool   // toggle debug mode
		PluginCache     bool   // toggle use of terraform's shared plugin cache
		TerraformBinDir string // destination directory for terraform binaries

		spawner  operationSpawner
		isRemote bool
	}

	DaemonOptions = Options
)

func NewDaemonOptionsFromFlags(flags *pflag.FlagSet) *DaemonOptions {
	opts := DaemonOptions{}
	flags.IntVar(&opts.Concurrency, "concurrency", DefaultConcurrency, "Number of runs that can be processed concurrently")
	flags.BoolVar(&opts.Sandbox, "sandbox", false, "Isolate terraform apply within sandbox for additional security")
	flags.BoolVar(&opts.Debug, "debug", false, "Enable agent debug mode which dumps additional info to terraform runs.")
	flags.BoolVar(&opts.PluginCache, "plugin-cache", false, "Enable shared plugin cache for terraform providers.")
	flags.StringVar(&opts.TerraformBinDir, "terraform-bins-dir", releases.DefaultTerraformBinDir, "Destination directory for terraform binary downloads.")
	return &opts
}

// newRunner constructs a runner.
func newRunner(opts Options) (*Runner, error) {
	remoteLogger := opts.Logger
	if !opts.isRemote {
		// disable logging for server runners otherwise the server logs are
		// likely to contain duplicate logs from both the runner daemon and the
		// runner service, but still make logger available to server runner when
		// it does need to log something.
		remoteLogger = logr.Discard()
	}
	if opts.Concurrency == 0 {
		opts.Concurrency = DefaultConcurrency
	}
	if opts.Debug {
		opts.Logger.V(0).Info("enabled debug mode")
	}
	if opts.Sandbox {
		if _, err := exec.LookPath("bwrap"); errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("sandbox mode requires bubblewrap: %w", err)
		}
		opts.Logger.V(0).Info("enabled sandbox mode")
	}
	d := &Runner{
		Client:       opts.Client,
		registered:   make(chan *Runner),
		remoteLogger: remoteLogger,
		logger:       opts.Logger,
		isRemote:     opts.isRemote,
		spawner:      opts.spawner,
	}
	if opts.PluginCache {
		if err := os.MkdirAll(PluginCacheDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating plugin cache directory: %w", err)
		}
		opts.Logger.V(0).Info("enabled plugin cache", "path", PluginCacheDir)
	}
	return d, nil
}

// Start the runner daemon.
func (r *Runner) Start(ctx context.Context) error {
	r.remoteLogger.Info("starting runner", "version", internal.Version)

	// initialize terminator
	terminator := &terminator{mapping: make(map[JobSpec]cancelable)}

	// register runner with server
	runner, err := r.Client.register(ctx, registerOptions{
		Name:        r.Name,
		Version:     internal.Version,
		Concurrency: r.Concurrency,
	})
	if err != nil {
		return fmt.Errorf("registering runner: %w", err)
	}
	r.remoteLogger.Info("registered successfully", "runner", runner)
	// send registered runner to channel, letting caller know runner has
	// registered.
	go func() {
		r.registered <- runner
	}()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		defer func() {
			// send final status update using a context that is still valid
			// for a further 10 seconds unless daemon is forcefully shutdown.
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if updateErr := r.Client.updateStatus(ctx, runner.ID, RunnerExited); updateErr != nil {
				err = fmt.Errorf("sending final status update: %w", updateErr)
			} else {
				r.logger.Info("sent final status update", "status", "exited")
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
				if err := r.Client.updateStatus(ctx, runner.ID, status); err != nil {
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
						r.remoteLogger.Error(err, "sending runner status update", "status", status)
					}
				} else {
					r.remoteLogger.V(9).Info("sent runner status update", "status", status)
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
				r.remoteLogger.Info("waiting for next job")
				// block on waiting for jobs
				jobs, err := r.Client.getJobs(ctx, runner.ID)
				if err != nil {
					return err
				}
				for _, j := range jobs {
					if j.Status == JobAllocated {
						r.remoteLogger.Info("received job", "job", j)
						// start job and receive job token in return
						token, err := r.Client.startJob(ctx, j.Spec)
						if err != nil {
							if ctx.Err() != nil {
								return nil
							}
							r.remoteLogger.Error(err, "starting job")
							continue
						}
						op, err := r.spawner.newOperation(j, token)
						if err != nil {
							r.logger.Error(err, "starting job")
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
						r.remoteLogger.Info("received cancelation signal", "force", *j.Signaled, "job", j)
						terminator.cancel(j.Spec, *j.Signaled, true)
					}
				}
				return nil
			}
			policy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
			_ = backoff.RetryNotify(processJobs, policy, func(err error, next time.Duration) {
				r.remoteLogger.Error(err, "waiting for next job", "backoff", next)
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
func (r *Runner) Registered() <-chan *Runner {
	return r.registered
}

func (a *Runner) String() string      { return a.ID }
func (a *Runner) IsSiteAdmin() bool   { return true }
func (a *Runner) IsOwner(string) bool { return true }

func (a *Runner) Organizations() []string {
	// a runner is not a member of any organizations (although its agent pool
	// is, if it has one).
	return nil
}

func (*Runner) CanAccessSite(action rbac.Action) bool {
	// runner cannot carry out site-level actions
	return false
}

func (*Runner) CanAccessTeam(rbac.Action, string) bool {
	// agent cannot carry out team-level actions
	return false
}
