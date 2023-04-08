package agent

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgent_SetDefaultConcurrency(t *testing.T) {
	agent, err := NewAgent(logr.Discard(), nil, Config{})
	require.NoError(t, err)
	assert.Equal(t, 5, agent.Concurrency)
}
