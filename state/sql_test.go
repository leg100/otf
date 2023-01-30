package state

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateVersion_Create(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	stateDB := &pgdb{db}
	org := sql.CreateTestOrganization(t, db)
	ws := sql.CreateTestWorkspace(t, db, org)

	sv := newTestVersion(t,
		StateOutput{"out1", "string", "val1", false},
		StateOutput{"out2", "string", "val2", false},
	)

	err := stateDB.createVersion(ctx, sv)
	require.NoError(t, err)
}

func TestStateVersion_Get(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := sql.CreateTestOrganization(t, db)
	ws := sql.CreateTestWorkspace(t, db, org)
	sv := newTestVersion(t,
		StateOutput{"out1", "string", "val1", false},
	)
	err := db.CreateStateVersion(ctx, ws.ID(), sv)
	require.NoError(t, err)

	tests := []struct {
		name string
		opts otf.StateVersionGetOptions
		want func(t *testing.T, got *otf.StateVersion, err error)
	}{
		{
			name: "by id",
			opts: otf.StateVersionGetOptions{ID: otf.String(sv.ID())},
			want: func(t *testing.T, got *otf.StateVersion, err error) {
				if assert.NoError(t, err) {
					assert.Equal(t, sv, got)
				}
			},
		},
		{
			name: "by id - missing",
			opts: otf.StateVersionGetOptions{ID: otf.String("sv-does-not-exist")},
			want: func(t *testing.T, got *otf.StateVersion, err error) {
				assert.Equal(t, otf.ErrResourceNotFound, err)
			},
		},
		{
			name: "by workspace",
			opts: otf.StateVersionGetOptions{WorkspaceID: otf.String(ws.ID())},
			want: func(t *testing.T, got *otf.StateVersion, err error) {
				if assert.NoError(t, err) {
					assert.Equal(t, sv, got)
				}
			},
		},
		{
			name: "by workspace - missing",
			opts: otf.StateVersionGetOptions{WorkspaceID: otf.String("ws-does-not-exist")},
			want: func(t *testing.T, got *otf.StateVersion, err error) {
				assert.Equal(t, otf.ErrResourceNotFound, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetStateVersion(ctx, tt.opts)
			tt.want(t, got, err)
		})
	}
}

func TestStateVersion_List(t *testing.T) {
	ctx := context.Background()
	db := newTestDB(t)
	org := sql.CreateTestOrganization(t, db)
	ws := sql.CreateTestWorkspace(t, db, org)
	sv1 := createTestStateVersion(t, db, ws)
	sv2 := createTestStateVersion(t, db, ws)

	tests := []struct {
		name string
		opts otf.StateVersionListOptions
		want func(*testing.T, *otf.StateVersionList, ...*otf.StateVersion)
	}{
		{
			name: "filter by workspace",
			opts: otf.StateVersionListOptions{Workspace: ws.Name(), Organization: org.Name()},
			want: func(t *testing.T, l *otf.StateVersionList, created ...*otf.StateVersion) {
				assert.Equal(t, 2, len(l.Items))
				for _, c := range created {
					assert.Contains(t, l.Items, c)
				}
			},
		},
		{
			name: "filter by non-existent workspace",
			opts: otf.StateVersionListOptions{Workspace: "non-existent", Organization: "non-existent"},
			want: func(t *testing.T, l *otf.StateVersionList, created ...*otf.StateVersion) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.ListStateVersions(ctx, tt.opts)
			require.NoError(t, err)

			tt.want(t, results, sv1, sv2)
		})
	}
}
