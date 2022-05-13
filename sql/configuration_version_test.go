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

	_, err := db.ConfigurationVersionStore().Create(newTestConfigurationVersion(ws))
	require.NoError(t, err)
}

func TestConfigurationVersion_Update(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)

	err := db.ConfigurationVersionStore().Upload(context.Background(), cv.ID, func(cv *otf.ConfigurationVersion, uploader otf.ConfigUploader) error {
		_, err := uploader.Upload(context.Background(), nil)
		return err
	})
	require.NoError(t, err)

	got, err := db.ConfigurationVersionStore().Get(otf.ConfigurationVersionGetOptions{ID: &cv.ID})
	require.NoError(t, err)

	assert.Equal(t, otf.ConfigurationUploaded, got.Status)
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
			opts: otf.ConfigurationVersionGetOptions{ID: otf.String(cv.ID)},
		},
		{
			name: "by workspace",
			opts: otf.ConfigurationVersionGetOptions{WorkspaceID: otf.String(ws.ID)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.ConfigurationVersionStore().Get(tt.opts)
			require.NoError(t, err)

			// Assertion won't succeed unless both have a workspace with a nil
			// org.
			cv.Workspace.Organization = nil

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
		want        func(*testing.T, *otf.ConfigurationVersionList, ...*otf.ConfigurationVersion)
	}{
		{
			name:        "filter by workspace",
			workspaceID: ws.ID,
			want: func(t *testing.T, l *otf.ConfigurationVersionList, created ...*otf.ConfigurationVersion) {
				assert.Equal(t, 2, len(l.Items))
				for _, cv := range created {
					// Assertion won't succeed unless both have a workspace with
					// a nil org.
					cv.Workspace.Organization = nil

					assert.Contains(t, l.Items, cv)
				}
			},
		},
		{
			name:        "filter by non-existent workspace",
			workspaceID: "ws-non-existent",
			want: func(t *testing.T, l *otf.ConfigurationVersionList, created ...*otf.ConfigurationVersion) {
				assert.Empty(t, l.Items)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.ConfigurationVersionStore().List(tt.workspaceID, otf.ConfigurationVersionListOptions{})
			require.NoError(t, err)

			tt.want(t, results, cv1, cv2)
		})
	}
}
