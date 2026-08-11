package integration

import (
	"fmt"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
	goexpect "github.com/google/goexpect"
	"golang.org/x/term"
)

// spawnPTY starts a command attached to a pseudo-terminal and returns an
// expecter for driving it, in place of goexpect's own spawners.
//
// goexpect.SpawnWithArgs allocates the terminal via github.com/google/goterm,
// which opens /dev/ptmx and then issues the TIOCSPTLCK and TIOCGPTN ioctls.
// Those ioctl numbers are Linux-only, so on macOS the call fails with ENOTTY
// ("inappropriate ioctl for device") and no engine CLI test can run. The
// terminal is therefore allocated here, portably, and handed to
// goexpect.SpawnGeneric, which leaves every Expect and Send call unchanged.
//
// Pass the command's environment via env rather than goexpect.SetEnv: that
// option dereferences a command which SpawnGeneric never creates, so it panics.
//
// The command is returned so that tests can signal it. GExpect.SendSignal is
// unavailable for the same reason: it is only implemented for the expecters that
// spawn the command themselves.
func spawnPTY(t *testing.T, args []string, env []string, timeout time.Duration, opts ...goexpect.Option) (*goexpect.GExpect, *exec.Cmd, <-chan error, error) {
	t.Helper()

	ptmx, tty, err := pty.Open()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("opening pseudo-terminal: %w", err)
	}
	// Put the terminal into raw mode, as goexpect's spawners do, so that it
	// neither echoes what tests send nor translates newlines - the patterns
	// tests match on assume both.
	if _, err := term.MakeRaw(int(tty.Fd())); err != nil {
		tty.Close()
		ptmx.Close()
		return nil, nil, nil, fmt.Errorf("putting pseudo-terminal into raw mode: %w", err)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = tty, tty, tty
	// The command must lead its own session and own the terminal, so that it
	// treats it as interactive and receives terminal-generated signals.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true, Setctty: true}
	if err := cmd.Start(); err != nil {
		tty.Close()
		ptmx.Close()
		return nil, nil, nil, fmt.Errorf("starting %s: %w", args[0], err)
	}
	// Release the parent's handle on the slave; the command holds its own. This
	// matters: goexpect's waitForSession waits for its reader goroutine to finish
	// before reporting the command's exit status, and that reader only finishes
	// when reads of the master reach EOF, which requires every slave handle to be
	// closed. Holding one here makes reads block indefinitely on Linux once the
	// command exits, so nothing is ever sent on the returned error channel.
	tty.Close()

	e, errCh, err := goexpect.SpawnGeneric(&goexpect.GenOptions{
		In:   ptmx,
		Out:  ptmx,
		Wait: cmd.Wait,
		Close: func() error {
			err := cmd.Process.Kill()
			ptmx.Close()
			return err
		},
		Check: func() bool {
			if cmd.Process == nil {
				return false
			}
			// Signalling with 0 reports whether the process can still be
			// signalled, without sending anything.
			return cmd.Process.Signal(syscall.Signal(0)) == nil
		},
	}, timeout, opts...)
	return e, cmd, errCh, err
}
