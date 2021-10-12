package sqlite

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db, "org-123", "automatize")

	wdb := NewWorkspaceDB(db)

	ws, err := wdb.Create(newTestWorkspace("ws-123", "default", org))
	require.NoError(t, err)

	assert.Equal(t, int64(1), ws.Model.ID)
}

func TestWorkspace_Update(t *testing.T) {
	tests := []struct {
		name string
		spec otf.WorkspaceSpecifier
	}{
		{
			name: "by id",
			spec: otf.WorkspaceSpecifier{ID: otf.String("ws-123")},
		},
		{
			name: "by name",
			spec: otf.WorkspaceSpecifier{Name: otf.String("default"), OrganizationName: otf.String("automatize")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			org := createTestOrganization(t, db, "org-123", "automatize")

			wdb := NewWorkspaceDB(db)

			_, err := wdb.Create(newTestWorkspace("ws-123", "default", org))
			require.NoError(t, err)

			ws, err := wdb.Update(tt.spec, func(ws *otf.Workspace) error {
				ws.Name = "newdefault"
				return nil
			})
			require.NoError(t, err)

			assert.Equal(t, "newdefault", ws.Name)
		})
	}
}

func TestWorkspace_Get(t *testing.T) {
	tests := []struct {
		name string
		spec otf.WorkspaceSpecifier
	}{
		{
			name: "by id",
			spec: otf.WorkspaceSpecifier{ID: otf.String("ws-123")},
		},
		{
			name: "by name",
			spec: otf.WorkspaceSpecifier{Name: otf.String("default"), OrganizationName: otf.String("automatize")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			org := createTestOrganization(t, db, "org-123", "automatize")

			wdb := NewWorkspaceDB(db)

			_, err := wdb.Create(newTestWorkspace("ws-123", "default", org))
			require.NoError(t, err)

			ws, err := wdb.Get(tt.spec)
			require.NoError(t, err)

			assert.Equal(t, int64(1), ws.Model.ID)
		})
	}
}

func TestWorkspace_List(t *testing.T) {
	tests := []struct {
		name string
		opts otf.WorkspaceListOptions
		want int
	}{
		{
			name: "default",
			opts: otf.WorkspaceListOptions{},
			want: 1,
		},
		{
			name: "filter by org",
			opts: otf.WorkspaceListOptions{OrganizationName: otf.String("automatize")},
			want: 1,
		},
		{
			name: "filter by prefix",
			opts: otf.WorkspaceListOptions{Prefix: otf.String("def")},
			want: 1,
		},
		{
			name: "filter by non-existent org",
			opts: otf.WorkspaceListOptions{OrganizationName: otf.String("non-existent")},
			want: 0,
		},
		{
			name: "filter by non-existent prefix",
			opts: otf.WorkspaceListOptions{Prefix: otf.String("xyz")},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			org := createTestOrganization(t, db, "org-123", "automatize")

			wdb := NewWorkspaceDB(db)

			_, err := wdb.Create(newTestWorkspace("ws-123", "default", org))
			require.NoError(t, err)

			results, err := wdb.List(tt.opts)
			require.NoError(t, err)

			assert.Equal(t, tt.want, len(results.Items))
		})
	}
}

func TestWorkspace_Delete(t *testing.T) {
	tests := []struct {
		name string
		spec otf.WorkspaceSpecifier
	}{
		{
			name: "by id",
			spec: otf.WorkspaceSpecifier{ID: otf.String("ws-123")},
		},
		{
			name: "by name",
			spec: otf.WorkspaceSpecifier{Name: otf.String("default"), OrganizationName: otf.String("automatize")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			org := createTestOrganization(t, db, "org-123", "automatize")

			wdb := NewWorkspaceDB(db)

			_, err := wdb.Create(newTestWorkspace("ws-123", "default", org))
			require.NoError(t, err)

			require.NoError(t, wdb.Delete(tt.spec))

			results, err := wdb.List(otf.WorkspaceListOptions{})
			require.NoError(t, err)

			assert.Equal(t, 0, len(results.Items))
		})
	}
}
