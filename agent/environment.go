package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/fatih/color"
	"github.com/go-logr/logr"
	"github.com/hashicorp/go-multierror"
	"github.com/leg100/otf"
	"github.com/pkg/errors"
)

// Environment is an implementation of an execution environment
var _ otf.Environment = (*Environment)(nil)

// Environment provides an execution environment for a run, providing a working
// directory, services, capturing logs etc.
type Environment struct {
	otf.Application

	logr.Logger

	// Current working directory
	path string

	// Whether cancelation has been triggered
	canceled bool

	// Cancel context func for currently running func
	cancel context.CancelFunc

	// Current process
	proc *os.Process

	// CLI process output is written to this
	out io.WriteCloser

	// Downloader for workers to download terraform cli on demand
	otf.Downloader
	// For looking up path to terraform cli
	Terraform
	// Terraform version to use for this environment
	version string

	envs []string

	// Environment context - should contain subject for authenticating to
	// services
	ctx context.Context

	Config
}

func NewEnvironment(
	ctx context.Context,
	logger logr.Logger,
	app otf.Application,
	run *otf.Run,
	envs []string,
	downloader otf.Downloader,
	cfg Config,
) (*Environment, error) {
	// create dedicated directory for environment
	root, err := os.MkdirTemp("", "otf-plan")
	if err != nil {
		return nil, err
	}

	ws, err := app.GetWorkspace(ctx, run.WorkspaceID())
	if err != nil {
		return nil, errors.Wrap(err, "retrieving workspace")
	}

	// create token for terraform for it to authenticate with the otf registry
	// when retrieving modules and providers
	//
	// TODO: TF_TOKEN_* environment variables are only supported from terraform
	// v1.20 onwards. We should set that as the min version for use with otf.
	session, err := app.CreateRegistrySession(ctx, ws.Organization())
	if err != nil {
		return nil, errors.Wrap(err, "creating registry session")
	}
	tokenEnvVar := fmt.Sprintf("%s=%s", otf.HostnameCredentialEnv(app.Hostname()), session.Token())
	envs = append(envs, tokenEnvVar)

	// retrieve workspace variables and add them to the environment
	variables, err := app.ListVariables(ctx, run.WorkspaceID())
	if err != nil {
		return nil, errors.Wrap(err, "retrieving workspace variables")
	}
	for _, v := range variables {
		if v.Category() == otf.CategoryEnv {
			ev := fmt.Sprintf("%s=%s", v.Key(), v.Value())
			envs = append(envs, ev)
		}
	}
	if err := writeTerraformVariables(root, variables); err != nil {
		return nil, errors.Wrap(err, "writing terraform variables")
	}

	// Create and store cancel func so func's context can be canceled
	ctx, cancel := context.WithCancel(ctx)

	return &Environment{
		Logger:      logger,
		Application: app,
		Downloader:  downloader,
		Terraform:   &TerraformPathFinder{},
		version:     ws.TerraformVersion(),
		out:         otf.NewJobWriter(ctx, app, logger, run),
		path:        root,
		envs:        envs,
		cancel:      cancel,
		ctx:         ctx,
		Config:      cfg,
	}, nil
}

func (e *Environment) Close() error {
	return os.RemoveAll(e.path)
}

// Execute executes a phase and regardless of whether it fails, it'll close the
// environment logs.
func (e *Environment) Execute(phase Doer) (err error) {
	var errors *multierror.Error

	if err := phase.Do(e); err != nil {
		errors = multierror.Append(errors, err)
	}

	// Mark the logs as fully uploaded
	if err := e.out.Close(); err != nil {
		errors = multierror.Append(errors, fmt.Errorf("closing logs: %w", err))
	}

	return errors.ErrorOrNil()
}

func (e *Environment) Path() string {
	return e.path
}

// Cancel terminates execution. Force controls whether termination is graceful
// or not. Performed on a best-effort basis - the func or process may have
// finished before they are cancelled, in which case only the next func or
// process will be stopped from executing.
func (e *Environment) Cancel(force bool) {
	e.canceled = true

	e.cancelCLI(force)
	e.cancelFunc(force)
}

// RunTerraform runs a terraform command in the environment
func (e *Environment) RunTerraform(cmd string, args ...string) error {
	// Dump info if in debug mode
	if e.Debug && (cmd == "plan" || cmd == "apply") {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		fmt.Fprintln(e.out)
		fmt.Fprintln(e.out, "Debug mode enabled")
		fmt.Fprintln(e.out, "------------------")
		fmt.Fprintf(e.out, "Hostname: %s\n", hostname)
		fmt.Fprintf(e.out, "External agent: %t\n", e.External)
		fmt.Fprintf(e.out, "Sandbox mode: %t\n", e.Sandbox)
		fmt.Fprintln(e.out, "------------------")
		fmt.Fprintln(e.out)
	}

	// optionally sandbox terraform apply using bubblewrap
	if e.Sandbox && cmd == "apply" {
		return e.RunCLI("bwrap", e.buildSandboxArgs(args)...)
	}

	return e.RunCLI(e.TerraformPath(), append([]string{cmd}, args...)...)
}

