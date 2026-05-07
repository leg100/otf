package runner

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/mitchellh/iochan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeOperationClientHostname is a minimal OperationClient stub that only
// satisfies the Hostname() call used in DoOperation.
type fakeOperationClientHostname struct {
	OperationClient
	hostname string
}

func (f *fakeOperationClientHostname) Hostname() string { return f.hostname }

func TestCredentialEnvs(t *testing.T) {
	credentialEnvs := func(opts OperationOptions) []string {
		var envs []string
		envs = append(envs, internal.CredentialEnv(opts.Client.Hostname(), opts.JobToken))
		if opts.CredentialHostname != "" && opts.CredentialHostname != opts.Client.Hostname() {
			envs = append(envs, internal.CredentialEnv(opts.CredentialHostname, opts.JobToken))
		}
		return envs
	}

	t.Run("sets only client hostname env when CredentialHostname is empty", func(t *testing.T) {
		opts := OperationOptions{
			Client:   &fakeOperationClientHostname{hostname: "app.otf.example.com"},
			JobToken: []byte("token"),
		}
		envs := credentialEnvs(opts)
		assert.Len(t, envs, 1)
		assert.Contains(t, envs, internal.CredentialEnv("app.otf.example.com", []byte("token")))
	})

	t.Run("sets both env vars when CredentialHostname differs from client hostname", func(t *testing.T) {
		opts := OperationOptions{
			Client:             &fakeOperationClientHostname{hostname: "otf.default:8080"},
			JobToken:           []byte("token"),
			CredentialHostname: "app.otf.example.com",
		}
		envs := credentialEnvs(opts)
		assert.Len(t, envs, 2)
		assert.Contains(t, envs, internal.CredentialEnv("otf.default:8080", []byte("token")))
		assert.Contains(t, envs, internal.CredentialEnv("app.otf.example.com", []byte("token")))
	})

	t.Run("sets only one env var when CredentialHostname matches client hostname", func(t *testing.T) {
		opts := OperationOptions{
			Client:             &fakeOperationClientHostname{hostname: "app.otf.example.com"},
			JobToken:           []byte("token"),
			CredentialHostname: "app.otf.example.com",
		}
		envs := credentialEnvs(opts)
		assert.Len(t, envs, 1)
		assert.Contains(t, envs, internal.CredentialEnv("app.otf.example.com", []byte("token")))
	})
}

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
			Logger:  logr.Discard(),
			out:     w,
			workdir: &workdir{root: ""},
		}
		done := make(chan error)
		go func() {
			done <- wkr.execute([]string{"./testdata/killme"})
		}()

		assert.Equal(t, "ok, you can kill me now\n", <-iochan.DelimReader(r, '\n'))
		wkr.cancel(false, true)
		assert.NoError(t, <-done)
	})

	t.Run("cancel forceably", func(t *testing.T) {
		r, w := io.Pipe()
		reader := iochan.DelimReader(r, '\n')
		op := &operation{
			Logger:   logr.Discard(),
			cancelfn: func() {},
			out:      w,
			workdir:  &workdir{root: ""},
		}
		done := make(chan error)
		go func() {
			done <- op.execute([]string{"./testdata/killme_harder"})
		}()

		// send graceful cancel
		assert.Equal(t, "ok, try killing me now\n", <-reader)
		op.cancel(false, true)
		assert.Equal(t, "you will have to try harder than that\n", <-reader)
		// send force cancel
		op.cancel(true, true)
		assert.Error(t, <-done)
	})
}
