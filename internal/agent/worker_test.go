package agent

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/mitchellh/iochan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_execute(t *testing.T) {
	t.Run("no options", func(t *testing.T) {
		var got bytes.Buffer
		w := &operation{
			out:     &got,
			workdir: &workdir{root: ""},
		}
		err := w.execute([]string{"./testdata/exe"})
		require.NoError(t, err)
		assert.Equal(t, "some output\n", got.String())
	})

	t.Run("redirect stdout", func(t *testing.T) {
		w := &operation{
			out:     io.Discard,
			workdir: &workdir{root: ""},
		}

		dst := path.Join(t.TempDir(), "dst")
		err := w.execute([]string{"./testdata/exe"}, redirectStdout(dst))
		require.NoError(t, err)

		got, err := os.ReadFile(dst)
		require.NoError(t, err)
		assert.Equal(t, "some output\n", string(got))
	})

	t.Run("sandbox", func(t *testing.T) {
		if _, err := exec.LookPath("bwrap"); err != nil {
			t.Skip("Skipping test that requires bwrap")
		}

		w := &operation{
			config:  Config{Sandbox: true},
			out:     io.Discard,
			workdir: &workdir{root: "."},
		}

		err := w.execute([]string{"./testdata/staticbin"}, sandboxIfEnabled())
		require.NoError(t, err)
	})

	t.Run("stderr", func(t *testing.T) {
		var got bytes.Buffer
		w := &operation{
			out:     &got,
			workdir: &workdir{root: ""},
		}
		err := w.execute([]string{"./testdata/badexe"})
		if assert.Error(t, err) {
			assert.Equal(t, "exit status 1: an error", err.Error())
		}
	})

	t.Run("cancel", func(t *testing.T) {
		r, w := io.Pipe()
		wkr := &operation{
			out:     w,
			workdir: &workdir{root: ""},
		}
		done := make(chan error)
		go func() {
			done <- wkr.execute([]string{"./testdata/killme"})
		}()

		assert.Equal(t, "ok, you can kill me now\n", <-iochan.DelimReader(r, '\n'))
		wkr.cancel(false)
		assert.NoError(t, <-done)
	})
}

func TestExecutor_addSandboxWrapper(t *testing.T) {
	t.Run("without plugin cache", func(t *testing.T) {
		w := operation{
			workdir: &workdir{root: "/root"},
		}
		want := []string{
			"bwrap",
			"--ro-bind", "/tmp/tf-bins/1.1.1/terraform", "/bin/terraform",
			"--bind", "/root", "/config",
			"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
			"--ro-bind", "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt",
			"--chdir", "/config",
			"--proc", "/proc",
			"--tmpfs", "/tmp",
			"/bin/terraform", "apply",
			"-input=false", "-no-color",
		}
		assert.Equal(t, want, w.addSandboxWrapper([]string{"/tmp/tf-bins/1.1.1/terraform", "apply", "-input=false", "-no-color"}))
	})

	t.Run("with plugin cache", func(t *testing.T) {
		w := operation{
			config: Config{
				PluginCache: true,
			},
			workdir: &workdir{root: "/root"},
		}
		want := []string{
			"bwrap",
			"--ro-bind", "/tmp/tf-bins/1.1.1/terraform", "/bin/terraform",
			"--bind", "/root", "/config",
			"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
			"--ro-bind", "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt",
			"--chdir", "/config",
			"--proc", "/proc",
			"--tmpfs", "/tmp",
			"--ro-bind", PluginCacheDir, PluginCacheDir,
			"/bin/terraform", "apply",
			"-input=false", "-no-color",
		}
		assert.Equal(t, want, w.addSandboxWrapper([]string{"/tmp/tf-bins/1.1.1/terraform", "apply", "-input=false", "-no-color"}))
	})

	t.Run("with relative working directory", func(t *testing.T) {
		w := operation{
			config: Config{
				PluginCache: true,
			},
			workdir: &workdir{root: "/root", relative: "/relative"},
		}
		want := []string{
			"bwrap",
			"--ro-bind", "/tmp/tf-bins/1.1.1/terraform", "/bin/terraform",
			"--bind", "/root", "/config",
			"--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf",
			"--ro-bind", "/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/certs/ca-certificates.crt",
			"--chdir", "/config/relative",
			"--proc", "/proc",
			"--tmpfs", "/tmp",
			"--ro-bind", PluginCacheDir, PluginCacheDir,
			"/bin/terraform", "apply",
			"-input=false", "-no-color",
		}
		assert.Equal(t, want, w.addSandboxWrapper([]string{"/tmp/tf-bins/1.1.1/terraform", "apply", "-input=false", "-no-color"}))
	})
}
