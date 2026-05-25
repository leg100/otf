package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/leg100/otf/internal/organization"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrganizationCommand(t *testing.T) {
	cmd := (&CLI{client: &fakeClient{}}).newOrganizationCommand()
	cmd.SetArgs([]string{"acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Successfully created organization acme-corp\n", got.String())
}

func TestDeleteOrganizationCommand(t *testing.T) {
	cmd := (&CLI{client: &fakeClient{}}).deleteOrganizationCommand()
	cmd.SetArgs([]string{"acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Successfully deleted organization acme-corp\n", got.String())
}

func TestOrganizationCommandMissingName(t *testing.T) {
	cmd := (&CLI{client: &fakeClient{}}).newOrganizationCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}

type fakeClient struct {
}

func (f *fakeClient) CreateOrganization(ctx context.Context, opts organization.CreateOptions) (*organization.Organization, error) {
	return organization.NewOrganization(opts)
}

func (f *fakeClient) DeleteOrganization(context.Context, organization.Name) error {
	return nil
}
