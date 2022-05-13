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
	ws := newTestWorkspace(org)

	_, err := db.WorkspaceStore().Create(ws)
	require.NoError(t, err)

	t.Run("Duplicate", func(t *testing.T) {
		_, err := db.WorkspaceStore().Create(ws)
		require.Equal(t, otf.ErrResourcesAlreadyExists, err)
	})
}

func TestWorkspace_Update(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	tests := []struct {
		name string
		spec func(ws *otf.Workspace) otf.WorkspaceSpec
	}{
		{
			name: "by id",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpec {
				return otf.WorkspaceSpec{ID: otf.String(ws.ID)}
			},
		},
		{
			name: "by name",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpec {
				return otf.WorkspaceSpec{Name: otf.String(ws.Name), OrganizationName: otf.String(org.Name)}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := createTestWorkspace(t, db, org)

			_, err := db.WorkspaceStore().Update(tt.spec(ws), func(ws *otf.Workspace) error {
				ws.Description = "updated description"
				return nil
			})
			require.NoError(t, err)

			got, err := db.WorkspaceStore().Get(tt.spec(ws))
			require.NoError(t, err)

			assert.Equal(t, "updated description", got.Description)
		})
	}
}

func TestWorkspace_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)

	tests := []struct {
		name string
		spec otf.WorkspaceSpec
	}{
		{
			name: "by id",
			spec: otf.WorkspaceSpec{ID: otf.String(ws.ID)},
		},
		{
			name: "by name",
			spec: otf.WorkspaceSpec{Name: otf.String(ws.Name), OrganizationName: otf.String(org.Name)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.WorkspaceStore().Get(tt.spec)
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

	tests := []struct {
		name string
		opts otf.WorkspaceListOptions
		want func(*testing.T, *otf.WorkspaceList)
	}{
		{
			name: "filter by org",
			opts: otf.WorkspaceListOptions{OrganizationName: otf.String(org.Name), Prefix: otf.String("")},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Contains(t, l.Items, ws1)
				assert.Contains(t, l.Items, ws2)
			},
		},
		{
			name: "filter by prefix",
			opts: otf.WorkspaceListOptions{Prefix: otf.String(ws1.Name[:5])},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 1, len(l.Items))
				assert.Equal(t, ws1, l.Items[0])
			},
		},
		{
			name: "filter by non-existent org",
			opts: otf.WorkspaceListOptions{OrganizationName: otf.String("non-existent")},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
		{
			name: "filter by non-existent prefix",
			opts: otf.WorkspaceListOptions{Prefix: otf.String("xyz")},
			want: func(t *testing.T, l *otf.WorkspaceList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.WorkspaceStore().List(tt.opts)
			require.NoError(t, err)

			tt.want(t, results)
		})
	}
}

func TestWorkspace_Delete(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	tests := []struct {
		name string
		spec func(ws *otf.Workspace) otf.WorkspaceSpec
	}{
		{
			name: "by id",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpec {
				return otf.WorkspaceSpec{ID: otf.String(ws.ID)}
			},
		},
		{
			name: "by name",
			spec: func(ws *otf.Workspace) otf.WorkspaceSpec {
				return otf.WorkspaceSpec{Name: otf.String(ws.Name), OrganizationName: otf.String(ws.Organization.Name)}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws := createTestWorkspace(t, db, org)
			cv := createTestConfigurationVersion(t, db, ws)
			_ = createTestRun(t, db, ws, cv)

			err := db.WorkspaceStore().Delete(tt.spec(ws))
			require.NoError(t, err)

			results, err := db.WorkspaceStore().List(otf.WorkspaceListOptions{OrganizationName: otf.String(org.Name)})
			require.NoError(t, err)

			assert.Equal(t, 0, len(results.Items))

			// Test ON CASCADE DELETE functionality for runs
			rl, err := db.RunStore().List(otf.RunListOptions{WorkspaceID: otf.String(ws.ID)})
			require.NoError(t, err)

			assert.Equal(t, 0, len(rl.Items))

			// Test ON CASCADE DELETE functionality for config versions
			cvl, err := db.ConfigurationVersionStore().List(ws.ID, otf.ConfigurationVersionListOptions{})
			require.NoError(t, err)

			assert.Equal(t, 0, len(cvl.Items))
		})
	}
}
