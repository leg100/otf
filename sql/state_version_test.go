package sql

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateVersion_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)

	out1 := appendOutput("out1", "string", "val1", false)
	out2 := appendOutput("out2", "string", "val2", false)

	_, err := db.StateVersionStore().Create(newTestStateVersion(ws, out1, out2))
	require.NoError(t, err)
}

func TestStateVersion_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	sv := createTestStateVersion(t, db, ws,
		appendOutput("out1", "string", "val1", false),
		appendOutput("out2", "string", "val2", false),
	)

	tests := []struct {
		name string
		opts otf.StateVersionGetOptions
	}{
		{
			name: "by id",
			opts: otf.StateVersionGetOptions{ID: otf.String(sv.ID)},
		},
		{
			name: "by workspace",
			opts: otf.StateVersionGetOptions{WorkspaceID: otf.String(ws.ID)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.StateVersionStore().Get(tt.opts)
			require.NoError(t, err)

			// Assertion won't succeed unless both have a workspace with a nil
			// org.
			sv.Workspace.Organization = nil

			assert.Equal(t, sv, got)
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
			opts: otf.StateVersionListOptions{Workspace: otf.String(ws.Name), Organization: otf.String(org.Name)},
			want: func(t *testing.T, l *otf.StateVersionList, created ...*otf.StateVersion) {
				assert.Equal(t, 2, len(l.Items))
				for _, c := range created {
					// Assertion won't succeed unless both have a workspace with
					// a nil org.
					c.Workspace.Organization = nil

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
			results, err := db.StateVersionStore().List(tt.opts)
			require.NoError(t, err)

			tt.want(t, results, sv1, sv2)
		})
	}
}
