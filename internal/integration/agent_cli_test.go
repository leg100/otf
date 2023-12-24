package integration

import (
	"encoding/json"
	"testing"

	"github.com/leg100/otf/internal/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_AgentCLI demonstrates managing agents via the CLI
func TestIntegration_AgentCLI(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t, nil)

	t.Run("create agent pool", func(t *testing.T) {
		// should output agent pool marshaled into json
		out := daemon.otfcli(t, ctx, "agents", "pools", "new",
			"--organization", org.Name, "--name", "pool1",
		)
		var got agent.Pool
		err := json.Unmarshal([]byte(out), &got)
		require.NoError(t, err)
		assert.Equal(t, "pool1", got.Name)
	})

	t.Run("create agent token", func(t *testing.T) {
		pool, err := daemon.Agents.CreateAgentPool(ctx, agent.CreateAgentPoolOptions{
			Name:         "pool-1",
			Organization: org.Name,
		})
		require.NoError(t, err)
		out := daemon.otfcli(t, ctx, "agents", "tokens", "new",
			"--agent-pool-id", pool.ID, "--description", "token-1",
		)
		// should return agent token (base64 encoded JWT)
		assert.Regexp(t, `^[\w-]+\.[\w-]+\.[\w-]+$`, out)
	})
}
