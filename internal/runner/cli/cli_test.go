package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/runner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClient struct {
	client
	token []byte
}

func (f *fakeClient) CreateAgentToken(context.Context, resource.TfeID, runner.CreateAgentTokenOptions) (*runner.AgentToken, []byte, error) {
	return nil, f.token, nil
}

func TestAgentTokenNewCommand(t *testing.T) {
	cli := &cli{
		client: &fakeClient{
			token: []byte("secret-token"),
		},
	}
	cmd := cli.agentTokenNewCommand()
	cmd.SetArgs([]string{"--agent-pool-id", "pool-123", "--description", "my new token"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Regexp(t, `Successfully created agent token: secret-token`, got.String())
}
