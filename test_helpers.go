package otf

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newTestOrganization(t *testing.T) *Organization {
	org, err := NewOrganization(OrganizationCreateOptions{
		Name: String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

func newTestWorkspace(t *testing.T, org *Organization) *Workspace {
	ws, err := NewWorkspace(org, WorkspaceCreateOptions{
		Name: uuid.NewString(),
	})
	require.NoError(t, err)
	return ws
}

func newTestConfigurationVersion(t *testing.T, ws *Workspace, opts ConfigurationVersionCreateOptions) *ConfigurationVersion {
	cv, err := NewConfigurationVersion(ws.ID(), opts)
	require.NoError(t, err)
	return cv
}
