package sqlite

import (
	"encoding/base64"
	"os"
	"testing"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestStateVersion(t *testing.T) {
	db, err := New(":memory:", &gorm.Config{})
	require.NoError(t, err)

	orgDB := NewOrganizationDB(db)
	wsDB := NewWorkspaceDB(db)
	svDB := NewStateVersionDB(db)

	// Create one org and three workspaces

	org, err := orgDB.Create(&ots.Organization{
		Name:  "automatize",
		Email: "sysadmin@automatize.co.uk",
	})
	require.NoError(t, err)

	var workspaces []*ots.Workspace
	for _, name := range []string{"dev", "staging", "prod"} {
		ws, err := wsDB.Create(&ots.Workspace{
			Name:         name,
			ID:           ots.NewWorkspaceID(),
			Organization: org,
		})
		require.NoError(t, err)

		require.Equal(t, name, ws.Name)
		require.Contains(t, ws.ID, "ws-")

		workspaces = append(workspaces, ws)
	}

	// Create state versions (3 WSs, 3 SVs per WS)

	// ...first get a state file to upload
	data, err := os.ReadFile("../testdata/terraform.tfstate")
	require.NoError(t, err)

	var stateVersionIDs []string
	for _, ws := range workspaces {
		for _, j := range []int{1, 2, 3} {
			sv, err := svDB.Create(&ots.StateVersion{
				ID:        ots.NewStateVersionID(),
				Serial:    int64(j),
				State:     base64.StdEncoding.EncodeToString(data),
				Workspace: ws,
			})
			require.NoError(t, err)

			require.Contains(t, sv.ID, "sv-")

			stateVersionIDs = append(stateVersionIDs, sv.ID)
		}
	}

	// List

	svl, err := svDB.List(tfe.StateVersionListOptions{
		ListOptions:  tfe.ListOptions{PageNumber: 1, PageSize: 20},
		Organization: ots.String("automatize"),
		Workspace:    ots.String("dev"),
	})
	require.NoError(t, err)

	require.Equal(t, 3, len(svl.Items))

	// Get

	_, err = svDB.Get(ots.StateVersionGetOptions{ID: &stateVersionIDs[0]})
	require.NoError(t, err)

	// Current

	sv, err := svDB.Get(ots.StateVersionGetOptions{WorkspaceID: &workspaces[0].ID})
	require.NoError(t, err)
	require.Equal(t, int64(3), sv.Serial)
}
