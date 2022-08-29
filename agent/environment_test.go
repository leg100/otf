package agent

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironment(t *testing.T) {
	env, _ := newTestEnvironment(t)
	err := env.Execute(&fakeJob{"sleep", []string{"1"}})
	require.NoError(t, err)
}

func TestEnvironment_Cancel(t *testing.T) {
	env, _ := newTestEnvironment(t)

	wait := make(chan error)
	go func() {
		wait <- env.Execute(&fakeJob{"sleep", []string{"10"}})
	}()
	// give the 'sleep' cmd above time to start
	time.Sleep(1000 * time.Millisecond)

	require.NoError(t, env.Cancel(true))
	err := <-wait
	assert.Error(t, err)
}

func TestEnvironment_Container(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantOut string
		wantErr error
	}{
		{
			name:    "terraform version",
			args:    []string{"version"},
			wantOut: "Terraform v1.0.10\non linux_amd64\n",
		},
		{
			name:    "invalid arg",
			args:    []string{"invalid"},
			wantErr: ErrNonZeroExitCode,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env, out := newTestEnvironment(t)
			err := env.RunCLI("terraform", tt.args...)
			if tt.wantErr != nil {
				assert.True(t, errors.Is(err, ErrNonZeroExitCode))
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantOut, out.String())
		})
	}
}

func newTestEnvironment(t *testing.T) (*Environment, *fakeWriteCloser) {
	pluginDir := t.TempDir()

	env, err := NewEnvironment(
		logr.Discard(),
		nil,
		"run-123",
		otf.PlanPhase,
		context.Background(),
		[]string{
			"TF_PLUGIN_CACHE_DIR=" + pluginDir,
			"TF_IN_AUTOMATION=true",
			"CHECKPOINT_DISABLE=true",
		},
	)
	require.NoError(t, err)

	// replace default job writer with fake
	out := &fakeWriteCloser{new(bytes.Buffer)}
	env.out = out

	return env, out
}
