package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/client"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_OrganizationTokens demonstrates the use of an organization
// token to authenticate via the API.
func TestIntegration_OrganizationTokens(t *testing.T) {
	integrationTest(t)

	daemon, org, ctx := setup(t)

	ot, token, err := daemon.Organizations.CreateToken(ctx, organization.CreateOrganizationTokenOptions{
		Organization: org.Name,
	})
	require.NoError(t, err)
	assert.Equal(t, org.Name, ot.Organization)

	client, err := client.New(otfhttp.ClientConfig{
		URL:   daemon.System.URL("/"),
		Token: string(token),
	})
	require.NoError(t, err)

	// create some workspaces and attempt to list them using client
	// authenticating with an organization token
	daemon.createWorkspace(t, ctx, org)
	daemon.createWorkspace(t, ctx, org)
	daemon.createWorkspace(t, ctx, org)

	got, err := client.Workspaces.List(ctx, workspace.ListOptions{
		Organization: &org.Name,
	})
	require.NoError(t, err)
	assert.Equal(t, 3, len(got.Items))

	// re-generate token
	_, _, err = daemon.Organizations.CreateToken(ctx, organization.CreateOrganizationTokenOptions{
		Organization: org.Name,
	})
	require.NoError(t, err)

	// access with previous token should now be refused
	_, err = client.Workspaces.List(ctx, workspace.ListOptions{
		Organization: &org.Name,
	})
	require.Equal(t, internal.ErrUnauthorized, err)
}
