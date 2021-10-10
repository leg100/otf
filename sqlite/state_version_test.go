package sqlite

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func TestStateVersion(t *testing.T) {
	db, err := New(logr.Discard(), ":memory:")
	require.NoError(t, err)

	orgDB := NewOrganizationDB(db)
	wsDB := NewWorkspaceDB(db)
	svDB := NewStateVersionDB(db)

	// Create one org and three workspaces

	org, err := orgDB.Create(&otf.Organization{
		Name:  "automatize",
		Email: "sysadmin@automatize.co.uk",
	})
	require.NoError(t, err)

	var workspaces []*otf.Workspace
	for _, name := range []string{"dev", "staging", "prod"} {
		ws, err := wsDB.Create(&otf.Workspace{
			Name:         name,
			ID:           otf.GenerateID("ws"),
			Organization: org,
		})
		require.NoError(t, err)

		require.Equal(t, name, ws.Name)
		require.Contains(t, ws.ID, "ws-")

		workspaces = append(workspaces, ws)
	}

	// Create state versions (3 WSs, 3 SVs per WS)

	var stateVersionIDs []string
	for _, ws := range workspaces {
		for _, j := range []int{1, 2, 3} {
			sv, err := svDB.Create(&otf.StateVersion{
				ID:        otf.GenerateID("sv"),
				Serial:    int64(j),
				Workspace: ws,
			})
			require.NoError(t, err)

			require.Contains(t, sv.ID, "sv-")

			stateVersionIDs = append(stateVersionIDs, sv.ID)
		}
	}

	// List

	svl, err := svDB.List(otf.StateVersionListOptions{
		ListOptions:  otf.ListOptions{PageNumber: 1, PageSize: 20},
		Organization: otf.String("automatize"),
		Workspace:    otf.String("dev"),
	})
	require.NoError(t, err)

	require.Equal(t, 3, len(svl.Items))

	// Get

	_, err = svDB.Get(otf.StateVersionGetOptions{ID: &stateVersionIDs[0]})
	require.NoError(t, err)

	// Current

	sv, err := svDB.Get(otf.StateVersionGetOptions{WorkspaceID: &workspaces[0].ID})
	require.NoError(t, err)
	require.Equal(t, int64(3), sv.Serial)
}
