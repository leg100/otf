package sqlite

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationVersion_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db, "org-123", "automatize")
	ws := createTestWorkspace(t, db, "ws-123", "default", org)

	cdb := NewConfigurationVersionDB(db)

	cv, err := cdb.Create(newTestConfigurationVersion("cv-123", ws))
	require.NoError(t, err)

	assert.Equal(t, int64(1), cv.Model.ID)
}

func TestConfigurationVersion_Update(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db, "org-123", "automatize")
	ws := createTestWorkspace(t, db, "ws-123", "default", org)
	cv := createTestConfigurationVersion(t, db, "cv-123", ws)

	cdb := NewConfigurationVersionDB(db)

	updated, err := cdb.Update(cv.ID, func(cv *otf.ConfigurationVersion) error {
		cv.Status = otf.ConfigurationUploaded
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, otf.ConfigurationUploaded, updated.Status)
}

func TestConfigurationVersion_Get(t *testing.T) {
	tests := []struct {
		name string
		opts otf.ConfigurationVersionGetOptions
	}{
		{
			name: "by id",
			opts: otf.ConfigurationVersionGetOptions{ID: otf.String("cv-123")},
		},
		{
			name: "by workspace",
			opts: otf.ConfigurationVersionGetOptions{WorkspaceID: otf.String("ws-123")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			org := createTestOrganization(t, db, "org-123", "automatize")
			ws := createTestWorkspace(t, db, "ws-123", "default", org)
			cv := createTestConfigurationVersion(t, db, "cv-123", ws)

			cdb := NewConfigurationVersionDB(db)

			got, err := cdb.Get(tt.opts)
			require.NoError(t, err)

			// Assertion won't succeed unless both have a workspace with a nil
			// org.
			cv.Workspace.Organization = nil

			assert.Equal(t, cv, got)
		})
	}
}

func TestConfigurationVersion_List(t *testing.T) {
	tests := []struct {
		name        string
		workspaceID string
		want        func(*testing.T, *otf.ConfigurationVersionList, ...*otf.ConfigurationVersion)
	}{
		{
			name:        "filter by workspace",
			workspaceID: "ws-123",
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
			db := newTestDB(t)
			org := createTestOrganization(t, db, "org-123", "automatize")
			ws := createTestWorkspace(t, db, "ws-123", "default", org)

			cv1 := createTestConfigurationVersion(t, db, "cv-1", ws)
			cv2 := createTestConfigurationVersion(t, db, "cv-2", ws)

			cdb := NewConfigurationVersionDB(db)

			results, err := cdb.List(tt.workspaceID, otf.ConfigurationVersionListOptions{})
			require.NoError(t, err)

			tt.want(t, results, cv1, cv2)
		})
	}
}