// RunCLI executes a CLI process in the executor.
func (e *Environment) RunCLI(name string, args ...string) error {
	if e.canceled {
		return fmt.Errorf("execution canceled")
	}

	logger := e.Logger.WithValues("name", name, "args", args, "path", e.path)

	cmd := exec.Command(name, args...)
	cmd.Dir = e.path
	cmd.Stdout = e.out
	cmd.Env = append(os.Environ(), e.envs...)

	// send stderr to both environment output (for sending to client) and to
	// local var so we can report on errors below
	stderr := new(bytes.Buffer)
	cmd.Stderr = io.MultiWriter(e.out, stderr)

	if err := cmd.Start(); err != nil {
		logger.Error(err, "starting command", "stderr", stderr.String())
		return err
	}
	// store process so that it can be canceled
	e.proc = cmd.Process

	logger.V(2).Info("running command")

	if err := cmd.Wait(); err != nil {
		logger.Error(err, "running command", "stderr", stderr.String())
		return err
	}

	return nil
}

// TerraformPath provides the path to the terraform bin
func (e *Environment) TerraformPath() string {
	return e.Terraform.TerraformPath(e.version)
}

// RunFunc invokes a func in the executor.
func (e *Environment) RunFunc(fn otf.EnvironmentFunc) error {
	if e.canceled {
		return fmt.Errorf("execution canceled")
	}

	if err := fn(e.ctx, e); err != nil {
		e.printRedErrorMessage(err)
		return err
	}
	return nil
}

func (e *Environment) Write(p []byte) (int, error) {
	return e.out.Write(p)
}

func (e *Environment) printRedErrorMessage(err error) {
	fmt.Fprintln(e.out)

	// Print "Error:" in bright red, overriding the behaviour to disable colors
	// on a non-tty output
	red := color.New(color.FgHiRed)
	red.EnableColor()
	red.Fprint(e.out, "Error: ")

	fmt.Fprint(e.out, err.Error())
	fmt.Fprintln(e.out)
}

func (e *Environment) cancelCLI(force bool) {
	if e.proc == nil {
		return
	}
	if force {
		e.proc.Signal(os.Kill)
	} else {
		e.proc.Signal(os.Interrupt)
	}
}

func (e *Environment) cancelFunc(force bool) {
	// Don't cancel a func's context unless forced
	if !force {
		return
	}
	if e.cancel == nil {
		return
	}
	e.cancel()
}

// buildBubblewrapArgs builds the args for running a terraform apply within a
// bubblewrap sandbox.
func (e *Environment) buildSandboxArgs(args []string) []string {
	bargs := []string{
		"--ro-bind", e.TerraformPath(), "/bin/terraform",
		"--bind", e.path, "/workspace",
		// for DNS lookups
		"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
		// for verifying SSL connections
		"--ro-bind", otf.SSLCertsDir(), otf.SSLCertsDir(),
		"--chdir", "/workspace",
		// terraform v1.0.10 (but not v1.2.2) reads /proc/self/exe.
		"--proc", "/proc",
		// avoids provider error "failed to read schema..."
		"--tmpfs", "/tmp",
	}
	if e.PluginCache {
		bargs = append(bargs, "--ro-bind", PluginCacheDir, PluginCacheDir)
	}
	bargs = append(bargs, "terraform", "apply")
	return append(bargs, args...)
}

type Doer interface {
	// TODO: environment is excessive; can we pass in something that exposes
	// fewer methods like an 'executor'?
	Do(otf.Environment) error
}

// writeTerraformVariables writes workspace variables to a file named
// terraform.tfvars located in the given path. If the file already exists it'll
// be appended to.
func writeTerraformVariables(dir string, vars []*otf.Variable) error {
	path := path.Join(dir, "terraform.tfvars")
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	var b strings.Builder
	// lazily start with a new line in case user uploaded terraform.tfvars with
	// content already
	b.WriteRune('\n')
	for _, v := range vars {
		if v.Category() == otf.CategoryTerraform {
			b.WriteString(v.Key())
			b.WriteString(" = ")
			if v.HCL() {
				b.WriteString(v.Value())
			} else {
				b.WriteString(`"`)
				b.WriteString(v.Value())
				b.WriteString(`"`)
			}
			b.WriteRune('\n')
		}
	}
	f.WriteString(b.String())

	return nil
}
