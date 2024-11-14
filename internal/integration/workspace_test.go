package integration

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		daemon, org, ctx := setup(t, nil)

		// watch workspace events
		sub, unsub := daemon.Workspaces.Watch(ctx)
		defer unsub()

		ws, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
			Name:         internal.String(uuid.NewString()),
			Organization: internal.String(org.Name),
		})
		require.NoError(t, err)

		t.Run("duplicate error", func(t *testing.T) {
			_, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
				Name:         internal.String(ws.Name),
				Organization: internal.String(org.Name),
			})
			require.Equal(t, internal.ErrResourceAlreadyExists, err)
		})

		t.Run("receive events", func(t *testing.T) {
			assert.Equal(t, pubsub.NewCreatedEvent(ws), <-sub)
		})
	})

	t.Run("create connected workspace", func(t *testing.T) {
		daemon, org, ctx := setup(t, nil, github.WithRepo("test/dummy"))

		vcsprov := daemon.createVCSProvider(t, ctx, org)
		ws, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
			Name:         internal.String(uuid.NewString()),
			Organization: &org.Name,
			ConnectOptions: &workspace.ConnectOptions{
				RepoPath:      internal.String("test/dummy"),
				VCSProviderID: &vcsprov.ID,
			},
		})
		require.NoError(t, err)

		// webhook should be registered with github
		hook := <-daemon.WebhookEvents
		require.Equal(t, github.WebhookCreated, hook.Action)

		t.Run("delete workspace connection", func(t *testing.T) {
			err := daemon.Connections.Disconnect(ctx, connections.DisconnectOptions{
				ResourceID: ws.ID,
			})
			require.NoError(t, err)
		})

		// webhook should now have been deleted from github
		hook = <-daemon.WebhookEvents
		require.Equal(t, github.WebhookDeleted, hook.Action)
	})

	t.Run("deleting connected workspace also deletes webhook", func(t *testing.T) {
		svc, org, ctx := setup(t, nil, github.WithRepo("test/dummy"))

		vcsprov := svc.createVCSProvider(t, ctx, org)
		ws, err := svc.Workspaces.Create(ctx, workspace.CreateOptions{
			Name:         internal.String(uuid.NewString()),
			Organization: &org.Name,
			ConnectOptions: &workspace.ConnectOptions{
				RepoPath:      internal.String("test/dummy"),
				VCSProviderID: &vcsprov.ID,
			},
		})
		require.NoError(t, err)

		// webhook should be registered with github
		hook := <-svc.WebhookEvents
		require.Equal(t, github.WebhookCreated, hook.Action)

		_, err = svc.Workspaces.Delete(ctx, ws.ID)
		require.NoError(t, err)

		// webhook should now have been deleted from github
		hook = <-svc.WebhookEvents
		require.Equal(t, github.WebhookDeleted, hook.Action)
	})

	t.Run("connect workspace", func(t *testing.T) {
		svc, org, ctx := setup(t, nil, github.WithRepo("test/dummy"))

		ws := svc.createWorkspace(t, ctx, org)
		vcsprov := svc.createVCSProvider(t, ctx, org)

		got, err := svc.Connections.Connect(ctx, connections.ConnectOptions{
			VCSProviderID: vcsprov.ID,
			ResourceID:    ws.ID,
			RepoPath:      "test/dummy",
		})
		require.NoError(t, err)
		want := &connections.Connection{VCSProviderID: vcsprov.ID, Repo: "test/dummy"}
		assert.Equal(t, want, got)

		t.Run("delete workspace connection", func(t *testing.T) {
			err := svc.Connections.Disconnect(ctx, connections.DisconnectOptions{
				ResourceID: ws.ID,
			})
			require.NoError(t, err)
		})
	})

	t.Run("update", func(t *testing.T) {
		daemon, org, ctx := setup(t, nil)

		// watch workspace events
		sub, unsub := daemon.Workspaces.Watch(ctx)
		defer unsub()

		ws := daemon.createWorkspace(t, ctx, org)
		assert.Equal(t, pubsub.NewCreatedEvent(ws), <-sub)

		got, err := daemon.Workspaces.Update(ctx, ws.ID, workspace.UpdateOptions{
			Description: internal.String("updated description"),
		})
		require.NoError(t, err)
		assert.Equal(t, "updated description", got.Description)
		assert.Equal(t, pubsub.NewUpdatedEvent(got), <-sub)

		// assert too that the WS returned by UpdateWorkspace is identical to one
		// returned by GetWorkspace
		want, err := daemon.Workspaces.Get(ctx, ws.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("get by id", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		want := svc.createWorkspace(t, ctx, nil)

		got, err := svc.Workspaces.Get(ctx, want.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("get by name", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		want := svc.createWorkspace(t, ctx, nil)

		got, err := svc.Workspaces.GetByName(ctx, want.Organization, want.Name)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)
		ws1 := svc.createWorkspace(t, ctx, org)
		ws2 := svc.createWorkspace(t, ctx, org)
		wsTagged, err := svc.Workspaces.Create(ctx, workspace.CreateOptions{
			Organization: internal.String(org.Name),
			Name:         internal.String("ws-tagged"),
			Tags:         []workspace.TagSpec{{Name: "foo"}, {Name: "bar"}},
		})
		require.NoError(t, err)

		tests := []struct {
			name string
			opts workspace.ListOptions
			want func(*testing.T, *resource.Page[*workspace.Workspace])
		}{
			{
				name: "filter by org",
				opts: workspace.ListOptions{Organization: internal.String(org.Name)},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
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
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 1, len(l.Items))
					assert.Equal(t, ws1, l.Items[0])
				},
			},
			{
				name: "filter by tag",
				opts: workspace.ListOptions{Tags: []string{"foo", "bar"}},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 1, len(l.Items))
					assert.Equal(t, wsTagged, l.Items[0])
				},
			},
			{
				name: "filter by non-existent org",
				opts: workspace.ListOptions{Organization: internal.String("non-existent")},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "filter by non-existent name regex",
				opts: workspace.ListOptions{Organization: internal.String(org.Name), Search: "xyz"},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "paginated results ordered by updated_at",
				opts: workspace.ListOptions{Organization: internal.String(org.Name), PageOptions: resource.PageOptions{PageNumber: 1, PageSize: 1}},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 1, len(l.Items))
					// results are in descending order so we expect wsTagged to be listed
					// first...unless - and this happens very occasionally - the
					// updated_at time is equal down to nearest millisecond.
					if !ws2.UpdatedAt.Equal(wsTagged.UpdatedAt) {
						assert.Equal(t, wsTagged, l.Items[0])
					}
					assert.Equal(t, 3, l.TotalCount)
				},
			},
			{
				name: "stray pagination",
				opts: workspace.ListOptions{Organization: internal.String(org.Name), PageOptions: resource.PageOptions{PageNumber: 999, PageSize: 10}},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					// zero results but count should ignore pagination
					assert.Equal(t, 0, len(l.Items))
					assert.Equal(t, 3, l.TotalCount)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// call endpoint using admin to avoid authz errors.
				results, err := svc.Workspaces.List(adminCtx, tt.opts)
				require.NoError(t, err)

				tt.want(t, results)
			})
		}
	})

	t.Run("list by tag", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)
		ws1, err := svc.Workspaces.Create(ctx, workspace.CreateOptions{
			Name:         internal.String(uuid.NewString()),
			Organization: &org.Name,
			Tags:         []workspace.TagSpec{{Name: "foo"}},
		})
		require.NoError(t, err)
		ws2, err := svc.Workspaces.Create(ctx, workspace.CreateOptions{
			Name:         internal.String(uuid.NewString()),
			Organization: &org.Name,
			Tags:         []workspace.TagSpec{{Name: "foo"}, {Name: "bar"}},
		})
		require.NoError(t, err)

		tests := []struct {
			name string
			tags []string
			want func(*testing.T, *resource.Page[*workspace.Workspace])
		}{
			{
				name: "foo",
				tags: []string{"foo"},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, ws1)
					assert.Contains(t, l.Items, ws2)

					// check pagination metadata
					assert.Equal(t, 2, l.TotalCount)
				},
			},
			{
				name: "foo and bar",
				tags: []string{"foo", "bar"},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 1, len(l.Items))
					assert.Contains(t, l.Items, ws2)

					// check pagination metadata
					assert.Equal(t, 1, l.TotalCount)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				results, err := svc.Workspaces.List(ctx, workspace.ListOptions{
					Organization: &org.Name,
					Tags:         tt.tags,
				})
				require.NoError(t, err)

				tt.want(t, results)
			})
		}
	})

	t.Run("list by user", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)
		ws1 := svc.createWorkspace(t, ctx, org)
		ws2 := svc.createWorkspace(t, ctx, org)

		team1 := svc.createTeam(t, ctx, org)
		team2 := svc.createTeam(t, ctx, org)
		_ = svc.Workspaces.SetPermission(ctx, ws1.ID, team1.ID, authz.WorkspacePlanRole)
		_ = svc.Workspaces.SetPermission(ctx, ws2.ID, team2.ID, authz.WorkspacePlanRole)
		user1 := svc.createUser(t, user.WithTeams(team1, team2))
		user2 := svc.createUser(t)

		tests := []struct {
			name string
			user *user.User
			opts workspace.ListOptions
			want func(*testing.T, *resource.Page[*workspace.Workspace])
		}{
			{
				name: "show both workspaces",
				user: user1,
				opts: workspace.ListOptions{Organization: internal.String(org.Name)},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, ws1)
					assert.Contains(t, l.Items, ws2)
				},
			},
			{
				name: "query non-existent org",
				user: user1,
				opts: workspace.ListOptions{Organization: internal.String("acme-corp")},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "user with no perms",
				user: user2,
				opts: workspace.ListOptions{Organization: internal.String(org.Name)},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "paginated results ordered by updated_at",
				user: user1,
				opts: workspace.ListOptions{
					Organization: internal.String(org.Name),
					PageOptions:  resource.PageOptions{PageNumber: 1, PageSize: 1},
				},
				want: func(t *testing.T, l *resource.Page[*workspace.Workspace]) {
					assert.Equal(t, 1, len(l.Items))
					// results are in descending order so we expect ws2 to be listed
					// first...unless - and this happens very occasionally - the
					// updated_at time is equal down to nearest millisecond.
					if !ws1.UpdatedAt.Equal(ws2.UpdatedAt) {
						assert.Equal(t, ws2, l.Items[0])
					}
					assert.Equal(t, 2, l.TotalCount)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ctx := authz.AddSubjectToContext(ctx, tt.user)
				results, err := svc.Workspaces.List(ctx, tt.opts)
				require.NoError(t, err)

				tt.want(t, results)
			})
		}
	})

	t.Run("lock", func(t *testing.T) {
		svc, org, ctx := setup(t, nil)
		ws := svc.createWorkspace(t, ctx, org)

		got, err := svc.Workspaces.Lock(ctx, ws.ID, nil)
		require.NoError(t, err)
		assert.True(t, got.Locked())

		t.Run("unlock", func(t *testing.T) {
			got, err := svc.Workspaces.Unlock(ctx, ws.ID, nil, false)
			require.NoError(t, err)
			assert.False(t, got.Locked())
		})
	})

	t.Run("delete", func(t *testing.T) {
		daemon, org, ctx := setup(t, nil)

		// watch workspace events
		sub, unsub := daemon.Workspaces.Watch(ctx)
		defer unsub()

		ws := daemon.createWorkspace(t, ctx, org)
		assert.Equal(t, pubsub.NewCreatedEvent(ws), <-sub)

		_, err := daemon.Workspaces.Delete(ctx, ws.ID)
		require.NoError(t, err)
		assert.Equal(t, pubsub.NewDeletedEvent(&workspace.Workspace{ID: ws.ID}), <-sub)

		results, err := daemon.Workspaces.List(ctx, workspace.ListOptions{Organization: internal.String(ws.Organization)})
		require.NoError(t, err)
		assert.Equal(t, 0, len(results.Items))
	})
}
