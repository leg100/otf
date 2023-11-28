package agent

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

	"github.com/fatih/color"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/logs"
	"github.com/leg100/otf/internal/releases"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/variable"
)

const (
	localStateFilename = "terraform.tfstate"
	planFilename       = "plan.out"
	jsonPlanFilename   = "plan.out.json"
	lockFilename       = ".terraform.lock.hcl"
)

var ascii = regexp.MustCompile("[[:^ascii:]]")

// operation performs the execution of a job
type operation struct {
	client
	*run.Run
	logr.Logger

	config        Config
	job           *Job
	debug         bool
	canceled      bool
	ctx           context.Context
	cancelfn      context.CancelFunc
	out           io.Writer
	terraformPath string
	envs          []string
	variables     []*variable.Variable // terraform variables
	proc          *os.Process          // current or last process
	downloader    releases.Downloader
	token         []byte

	*workdir
}

type newOperationOptions struct {
	logger     logr.Logger
	client     client
	job        *Job
	downloader releases.Downloader
	envs       []string
	token      []byte
}

func newOperation(opts newOperationOptions) *operation {
	// an operation has its own uninherited context; the operation is instead
	// canceled via its cancel() method.
	ctx, cancelfn := context.WithCancel(context.Background())
	return &operation{
		Logger:     opts.logger.WithValues("job", opts.job),
		client:     opts.client,
		job:        opts.job,
		envs:       opts.envs,
		downloader: opts.downloader,
		token:      opts.token,
		ctx:        ctx,
		cancelfn:   cancelfn,
	}
}

// doAndFinish executes the job and marks the job as complete with the
// appropriate status.
func (o *operation) doAndFinish() {
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
	if err := o.finishJob(o.ctx, o.job.Spec, opts); err != nil {
		o.Error(err, "sending job status", "status", opts.Status)
	}
}

