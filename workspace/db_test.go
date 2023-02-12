package workspace

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_Create(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	ws := otf.NewTestWorkspace(t, org)

	err := db.CreateWorkspace(ctx, ws)
	require.NoError(t, err)

	t.Run("Duplicate", func(t *testing.T) {
		err := db.CreateWorkspace(ctx, ws)
		require.Equal(t, otf.ErrResourceAlreadyExists, err)
	})
}

func TestWorkspace_Update(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)

	got, err := db.UpdateWorkspace(ctx, ws.ID(), func(ws *otf.Workspace) error {
		return ws.Update(otf.UpdateWorkspaceOptions{
			Description: otf.String("updated description"),
		})
	})
	require.NoError(t, err)
	assert.Equal(t, "updated description", got.Description())

	// assert too that the WS returned by UpdateWorkspace is identical to one
	// returned by GetWorkspace
	want, err := db.GetWorkspace(ctx, ws.ID())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestWorkspace_GetByID(t *testing.T) {
	db := sql.NewTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	want := CreateTestWorkspace(t, db, org)

	got, err := db.GetWorkspace(context.Background(), want.ID())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestWorkspace_GetByName(t *testing.T) {
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	want := CreateTestWorkspace(t, db, org)

	got, err := db.GetWorkspaceByName(context.Background(), org.Name(), want.Name())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestWorkspace_Lock(t *testing.T) {
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	user := CreateTestUser(t, db)
	ctx := otf.AddSubjectToContext(context.Background(), user)

	t.Run("lock by id", func(t *testing.T) {
		ws := CreateTestWorkspace(t, db, org)
		got, err := db.LockWorkspace(ctx, ws.ID(), otf.WorkspaceLockOptions{})
		require.NoError(t, err)
		assert.True(t, got.Locked())
	})

	t.Run("lock by name", func(t *testing.T) {
		ws := CreateTestWorkspace(t, db, org)
		got, err := db.LockWorkspace(ctx, ws.ID(), otf.WorkspaceLockOptions{})
		require.NoError(t, err)
		assert.True(t, got.Locked())
	})

	t.Run("unlock by id", func(t *testing.T) {
		ws := CreateTestWorkspace(t, db, org)
		_, err := db.LockWorkspace(ctx, ws.ID(), otf.WorkspaceLockOptions{})
		require.NoError(t, err)
		got, err := db.UnlockWorkspace(ctx, ws.ID(), otf.WorkspaceUnlockOptions{})
		require.NoError(t, err)
		assert.False(t, got.Locked())
	})

	t.Run("unlock by name", func(t *testing.T) {
		ws := CreateTestWorkspace(t, db, org)
		_, err := db.LockWorkspace(ctx, ws.ID(), otf.WorkspaceLockOptions{})
		require.NoError(t, err)
		got, err := db.UnlockWorkspace(ctx, ws.ID(), otf.WorkspaceUnlockOptions{})
		require.NoError(t, err)
		assert.False(t, got.Locked())
	})
}

func TestWorkspace_ListByUserID(t *testing.T) {
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws1 := CreateTestWorkspace(t, db, org)
	ws2 := CreateTestWorkspace(t, db, org)
	team1 := createTestTeam(t, db, org)
	team2 := createTestTeam(t, db, org)
	_ = createTestWorkspacePermission(t, db, ws1, team1, rbac.WorkspaceAdminRole)
	_ = createTestWorkspacePermission(t, db, ws2, team2, rbac.WorkspacePlanRole)
	user := CreateTestUser(t, db, otf.WithTeamMemberships(team1, team2))

	tests := []struct {
		name         string
		userID       string
		organization string
		opts         otf.ListOptions
		want         func(*testing.T, *otf.WorkspaceList)
	}{
		{
			name:         "show both workspaces",
			userID:       user.ID(),
			organization: org.Name(),
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Contains(t, l.Items, ws1)
				assert.Contains(t, l.Items, ws2)
			},
		},
		{
			name:         "query non-existent org",
			userID:       user.ID(),
			organization: "acme-corp",
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
		{
			name:         "query non-existent user",
			userID:       "mr-invisible",
			organization: org.Name(),
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
		{
			name:         "paginated results ordered by updated_at",
			userID:       user.ID(),
			organization: org.Name(),
			opts:         otf.ListOptions{PageNumber: 1, PageSize: 1},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 1, len(l.Items))
				// results are in descending order so we expect ws2 to be listed
				// first.
				assert.Equal(t, ws2, l.Items[0])
				assert.Equal(t, 2, l.TotalCount())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.ListWorkspacesByUserID(context.Background(), tt.userID, tt.organization, tt.opts)
			require.NoError(t, err)

			tt.want(t, results)
		})
	}
}

func TestWorkspace_List(t *testing.T) {
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws1 := CreateTestWorkspace(t, db, org)
	ws2 := CreateTestWorkspace(t, db, org)

	tests := []struct {
		name string
		opts otf.WorkspaceListOptions
		want func(*testing.T, *otf.WorkspaceList)
	}{
		{
			name: "filter by org",
			opts: otf.WorkspaceListOptions{Organization: otf.String(org.Name())},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Contains(t, l.Items, ws1)
				assert.Contains(t, l.Items, ws2)
			},
		},
		{
			name: "filter by prefix",
			opts: otf.WorkspaceListOptions{Organization: otf.String(org.Name()), Prefix: ws1.Name()[:5]},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 1, len(l.Items))
				assert.Equal(t, ws1, l.Items[0])
			},
		},
		{
			name: "filter by non-existent org",
			opts: otf.WorkspaceListOptions{Organization: otf.String("non-existent")},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
		{
			name: "filter by non-existent prefix",
			opts: otf.WorkspaceListOptions{Organization: otf.String(org.Name()), Prefix: "xyz"},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
		{
			name: "paginated results ordered by updated_at",
			opts: otf.WorkspaceListOptions{Organization: otf.String(org.Name()), ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 1, len(l.Items))
				// results are in descending order so we expect ws2 to be listed
				// first.
				assert.Equal(t, ws2, l.Items[0])
				assert.Equal(t, 2, l.TotalCount())
			},
		},
		{
			name: "stray pagination",
			opts: otf.WorkspaceListOptions{Organization: otf.String(org.Name()), ListOptions: otf.ListOptions{PageNumber: 999, PageSize: 10}},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				// zero results but count should ignore pagination
				assert.Equal(t, 0, len(l.Items))
				assert.Equal(t, 2, l.TotalCount())
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
	db := NewTestDB(t)
	ctx := context.Background()
	org := CreateTestOrganization(t, db)

	ws := CreateTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})
	_ = createTestRun(t, db, ws, cv)

	err := db.DeleteWorkspace(ctx, ws.ID())
	require.NoError(t, err)

	results, err := db.ListWorkspaces(ctx, otf.WorkspaceListOptions{Organization: otf.String(org.Name())})
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
}

