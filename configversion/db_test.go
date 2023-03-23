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

func TestDB(t *testing.T) {
	ctx := context.Background()
	db := &pgdb{sql.NewTestDB(t)}

	t.Run("create", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		cv := NewTestConfigurationVersion(t, ws, ConfigurationVersionCreateOptions{})

		err := db.CreateConfigurationVersion(ctx, cv)
		require.NoError(t, err)
	})
	t.Run("update", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		cv := CreateTestConfigurationVersion(t, db, ws, ConfigurationVersionCreateOptions{})

		err := db.UploadConfigurationVersion(ctx, cv.ID, func(cv *ConfigurationVersion, uploader ConfigUploader) error {
			_, err := uploader.Upload(ctx, nil)
			return err
		})
		require.NoError(t, err)

		got, err := db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{ID: otf.String(cv.ID)})
		require.NoError(t, err)

		assert.Equal(t, ConfigurationUploaded, got.Status)
	})

	t.Run("get", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		cv := CreateTestConfigurationVersion(t, db, ws, ConfigurationVersionCreateOptions{})

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
				got, err := db.GetConfigurationVersion(ctx, tt.opts)
				require.NoError(t, err)
				assert.Equal(t, cv, got)
			})
		}
	})

	t.Run("list", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)

		cv1 := CreateTestConfigurationVersion(t, db, ws, ConfigurationVersionCreateOptions{})
		cv2 := CreateTestConfigurationVersion(t, db, ws, ConfigurationVersionCreateOptions{})

		tests := []struct {
			name        string
			workspaceID string
			opts        ConfigurationVersionListOptions
			want        func(*testing.T, *ConfigurationVersionList)
		}{
			{
				name:        "no pagination",
				workspaceID: ws.ID,
				want: func(t *testing.T, got *ConfigurationVersionList) {
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
				want: func(t *testing.T, got *ConfigurationVersionList) {
					assert.Equal(t, 1, len(got.Items))
					assert.Equal(t, 2, got.TotalCount())
				},
			},
			{
				name:        "stray pagination",
				workspaceID: ws.ID,
				opts:        ConfigurationVersionListOptions{ListOptions: otf.ListOptions{PageNumber: 999, PageSize: 10}},
				want: func(t *testing.T, got *ConfigurationVersionList) {
					// Zero items but total count should ignore pagination
					assert.Equal(t, 0, len(got.Items))
					assert.Equal(t, 2, got.TotalCount())
				},
			},
			{
				name:        "query non-existent workspace",
				workspaceID: "ws-non-existent",
				want: func(t *testing.T, got *ConfigurationVersionList) {
					assert.Empty(t, got.Items)
					assert.Equal(t, 0, got.TotalCount())
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				results, err := db.ListConfigurationVersions(ctx, tt.workspaceID, tt.opts)
				require.NoError(t, err)

				tt.want(t, results)
			})
		}
	})
}
