package main

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentTokenNewCommand(t *testing.T) {
	cmd := AgentTokenNewCommand(&fakeClientFactory{})
	cmd.SetArgs([]string{"testing", "--organization", "automatize"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Regexp(t,
		regexp.MustCompile(`Successfully created agent token: [a-zA-Z0-9\-_]+`),
		got.String())
}
