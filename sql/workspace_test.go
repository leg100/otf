package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := newTestWorkspace(t, org)

	err := db.CreateWorkspace(context.Background(), ws)
	require.NoError(t, err)

	t.Run("Duplicate", func(t *testing.T) {
		err := db.CreateWorkspace(context.Background(), ws)
		require.Equal(t, otf.ErrResourcesAlreadyExists, err)
	})
}

func TestWorkspace_Update(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	org := createTestOrganization(t, db)

	tests := []struct {
		name string
		spec func(ws *otf.Workspace) otf.WorkspaceSpec
	}{
		{
			name: "by id",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpec {
				return ws.SpecID()
			},
		},
		{
			name: "by name",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpec {
				return ws.SpecName()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := createTestWorkspace(t, db, org)
			_, err := db.UpdateWorkspace(ctx, tt.spec(ws), func(ws *otf.Workspace) error {
				return ws.UpdateWithOptions(context.Background(), otf.WorkspaceUpdateOptions{
					Description: otf.String("updated description"),
				})
			})
			require.NoError(t, err)
			got, err := db.GetWorkspace(ctx, tt.spec(ws))
			require.NoError(t, err)
			assert.Equal(t, "updated description", got.Description())
		})
	}
}

func TestWorkspace_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)

	tests := []struct {
		name string
		spec otf.WorkspaceSpec
	}{
		{
			name: "by id",
			spec: ws.SpecID(),
		},
		{
			name: "by name",
			spec: ws.SpecName(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetWorkspace(context.Background(), tt.spec)
			require.NoError(t, err)
			assert.Equal(t, ws, got)
		})
	}
}

func TestWorkspace_Lock(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	org := createTestOrganization(t, db)

	t.Run("lock by id", func(t *testing.T) {
		ws := createTestWorkspace(t, db, org)
		got, err := db.LockWorkspace(ctx, ws.SpecID(), otf.WorkspaceLockOptions{
			Requestor: &otf.AnonymousUser,
		})
		require.NoError(t, err)
		assert.True(t, got.Locked())
	})

	t.Run("lock by name", func(t *testing.T) {
		ws := createTestWorkspace(t, db, org)
		got, err := db.LockWorkspace(ctx, ws.SpecName(), otf.WorkspaceLockOptions{
			Requestor: &otf.AnonymousUser,
		})
		require.NoError(t, err)
		assert.True(t, got.Locked())
	})

	t.Run("unlock by id", func(t *testing.T) {
		ws := createTestWorkspace(t, db, org)
		_, err := db.LockWorkspace(ctx, ws.SpecID(), otf.WorkspaceLockOptions{
			Requestor: &otf.AnonymousUser,
		})
		require.NoError(t, err)
		got, err := db.UnlockWorkspace(ctx, ws.SpecID(), otf.WorkspaceUnlockOptions{
			Requestor: &otf.AnonymousUser,
		})
		require.NoError(t, err)
		assert.False(t, got.Locked())
	})

	t.Run("unlock by name", func(t *testing.T) {
		ws := createTestWorkspace(t, db, org)
		_, err := db.LockWorkspace(ctx, ws.SpecName(), otf.WorkspaceLockOptions{
			Requestor: &otf.AnonymousUser,
		})
		require.NoError(t, err)
		got, err := db.UnlockWorkspace(ctx, ws.SpecID(), otf.WorkspaceUnlockOptions{
			Requestor: &otf.AnonymousUser,
		})
		require.NoError(t, err)
		assert.False(t, got.Locked())
	})
}

func TestWorkspace_List(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws1 := createTestWorkspace(t, db, org)
	ws2 := createTestWorkspace(t, db, org)

	tests := []struct {
		name string
		opts otf.WorkspaceListOptions
		want func(*testing.T, *otf.WorkspaceList)
	}{
		{
			name: "filter by org",
			opts: otf.WorkspaceListOptions{OrganizationName: org.Name()},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Contains(t, l.Items, ws1)
				assert.Contains(t, l.Items, ws2)
			},
		},
		{
			name: "filter by prefix",
			opts: otf.WorkspaceListOptions{OrganizationName: org.Name(), Prefix: ws1.Name()[:5]},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 1, len(l.Items))
				assert.Equal(t, ws1, l.Items[0])
			},
		},
		{
			name: "filter by non-existent org",
			opts: otf.WorkspaceListOptions{OrganizationName: "non-existent"},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
		{
			name: "filter by non-existent prefix",
			opts: otf.WorkspaceListOptions{OrganizationName: org.Name(), Prefix: "xyz"},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
		{
			name: "stray pagination",
			opts: otf.WorkspaceListOptions{OrganizationName: org.Name(), ListOptions: otf.ListOptions{PageNumber: 999, PageSize: 10}},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				// zero results but count should ignore pagination
				assert.Equal(t, 0, len(l.Items))
				assert.Equal(t, 2, l.TotalCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.ListWorkspaces(context.Background(), tt.opts)
			require.NoError(t, err)

			tt.want(t, results)
		})
	}
}

func TestWorkspace_Delete(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	org := createTestOrganization(t, db)

	tests := []struct {
		name string
		spec func(ws *otf.Workspace) otf.WorkspaceSpec
	}{
		{
			name: "by id",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpec {
				return ws.SpecID()
			},
		},
		{
			name: "by name",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpec {
				return ws.SpecName()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := createTestWorkspace(t, db, org)
			cv := createTestConfigurationVersion(t, db, ws)
			_ = createTestRun(t, db, ws, cv)

			err := db.DeleteWorkspace(ctx, tt.spec(ws))
			require.NoError(t, err)

			results, err := db.ListWorkspaces(ctx, otf.WorkspaceListOptions{OrganizationName: org.Name()})
			require.NoError(t, err)

			assert.Equal(t, 0, len(results.Items))

			// Test ON CASCADE DELETE functionality for runs
			rl, err := db.ListRuns(ctx, otf.RunListOptions{WorkspaceID: otf.String(ws.ID())})
			require.NoError(t, err)

			assert.Equal(t, 0, len(rl.Items))

			// Test ON CASCADE DELETE functionality for config versions
			cvl, err := db.ListConfigurationVersions(ctx, ws.ID(), otf.ConfigurationVersionListOptions{})
			require.NoError(t, err)

			assert.Equal(t, 0, len(cvl.Items))
		})
	}
}
