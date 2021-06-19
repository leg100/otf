package sqlite

import (
	"testing"

	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestConfigurationVersion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"))
	require.NoError(t, err)

	svc := NewConfigurationVersionService(db)

	// Create

	cv, err := svc.CreateConfigurationVersion(&tfe.ConfigurationVersionCreateOptions{})
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

	cvs, err := svc.ListConfigurationVersions(ots.ConfigurationVersionListOptions{})
	require.NoError(t, err)

	require.Equal(t, 1, len(cvs.Items))
}
