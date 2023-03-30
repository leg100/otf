package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrganizationCommand(t *testing.T) {
	cmd := fakeApp().newOrganizationCommand()
	cmd.SetArgs([]string{"acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Successfully created organization acme-corp\n", got.String())
}

func TestDeleteOrganizationCommand(t *testing.T) {
	cmd := fakeApp().deleteOrganizationCommand()
	cmd.SetArgs([]string{"acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Successfully deleted organization acme-corp\n", got.String())
}

func TestOrganizationCommandMissingName(t *testing.T) {
	cmd := fakeApp().newOrganizationCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}
