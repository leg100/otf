package sqlite

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

func TestConfigurationVersion(t *testing.T) {
	db, err := New(":memory:")
	require.NoError(t, err)

	cvDB := NewConfigurationVersionDB(db)
	wsDB := NewWorkspaceDB(db)
	orgDB := NewOrganizationDB(db)

	// Create 1 org, 1 ws, 1 cv

	org, err := orgDB.Create(&otf.Organization{
		ID:    "org-123",
		Name:  "automatize",
		Email: "sysadmin@automatize.co.uk",
	})
	require.NoError(t, err)

	ws, err := wsDB.Create(&otf.Workspace{
		Name:         "dev",
		ID:           "ws-123",
		Organization: org,
	})
	require.NoError(t, err)

	cv, err := cvDB.Create(&otf.ConfigurationVersion{
		ID:        "cv-123",
		Status:    otf.ConfigurationPending,
		Workspace: ws,
	})
	require.NoError(t, err)

	require.Equal(t, otf.ConfigurationPending, cv.Status)

	// Update

	cv, err = cvDB.Update(cv.ID, func(cv *otf.ConfigurationVersion) error {
		cv.Status = otf.ConfigurationUploaded
		return nil
	})
	require.NoError(t, err)

	// Get

	cv, err = cvDB.Get(otf.ConfigurationVersionGetOptions{ID: &cv.ID})
	require.NoError(t, err)

	require.Equal(t, otf.ConfigurationUploaded, cv.Status)

	// List

	cvs, err := cvDB.List(ws.ID, otf.ConfigurationVersionListOptions{})
	require.NoError(t, err)

	require.Equal(t, 1, len(cvs.Items))
}
