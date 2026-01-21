package organization

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrganizationCommand(t *testing.T) {
	cmd := newFakeCLI(nil).newOrganizationCommand()
	cmd.SetArgs([]string{"acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Successfully created organization acme-corp\n", got.String())
}

func TestDeleteOrganizationCommand(t *testing.T) {
	cmd := newFakeCLI(nil).deleteOrganizationCommand()
	cmd.SetArgs([]string{"acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	assert.Equal(t, "Successfully deleted organization acme-corp\n", got.String())
}

func TestOrganizationCommandMissingName(t *testing.T) {
	cmd := newFakeCLI(nil).newOrganizationCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	assert.EqualError(t, err, "accepts 1 arg(s), received 0")
}

type fakeCLIService struct {
	org *Organization
}

func newFakeCLI(org *Organization) *CLI {
	return &CLI{cliService: &fakeCLIService{org: org}}
}

func (f *fakeCLIService) CreateOrganization(ctx context.Context, opts CreateOptions) (*Organization, error) {
	return NewOrganization(opts)
}

func (f *fakeCLIService) DeleteOrganization(context.Context, Name) error {
	return nil
}
