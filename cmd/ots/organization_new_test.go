package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationCommand(t *testing.T) {
	cmd := OrganizationNewCommand(&FakeClientConfig{})
	cmd.SetArgs([]string{"automatize", "--email", "sysadmin@automatize.co"})
	require.NoError(t, cmd.Execute())
}

func TestOrganizationCommandMissingName(t *testing.T) {
	cmd := OrganizationNewCommand(&FakeClientConfig{})
	cmd.SetArgs([]string{"--email", "sysadmin@automatize.co"})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}

func TestOrganizationCommandMissingEmail(t *testing.T) {
	cmd := OrganizationNewCommand(&FakeClientConfig{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"email\" not set")
}
