package remoteops

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDaemon_SetDefaultConcurrency(t *testing.T) {
	daemon, err := NewDaemon(logr.Discard(), nil, Config{})
	require.NoError(t, err)
	assert.Equal(t, 5, daemon.Concurrency)
}
