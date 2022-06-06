package sql

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateVersion_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)

	sv := otf.NewTestStateVersion(t,
		otf.StateOutput{"out1", "string", "val1", false},
		otf.StateOutput{"out2", "string", "val2", false},
	)

	err := db.CreateStateVersion(context.Background(), ws.ID(), sv)
	require.NoError(t, err)
}

func TestStateVersion_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	sv := createTestStateVersion(t, db, ws,
		otf.StateOutput{"out1", "string", "val1", false},
	)

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
			got, err := db.GetStateVersion(context.Background(), tt.opts)
			tt.want(t, got, err)
		})
	}
}

func TestStateVersion_List(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	sv1 := createTestStateVersion(t, db, ws)
	sv2 := createTestStateVersion(t, db, ws)

	tests := []struct {
		name string
		opts otf.StateVersionListOptions
		want func(*testing.T, *otf.StateVersionList, ...*otf.StateVersion)
	}{
		{
			name: "filter by workspace",
			opts: otf.StateVersionListOptions{Workspace: otf.String(ws.Name()), Organization: otf.String(org.Name())},
			want: func(t *testing.T, l *otf.StateVersionList, created ...*otf.StateVersion) {
				assert.Equal(t, 2, len(l.Items))
				for _, c := range created {
					assert.Contains(t, l.Items, c)
				}
			},
		},
		{
			name: "filter by non-existent workspace",
			opts: otf.StateVersionListOptions{Workspace: otf.String("non-existent"), Organization: otf.String("non-existent")},
			want: func(t *testing.T, l *otf.StateVersionList, created ...*otf.StateVersion) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.ListStateVersions(context.Background(), tt.opts)
			require.NoError(t, err)

			tt.want(t, results, sv1, sv2)
		})
	}
}
