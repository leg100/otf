package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/leg100/otf/internal/logr"
	"github.com/spf13/pflag"
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

	agentID string   // unique ID assigned by server
	envs    []string // terraform environment variables
	config  Config
}

// New constructs a new agent daemon.
func New(logger logr.Logger, app client, cfg Config) (*daemon, error) {
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
		terminator: &terminator{mapping: make(map[JobSpec]cancelable)},
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

func (d *daemon) Start(ctx context.Context) error {
	// register agent with server
	agent, err := d.registerAgent(ctx, registerAgentOptions{
		Name:        d.config.Name,
		Concurrency: d.config.Concurrency,
	})
	if err != nil {
		return err
	}
	d.agentID = agent.ID
	d.Logger = d.WithValues("agent_id", agent.ID)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				// send agent status update
				status := AgentIdle
				if d.totalJobs() > 0 {
					status = AgentBusy
				}
				if err := d.updateAgentStatus(ctx, d.agentID, status); err != nil {
					d.Error(err, "sending agent status update", "status", status)
				}
			case <-ctx.Done():
				// send exited status update
				if err := d.updateAgentStatus(ctx, d.agentID, AgentExited); err != nil {
					d.Error(err, "sending agent status update", "status", "exited")
				}
				return
			}
		}
	}()
	for {
		// block on waiting for jobs
		jobs, err := d.getAgentJobs(ctx, d.agentID)
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			break
		}
		for _, j := range jobs {
			if j.Status == JobAllocated {
				d.Info("received allocated job", "job", j)
				token, err := d.createJobToken(ctx, j.JobSpec)
				if err != nil {
					return err
				}
				if err := d.updateJobStatus(ctx, j.JobSpec, JobRunning); err != nil {
					d.Error(err, "sending job status", "status", JobRunning, "job", j)
					continue
				}
				w := &worker{
					Logger: d.Logger,
					client: d.client,
					job:    j,
					token:  token,
					envs:   d.envs,
				}
				// check worker in with the terminator so that it can receive a
				// cancelation signal should one arrive.
				d.checkIn(j.JobSpec, w)
				wg.Add(1)
				go func(j *Job) {
					defer wg.Done()
					defer d.checkOut(j.JobSpec)

					if err := w.do(ctx); err != nil {
						d.Error(err, "job returned an error", "job", j)
					}
					var status JobStatus
					switch {
					case w.canceled:
						status = JobCanceled
					case err != nil:
						status = JobErrored
					default:
						status = JobFinished
					}
					if err := d.updateJobStatus(ctx, j.JobSpec, status); err != nil {
						d.Error(err, "sending job status", "status", status, "job", j)
					}
				}(j)
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
	wg.Wait()
	return nil
}
