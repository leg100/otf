package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/repo"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &otf.Superuser{})

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)

		ws, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         otf.String(uuid.NewString()),
			Organization: otf.String(org.Name),
		})
		require.NoError(t, err)

		t.Run("duplicate error", func(t *testing.T) {
			_, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
				Name:         otf.String(ws.Name),
				Organization: otf.String(org.Name),
			})
			require.Equal(t, otf.ErrResourceAlreadyExists, err)
		})
	})

	t.Run("create connected workspace", func(t *testing.T) {
		svc := setup(t, &config{repo: "test/dummy"})

		org := svc.createOrganization(t, ctx)
		vcsprov := svc.createVCSProvider(t, ctx, org)
		ws, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         otf.String(uuid.NewString()),
			Organization: &org.Name,
			ConnectOptions: &workspace.ConnectOptions{
				RepoPath:      "test/dummy",
				VCSProviderID: vcsprov.ID,
			},
		})
		require.NoError(t, err)

		// webhook should be registered with github
		require.True(t, svc.githubServer.HasWebhook())

		t.Run("delete workspace connection", func(t *testing.T) {
			err := svc.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.WorkspaceConnection,
				ResourceID:     ws.ID,
			})
			require.NoError(t, err)
		})

		// webhook should now have been deleted from github
		require.False(t, svc.githubServer.HasWebhook())
	})

	t.Run("deleting connected workspace also deletes webhook", func(t *testing.T) {
		svc := setup(t, &config{repo: "test/dummy"})

		org := svc.createOrganization(t, ctx)
		vcsprov := svc.createVCSProvider(t, ctx, org)
		ws, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         otf.String(uuid.NewString()),
			Organization: &org.Name,
			ConnectOptions: &workspace.ConnectOptions{
				RepoPath:      "test/dummy",
				VCSProviderID: vcsprov.ID,
			},
		})
		require.NoError(t, err)

		// webhook should be registered with github
		require.True(t, svc.githubServer.HasWebhook())

		_, err = svc.DeleteWorkspace(ctx, ws.ID)
		require.NoError(t, err)

		// webhook should now have been deleted from github
		require.False(t, svc.githubServer.HasWebhook())
	})

	t.Run("connect workspace", func(t *testing.T) {
		svc := setup(t, &config{repo: "test/dummy"})

		org := svc.createOrganization(t, ctx)
		ws := svc.createWorkspace(t, ctx, org)
		vcsprov := svc.createVCSProvider(t, ctx, org)

		got, err := svc.Connect(ctx, repo.ConnectOptions{
			ConnectionType: repo.WorkspaceConnection,
			VCSProviderID:  vcsprov.ID,
			ResourceID:     ws.ID,
			RepoPath:       "test/dummy",
		})
		require.NoError(t, err)
		want := &repo.Connection{VCSProviderID: vcsprov.ID, Repo: "test/dummy"}
		assert.Equal(t, want, got)

		t.Run("delete workspace connection", func(t *testing.T) {
			err := svc.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.WorkspaceConnection,
				ResourceID:     ws.ID,
			})
			require.NoError(t, err)
		})
	})

	t.Run("update", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)

		got, err := svc.UpdateWorkspace(ctx, ws.ID, workspace.UpdateOptions{
			Description: otf.String("updated description"),
		})
		require.NoError(t, err)
		assert.Equal(t, "updated description", got.Description)

		// assert too that the WS returned by UpdateWorkspace is identical to one
		// returned by GetWorkspace
		want, err := svc.GetWorkspace(ctx, ws.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("get by id", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createWorkspace(t, ctx, nil)

		got, err := svc.GetWorkspace(ctx, want.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("get by name", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createWorkspace(t, ctx, nil)

		got, err := svc.GetWorkspaceByName(ctx, want.Organization, want.Name)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		ws1 := svc.createWorkspace(t, ctx, org)
		ws2 := svc.createWorkspace(t, ctx, org)

		tests := []struct {
			name string
			opts workspace.ListOptions
			want func(*testing.T, *workspace.WorkspaceList)
		}{
			{
				name: "filter by org",
				opts: workspace.ListOptions{Organization: otf.String(org.Name)},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, ws1)
					assert.Contains(t, l.Items, ws2)
				},
			},
			{
				name: "filter by prefix",
				opts: workspace.ListOptions{Organization: otf.String(org.Name), Prefix: ws1.Name[:5]},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 1, len(l.Items))
					assert.Equal(t, ws1, l.Items[0])
				},
			},
			{
				name: "filter by non-existent org",
				opts: workspace.ListOptions{Organization: otf.String("non-existent")},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "filter by non-existent prefix",
				opts: workspace.ListOptions{Organization: otf.String(org.Name), Prefix: "xyz"},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "paginated results ordered by updated_at",
				opts: workspace.ListOptions{Organization: otf.String(org.Name), ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 1, len(l.Items))
					// results are in descending order so we expect ws2 to be listed
					// first...unless - and this happens very occasionally - the
					// updated_at time is equal down to nearest millisecond.
					if !ws1.UpdatedAt.Equal(ws2.UpdatedAt) {
						assert.Equal(t, ws2, l.Items[0])
					}
					assert.Equal(t, 2, l.TotalCount())
				},
			},
			{
				name: "stray pagination",
				opts: workspace.ListOptions{Organization: otf.String(org.Name), ListOptions: otf.ListOptions{PageNumber: 999, PageSize: 10}},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
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
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		ws1 := svc.createWorkspace(t, ctx, org)
		ws2 := svc.createWorkspace(t, ctx, org)

		team1 := svc.createTeam(t, ctx, org)
		team2 := svc.createTeam(t, ctx, org)
		_ = svc.SetPermission(ctx, ws1.ID, team1.Name, rbac.WorkspacePlanRole)
		_ = svc.SetPermission(ctx, ws2.ID, team2.Name, rbac.WorkspacePlanRole)
		user1 := svc.createUser(t, ctx, auth.WithTeams(team1, team2))
		user2 := svc.createUser(t, ctx)

		tests := []struct {
			name string
			user *auth.User
			opts workspace.ListOptions
			want func(*testing.T, *workspace.WorkspaceList)
		}{
			{
				name: "show both workspaces",
				user: user1,
				opts: workspace.ListOptions{Organization: otf.String(org.Name)},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, ws1)
					assert.Contains(t, l.Items, ws2)
				},
			},
			{
				name: "query non-existent org",
				user: user1,
				opts: workspace.ListOptions{Organization: otf.String("acme-corp")},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "user with no perms",
				user: user2,
				opts: workspace.ListOptions{Organization: otf.String(org.Name)},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "paginated results ordered by updated_at",
				user: user1,
				opts: workspace.ListOptions{
					Organization: otf.String(org.Name),
					ListOptions:  otf.ListOptions{PageNumber: 1, PageSize: 1},
				},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 1, len(l.Items))
					// results are in descending order so we expect ws2 to be listed
					// first...unless - and this happens very occasionally - the
					// updated_at time is equal down to nearest millisecond.
					if !ws1.UpdatedAt.Equal(ws2.UpdatedAt) {
						assert.Equal(t, ws2, l.Items[0])
					}
					assert.Equal(t, 2, l.TotalCount())
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx := otf.AddSubjectToContext(ctx, tt.user)
				results, err := svc.ListWorkspaces(ctx, tt.opts)
				require.NoError(t, err)

				tt.want(t, results)
			})
		}
	})

	t.Run("lock", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		ws := svc.createWorkspace(t, ctx, org)

		// create user with owner perms
		team, err := svc.CreateTeam(ctx, auth.NewTeamOptions{
			Name:         "owners", // has perm to lock/unlock workspace
			Organization: org.Name,
		})
		require.NoError(t, err)
		user := svc.createUser(t, ctx, auth.WithTeams(team))
		ctx := otf.AddSubjectToContext(ctx, user)

		got, err := svc.LockWorkspace(ctx, ws.ID, nil)
		require.NoError(t, err)
		assert.True(t, got.Locked())

		t.Run("unlock", func(t *testing.T) {
			got, err := svc.UnlockWorkspace(ctx, ws.ID, nil, false)
			require.NoError(t, err)
			assert.False(t, got.Locked())
		})
	})

	t.Run("delete", func(t *testing.T) {
		svc := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, nil)

		_, err := svc.DeleteWorkspace(ctx, ws.ID)
		require.NoError(t, err)

		results, err := svc.ListWorkspaces(ctx, workspace.ListOptions{Organization: otf.String(ws.Organization)})
		require.NoError(t, err)
		assert.Equal(t, 0, len(results.Items))

		// TODO: Test ON CASCADE DELETE functionality for config versions,
		// runs, etc
	})
}
