package configversion

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationVersion_Create(t *testing.T) {
	db := sql.NewTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	cv := otf.NewTestConfigurationVersion(t, ws, otf.ConfigurationVersionCreateOptions{})

	err := db.CreateConfigurationVersion(context.Background(), cv)
	require.NoError(t, err)
}

func TestConfigurationVersion_Update(t *testing.T) {
	db := sql.NewTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})

	err := db.UploadConfigurationVersion(context.Background(), cv.ID, func(cv *otf.ConfigurationVersion, uploader otf.ConfigUploader) error {
		_, err := uploader.Upload(context.Background(), nil)
		return err
	})
	require.NoError(t, err)

	got, err := db.GetConfigurationVersion(context.Background(), otf.ConfigurationVersionGetOptions{ID: otf.String(cv.ID)})
	require.NoError(t, err)

	assert.Equal(t, otf.ConfigurationUploaded, got.Status())
}

func TestConfigurationVersion_Get(t *testing.T) {
	db := sql.NewTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})

	tests := []struct {
		name string
		opts ConfigurationVersionGetOptions
	}{
		{
			name: "by id",
			opts: ConfigurationVersionGetOptions{ID: otf.String(cv.ID)},
		},
		{
			name: "by workspace",
			opts: ConfigurationVersionGetOptions{WorkspaceID: otf.String(ws.ID)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetConfigurationVersion(context.Background(), tt.opts)
			require.NoError(t, err)
			assert.Equal(t, cv, got)
		})
	}
}

func TestConfigurationVersion_List(t *testing.T) {
	db := sql.NewTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	ws := workspace.CreateTestWorkspace(t, db, org.Name())

	cv1 := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})
	cv2 := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})

	tests := []struct {
		name        string
		workspaceID string
		opts        ConfigurationVersionListOptions
		want        func(*testing.T, *otf.ConfigurationVersionList)
	}{
		{
			name:        "no pagination",
			workspaceID: ws.ID,
			want: func(t *testing.T, got *otf.ConfigurationVersionList) {
				assert.Equal(t, 2, len(got.Items))
				assert.Equal(t, 2, got.TotalCount())
				assert.Contains(t, got.Items, cv1)
				assert.Contains(t, got.Items, cv2)
			},
		},
		{
			name:        "pagination",
			workspaceID: ws.ID,
			opts:        ConfigurationVersionListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}},
			want: func(t *testing.T, got *otf.ConfigurationVersionList) {
				assert.Equal(t, 1, len(got.Items))
				assert.Equal(t, 2, got.TotalCount())
			},
		},
		{
			name:        "stray pagination",
			workspaceID: ws.ID,
			opts:        ConfigurationVersionListOptions{ListOptions: otf.ListOptions{PageNumber: 999, PageSize: 10}},
			want: func(t *testing.T, got *otf.ConfigurationVersionList) {
				// Zero items but total count should ignore pagination
				assert.Equal(t, 0, len(got.Items))
				assert.Equal(t, 2, got.TotalCount())
			},
		},
		{
			name:        "query non-existent workspace",
			workspaceID: "ws-non-existent",
			want: func(t *testing.T, got *otf.ConfigurationVersionList) {
				assert.Empty(t, got.Items)
				assert.Equal(t, 0, got.TotalCount())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.ListConfigurationVersions(context.Background(), tt.workspaceID, tt.opts)
			require.NoError(t, err)

			tt.want(t, results)
		})
	}
}
