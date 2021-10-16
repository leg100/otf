package sql

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	wdb := NewWorkspaceDB(db)

	ws, err := wdb.Create(newTestWorkspace(org))
	require.NoError(t, err)

	assert.Equal(t, int64(1), ws.Model.ID)
}

func TestWorkspace_Update(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	tests := []struct {
		name string
		spec func(ws *otf.Workspace) otf.WorkspaceSpecifier
	}{
		{
			name: "by id",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpecifier {
				return otf.WorkspaceSpecifier{ID: otf.String(ws.ID)}
			},
		},
		{
			name: "by name",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpecifier {
				return otf.WorkspaceSpecifier{Name: otf.String(ws.Name), OrganizationName: otf.String(org.Name)}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := createTestWorkspace(t, db, org)

			wdb := NewWorkspaceDB(db)

			_, err := wdb.Update(tt.spec(ws), func(ws *otf.Workspace) error {
				ws.Description = "updated description"
				return nil
			})
			require.NoError(t, err)

			got, err := wdb.Get(tt.spec(ws))
			require.NoError(t, err)

			assert.Equal(t, "updated description", got.Description)
		})
	}
}

func TestWorkspace_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)

	wdb := NewWorkspaceDB(db)

	tests := []struct {
		name string
		spec otf.WorkspaceSpecifier
	}{
		{
			name: "by id",
			spec: otf.WorkspaceSpecifier{ID: otf.String(ws.ID)},
		},
		{
			name: "by name",
			spec: otf.WorkspaceSpecifier{Name: otf.String(ws.Name), OrganizationName: otf.String(org.Name)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := wdb.Get(tt.spec)
			require.NoError(t, err)

			assert.Equal(t, ws, got)
		})
	}
}

func TestWorkspace_List(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws1 := createTestWorkspace(t, db, org)
	ws2 := createTestWorkspace(t, db, org)

	wdb := NewWorkspaceDB(db)

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
			opts: otf.WorkspaceListOptions{OrganizationName: otf.String(org.Name)},
			want: func(t *testing.T, l *otf.WorkspaceList, created ...*otf.Workspace) {
				assert.Equal(t, 2, len(l.Items))
				assert.Equal(t, created, l.Items)
			},
		},
		{
			name: "filter by prefix",
			opts: otf.WorkspaceListOptions{Prefix: otf.String(ws1.Name[:2])},
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
			results, err := wdb.List(tt.opts)
			require.NoError(t, err)

			tt.want(t, results, ws1, ws2)
		})
	}
}

func TestWorkspace_Delete(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	wdb := NewWorkspaceDB(db)
	rdb := NewRunDB(db)
	cdb := NewConfigurationVersionDB(db)

	tests := []struct {
		name string
		spec func(ws *otf.Workspace) otf.WorkspaceSpecifier
	}{
		{
			name: "by id",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpecifier {
				return otf.WorkspaceSpecifier{ID: otf.String(ws.ID)}
			},
		},
		{
			name: "by name",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpecifier {
				return otf.WorkspaceSpecifier{Name: otf.String(ws.Name), OrganizationName: otf.String(ws.Organization.Name)}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := createTestWorkspace(t, db, org)
			cv := createTestConfigurationVersion(t, db, ws)
			_ = createTestRun(t, db, ws, cv)

			err := wdb.Delete(tt.spec(ws))
			require.NoError(t, err)

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
