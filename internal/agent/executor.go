package agent

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/leg100/otf/internal"
)

var ascii = regexp.MustCompile("[[:^ascii:]]")

type (
	// executor executes processes.
	executor struct {
		Config
		terraformPathFinder

		version string // terraform cli version
		out     io.Writer
		envs    []string
		workdir *workdir

		*execution // current or last execution of a process
	}

	// execution is an execution of a process.
	execution struct {
		Config

		out     io.Writer
		envs    []string
		workdir *workdir
		proc    *os.Process // current or last process

		// options
		redirectStdout   *string
		sandboxIfEnabled bool
	}

	executionOption func(*execution)
)

// sandboxIfEnabled sandboxes the execution process *if* the agent is configured
// with a sandbox.
func sandboxIfEnabled() executionOption {
	return func(e *execution) {
		e.sandboxIfEnabled = true
	}
}

// redirectStdout redirects stdout to the destination path.
func redirectStdout(dst string) executionOption {
	return func(e *execution) {
		e.redirectStdout = &dst
	}
}

// execute executes a process.
func (e *executor) execute(args []string, opts ...executionOption) error {
	exe := execution{
		Config:  e.Config,
		out:     e.out,
		envs:    e.envs,
		workdir: e.workdir,
	}
	for _, fn := range opts {
		fn(&exe)
	}
	e.execution = &exe
	if err := exe.execute(args); err != nil {
		return err
	}
	return nil
}

// executeTerraform executes a terraform process
func (e *executor) executeTerraform(args []string, opts ...executionOption) error {
	args = append([]string{e.TerraformPath(e.version)}, args...)
	return e.execute(args, opts...)
}

func (e *execution) execute(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing command name")
	}
	if e.sandboxIfEnabled && e.Sandbox {
		args = e.addSandboxWrapper(args)
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = e.workdir.String()
	cmd.Env = append(os.Environ(), e.envs...)

	if e.redirectStdout != nil {
		dst, err := os.Create(path.Join(e.workdir.String(), *e.redirectStdout))
		if err != nil {
			return err
		}
		defer dst.Close()
		cmd.Stdout = dst
	} else {
		cmd.Stdout = e.out
	}

	// send stderr to both output (for sending to client) and to
	// buffer, so that upon error its contents can be relayed.
	stderr := new(bytes.Buffer)
	cmd.Stderr = io.MultiWriter(e.out, stderr)

	if err := cmd.Start(); err != nil {
		return err
	}
	e.proc = cmd.Process

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%w: %s", err, cleanStderr(stderr.String()))
	}
	return nil
}

// cancel sends a termination signal to the current process
func (e *execution) cancel(force bool) {
	if e.proc != nil {
		if force {
			e.proc.Signal(os.Kill)
		} else {
			e.proc.Signal(os.Interrupt)
		}
	}
}

// addSandboxWrapper wraps the args within a bubblewrap sandbox.
func (e *execution) addSandboxWrapper(args []string) []string {
	bargs := []string{
		"bwrap",
		"--ro-bind", args[0], path.Join("/bin", path.Base(args[0])),
		"--bind", e.workdir.root, "/config",
		// for DNS lookups
		"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
		// for verifying SSL connections
		"--ro-bind", internal.SSLCertsDir(), internal.SSLCertsDir(),
		"--chdir", path.Join("/config", e.workdir.relative),
		// terraform v1.0.10 (but not v1.2.2) reads /proc/self/exe.
		"--proc", "/proc",
		// avoids provider error "failed to read schema..."
		"--tmpfs", "/tmp",
	}
	if e.PluginCache {
		bargs = append(bargs, "--ro-bind", PluginCacheDir, PluginCacheDir)
	}
	bargs = append(bargs, path.Join("/bin", path.Base(args[0])))
	return append(bargs, args[1:]...)
}

// cleanStderr cleans up stderr output to make it suitable for logging:
// newlines, ansi escape sequences, and non-ascii characters are removed
func cleanStderr(stderr string) string {
	stderr = stripAnsi(stderr)
	stderr = ascii.ReplaceAllLiteralString(stderr, "")
	stderr = strings.Join(strings.Fields(stderr), " ")
	return stderr
}
