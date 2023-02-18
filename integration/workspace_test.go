package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/testutil"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_Create(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	svc := testutil.NewWorkspaceService(t, db)
	org := testutil.CreateOrganization(t, db)

	t.Run("create", func(t *testing.T) {
		name := uuid.NewString()

		_, err := svc.CreateWorkspace(ctx, workspace.CreateWorkspaceOptions{
			Name:         otf.String(name),
			Organization: otf.String(org.Name()),
		})
		require.NoError(t, err)

		t.Run("Duplicate", func(t *testing.T) {
			_, err := svc.CreateWorkspace(ctx, workspace.CreateWorkspaceOptions{
				Name:         otf.String(name),
				Organization: otf.String(org.Name()),
			})
			require.Equal(t, otf.ErrResourceAlreadyExists, err)
		})
	})

	t.Run("update", func(t *testing.T) {
		org := testutil.CreateOrganization(t, db)
		ws := testutil.CreateWorkspace(t, db, org.Name())

		got, err := svc.UpdateWorkspace(ctx, ws.ID(), workspace.UpdateWorkspaceOptions{
			Description: otf.String("updated description"),
		})
		require.NoError(t, err)
		assert.Equal(t, "updated description", got.Description())

		// assert too that the WS returned by UpdateWorkspace is identical to one
		// returned by GetWorkspace
		want, err := svc.GetWorkspace(ctx, ws.ID())
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("get by id", func(t *testing.T) {
		want := testutil.CreateWorkspace(t, db, org.Name())

		got, err := svc.GetWorkspace(ctx, want.ID())
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("get by name", func(t *testing.T) {
		want := testutil.CreateWorkspace(t, db, org.Name())

		got, err := svc.GetWorkspaceByName(ctx, org.Name(), want.Name())
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("lock", func(t *testing.T) {
		user := testutil.CreateUser(t, db)
		ctx := otf.AddSubjectToContext(ctx, user)

		ws := testutil.CreateWorkspace(t, db, org.Name())
		got, err := svc.LockWorkspace(ctx, ws.ID())
		require.NoError(t, err)
		assert.True(t, got.Locked())
	})

	t.Run("unlock", func(t *testing.T) {
		ws := testutil.CreateWorkspace(t, db, org.Name())
		_, err := svc.LockWorkspace(ctx, ws.ID())
		require.NoError(t, err)
		got, err := svc.UnlockWorkspace(ctx, ws.ID(), false)
		require.NoError(t, err)
		assert.False(t, got.Locked())
	})

	t.Run("list", func(t *testing.T) {
		ws1 := testutil.CreateWorkspace(t, db, org.Name())
		ws2 := testutil.CreateWorkspace(t, db, org.Name())

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
				results, err := svc.ListWorkspaces(ctx, tt.opts)
				require.NoError(t, err)

				tt.want(t, results)
			})
		}
	})

	t.Run("list by user", func(t *testing.T) {
		ws1 := testutil.CreateWorkspace(t, db, org.Name())
		ws2 := testutil.CreateWorkspace(t, db, org.Name())
		team1 := createTestTeam(t, db, org)
		team2 := createTestTeam(t, db, org)
		_ = createTestWorkspacePermission(t, db, ws1, team1, rbac.WorkspaceAdminRole)
		_ = createTestWorkspacePermission(t, db, ws2, team2, rbac.WorkspacePlanRole)
		user := testutil.CreateUser(t, db, auth.WithTeams(team1, team2))

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
				results, err := db.ListWorkspacesByUserID(ctx, tt.userID, tt.organization, tt.opts)
				require.NoError(t, err)

				tt.want(t, results)
			})
		}
	})

	t.Run("delete", func(t *testing.T) {
		ws := testutil.CreateWorkspace(t, db, org.Name())
		cv := testutil.CreateConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})
		_ = testutil.CreateRun(t, db, ws, cv)
		runService := testutil.NewRunService(db)
		configService := testutil.NewConfigVersionService(db)

		_, err := svc.DeleteWorkspace(ctx, ws.ID())
		require.NoError(t, err)

		results, err := svc.ListWorkspaces(ctx, otf.WorkspaceListOptions{Organization: otf.String(org.Name())})
		require.NoError(t, err)

		assert.Equal(t, 0, len(results.Items))

		// Test ON CASCADE DELETE functionality for runs
		rl, err := runService.ListRuns(ctx, otf.RunListOptions{WorkspaceID: otf.String(ws.ID())})
		require.NoError(t, err)

		assert.Equal(t, 0, len(rl.Items))

		// Test ON CASCADE DELETE functionality for config versions
		cvl, err := configService.ListConfigurationVersions(ctx, ws.ID(), otf.ConfigurationVersionListOptions{})
		require.NoError(t, err)

		assert.Equal(t, 0, len(cvl.Items))
	})

	t.Run("set permission", func(t *testing.T) {
		org := testutil.CreateOrganization(t, db)
		ws := testutil.CreateWorkspace(t, db, org.Name())
		team := createTestTeam(t, db, org)

		err := svc.SetWorkspacePermission(ctx, ws.ID(), team.Name(), rbac.WorkspacePlanRole)
		require.NoError(t, err)

		t.Run("Update", func(t *testing.T) {
			err := svc.SetWorkspacePermission(ctx, ws.ID(), team.Name(), rbac.WorkspaceAdminRole)
			require.NoError(t, err)
		})
	})

	t.Run("list permissions", func(t *testing.T) {
		org := testutil.CreateOrganization(t, db)
		ws := testutil.CreateWorkspace(t, db, org.Name())
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
	})
	t.Run("unset permission", func(t *testing.T) {
		org := testutil.CreateOrganization(t, db)
		ws := testutil.CreateWorkspace(t, db, org.Name())
		team := createTestTeam(t, db, org)
		_ = createTestWorkspacePermission(t, db, ws, team, rbac.WorkspaceAdminRole)

		err := db.UnsetWorkspacePermission(ctx, ws.ID(), team.Name())
		require.NoError(t, err)

		perms, err := db.ListWorkspacePermissions(ctx, ws.ID())
		require.NoError(t, err)
		assert.Equal(t, 0, len(perms))
	})
}
