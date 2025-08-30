package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/fatih/color"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/dynamiccreds"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	runpkg "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
	"github.com/leg100/otf/internal/workspace"
)

const (
	localStateFilename = "terraform.tfstate"
	planFilename       = "plan.out"
	jsonPlanFilename   = "plan.out.json"
	lockFilename       = ".terraform.lock.hcl"
)

var (
	defaultEnvs = []string{
		"TF_IN_AUTOMATION=true",
		"CHECKPOINT_DISABLE=true",
	}
	ascii = regexp.MustCompile("[[:^ascii:]]")
)

type (
	// operation performs the execution of a job
	operation struct {
		logr.Logger
		*workdir

		Sandbox     bool // isolate privileged ops within sandbox
		Debug       bool // toggle debug mode
		PluginCache bool // toggle use of engine's shared plugin cache

		job           *Job
		run           *runpkg.Run
		canceled      bool
		ctx           context.Context
		cancelfn      context.CancelFunc
		out           io.Writer
		enginePath    string
		envs          []string
		terraformVars []*variable.Variable
		proc          *os.Process
		downloader    downloader
		isAgent       bool
		engineBinDir  string

		runs       runClient
		workspaces workspaceClient
		variables  variablesClient
		state      stateClient
		configs    configClient
		server     hostnameClient
		jobs       operationJobsClient
	}

	operationOptions struct {
		Sandbox      bool   // isolate privileged ops within sandbox
		Debug        bool   // toggle debug mode
		PluginCache  bool   // toggle use of engine's shared plugin cache
		engineBinDir string // destination directory for engine binaries

		logger   logr.Logger
		job      *Job
		jobToken []byte
		isAgent  bool

		runs       runClient
		workspaces workspaceClient
		variables  variablesClient
		state      stateClient
		configs    configClient
		server     hostnameClient
		jobs       operationJobsClient
	}

	operationSpawner interface {
		newOperation(job *Job, jobToken []byte) (*operation, error)
	}

	operationJobsClient interface {
		awaitJobSignal(ctx context.Context, jobID resource.TfeID) func() (jobSignal, error)
		finishJob(ctx context.Context, jobID resource.TfeID, opts finishJobOptions) error
		GenerateDynamicCredentialsToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error)
	}

	// downloader downloads engine versions
	downloader interface {
		Download(ctx context.Context, version string, w io.Writer) (string, error)
	}

	runClient interface {
		Get(ctx context.Context, runID resource.TfeID) (*runpkg.Run, error)
		GetPlanFile(ctx context.Context, id resource.TfeID, format runpkg.PlanFormat) ([]byte, error)
		UploadPlanFile(ctx context.Context, id resource.TfeID, plan []byte, format runpkg.PlanFormat) error
		GetLockFile(ctx context.Context, id resource.TfeID) ([]byte, error)
		UploadLockFile(ctx context.Context, id resource.TfeID, lockFile []byte) error
		PutChunk(ctx context.Context, opts runpkg.PutChunkOptions) error
	}

	workspaceClient interface {
		Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
	}

	variablesClient interface {
		ListEffectiveVariables(ctx context.Context, runID resource.TfeID) ([]*variable.Variable, error)
	}

	configClient interface {
		DownloadConfig(ctx context.Context, id resource.TfeID) ([]byte, error)
	}

	stateClient interface {
		Create(ctx context.Context, opts state.CreateStateVersionOptions) (*state.Version, error)
		DownloadCurrent(ctx context.Context, workspaceID resource.TfeID) ([]byte, error)
	}

	hostnameClient interface {
		Hostname() string
	}
)

