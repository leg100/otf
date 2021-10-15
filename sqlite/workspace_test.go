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
			_ = createTestWorkspace(t, db, "ws-123", "default", org)

			wdb := NewWorkspaceDB(db)

			_, err := wdb.Update(tt.spec, func(ws *otf.Workspace) error {
				ws.Description = "updated description"
				return nil
			})
			require.NoError(t, err)

			got, err := wdb.Get(tt.spec)
			require.NoError(t, err)

			assert.Equal(t, "updated description", got.Description)
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
			ws := createTestWorkspace(t, db, "ws-123", "default", org)

			wdb := NewWorkspaceDB(db)

			got, err := wdb.Get(tt.spec)
			require.NoError(t, err)

			assert.Equal(t, ws, got)
		})
	}
}

func TestWorkspace_List(t *testing.T) {
	tests := []struct {
		name string
		opts otf.WorkspaceListOptions
		want func(*testing.T, *otf.WorkspaceList, ...*otf.Workspace)
	}{
		{
			name: "default",
			opts: otf.WorkspaceListOptions{},
			want: func(t *testing.T, l *otf.WorkspaceList, created ...*otf.Workspace) {
				assert.Equal(t, 2, len(l.Items))
				assert.Equal(t, created, l.Items)
			},
		},
		{
			name: "filter by org",
			opts: otf.WorkspaceListOptions{OrganizationName: otf.String("automatize")},
			want: func(t *testing.T, l *otf.WorkspaceList, created ...*otf.Workspace) {
				assert.Equal(t, 2, len(l.Items))
				assert.Equal(t, created, l.Items)
			},
		},
		{
			name: "filter by prefix",
			opts: otf.WorkspaceListOptions{Prefix: otf.String("dev")},
			want: func(t *testing.T, l *otf.WorkspaceList, created ...*otf.Workspace) {
				assert.Equal(t, 1, len(l.Items))
			},
		},
		{
			name: "filter by non-existent org",
			opts: otf.WorkspaceListOptions{OrganizationName: otf.String("non-existent")},
			want: func(t *testing.T, l *otf.WorkspaceList, created ...*otf.Workspace) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
		{
			name: "filter by non-existent prefix",
			opts: otf.WorkspaceListOptions{Prefix: otf.String("xyz")},
			want: func(t *testing.T, l *otf.WorkspaceList, created ...*otf.Workspace) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			org := createTestOrganization(t, db, "org-123", "automatize")
			ws1 := createTestWorkspace(t, db, "ws-1", "dev", org)
			ws2 := createTestWorkspace(t, db, "ws-2", "prod", org)

			wdb := NewWorkspaceDB(db)

			results, err := wdb.List(tt.opts)
			require.NoError(t, err)

			tt.want(t, results, ws1, ws2)
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
			ws := createTestWorkspace(t, db, "ws-123", "default", org)
			cv := createTestConfigurationVersion(t, db, "cv-123", ws)
			_ = createTestRun(t, db, "run-123", ws, cv)

			wdb := NewWorkspaceDB(db)
			rdb := NewRunDB(db)
			cdb := NewConfigurationVersionDB(db)

			require.NoError(t, wdb.Delete(tt.spec))

			results, err := wdb.List(otf.WorkspaceListOptions{})
			require.NoError(t, err)

			assert.Equal(t, 0, len(results.Items))

			// Test ON CASCADE DELETE functionality for runs
			rl, err := rdb.List(otf.RunListOptions{})
			require.NoError(t, err)

			assert.Equal(t, 0, len(rl.Items))

			// Test ON CASCADE DELETE functionality for config versions
			cvl, err := cdb.List(ws.ID, otf.ConfigurationVersionListOptions{})
			require.NoError(t, err)

			assert.Equal(t, 0, len(cvl.Items))
		})
	}
}
