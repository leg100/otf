package runner

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentTokenNewCommand(t *testing.T) {
	cli := &agentCLI{
		agentCLIService: &fakeService{
			token: []byte("secret-token"),
		},
	}
	cmd := cli.agentTokenNewCommand()
	cmd.SetArgs([]string{"testing", "--agent-pool-id", "pool-123", "--description", "my new token"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Regexp(t, `Successfully created agent token: secret-token`, got.String())
}
