package sqlite

import (
	"testing"

	"github.com/leg100/go-tfe"
	"github.com/leg100/ots"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConfigurationVersion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	require.NoError(t, err)

	svc := NewConfigurationVersionService(db)
	wsSvc := NewWorkspaceService(db)
	orgService := NewOrganizationService(db)

	// Create 1 org, 1 ws, 1 cv

	_, err = orgService.CreateOrganization(&tfe.OrganizationCreateOptions{
		Name:  ots.String("automatize"),
		Email: ots.String("sysadmin@automatize.co.uk"),
	})
	require.NoError(t, err)

	ws, err := wsSvc.CreateWorkspace("automatize", &tfe.WorkspaceCreateOptions{
		Name: ots.String("dev"),
	})
	require.NoError(t, err)

	cv, err := svc.CreateConfigurationVersion(ws.ID, &tfe.ConfigurationVersionCreateOptions{})
	require.NoError(t, err)

	require.Equal(t, tfe.ConfigurationPending, cv.Status)

	// Upload

	err = svc.UploadConfigurationVersion(cv.ID, []byte("testdata"))
	require.NoError(t, err)

	// Get

	cv, err = svc.GetConfigurationVersion(cv.ID)
	require.NoError(t, err)

	require.Equal(t, tfe.ConfigurationUploaded, cv.Status)

	// List

	cvs, err := svc.ListConfigurationVersions(ws.ID, tfe.ConfigurationVersionListOptions{})
	require.NoError(t, err)

	require.Equal(t, 1, len(cvs.Items))
}
