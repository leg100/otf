package state

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

func TestStateVersion_Create(t *testing.T) {
	ctx := context.Background()
	db := &pgdb{sql.NewTestDB(t)}

	t.Run("create", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)

		sv := newTestVersion(t, ws,
			"out1", `"val1"`,
			"out2", `"val2"`,
		)

		err := stateDB.createVersion(ctx, sv)
		require.NoError(t, err)
	})

	t.Run("get", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name())
		sv := createTestStateVersion(t, stateDB, ws, "out1", `"val1"`)
		sv := createTestStateVersion(t, stateDB, ws, "out1", `"val1"`)

		tests := []struct {
			name string
			opts stateVersionGetOptions
			want func(t *testing.T, got *version, err error)
		}{
			{
				name: "by id",
				opts: stateVersionGetOptions{ID: otf.String(sv.ID)},
				want: func(t *testing.T, got *version, err error) {
					if assert.NoError(t, err) {
						assert.Equal(t, sv, got)
					}
				},
			},
			{
				name: "by id - missing",
				opts: stateVersionGetOptions{ID: otf.String("sv-does-not-exist")},
				want: func(t *testing.T, got *version, err error) {
					assert.Equal(t, otf.ErrResourceNotFound, err)
				},
			},
			{
				name: "by workspace",
				opts: stateVersionGetOptions{WorkspaceID: otf.String(ws.ID)},
				want: func(t *testing.T, got *version, err error) {
					if assert.NoError(t, err) {
						assert.Equal(t, sv, got)
					}
				},
			},
			{
				name: "by workspace - missing",
				opts: stateVersionGetOptions{WorkspaceID: otf.String("ws-does-not-exist")},
				want: func(t *testing.T, got *version, err error) {
					assert.Equal(t, otf.ErrResourceNotFound, err)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := stateDB.getVersion(ctx, tt.opts)
				tt.want(t, got, err)
			})
		}
	})

	t.Run("list", func(t *testing.T) {
		org := sql.CreateTestOrganization(t, db)
		ws := sql.CreateTestWorkspace(t, db, org)

		sv1 := createTestStateVersion(t, stateDB, ws)
		sv2 := createTestStateVersion(t, stateDB, ws)

		tests := []struct {
			name string
			opts stateVersionListOptions
			want func(*testing.T, *versionList, ...*version)
		}{
			{
				name: "filter by workspace",
				opts: stateVersionListOptions{Workspace: ws.Name(), Organization: org.Name()},
				want: func(t *testing.T, l *versionList, created ...*version) {
					assert.Equal(t, 2, len(l.Items))
					for _, c := range created {
						assert.Contains(t, l.Items, c)
					}
				},
			},
			{
				name: "filter by non-existent workspace",
				opts: stateVersionListOptions{Workspace: "non-existent", Organization: "non-existent"},
				want: func(t *testing.T, l *versionList, created ...*version) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				results, err := stateDB.listVersions(ctx, tt.opts)
				require.NoError(t, err)

				tt.want(t, results, sv1, sv2)
			})
		}
	})
}

func createTestStateVersion(t *testing.T, stateDB *pgdb, ws *otf.Workspace, outputKVs ...string) *version {
	ctx := context.Background()
	sv := newTestVersion(t, ws, outputKVs...)
	err := stateDB.createVersion(ctx, sv)
	require.NoError(t, err)
	t.Cleanup(func() {
		stateDB.deleteVersion(ctx, sv.ID)
	})
	return sv
}
