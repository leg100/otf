package sqlite

import (
	"testing"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestWorkspace(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	require.NoError(t, err)

	orgDB := NewOrganizationDB(db)
	wsDB := NewWorkspaceDB(db)

	// Create one org and three workspaces

	org, err := orgDB.Create(&ots.Organization{
		ExternalID: "org-123",
		Name:       "automatize",
		Email:      "sysadmin@automatize.co.uk",
	})
	require.NoError(t, err)

	for _, name := range []string{"dev", "staging", "prod"} {
		ws, err := wsDB.Create(&ots.Workspace{
			Name:           name,
			ExternalID:     ots.NewWorkspaceID(),
			Organization:   org,
			OrganizationID: org.InternalID,
		})
		require.NoError(t, err)

		require.Equal(t, name, ws.Name)
		require.Contains(t, ws.ExternalID, "ws-")
	}

	// Update

	spec := ots.WorkspaceSpecifier{Name: ots.String("dev"), OrganizationName: ots.String("automatize")}
	ws, err := wsDB.Update(spec, func(ws *ots.Workspace) error {
		ws.Name = "newdev"
		return nil
	})
	require.NoError(t, err)

	require.Equal(t, "newdev", ws.Name)

	// Get

	ws, err = wsDB.Get(ots.WorkspaceSpecifier{Name: ots.String("newdev"), OrganizationName: ots.String("automatize")})
	require.NoError(t, err)

	require.Equal(t, "newdev", ws.Name)

	// List

	workspaces, err := wsDB.List("automatize", ots.WorkspaceListOptions{})
	require.NoError(t, err)

	require.Equal(t, 3, len(workspaces.Items))

	// List with pagination

	workspaces, err = wsDB.List("automatize", ots.WorkspaceListOptions{ListOptions: tfe.ListOptions{PageNumber: 1, PageSize: 2}})
	require.NoError(t, err)

	require.Equal(t, 2, len(workspaces.Items))

	workspaces, err = wsDB.List("automatize", ots.WorkspaceListOptions{ListOptions: tfe.ListOptions{PageNumber: 2, PageSize: 2}})
	require.NoError(t, err)

	require.Equal(t, 1, len(workspaces.Items))

	// List with search

	workspaces, err = wsDB.List("automatize", ots.WorkspaceListOptions{Prefix: ots.String("new")})
	require.NoError(t, err)

	require.Equal(t, 1, len(workspaces.Items))
	require.Equal(t, "newdev", workspaces.Items[0].Name)

	// Delete

	require.NoError(t, wsDB.Delete(ots.WorkspaceSpecifier{Name: ots.String("newdev"), OrganizationName: ots.String("automatize")}))

	// Re-create

	ws, err = wsDB.Create(&ots.Workspace{
		Name:         "dev",
		Organization: org,
	})
	require.NoError(t, err)

	require.Equal(t, "dev", ws.Name)

	// Update by ID

	ws, err = wsDB.Update(ots.WorkspaceSpecifier{ID: ots.String(ws.ExternalID)}, func(ws *ots.Workspace) error {
		ws.Name = "staging"
		return nil
	})
	require.NoError(t, err)

	require.Equal(t, "staging", ws.Name)

	// Get by ID

	ws, err = wsDB.Get(ots.WorkspaceSpecifier{ID: ots.String(ws.ExternalID)})
	require.NoError(t, err)

	require.Equal(t, "staging", ws.Name)

	// Delete by ID

	require.NoError(t, wsDB.Delete(ots.WorkspaceSpecifier{ID: ots.String(ws.ExternalID)}))
}
