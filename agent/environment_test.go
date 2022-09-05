package agent

import (
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvironment(t *testing.T) {
	env := Environment{
		Logger: logr.Discard(),
		out:    discard{},
	}
	err := env.Execute(&fakeJob{"sleep", []string{"1"}})
	require.NoError(t, err)
}

func TestEnvironment_Cancel(t *testing.T) {
	env := Environment{
		Logger: logr.Discard(),
		out:    discard{},
	}

	wait := make(chan error)
	go func() {
		wait <- env.Execute(&fakeJob{"sleep", []string{"100"}})
	}()
	// give the 'sleep' cmd above time to start
	time.Sleep(100 * time.Millisecond)

	env.Cancel(false)
	err := <-wait
	assert.Error(t, err)
}
