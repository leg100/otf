package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/repo"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace(t *testing.T) {
	t.Parallel()

	t.Run("create", func(t *testing.T) {
		svc := setup(t, nil)
		sub := svc.createSubscriber(t, ctx)
		org := svc.createOrganization(t, ctx)

		ws, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         internal.String(uuid.NewString()),
			Organization: internal.String(org.Name),
		})
		require.NoError(t, err)

		t.Run("duplicate error", func(t *testing.T) {
			_, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
				Name:         internal.String(ws.Name),
				Organization: internal.String(org.Name),
			})
			require.Equal(t, internal.ErrResourceAlreadyExists, err)
		})

		t.Run("receive events", func(t *testing.T) {
			assert.Equal(t, pubsub.NewCreatedEvent(org), <-sub)
			assert.Equal(t, pubsub.NewCreatedEvent(ws), <-sub)
		})
	})

	t.Run("create connected workspace", func(t *testing.T) {
		svc := setup(t, nil, github.WithRepo("test/dummy"))

		org := svc.createOrganization(t, ctx)
		vcsprov := svc.createVCSProvider(t, ctx, org)
		ws, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         internal.String(uuid.NewString()),
			Organization: &org.Name,
			ConnectOptions: &workspace.ConnectOptions{
				RepoPath:      "test/dummy",
				VCSProviderID: vcsprov.ID,
			},
		})
		require.NoError(t, err)

		// webhook should be registered with github
		require.True(t, svc.HasWebhook())

		t.Run("delete workspace connection", func(t *testing.T) {
			err := svc.Disconnect(ctx, repo.DisconnectOptions{
				ConnectionType: repo.WorkspaceConnection,
				ResourceID:     ws.ID,
			})
			require.NoError(t, err)
		})

		// webhook should now have been deleted from github
		require.False(t, svc.HasWebhook())
	})

	t.Run("deleting connected workspace also deletes webhook", func(t *testing.T) {
		svc := setup(t, nil, github.WithRepo("test/dummy"))

		org := svc.createOrganization(t, ctx)
		vcsprov := svc.createVCSProvider(t, ctx, org)
		ws, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         internal.String(uuid.NewString()),
			Organization: &org.Name,
			ConnectOptions: &workspace.ConnectOptions{
				RepoPath:      "test/dummy",
				VCSProviderID: vcsprov.ID,
			},
		})
		require.NoError(t, err)

		// webhook should be registered with github
		require.True(t, svc.HasWebhook())

		_, err = svc.DeleteWorkspace(ctx, ws.ID)
		require.NoError(t, err)

		// webhook should now have been deleted from github
		require.False(t, svc.HasWebhook())
	})

	t.Run("connect workspace", func(t *testing.T) {
		svc := setup(t, nil, github.WithRepo("test/dummy"))

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
		sub := svc.createSubscriber(t, ctx)
		org := svc.createOrganization(t, ctx)
		ws := svc.createWorkspace(t, ctx, org)
		assert.Equal(t, pubsub.NewCreatedEvent(org), <-sub)
		assert.Equal(t, pubsub.NewCreatedEvent(ws), <-sub)

		got, err := svc.UpdateWorkspace(ctx, ws.ID, workspace.UpdateOptions{
			Description: internal.String("updated description"),
		})
		require.NoError(t, err)
		assert.Equal(t, "updated description", got.Description)
		assert.Equal(t, pubsub.NewUpdatedEvent(got), <-sub)

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
		wsTagged, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
			Organization: internal.String(org.Name),
			Name:         internal.String("ws-tagged"),
			Tags:         []workspace.TagSpec{{Name: "foo"}, {Name: "bar"}},
		})
		require.NoError(t, err)

		tests := []struct {
			name string
			opts workspace.ListOptions
			want func(*testing.T, *workspace.WorkspaceList)
		}{
			{
				name: "filter by org",
				opts: workspace.ListOptions{Organization: internal.String(org.Name)},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 3, len(l.Items))
					assert.Contains(t, l.Items, ws1)
					assert.Contains(t, l.Items, ws2)
				},
			},
			{
				name: "filter by name regex",
				// test workspaces are named `workspace-<random 6 alphanumerals>`, so prefix with 14
				// characters to be pretty damn sure only ws1 is selected.
				opts: workspace.ListOptions{Organization: internal.String(org.Name), Search: ws1.Name[:14]},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 1, len(l.Items))
					assert.Equal(t, ws1, l.Items[0])
				},
			},
			{
				name: "filter by tag",
				opts: workspace.ListOptions{Tags: []string{"foo", "bar"}},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 1, len(l.Items))
					assert.Equal(t, wsTagged, l.Items[0])
				},
			},
			{
				name: "filter by non-existent org",
				opts: workspace.ListOptions{Organization: internal.String("non-existent")},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "filter by non-existent name regex",
				opts: workspace.ListOptions{Organization: internal.String(org.Name), Search: "xyz"},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "paginated results ordered by updated_at",
				opts: workspace.ListOptions{Organization: internal.String(org.Name), ListOptions: internal.ListOptions{PageNumber: 1, PageSize: 1}},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 1, len(l.Items))
					// results are in descending order so we expect wsTagged to be listed
					// first...unless - and this happens very occasionally - the
					// updated_at time is equal down to nearest millisecond.
					if !ws2.UpdatedAt.Equal(wsTagged.UpdatedAt) {
						assert.Equal(t, wsTagged, l.Items[0])
					}
					assert.Equal(t, 3, l.TotalCount())
				},
			},
			{
				name: "stray pagination",
				opts: workspace.ListOptions{Organization: internal.String(org.Name), ListOptions: internal.ListOptions{PageNumber: 999, PageSize: 10}},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					// zero results but count should ignore pagination
					assert.Equal(t, 0, len(l.Items))
					assert.Equal(t, 3, l.TotalCount())
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

	t.Run("list by tag", func(t *testing.T) {
		svc := setup(t, nil)
		org := svc.createOrganization(t, ctx)
		ws1, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         internal.String(uuid.NewString()),
			Organization: &org.Name,
			Tags:         []workspace.TagSpec{{Name: "foo"}},
		})
		require.NoError(t, err)
		ws2, err := svc.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         internal.String(uuid.NewString()),
			Organization: &org.Name,
			Tags:         []workspace.TagSpec{{Name: "foo"}, {Name: "bar"}},
		})
		require.NoError(t, err)

		tests := []struct {
			name string
			tags []string
			want func(*testing.T, *workspace.WorkspaceList)
		}{
			{
				name: "foo",
				tags: []string{"foo"},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, ws1)
					assert.Contains(t, l.Items, ws2)
				},
			},
			{
				name: "foo and bar",
				tags: []string{"foo", "bar"},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 1, len(l.Items))
					assert.Contains(t, l.Items, ws2)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				results, err := svc.ListWorkspaces(ctx, workspace.ListOptions{
					Tags: tt.tags,
				})
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
				opts: workspace.ListOptions{Organization: internal.String(org.Name)},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, ws1)
					assert.Contains(t, l.Items, ws2)
				},
			},
			{
				name: "query non-existent org",
				user: user1,
				opts: workspace.ListOptions{Organization: internal.String("acme-corp")},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "user with no perms",
				user: user2,
				opts: workspace.ListOptions{Organization: internal.String(org.Name)},
				want: func(t *testing.T, l *workspace.WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "paginated results ordered by updated_at",
				user: user1,
				opts: workspace.ListOptions{
					Organization: internal.String(org.Name),
					ListOptions:  internal.ListOptions{PageNumber: 1, PageSize: 1},
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
				ctx := internal.AddSubjectToContext(ctx, tt.user)
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
		sub := svc.createSubscriber(t, ctx)
		org := svc.createOrganization(t, ctx)
		ws := svc.createWorkspace(t, ctx, org)
		assert.Equal(t, pubsub.NewCreatedEvent(org), <-sub)
		assert.Equal(t, pubsub.NewCreatedEvent(ws), <-sub)

		_, err := svc.DeleteWorkspace(ctx, ws.ID)
		require.NoError(t, err)
		assert.Equal(t, pubsub.NewDeletedEvent(&workspace.Workspace{ID: ws.ID}), <-sub)

		results, err := svc.ListWorkspaces(ctx, workspace.ListOptions{Organization: internal.String(ws.Organization)})
		require.NoError(t, err)
		assert.Equal(t, 0, len(results.Items))
	})
}
