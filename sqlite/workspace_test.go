package sqlite

import (
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWorkspace(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	require.NoError(t, err)

	orgService := NewOrganizationService(db)
	svc := NewWorkspaceService(db)

	// Create one org and three workspaces

	_, err = orgService.CreateOrganization(&tfe.OrganizationCreateOptions{
		Name:  ots.String("automatize"),
		Email: ots.String("sysadmin@automatize.co.uk"),
	})
	require.NoError(t, err)

	for _, name := range []string{"dev", "staging", "prod"} {
		ws, err := svc.CreateWorkspace("automatize", &tfe.WorkspaceCreateOptions{
			Name: ots.String(name),
		})
		require.NoError(t, err)

		require.Equal(t, name, ws.Name)
		require.Contains(t, ws.ID, "ws-")
	}

	// Update

	ws, err := svc.UpdateWorkspace("dev", "automatize", &tfe.WorkspaceUpdateOptions{
		Name: ots.String("newdev"),
	})
	require.NoError(t, err)

	require.Equal(t, "newdev", ws.Name)

	// Get

	ws, err = svc.GetWorkspace("newdev", "automatize")
	require.NoError(t, err)

	require.Equal(t, "newdev", ws.Name)

	// List

	workspaces, err := svc.ListWorkspaces("automatize", ots.WorkspaceListOptions{})
	require.NoError(t, err)

	require.Equal(t, 3, len(workspaces.Items))

	// List with pagination

	workspaces, err = svc.ListWorkspaces("automatize", ots.WorkspaceListOptions{ListOptions: ots.ListOptions{PageNumber: 1, PageSize: 2}})
	require.NoError(t, err)

	require.Equal(t, 2, len(workspaces.Items))

	workspaces, err = svc.ListWorkspaces("automatize", ots.WorkspaceListOptions{ListOptions: ots.ListOptions{PageNumber: 2, PageSize: 2}})
	require.NoError(t, err)

	require.Equal(t, 1, len(workspaces.Items))

	// List with search

	workspaces, err = svc.ListWorkspaces("automatize", ots.WorkspaceListOptions{Search: ots.String("new")})
	require.NoError(t, err)

	require.Equal(t, 1, len(workspaces.Items))
	require.Equal(t, "newdev", workspaces.Items[0].Name)

	// Delete

	require.NoError(t, svc.DeleteWorkspace("newdev", "automatize"))

	// Re-create

	ws, err = svc.CreateWorkspace("automatize", &tfe.WorkspaceCreateOptions{
		Name: ots.String("dev"),
	})
	require.NoError(t, err)

	require.Equal(t, "dev", ws.Name)

	// Update by ID

	ws, err = svc.UpdateWorkspaceByID(ws.ID, &tfe.WorkspaceUpdateOptions{
		Name: ots.String("staging"),
	})
	require.NoError(t, err)

	require.Equal(t, "staging", ws.Name)

	// Get by ID

	ws, err = svc.GetWorkspaceByID(ws.ID)
	require.NoError(t, err)

	require.Equal(t, "staging", ws.Name)

	// Delete by ID

	require.NoError(t, svc.DeleteWorkspaceByID(ws.ID))
}
