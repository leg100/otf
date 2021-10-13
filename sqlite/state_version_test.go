package sqlite

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateVersion_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db, "org-123", "automatize")
	ws := createTestWorkspace(t, db, "ws-123", "default", org)

	sdb := NewStateVersionDB(db)

	out1 := appendOutput("svo-1", "out1", "string", "val1", false)
	out2 := appendOutput("svo-2", "out2", "string", "val2", false)

	sv, err := sdb.Create(newTestStateVersion("sv-123", ws, out1, out2))
	require.NoError(t, err)

	assert.Equal(t, int64(1), sv.Model.ID)
}

func TestStateVersion_Get(t *testing.T) {
	tests := []struct {
		name string
		opts otf.StateVersionGetOptions
	}{
		{
			name: "by id",
			opts: otf.StateVersionGetOptions{ID: otf.String("cv-123")},
		},
		{
			name: "by workspace",
			opts: otf.StateVersionGetOptions{WorkspaceID: otf.String("ws-123")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			org := createTestOrganization(t, db, "org-123", "automatize")
			ws := createTestWorkspace(t, db, "ws-123", "default", org)
			_ = createTestStateVersion(t, db, "cv-123", ws,
				appendOutput("svo-1", "out1", "string", "val1", false),
				appendOutput("svo-2", "out2", "string", "val2", false),
			)

			sdb := NewStateVersionDB(db)

			sv, err := sdb.Get(tt.opts)
			require.NoError(t, err)

			assert.Equal(t, int64(1), sv.Model.ID)
			assert.Equal(t, 2, len(sv.Outputs))
		})
	}
}

func TestStateVersion_List(t *testing.T) {
	tests := []struct {
		name string
		opts otf.StateVersionListOptions
		want int
	}{
		{
			name: "filter by workspace",
			opts: otf.StateVersionListOptions{Workspace: otf.String("default"), Organization: otf.String("automatize")},
			want: 1,
		},
		{
			name: "filter by non-existent workspace",
			opts: otf.StateVersionListOptions{Workspace: otf.String("non-existent"), Organization: otf.String("automatize")},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			org := createTestOrganization(t, db, "org-123", "automatize")
			ws := createTestWorkspace(t, db, "ws-123", "default", org)
			_ = createTestStateVersion(t, db, "sv-123", ws)

			sdb := NewStateVersionDB(db)

			results, err := sdb.List(tt.opts)
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(results.Items))
		})
	}
}
