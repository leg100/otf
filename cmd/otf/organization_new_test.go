package main

import (
	"testing"

	"github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationCommand(t *testing.T) {
	cmd := OrganizationNewCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	require.NoError(t, cmd.Execute())
}

func TestOrganizationCommandMissingName(t *testing.T) {
	cmd := OrganizationNewCommand(&http.FakeClientFactory{})
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}