// do executes the job
func (o *operation) do() error {
	// if this is a pool agent using RPC to communicate with the server
	// then use a new ac for this job, configured to authenticate using the
	// job token and to retry requests upon encountering transient errors.
	if ac, ok := o.client.(*rpcClient); ok {
		jc, err := ac.NewJobClient(o.token, o.Logger)
		if err != nil {
			return fmt.Errorf("initializing job client: %w", err)
		}
		o.client = jc
	} else {
		// this is a server agent: directly authenticate as job with services
		o.ctx = internal.AddSubjectToContext(o.ctx, o.job)
	}

	// make token available to terraform CLI
	o.envs = append(o.envs, internal.CredentialEnv(o.Hostname(), o.token))

	run, err := o.GetRun(o.ctx, o.job.Spec.RunID)
	if err != nil {
		return err
	}
	o.Run = run

	// Get workspace in order to get working directory path
	//
	// TODO: add working directory to run.Run
	ws, err := o.GetWorkspace(o.ctx, o.job.WorkspaceID)
	if err != nil {
		return fmt.Errorf("retreiving workspace: %w", err)
	}
	wd, err := newWorkdir(ws.WorkingDirectory)
	if err != nil {
		return fmt.Errorf("constructing working directory: %w", err)
	}
	defer wd.close()
	o.workdir = wd
	// retrieve variables and add them to the environment
	variables, err := o.ListEffectiveVariables(o.ctx, run.ID)
	if err != nil {
		return fmt.Errorf("retrieving variables: %w", err)
	}
	// append variables that are environment variables to the list of
	// environment variables
	for _, v := range variables {
		if v.Category == variable.CategoryEnv {
			ev := fmt.Sprintf("%s=%s", v.Key, v.Value)
			o.envs = append(o.envs, ev)
		}
	}
	writer := logs.NewPhaseWriter(o.ctx, logs.PhaseWriterOptions{
		RunID:  run.ID,
		Phase:  run.Phase(),
		Writer: o,
	})
	defer writer.Close()
	o.out = writer

	// dump info if in debug mode
	if o.debug {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		fmt.Fprintln(o.out)
		fmt.Fprintln(o.out, "Debug mode enabled")
		fmt.Fprintln(o.out, "------------------")
		fmt.Fprintf(o.out, "Hostname: %s\n", hostname)
		fmt.Fprintf(o.out, "External agent: %t\n", !o.config.server)
		fmt.Fprintf(o.out, "Sandbox mode: %t\n", !o.config.Sandbox)
		fmt.Fprintln(o.out, "------------------")
		fmt.Fprintln(o.out)
	}

	// compile list of steps comprising operation
	type step func(context.Context) error
	steps := []step{
		o.downloadTerraform,
		o.downloadConfig,
		o.writeTerraformVars,
		o.deleteBackendConfig,
		o.downloadState,
	}
	switch run.Phase() {
	case internal.PlanPhase:
		steps = append(steps, o.terraformInit)
		steps = append(steps, o.terraformPlan)
		steps = append(steps, o.convertPlanToJSON)
		steps = append(steps, o.uploadPlan)
		steps = append(steps, o.uploadJSONPlan)
		steps = append(steps, o.uploadLockFile)
	case internal.ApplyPhase:
		// Download lock file from plan phase for the apply phase, to ensure
		// same providers are used in both phases.
		steps = append(steps, o.downloadLockFile)
		steps = append(steps, o.downloadPlanFile)
		steps = append(steps, o.terraformInit)
		steps = append(steps, o.terraformApply)
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

func (o *operation) cancel(force bool) {
	o.canceled = true
	// cancel context only if forced and if there is a context to cancel
	if force && o.cancelfn != nil {
		o.cancelfn()
	}
	// signal current process if there is one.
	if o.proc != nil {
		if force {
			o.Info("sending SIGKILL to terraform process", "pid", o.proc.Pid)
			o.proc.Signal(os.Kill)
		} else {
			o.Info("sending SIGINT to terraform process", "pid", o.proc.Pid)
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
	if opts.sandboxIfEnabled && o.config.Sandbox {
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
	if o.config.PluginCache {
		bargs = append(bargs, "--ro-bind", PluginCacheDir, PluginCacheDir)
	}
	bargs = append(bargs, path.Join("/bin", path.Base(args[0])))
	return append(bargs, args[1:]...)
}

func (o *operation) downloadTerraform(ctx context.Context) error {
	var err error
	o.terraformPath, err = o.downloader.Download(ctx, o.TerraformVersion, o.out)
	return err
}

func (o *operation) downloadConfig(ctx context.Context) error {
	cv, err := o.DownloadConfig(ctx, o.ConfigurationVersionID)
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
	statefile, err := o.DownloadCurrentState(ctx, o.WorkspaceID)
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
	lockFile, err := o.GetLockFile(ctx, o.ID)
	if err != nil {
		return err
	}
	return o.writeFile(lockFilename, lockFile)
}

func (o *operation) writeTerraformVars(ctx context.Context) error {
	if err := variable.WriteTerraformVars(o.workdir.String(), o.variables); err != nil {
		return fmt.Errorf("writing terraform.fvars: %w", err)
	}
	return nil
}

func (o *operation) terraformInit(ctx context.Context) error {
	return o.execute([]string{o.terraformPath, "init"})
}

func (o *operation) terraformPlan(ctx context.Context) error {
	args := []string{"plan"}
	if o.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+planFilename)
	return o.execute(append([]string{o.terraformPath}, args...))
}

func (o *operation) terraformApply(ctx context.Context) (err error) {
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

	args := []string{"apply"}
	if o.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, planFilename)
	return o.execute(append([]string{o.terraformPath}, args...), sandboxIfEnabled())
}

func (o *operation) convertPlanToJSON(ctx context.Context) error {
	args := []string{"show", "-json", planFilename}
	return o.execute(
		append([]string{o.terraformPath}, args...),
		redirectStdout(jsonPlanFilename),
	)
}

func (o *operation) uploadPlan(ctx context.Context) error {
	file, err := o.readFile(planFilename)
	if err != nil {
		return err
	}

	if err := o.UploadPlanFile(ctx, o.ID, file, run.PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (o *operation) uploadJSONPlan(ctx context.Context) error {
	jsonFile, err := o.readFile(jsonPlanFilename)
	if err != nil {
		return err
	}
	if err := o.UploadPlanFile(ctx, o.ID, jsonFile, run.PlanFormatJSON); err != nil {
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
	if err := o.UploadLockFile(ctx, o.ID, lockFile); err != nil {
		return fmt.Errorf("unable to upload lock file: %w", err)
	}
	return nil
}

func (o *operation) downloadPlanFile(ctx context.Context) error {
	plan, err := o.GetPlanFile(ctx, o.ID, run.PlanFormatBinary)
	if err != nil {
		return err
	}

	return o.writeFile(planFilename, plan)
}

// uploadState reads, parses, and uploads terraform state
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
	_, err = o.CreateStateVersion(ctx, state.CreateStateVersionOptions{
		WorkspaceID: &o.WorkspaceID,
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
