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

// worker does a Job
type worker struct {
	client
	logr.Logger
	*run.Run
	releases.Downloader

	config        Config
	agent         *Agent
	job           *Job
	token         []byte
	debug         bool
	canceled      bool
	cancelfn      context.CancelFunc
	out           io.Writer
	terraformPath string
	envs          []string
	variables     []*variable.Variable // terraform variables
	proc          *os.Process          // current or last process

	*workdir
}

type step func(context.Context) error

// do the Job
func (w *worker) do(ctx context.Context) error {
	// allow parent to cancel op
	ctx, cancel := context.WithCancel(ctx)
	w.cancelfn = cancel

	// setup appropriate auth depending on whether agent is part of the server
	// or talking to server via RPC
	if w.config.server {
		// directly authenticate as job with services
		ctx = internal.AddSubjectToContext(ctx, w.job)
	} else {
		// set jwt token as bearer in RPC requests
		w.client = w.client.(*rpcClient).newClientWithToken(w.token)
	}

	// make token available to terraform cli
	w.envs = append(w.envs, internal.CredentialEnv(w.Hostname(), w.token))

	run, err := w.GetRun(ctx, w.job.RunID)
	if err != nil {
		return err
	}
	// Get workspace in order to get working directory path
	//
	// TODO: add working directory to run.Run
	ws, err := w.GetWorkspace(ctx, w.job.RunID)
	if err != nil {
		return err
	}
	wd, err := newWorkdir(ws.WorkingDirectory)
	if err != nil {
		return err
	}
	defer wd.close()
	w.workdir = wd
	// retrieve variables and add them to the environment
	variables, err := w.ListEffectiveVariables(ctx, run.ID)
	if err != nil {
		return fmt.Errorf("retrieving variables: %w", err)
	}
	// append variables that are environment variables to the list of
	// environment variables
	for _, v := range variables {
		if v.Category == variable.CategoryEnv {
			ev := fmt.Sprintf("%s=%s", v.Key, v.Value)
			w.envs = append(w.envs, ev)
		}
	}
	writer := logs.NewPhaseWriter(ctx, logs.PhaseWriterOptions{
		RunID:  run.ID,
		Phase:  run.Phase(),
		Writer: w,
	})
	defer writer.Close()
	w.out = writer

	// dump info if in debug mode
	if w.debug {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		fmt.Fprintln(w.out)
		fmt.Fprintln(w.out, "Debug mode enabled")
		fmt.Fprintln(w.out, "------------------")
		fmt.Fprintf(w.out, "Hostname: %s\n", hostname)
		fmt.Fprintf(w.out, "External agent: %t\n", !w.agent.Server)
		fmt.Fprintf(w.out, "Sandbox mode: %t\n", !w.config.Sandbox)
		fmt.Fprintln(w.out, "------------------")
		fmt.Fprintln(w.out)
	}

	// compile list of steps that comprise operation
	steps := []step{
		w.downloadTerraform,
		w.downloadConfig,
		w.writeTerraformVars,
		w.deleteBackendConfig,
		w.downloadState,
	}
	switch run.Phase() {
	case internal.PlanPhase:
		steps = append(steps, w.terraformInit)
		steps = append(steps, w.terraformPlan)
		steps = append(steps, w.convertPlanToJSON)
		steps = append(steps, w.uploadPlan)
		steps = append(steps, w.uploadJSONPlan)
		steps = append(steps, w.uploadLockFile)
	case internal.ApplyPhase:
		// Download lock file from plan phase for the apply phase, to ensure
		// same providers are used in both phases.
		steps = append(steps, w.downloadLockFile)
		steps = append(steps, w.downloadPlanFile)
		steps = append(steps, w.terraformInit)
		steps = append(steps, w.terraformApply)
	}

	// do each step
	for _, step := range steps {
		// skip remaining steps if op is canceled
		if w.canceled {
			return fmt.Errorf("execution canceled")
		}
		// do step
		if err := step(ctx); err != nil {
			// write error message to output
			errbuilder := strings.Builder{}
			errbuilder.WriteRune('\n')

			red := color.New(color.FgHiRed)
			red.EnableColor() // force color on non-tty output
			red.Fprint(&errbuilder, "Error: ")

			errbuilder.WriteString(err.Error())
			errbuilder.WriteRune('\n')
			fmt.Fprint(w.out, errbuilder.String())
			return err
		}
	}
	return nil
}

