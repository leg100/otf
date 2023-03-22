package workspace

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	ctx := context.Background()
	db := &pgdb{sql.NewTestDB(t)}

	t.Run("create", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := newTestWorkspace(t, org.Name)
		err := db.create(ctx, ws)
		require.NoError(t, err)

		t.Run("duplicate", func(t *testing.T) {
			err := db.create(ctx, ws)
			require.Equal(t, otf.ErrResourceAlreadyExists, err)
		})

		t.Run("duplicate name", func(t *testing.T) {
			dup := newTestWorkspace(t, org.Name)
			dup.Name = ws.Name
			err := db.create(ctx, dup)
			require.Equal(t, otf.ErrResourceAlreadyExists, err)
		})
	})

	t.Run("update", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := CreateTestWorkspace(t, db, org.Name)

		got, err := db.update(ctx, ws.ID, func(ws *Workspace) error {
			return ws.Update(UpdateOptions{
				Description: otf.String("updated description"),
			})
		})
		require.NoError(t, err)
		assert.Equal(t, "updated description", got.Description)

		// assert too that the WS returned by UpdateWorkspace is identical to one
		// returned by GetWorkspace
		want, err := db.get(ctx, ws.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("get by id", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		want := CreateTestWorkspace(t, db, org.Name)

		got, err := db.get(ctx, want.ID)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("get by name", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		want := CreateTestWorkspace(t, db, org.Name)

		got, err := db.getByName(ctx, org.Name, want.Name)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("lock", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		user := auth.CreateTestUser(t, db)
		ctx := otf.AddSubjectToContext(ctx, user)

		t.Run("lock", func(t *testing.T) {
			ws := CreateTestWorkspace(t, db, org.Name)
			lockholder := UserLock{ID: user.ID, Username: user.Username}
			got, err := db.toggleLock(ctx, ws.ID, func(lock *Lock) error {
				return lock.Lock(lockholder)
			})
			require.NoError(t, err)
			assert.True(t, got.Locked())

			t.Run("unlock", func(t *testing.T) {
				got, err := db.toggleLock(ctx, ws.ID, func(lock *Lock) error {
					return lock.Unlock(lockholder, false)
				})
				require.NoError(t, err)
				assert.False(t, got.Locked())
			})
		})
	})

	t.Run("list", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws1 := CreateTestWorkspace(t, db, org.Name)
		ws2 := CreateTestWorkspace(t, db, org.Name)

		tests := []struct {
			name string
			opts ListOptions
			want func(*testing.T, *WorkspaceList)
		}{
			{
				name: "filter by org",
				opts: ListOptions{Organization: otf.String(org.Name)},
				want: func(t *testing.T, l *WorkspaceList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, ws1)
					assert.Contains(t, l.Items, ws2)
				},
			},
			{
				name: "filter by prefix",
				opts: ListOptions{Organization: otf.String(org.Name), Prefix: ws1.Name[:5]},
				want: func(t *testing.T, l *WorkspaceList) {
					assert.Equal(t, 1, len(l.Items))
					assert.Equal(t, ws1, l.Items[0])
				},
			},
			{
				name: "filter by non-existent org",
				opts: ListOptions{Organization: otf.String("non-existent")},
				want: func(t *testing.T, l *WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "filter by non-existent prefix",
				opts: ListOptions{Organization: otf.String(org.Name), Prefix: "xyz"},
				want: func(t *testing.T, l *WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "paginated results ordered by updated_at",
				opts: ListOptions{Organization: otf.String(org.Name), ListOptions: otf.ListOptions{PageNumber: 1, PageSize: 1}},
				want: func(t *testing.T, l *WorkspaceList) {
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
				opts: ListOptions{Organization: otf.String(org.Name), ListOptions: otf.ListOptions{PageNumber: 999, PageSize: 10}},
				want: func(t *testing.T, l *WorkspaceList) {
					// zero results but count should ignore pagination
					assert.Equal(t, 0, len(l.Items))
					assert.Equal(t, 2, l.TotalCount())
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				results, err := db.list(ctx, tt.opts)
				require.NoError(t, err)

				tt.want(t, results)
			})
		}
	})

	t.Run("list by user", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws1 := CreateTestWorkspace(t, db, org.Name)
		ws2 := CreateTestWorkspace(t, db, org.Name)
		team1 := auth.CreateTestTeam(t, db, org)
		team2 := auth.CreateTestTeam(t, db, org)
		_ = CreateTestWorkspacePermission(t, db, ws1, team1, rbac.WorkspaceAdminRole)
		_ = CreateTestWorkspacePermission(t, db, ws2, team2, rbac.WorkspacePlanRole)
		user := auth.CreateTestUser(t, db, auth.WithTeams(team1, team2))

		tests := []struct {
			name         string
			userID       string
			organization string
			opts         otf.ListOptions
			want         func(*testing.T, *WorkspaceList)
		}{
			{
				name:         "show both workspaces",
				userID:       user.ID,
				organization: org.Name,
				want: func(t *testing.T, l *WorkspaceList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, ws1)
					assert.Contains(t, l.Items, ws2)
				},
			},
			{
				name:         "query non-existent org",
				userID:       user.ID,
				organization: "acme-corp",
				want: func(t *testing.T, l *WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name:         "query non-existent user",
				userID:       "mr-invisible",
				organization: org.Name,
				want: func(t *testing.T, l *WorkspaceList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name:         "paginated results ordered by updated_at",
				userID:       user.ID,
				organization: org.Name,
				opts:         otf.ListOptions{PageNumber: 1, PageSize: 1},
				want: func(t *testing.T, l *WorkspaceList) {
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
				results, err := db.listByUserID(ctx, tt.userID, tt.organization, tt.opts)
				require.NoError(t, err)

				tt.want(t, results)
			})
		}
	})

	t.Run("delete", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := CreateTestWorkspace(t, db, org.Name)

		err := db.delete(ctx, ws.ID)
		require.NoError(t, err)

		results, err := db.list(ctx, ListOptions{Organization: otf.String(org.Name)})
		require.NoError(t, err)
		assert.Equal(t, 0, len(results.Items))

		// TODO: Test ON CASCADE DELETE functionality for config versions,
		// runs, etc
	})
}
