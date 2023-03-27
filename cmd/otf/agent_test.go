package main

import (
	"bytes"
	"testing"

	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentTokenNewCommand(t *testing.T) {
	at := &auth.AgentToken{Token: "secret-token"}
	cmd := fakeApp(withAgentToken(at)).agentTokenNewCommand()
	cmd.SetArgs([]string{"testing", "--organization", "automatize"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Regexp(t, `Successfully created agent token: secret-token`, got.String())
}