func newOperation(opts operationOptions) *operation {
	// An operation has its own uninherited context; the operation is instead
	// canceled via its cancel() method, which provides more control, with the
	// ability to gracefully or forcefully cancel an operation.
	ctx, cancelfn := context.WithCancel(context.Background())
	// Authenticate as the job (only effective on server runner; the agent
	// runner instead authenticates remotely via its job token).
	ctx = authz.AddSubjectToContext(ctx, opts.job)

	envs := defaultEnvs
	if opts.PluginCache {
		envs = append(envs, "TF_PLUGIN_CACHE_DIR="+PluginCacheDir)
	}
	// make token available to engine CLI
	envs = append(envs, internal.CredentialEnv(opts.server.Hostname(), opts.jobToken))

	return &operation{
		Logger:       opts.logger.WithValues("job", opts.job),
		Sandbox:      opts.Sandbox,
		Debug:        opts.Debug,
		job:          opts.job,
		engineBinDir: opts.engineBinDir,
		envs:         envs,
		ctx:          ctx,
		cancelfn:     cancelfn,
		runs:         opts.runs,
		workspaces:   opts.workspaces,
		variables:    opts.variables,
		jobs:         opts.jobs,
		state:        opts.state,
		configs:      opts.configs,
		server:       opts.server,
		isAgent:      opts.isAgent,
	}
}

// doAndFinish executes the job and marks the job as complete with the
// appropriate status.
func (o *operation) doAndFinish() {
	// Whilst operation is being done relay any cancelation signals
	go func() {
		handleJobSignal := func() error {
			for {
				// blocks until signal received
				signal, err := o.jobs.awaitJobSignal(o.ctx, o.job.ID)()
				if err != nil {
					// If context has closed then the op has finished and we can
					// exit.
					if o.ctx.Err() != nil {
						return nil
					}
					return err
				}
				o.cancel(signal.Force, true)
			}
		}
		policy := backoff.WithContext(backoff.NewExponentialBackOff(), o.ctx)
		_ = backoff.RetryNotify(handleJobSignal, policy, func(err error, next time.Duration) {
			// An error is likely to do with proxies timing out long lived
			// connections like this one.
			o.V(8).Info("awaiting job signal", "error", err, "backoff", next)
		})
	}()
	// Upon finish cancel the context to ensure everything is cleaned up,
	// including stopping the job signaling go routine above.
	defer o.cancelfn()

	// do the job, and then handle any error and send appropriate job status
	// update
	err := o.do()

	var opts finishJobOptions
	switch {
	case o.canceled:
		if o.ctx.Err() != nil {
			// the context is closed, which only occurs when the server has
			// already canceled the job and the server has sent the operation a
			// force-cancel signal. In which case there is nothing more to be
			// done other than tell the user what happened.
			o.Error(err, "job forceably canceled")
			return
		}
		opts.Status = JobCanceled
		o.Error(err, "job canceled")
	case err != nil:
		opts.Status = JobErrored
		opts.Error = err.Error()
		o.Error(err, "finished job with error")
	default:
		opts.Status = JobFinished
		o.V(0).Info("finished job successfully")
	}
	if err := o.jobs.finishJob(o.ctx, o.job.ID, opts); err != nil {
		o.Error(err, "sending job status", "status", opts.Status)
	}
}

