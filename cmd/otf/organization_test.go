package main

import (
	"bytes"
	"testing"

	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationCommand(t *testing.T) {
	org := &organization.Organization{Name: "acme-corp"}
	cmd := fakeApp(withOrganization(org)).organizationNewCommand()
	cmd.SetArgs([]string{"automatize"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Successfully created organization acme-corp\n", got.String())
}

func TestOrganizationCommandMissingName(t *testing.T) {
	cmd := fakeApp().organizationNewCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}
