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

	cvDB := NewConfigurationVersionDB(db)
	wsDB := NewWorkspaceDB(db)
	orgDB := NewOrganizationDB(db)

	// Create 1 org, 1 ws, 1 cv

	org, err := orgDB.Create(&ots.Organization{
		ExternalID: "org-123",
		Name:       "automatize",
		Email:      "sysadmin@automatize.co.uk",
	})
	require.NoError(t, err)

	ws, err := wsDB.Create(&ots.Workspace{
		Name:           "dev",
		ExternalID:     "ws-123",
		OrganizationID: org.InternalID,
		Organization:   org,
	})
	require.NoError(t, err)

	cv, err := cvDB.Create(&ots.ConfigurationVersion{
		ExternalID:  "cv-123",
		Status:      tfe.ConfigurationPending,
		WorkspaceID: ws.InternalID,
		Workspace:   ws,
	})
	require.NoError(t, err)

	require.Equal(t, tfe.ConfigurationPending, cv.Status)

	// Update

	cv, err = cvDB.Update(cv.ExternalID, func(cv *ots.ConfigurationVersion) error {
		cv.Configuration = []byte("testdata")
		cv.Status = tfe.ConfigurationUploaded
		return nil
	})
	require.NoError(t, err)

	// Get

	cv, err = cvDB.Get(ots.ConfigurationVersionGetOptions{ID: &cv.ExternalID})
	require.NoError(t, err)

	require.Equal(t, tfe.ConfigurationUploaded, cv.Status)

	// List

	cvs, err := cvDB.List(ws.ExternalID, ots.ConfigurationVersionListOptions{})
	require.NoError(t, err)

	require.Equal(t, 1, len(cvs.Items))
}