// do executes the job
func (o *operation) do() error {
	run, err := o.runs.Get(o.ctx, o.job.RunID)
	if err != nil {
		return err
	}
	o.run = run
	o.downloader = engine.NewDownloader(o.run.Engine, o.engineBinDir)

	// Get workspace in order to get working directory path
	//
	// TODO: add working directory to run.Run so we skip having to retrieve
	// workspace.
	ws, err := o.workspaces.Get(o.ctx, o.job.WorkspaceID)
	if err != nil {
		return fmt.Errorf("retreiving workspace: %w", err)
	}
	wd, err := newWorkdir(ws.WorkingDirectory, o.job.RunID.String())
	if err != nil {
		return fmt.Errorf("constructing working directory: %w", err)
	}
	defer func() {
		if err := wd.close(); err != nil {
			o.Error(err, "deleting files after job completion", "job", o.job, "path", wd)
		}
	}()
	o.workdir = wd
	writer := runpkg.NewPhaseWriter(o.ctx, runpkg.PhaseWriterOptions{
		RunID:  run.ID,
		Phase:  run.Phase(),
		Writer: o.runs,
	})
	defer writer.Close()
	o.out = writer

	// dump info if in debug mode
	if o.Debug {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		fmt.Fprintln(o.out)
		fmt.Fprintln(o.out, "Debug mode enabled")
		fmt.Fprintln(o.out, "------------------")
		fmt.Fprintf(o.out, "Hostname: %s\n", hostname)
		fmt.Fprintf(o.out, "External agent: %t\n", o.isAgent)
		fmt.Fprintf(o.out, "Sandbox mode: %t\n", o.Sandbox)
		fmt.Fprintln(o.out, "------------------")
		fmt.Fprintln(o.out)
	}

	// compile list of steps comprising operation
	type step func(context.Context) error
	steps := []step{
		o.downloadEngine,
		o.downloadConfig,
		o.readVars,
		o.setupDynamicCredentials,
		o.writeTerraformVars,
		o.deleteBackendConfig,
		o.downloadState,
	}
	switch run.Phase() {
	case runpkg.PlanPhase:
		steps = append(steps, o.init)
		steps = append(steps, o.plan)
		steps = append(steps, o.convertPlanToJSON)
		steps = append(steps, o.uploadPlan)
		steps = append(steps, o.uploadJSONPlan)
		steps = append(steps, o.uploadLockFile)
	case runpkg.ApplyPhase:
		// Download lock file from plan phase for the apply phase, to ensure
		// same providers are used in both phases.
		steps = append(steps, o.downloadLockFile)
		steps = append(steps, o.downloadPlanFile)
		steps = append(steps, o.init)
		steps = append(steps, o.apply)
	}

	// do each step
	for _, step := range steps {
		// skip remaining steps if op is canceled
		if o.canceled {
			return fmt.Errorf("execution canceled")
		}
		// do step
		if err := step(o.ctx); err != nil {
			// write error message to output
			errbuilder := strings.Builder{}
			errbuilder.WriteRune('\n')

			red := color.New(color.FgHiRed)
			red.EnableColor() // force color on non-tty output
			red.Fprint(&errbuilder, "Error: ")

			errbuilder.WriteString(err.Error())
			errbuilder.WriteRune('\n')
			fmt.Fprint(o.out, errbuilder.String())
			return err
		}
	}
	return nil
}

func (o *operation) cancel(force, sendSignal bool) {
	o.canceled = true
	// cancel context only if forced
	if force {
		o.cancelfn()
	}
	// signal current process if there is one.
	if sendSignal && o.proc != nil {
		if force {
			o.V(2).Info("sending SIGKILL to engine process", "pid", o.proc.Pid)
			o.proc.Signal(os.Kill)
		} else {
			o.V(2).Info("sending SIGINT to engine process", "pid", o.proc.Pid)
			o.proc.Signal(os.Interrupt)
		}
	}
}

type (
	// executionOptions are options that modify the execution of a process.
	executionOptions struct {
		sandboxIfEnabled bool
		redirectStdout   *string
	}

	executionOptionFunc func(*executionOptions)
)

// sandboxIfEnabled sandboxes the execution process *if* the daemon is configured
// with a sandbox.
func sandboxIfEnabled() executionOptionFunc {
	return func(e *executionOptions) {
		e.sandboxIfEnabled = true
	}
}

// redirectStdout redirects stdout to the destination path.
func redirectStdout(dst string) executionOptionFunc {
	return func(e *executionOptions) {
		e.redirectStdout = &dst
	}
}

// execute executes a process.
func (o *operation) execute(args []string, funcs ...executionOptionFunc) error {
	if len(args) == 0 {
		return fmt.Errorf("missing command name")
	}
	var opts executionOptions
	for _, fn := range funcs {
		fn(&opts)
	}
	if opts.sandboxIfEnabled && o.Sandbox {
		args = o.addSandboxWrapper(args)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = o.workdir.String()
	cmd.Env = os.Environ()
	cmd.Env = append(os.Environ(), o.envs...)

	if opts.redirectStdout != nil {
		dst, err := os.Create(path.Join(o.workdir.String(), *opts.redirectStdout))
		if err != nil {
			return err
		}
		defer dst.Close()
		cmd.Stdout = dst
	} else {
		cmd.Stdout = o.out
	}

	// send stderr to both output (for sending to client) and to
	// buffer, so that upon error its contents can be relayed.
	stderr := new(bytes.Buffer)
	cmd.Stderr = io.MultiWriter(o.out, stderr)

	if err := cmd.Start(); err != nil {
		return err
	}
	o.proc = cmd.Process

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%w: %s", err, cleanStderr(stderr.String()))
	}
	return nil
}

