package agent

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentTokenNewCommand(t *testing.T) {
	cli := &agentCLI{
		Service: &fakeService{
			token: []byte("secret-token"),
		},
	}
	cmd := cli.agentTokenNewCommand()
	cmd.SetArgs([]string{"testing", "--organization", "automatize"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Regexp(t, `Successfully created agent token: secret-token`, got.String())
}
