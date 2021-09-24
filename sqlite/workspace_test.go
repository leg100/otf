package sqlite

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func TestWorkspace(t *testing.T) {
	db, err := New(":memory:")
	require.NoError(t, err)

	orgDB := NewOrganizationDB(db)
	wsDB := NewWorkspaceDB(db)
	_ = NewRunDB(db)
	_ = NewStateVersionDB(db)
	_ = NewConfigurationVersionDB(db)

	// Create one org and three workspaces

	org, err := orgDB.Create(&otf.Organization{
		ID:    "org-123",
		Name:  "automatize",
		Email: "sysadmin@automatize.co.uk",
	})
	require.NoError(t, err)

	for _, name := range []string{"dev", "staging", "prod"} {
		ws, err := wsDB.Create(&otf.Workspace{
			Name:         name,
			ID:           otf.GenerateID("ws"),
			Organization: org,
		})
		require.NoError(t, err)

		require.Equal(t, name, ws.Name)
		require.Contains(t, ws.ID, "ws-")
	}

	// Update

	spec := otf.WorkspaceSpecifier{Name: otf.String("dev"), OrganizationName: otf.String("automatize")}
	ws, err := wsDB.Update(spec, func(ws *otf.Workspace) error {
		ws.Name = "newdev"
		return nil
	})
	require.NoError(t, err)

	require.Equal(t, "newdev", ws.Name)

	// Get

	ws, err = wsDB.Get(otf.WorkspaceSpecifier{Name: otf.String("newdev"), OrganizationName: otf.String("automatize")})
	require.NoError(t, err)

	require.Equal(t, "newdev", ws.Name)

	// List

	workspaces, err := wsDB.List(otf.WorkspaceListOptions{OrganizationName: otf.String("automatize")})
	require.NoError(t, err)

	require.Equal(t, 3, len(workspaces.Items))

	// List with pagination

	workspaces, err = wsDB.List(otf.WorkspaceListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 2}})
	require.NoError(t, err)

	require.Equal(t, 2, len(workspaces.Items))

	workspaces, err = wsDB.List(otf.WorkspaceListOptions{ListOptions: otf.ListOptions{PageNumber: 2, PageSize: 2}})
	require.NoError(t, err)

	require.Equal(t, 1, len(workspaces.Items))

	// List with search

	workspaces, err = wsDB.List(otf.WorkspaceListOptions{Prefix: otf.String("new")})
	require.NoError(t, err)

	require.Equal(t, 1, len(workspaces.Items))
	require.Equal(t, "newdev", workspaces.Items[0].Name)

	// Delete

	require.NoError(t, wsDB.Delete(otf.WorkspaceSpecifier{Name: otf.String("newdev"), OrganizationName: otf.String("automatize")}))

	// Re-create

	ws, err = wsDB.Create(&otf.Workspace{
		Name:         "dev",
		Organization: org,
	})
	require.NoError(t, err)

	require.Equal(t, "dev", ws.Name)

	// Update by ID

	ws, err = wsDB.Update(otf.WorkspaceSpecifier{ID: otf.String(ws.ID)}, func(ws *otf.Workspace) error {
		ws.Name = "staging"
		return nil
	})
	require.NoError(t, err)

	require.Equal(t, "staging", ws.Name)

	// Get by ID

	ws, err = wsDB.Get(otf.WorkspaceSpecifier{ID: otf.String(ws.ID)})
	require.NoError(t, err)

	require.Equal(t, "staging", ws.Name)

	// Delete by ID

	require.NoError(t, wsDB.Delete(otf.WorkspaceSpecifier{ID: otf.String(ws.ID)}))
}