// addSandboxWrapper wraps the args within a bubblewrap sandbox.
func (o *operation) addSandboxWrapper(args []string) []string {
	bargs := []string{
		"bwrap",
		"--ro-bind", args[0], path.Join("/bin", path.Base(args[0])),
		"--bind", o.root, "/config",
		// for DNS lookups
		"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
		// for verifying SSL connections
		"--ro-bind", internal.SSLCertsDir(), internal.SSLCertsDir(),
		"--chdir", path.Join("/config", o.relative),
		// terraform v1.0.10 (but not v1.2.2) reads /proc/self/exe.
		"--proc", "/proc",
		// avoids provider error "failed to read schema..."
		"--tmpfs", "/tmp",
	}
	if o.PluginCache {
		bargs = append(bargs, "--ro-bind", PluginCacheDir, PluginCacheDir)
	}
	bargs = append(bargs, path.Join("/bin", path.Base(args[0])))
	return append(bargs, args[1:]...)
}

func (o *operation) downloadEngine(ctx context.Context) error {
	var err error
	o.enginePath, err = o.downloader.Download(ctx, o.run.EngineVersion, o.out)
	return err
}

func (o *operation) downloadConfig(ctx context.Context) error {
	cv, err := o.configs.DownloadConfig(ctx, o.run.ConfigurationVersionID)
	if err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}
	// Decompress and untar config into root dir
	if err := internal.Unpack(bytes.NewBuffer(cv), o.root); err != nil {
		return fmt.Errorf("unable to unpack config: %w", err)
	}
	return nil
}

func (o *operation) deleteBackendConfig(ctx context.Context) error {
	if err := internal.RewriteHCL(o.workdir.String(), internal.RemoveBackendBlock); err != nil {
		return fmt.Errorf("removing backend config: %w", err)
	}
	return nil
}

// downloadState downloads current state to disk. If there is no state yet then
// nothing will be downloaded and no error will be reported.
func (o *operation) downloadState(ctx context.Context) error {
	statefile, err := o.state.DownloadCurrent(ctx, o.run.WorkspaceID)
	if errors.Is(err, internal.ErrResourceNotFound) {
		return nil
	} else if err != nil {
		return fmt.Errorf("downloading state version: %w", err)
	}
	if err := o.writeFile(localStateFilename, statefile); err != nil {
		return fmt.Errorf("saving state to local disk: %w", err)
	}
	return nil
}

// downloadLockFile downloads the .terraform.lock.hcl file into the working
// directory. If one has not been uploaded then this will simply write an empty
// file, which is harmless.
func (o *operation) downloadLockFile(ctx context.Context) error {
	lockFile, err := o.runs.GetLockFile(ctx, o.run.ID)
	if err != nil {
		return err
	}
	return o.writeFile(lockFilename, lockFile)
}

// readVars retrieves terraform and environment variables and adds them to the
// operation
func (o *operation) readVars(ctx context.Context) error {
	variables, err := o.variables.ListEffectiveVariables(o.ctx, o.run.ID)
	if err != nil {
		return fmt.Errorf("retrieving variables: %w", err)
	}
	for _, v := range variables {
		if v.Category == variable.CategoryEnv {
			ev := fmt.Sprintf("%s=%s", v.Key, v.Value)
			o.envs = append(o.envs, ev)
		}
		if v.Category == variable.CategoryTerraform {
			o.terraformVars = append(o.terraformVars, v)
		}
	}
	return nil
}

// writeTerraformVars writes out terraform variables to a local hcl file
func (o *operation) writeTerraformVars(ctx context.Context) error {
	if err := variable.WriteTerraformVars(o.workdir.String(), o.terraformVars); err != nil {
		return fmt.Errorf("writing terraform.fvars: %w", err)
	}
	return nil
}

func (o *operation) setupDynamicCredentials(ctx context.Context) error {
	envs, err := dynamiccreds.Setup(
		ctx,
		o.jobs,
		o.workdir.String(),
		o.job.ID,
		o.job.Phase,
		o.envs,
	)
	if err != nil {
		return err
	}
	o.envs = append(o.envs, envs...)
	return nil
}