func TestWorkspacePermissions_Set(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	team := createTestTeam(t, db, org)

	err := db.SetWorkspacePermission(ctx, ws.ID(), team.Name(), rbac.WorkspacePlanRole)
	require.NoError(t, err)

	t.Run("Update", func(t *testing.T) {
		err := db.SetWorkspacePermission(ctx, ws.ID(), team.Name(), rbac.WorkspaceAdminRole)
		require.NoError(t, err)
	})
}

func TestWorkspacePermissions_List(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := organization.CreateTestOrganization(t, db)
	ws := workspace.CreateTestWorkspace(t, db, org)
	team1 := createTestTeam(t, db, org)
	team2 := createTestTeam(t, db, org)
	perm1 := createTestWorkspacePermission(t, db, ws, team1, rbac.WorkspaceAdminRole)
	perm2 := createTestWorkspacePermission(t, db, ws, team2, rbac.WorkspacePlanRole)

	perms, err := db.ListWorkspacePermissions(ctx, ws.ID())
	require.NoError(t, err)
	if assert.Equal(t, 2, len(perms)) {
		assert.Contains(t, perms, perm1)
		assert.Contains(t, perms, perm2)
	}
}

func TestWorkspacePermissions_Unset(t *testing.T) {
	ctx := context.Background()
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	team := createTestTeam(t, db, org)
	_ = createTestWorkspacePermission(t, db, ws, team, rbac.WorkspaceAdminRole)

	err := db.UnsetWorkspacePermission(ctx, ws.ID(), team.Name())
	require.NoError(t, err)

	perms, err := db.ListWorkspacePermissions(ctx, ws.ID())
	require.NoError(t, err)
	assert.Equal(t, 0, len(perms))
}
