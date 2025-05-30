package integration

import (
	"os"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigurationVersion(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, _, ctx := setup(t)
		ws := svc.createWorkspace(t, ctx, nil)

		_, err := svc.Configs.Create(ctx, ws.ID, configversion.CreateOptions{})
		require.NoError(t, err)
	})

	t.Run("upload config", func(t *testing.T) {
		svc, _, ctx := setup(t)
		cv := svc.createConfigurationVersion(t, ctx, nil, nil)
		tarball, err := os.ReadFile("./testdata/tarball.tar.gz")
		require.NoError(t, err)

		err = svc.Configs.UploadConfig(ctx, cv.ID, tarball)
		require.NoError(t, err)

		got, err := svc.Configs.Get(ctx, cv.ID)
		require.NoError(t, err)

		assert.Equal(t, configversion.ConfigurationUploaded, got.Status)

		t.Run("download config", func(t *testing.T) {
			gotConfig, err := svc.Configs.DownloadConfig(ctx, cv.ID)
			require.NoError(t, err)
			assert.Equal(t, tarball, gotConfig)
		})
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createConfigurationVersion(t, ctx, nil, nil)

		got, err := svc.Configs.Get(ctx, want.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("get latest", func(t *testing.T) {
		svc, _, ctx := setup(t)
		want := svc.createConfigurationVersion(t, ctx, nil, nil)

		got, err := svc.Configs.GetLatest(ctx, want.WorkspaceID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t)
		ws := svc.createWorkspace(t, ctx, nil)
		cv1 := svc.createConfigurationVersion(t, ctx, ws, nil)
		cv2 := svc.createConfigurationVersion(t, ctx, ws, nil)

		tests := []struct {
			name        string
			workspaceID resource.TfeID
			opts        configversion.ListOptions
			want        func(*testing.T, *resource.Page[*configversion.ConfigurationVersion], error)
		}{
			{
				name:        "no pagination",
				workspaceID: ws.ID,
				want: func(t *testing.T, got *resource.Page[*configversion.ConfigurationVersion], err error) {
					require.NoError(t, err)
					assert.Equal(t, 2, len(got.Items))
					assert.Equal(t, 2, got.TotalCount)
					assert.Contains(t, got.Items, cv1)
					assert.Contains(t, got.Items, cv2)
				},
			},
			{
				name:        "pagination",
				workspaceID: ws.ID,
				opts:        configversion.ListOptions{PageOptions: resource.PageOptions{PageNumber: 1, PageSize: 1}},
				want: func(t *testing.T, got *resource.Page[*configversion.ConfigurationVersion], err error) {
					require.NoError(t, err)
					assert.Equal(t, 1, len(got.Items))
					assert.Equal(t, 2, got.TotalCount)
				},
			},
			{
				name:        "stray pagination",
				workspaceID: ws.ID,
				opts:        configversion.ListOptions{PageOptions: resource.PageOptions{PageNumber: 999, PageSize: 10}},
				want: func(t *testing.T, got *resource.Page[*configversion.ConfigurationVersion], err error) {
					require.NoError(t, err)
					// Zero items but total count should ignore pagination
					assert.Equal(t, 0, len(got.Items))
					assert.Equal(t, 2, got.TotalCount)
				},
			},
			{
				name:        "query non-existent workspace",
				workspaceID: resource.NewTfeID(resource.WorkspaceKind),
				want: func(t *testing.T, got *resource.Page[*configversion.ConfigurationVersion], err error) {
					require.ErrorIs(t, err, internal.ErrResourceNotFound)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				results, err := svc.Configs.List(ctx, tt.workspaceID, tt.opts)
				tt.want(t, results, err)
			})
		}
	})
}