func (o *operation) init(ctx context.Context) error {
	return o.execute([]string{o.enginePath, "init", "-input=false"})
}

func (o *operation) plan(ctx context.Context) error {
	args := []string{"plan", "-input=false"}
	if o.run.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+planFilename)
	return o.execute(append([]string{o.enginePath}, args...))
}

func (o *operation) apply(ctx context.Context) (err error) {
	// prior to running an apply, capture info about local state file
	// so we can detect changes...
	statePath := filepath.Join(o.workdir.String(), localStateFilename)
	stateInfoBefore, _ := os.Stat(statePath)
	// ...and after the apply finishes, determine if there were changes, and if
	// so, create a new state version. We do this even if the apply failed
	// because since terraform v1.5, an apply can persist partial updates:
	//
	// https://github.com/hashicorp/terraform/pull/32680
	defer func() {
		stateInfoAfter, _ := os.Stat(statePath)
		if stateInfoAfter == nil {
			// no state file found
			return
		}
		if stateInfoBefore != nil && stateInfoAfter.ModTime().Equal(stateInfoBefore.ModTime()) {
			// no change to state file
			return
		}
		// either there was no state file before and there is one now, or the
		// state file modification time has changed. In either case we upload
		// the new state.
		if stateErr := o.uploadState(ctx); stateErr != nil {
			err = errors.Join(err, stateErr)
		}
	}()

	args := []string{"apply", "-input=false"}
	if o.run.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, planFilename)
	return o.execute(append([]string{o.enginePath}, args...), sandboxIfEnabled())
}

func (o *operation) convertPlanToJSON(ctx context.Context) error {
	args := []string{"show", "-json", planFilename}
	return o.execute(
		append([]string{o.enginePath}, args...),
		redirectStdout(jsonPlanFilename),
	)
}

func (o *operation) uploadPlan(ctx context.Context) error {
	file, err := o.readFile(planFilename)
	if err != nil {
		return err
	}

	if err := o.runs.UploadPlanFile(ctx, o.run.ID, file, runpkg.PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (o *operation) uploadJSONPlan(ctx context.Context) error {
	jsonFile, err := o.readFile(jsonPlanFilename)
	if err != nil {
		return err
	}
	if err := o.runs.UploadPlanFile(ctx, o.run.ID, jsonFile, runpkg.PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}
	return nil
}

func (o *operation) uploadLockFile(ctx context.Context) error {
	lockFile, err := o.readFile(lockFilename)
	if errors.Is(err, fs.ErrNotExist) {
		// there is no lock file to upload, which is ok
		return nil
	} else if err != nil {
		return fmt.Errorf("reading lock file: %w", err)
	}
	if err := o.runs.UploadLockFile(ctx, o.run.ID, lockFile); err != nil {
		return fmt.Errorf("unable to upload lock file: %w", err)
	}
	return nil
}

func (o *operation) downloadPlanFile(ctx context.Context) error {
	plan, err := o.runs.GetPlanFile(ctx, o.run.ID, runpkg.PlanFormatBinary)
	if err != nil {
		return err
	}

	return o.writeFile(planFilename, plan)
}

// uploadState reads, parses, and uploads state
func (o *operation) uploadState(ctx context.Context) error {
	statefile, err := o.readFile(localStateFilename)
	if err != nil {
		return err
	}
	// extract serial from state file
	var f state.File
	if err := json.Unmarshal(statefile, &f); err != nil {
		return err
	}
	_, err = o.state.Create(ctx, state.CreateStateVersionOptions{
		WorkspaceID: o.run.WorkspaceID,
		State:       statefile,
		Serial:      &f.Serial,
	})
	return err
}

// cleanStderr cleans up stderr output to make it suitable for logging:
// newlines, ansi escape sequences, and non-ascii characters are removed
func cleanStderr(stderr string) string {
	stderr = internal.StripAnsi(stderr)
	stderr = ascii.ReplaceAllLiteralString(stderr, "")
	stderr = strings.Join(strings.Fields(stderr), " ")
	return stderr
}
