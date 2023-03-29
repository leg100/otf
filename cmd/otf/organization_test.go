package main

import (
	"bytes"
	"testing"

	"github.com/leg100/otf/organization"
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

func TestAddOrganizationMembershipCommand(t *testing.T) {
	org := &organization.Organization{Name: "acme-corp"}
	cmd := fakeApp(withOrganization(org)).addOrganizationMembershipCommand()
	cmd.SetArgs([]string{"bobby", "--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Successfully added bobby to acme-corp\n", got.String())
}

func TestDeleteOrganizationMembershipCommand(t *testing.T) {
	org := &organization.Organization{Name: "acme-corp"}
	cmd := fakeApp(withOrganization(org)).deleteOrganizationMembershipCommand()
	cmd.SetArgs([]string{"bobby", "--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Successfully removed bobby from acme-corp\n", got.String())
}