func (w *worker) cancel(force bool) {
	w.canceled = true
	// cancel context only if forced and if there is a context to cancel
	if force && w.cancelfn != nil {
		w.cancelfn()
	}
	// signal current process if there is one.
	if w.proc != nil {
		if force {
			w.proc.Signal(os.Kill)
		} else {
			w.proc.Signal(os.Interrupt)
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
func (w *worker) execute(args []string, funcs ...executionOptionFunc) error {
	if len(args) == 0 {
		return fmt.Errorf("missing command name")
	}
	var opts executionOptions
	for _, fn := range funcs {
		fn(&opts)
	}
	if opts.sandboxIfEnabled && w.config.Sandbox {
		args = w.addSandboxWrapper(args)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = w.workdir.String()
	cmd.Env = os.Environ()
	cmd.Env = append(os.Environ(), w.envs...)

	if opts.redirectStdout != nil {
		dst, err := os.Create(path.Join(w.workdir.String(), *opts.redirectStdout))
		if err != nil {
			return err
		}
		defer dst.Close()
		cmd.Stdout = dst
	} else {
		cmd.Stdout = w.out
	}

	// send stderr to both output (for sending to client) and to
	// buffer, so that upon error its contents can be relayed.
	stderr := new(bytes.Buffer)
	cmd.Stderr = io.MultiWriter(w.out, stderr)

	if err := cmd.Start(); err != nil {
		return err
	}
	w.proc = cmd.Process

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%w: %s", err, cleanStderr(stderr.String()))
	}
	return nil
}

// addSandboxWrapper wraps the args within a bubblewrap sandbox.
func (w *worker) addSandboxWrapper(args []string) []string {
	bargs := []string{
		"bwrap",
		"--ro-bind", args[0], path.Join("/bin", path.Base(args[0])),
		"--bind", w.root, "/config",
		// for DNS lookups
		"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
		// for verifying SSL connections
		"--ro-bind", internal.SSLCertsDir(), internal.SSLCertsDir(),
		"--chdir", path.Join("/config", w.relative),
		// terraform v1.0.10 (but not v1.2.2) reads /proc/self/exe.
		"--proc", "/proc",
		// avoids provider error "failed to read schema..."
		"--tmpfs", "/tmp",
	}
	if w.config.PluginCache {
		bargs = append(bargs, "--ro-bind", PluginCacheDir, PluginCacheDir)
	}
	bargs = append(bargs, path.Join("/bin", path.Base(args[0])))
	return append(bargs, args[1:]...)
}

func (b *worker) downloadTerraform(ctx context.Context) error {
	var err error
	b.terraformPath, err = b.Download(ctx, b.TerraformVersion, b.out)
	return err
}

func (b *worker) downloadConfig(ctx context.Context) error {
	cv, err := b.DownloadConfig(ctx, b.ConfigurationVersionID)
	if err != nil {
		return fmt.Errorf("unable to download config: %w", err)
	}
	// Decompress and untar config into root dir
	if err := internal.Unpack(bytes.NewBuffer(cv), b.root); err != nil {
		return fmt.Errorf("unable to unpack config: %w", err)
	}
	return nil
}

func (b *worker) deleteBackendConfig(ctx context.Context) error {
	if err := internal.RewriteHCL(b.workdir.String(), internal.RemoveBackendBlock); err != nil {
		return fmt.Errorf("removing backend config: %w", err)
	}
	return nil
}

// downloadState downloads current state to disk. If there is no state yet then
// nothing will be downloaded and no error will be reported.
func (b *worker) downloadState(ctx context.Context) error {
	statefile, err := b.DownloadCurrentState(ctx, b.WorkspaceID)
	if errors.Is(err, internal.ErrResourceNotFound) {
		return nil
	} else if err != nil {
		return fmt.Errorf("downloading state version: %w", err)
	}
	if err := b.writeFile(localStateFilename, statefile); err != nil {
		return fmt.Errorf("saving state to local disk: %w", err)
	}
	return nil
}

// downloadLockFile downloads the .terraform.lock.hcl file into the working
// directory. If one has not been uploaded then this will simply write an empty
// file, which is harmless.
func (b *worker) downloadLockFile(ctx context.Context) error {
	lockFile, err := b.GetLockFile(ctx, b.ID)
	if err != nil {
		return err
	}
	return b.writeFile(lockFilename, lockFile)
}

func (b *worker) writeTerraformVars(ctx context.Context) error {
	if err := variable.WriteTerraformVars(b.workdir.String(), b.variables); err != nil {
		return fmt.Errorf("writing terraform.fvars: %w", err)
	}
	return nil
}

func (b *worker) terraformInit(ctx context.Context) error {
	return b.execute([]string{b.terraformPath, "init"})
}

func (b *worker) terraformPlan(ctx context.Context) error {
	args := []string{"plan"}
	if b.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, "-out="+planFilename)
	return b.execute(append([]string{b.terraformPath}, args...))
}

func (b *worker) terraformApply(ctx context.Context) (err error) {
	// prior to running an apply, capture info about local state file
	// so we can detect changes...
	statePath := filepath.Join(b.workdir.String(), localStateFilename)
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
		if stateErr := b.uploadState(ctx); stateErr != nil {
			err = errors.Join(err, stateErr)
		}
	}()

	args := []string{"apply"}
	if b.IsDestroy {
		args = append(args, "-destroy")
	}
	args = append(args, planFilename)
	return b.execute(append([]string{b.terraformPath}, args...), sandboxIfEnabled())
}

func (b *worker) convertPlanToJSON(ctx context.Context) error {
	args := []string{"show", "-json", planFilename}
	return b.execute(
		append([]string{b.terraformPath}, args...),
		redirectStdout(jsonPlanFilename),
	)
}

func (b *worker) uploadPlan(ctx context.Context) error {
	file, err := b.readFile(planFilename)
	if err != nil {
		return err
	}

	if err := b.UploadPlanFile(ctx, b.ID, file, run.PlanFormatBinary); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

func (b *worker) uploadJSONPlan(ctx context.Context) error {
	jsonFile, err := b.readFile(jsonPlanFilename)
	if err != nil {
		return err
	}
	if err := b.UploadPlanFile(ctx, b.ID, jsonFile, run.PlanFormatJSON); err != nil {
		return fmt.Errorf("unable to upload JSON plan: %w", err)
	}
	return nil
}

func (b *worker) uploadLockFile(ctx context.Context) error {
	lockFile, err := b.readFile(lockFilename)
	if errors.Is(err, fs.ErrNotExist) {
		// there is no lock file to upload, which is ok
		return nil
	} else if err != nil {
		return fmt.Errorf("reading lock file: %w", err)
	}
	if err := b.UploadLockFile(ctx, b.ID, lockFile); err != nil {
		return fmt.Errorf("unable to upload lock file: %w", err)
	}
	return nil
}

func (b *worker) downloadPlanFile(ctx context.Context) error {
	plan, err := b.GetPlanFile(ctx, b.ID, run.PlanFormatBinary)
	if err != nil {
		return err
	}

	return b.writeFile(planFilename, plan)
}

// uploadState reads, parses, and uploads terraform state
func (b *worker) uploadState(ctx context.Context) error {
	statefile, err := b.readFile(localStateFilename)
	if err != nil {
		return err
	}
	// extract serial from state file
	var f state.File
	if err := json.Unmarshal(statefile, &f); err != nil {
		return err
	}
	_, err = b.CreateStateVersion(ctx, state.CreateStateVersionOptions{
		WorkspaceID: &b.WorkspaceID,
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
