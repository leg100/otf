package sqlite

import (
	"encoding/base64"
	"testing"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestStateVersion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	require.NoError(t, err)

	orgService := NewOrganizationService(db)
	wsSvc := NewWorkspaceService(db)
	_ = NewStateVersionOutputService(db)
	svc := NewStateVersionService(db)

	// Create one org and three workspaces

	_, err = orgService.CreateOrganization(&tfe.OrganizationCreateOptions{
		Name:  ots.String("automatize"),
		Email: ots.String("sysadmin@automatize.co.uk"),
	})
	require.NoError(t, err)

	var workspaceIDs []string
	for _, name := range []string{"dev", "staging", "prod"} {
		ws, err := wsSvc.CreateWorkspace("automatize", &tfe.WorkspaceCreateOptions{
			Name: ots.String(name),
		})
		require.NoError(t, err)

		require.Equal(t, name, ws.Name)
		require.Contains(t, ws.ID, "ws-")

		workspaceIDs = append(workspaceIDs, ws.ID)
	}

	// Create state versions (3 WSs, 3 SVs per WS)

	var stateVersionIDs []string
	for _, ws := range workspaceIDs {
		for _, j := range []int{1, 2, 3} {
			sv, err := svc.CreateStateVersion(ws, &tfe.StateVersionCreateOptions{
				Serial: tfe.Int64(int64(j)),
				State:  ots.String(base64.StdEncoding.EncodeToString([]byte("test state"))),
			})
			require.NoError(t, err)

			require.Contains(t, sv.ID, "sv-")

			stateVersionIDs = append(stateVersionIDs, sv.ID)
		}
	}

	// List

	svl, err := svc.ListStateVersions("automatize", "dev", tfe.StateVersionListOptions{ListOptions: tfe.ListOptions{PageNumber: 1, PageSize: 20}})
	require.NoError(t, err)

	require.Equal(t, 3, len(svl.Items))

	// Get

	_, err = svc.GetStateVersion(stateVersionIDs[0])
	require.NoError(t, err)

	// Current

	sv, err := svc.CurrentStateVersion(workspaceIDs[0])
	require.NoError(t, err)
	require.Equal(t, int64(3), sv.Serial)
	// Current

	state, err := svc.DownloadStateVersion(stateVersionIDs[0])
	require.NoError(t, err)
	require.Equal(t, "test state", string(state))
}
