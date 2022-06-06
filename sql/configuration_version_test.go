package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationVersion_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)

	err := db.CreateConfigurationVersion(context.Background(), newTestConfigurationVersion(t, ws))
	require.NoError(t, err)
}

func TestConfigurationVersion_Update(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)

	err := db.UploadConfigurationVersion(context.Background(), cv.ID(), func(cv *otf.ConfigurationVersion, uploader otf.ConfigUploader) error {
		_, err := uploader.Upload(context.Background(), nil)
		return err
	})
	require.NoError(t, err)

	got, err := db.GetConfigurationVersion(context.Background(), otf.ConfigurationVersionGetOptions{ID: otf.String(cv.ID())})
	require.NoError(t, err)

	assert.Equal(t, otf.ConfigurationUploaded, got.Status())
}

func TestConfigurationVersion_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)

	tests := []struct {
		name string
		opts otf.ConfigurationVersionGetOptions
	}{
		{
			name: "by id",
			opts: otf.ConfigurationVersionGetOptions{ID: otf.String(cv.ID())},
		},
		{
			name: "by workspace",
			opts: otf.ConfigurationVersionGetOptions{WorkspaceID: otf.String(ws.ID())},
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
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)

	cv1 := createTestConfigurationVersion(t, db, ws)
	cv2 := createTestConfigurationVersion(t, db, ws)

	tests := []struct {
		name        string
		workspaceID string
		opts        otf.ConfigurationVersionListOptions
		want        func(*testing.T, *otf.ConfigurationVersionList)
	}{
		{
			name:        "no pagination",
			workspaceID: ws.ID(),
			want: func(t *testing.T, got *otf.ConfigurationVersionList) {
				assert.Equal(t, 2, len(got.Items))
				assert.Equal(t, 2, got.TotalCount)
				assert.Contains(t, got.Items, cv1)
				assert.Contains(t, got.Items, cv2)
			},
		},
		{
			name:        "pagination",
			workspaceID: ws.ID(),
			opts:        otf.ConfigurationVersionListOptions{ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}},
			want: func(t *testing.T, got *otf.ConfigurationVersionList) {
				assert.Equal(t, 1, len(got.Items))
				assert.Equal(t, 2, got.TotalCount)
			},
		},
		{
			name:        "stray pagination",
			workspaceID: ws.ID(),
			opts:        otf.ConfigurationVersionListOptions{ListOptions: otf.ListOptions{PageNumber: 999, PageSize: 10}},
			want: func(t *testing.T, got *otf.ConfigurationVersionList) {
				// Zero items but total count should ignore pagination
				assert.Equal(t, 0, len(got.Items))
				assert.Equal(t, 2, got.TotalCount)
			},
		},
		{
			name:        "query non-existent workspace",
			workspaceID: "ws-non-existent",
			want: func(t *testing.T, got *otf.ConfigurationVersionList) {
				assert.Empty(t, got.Items)
				assert.Equal(t, 0, got.TotalCount)
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
