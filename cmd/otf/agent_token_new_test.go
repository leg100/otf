package main

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentTokenNewCommand(t *testing.T) {
	cmd := AgentTokenNewCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"testing", "--organization", "automatize"})
	buf := bytes.Buffer{}
	cmd.SetOut(&buf)
	require.NoError(t, cmd.Execute())
	assert.Regexp(t,
		regexp.MustCompile(`Successfully created agent token: [a-zA-Z0-9\-_]+`),
		buf.String())
}
