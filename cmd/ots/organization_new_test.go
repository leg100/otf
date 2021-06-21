package main

import (
	"context"
	"testing"

	"github.com/hashicorp/go-tfe"
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

type FakeClientConfig struct{}

func (f FakeClientConfig) NewClient() (Client, error) { return &FakeClient{}, nil }

type FakeClient struct{}

func (f FakeClient) Organizations() tfe.Organizations { return &FakeOrganizationsClient{} }

type FakeOrganizationsClient struct {
	tfe.Organizations
}

func (f *FakeOrganizationsClient) Create(ctx context.Context, opts tfe.OrganizationCreateOptions) (*tfe.Organization, error) {
	return &tfe.Organization{
		Name:  *opts.Name,
		Email: *opts.Email,
	}, nil
}
