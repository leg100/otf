package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationCommand(t *testing.T) {
	cmd := OrganizationNewCommand(&fakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Successfully created organization automatize\n", got.String())
}

func TestOrganizationCommandMissingName(t *testing.T) {
	cmd := OrganizationNewCommand(&fakeClientFactory{})
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}
