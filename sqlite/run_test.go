package sqlite

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db, "org-123", "automatize")
	ws := createTestWorkspace(t, db, "ws-123", "default", org)
	cv := createTestConfigurationVersion(t, db, "cv-123", ws)

	rdb := NewRunDB(db)

	run, err := rdb.Create(newTestRun("run-123", ws, cv))
	require.NoError(t, err)

	// Ensure primary keys populated
	assert.Equal(t, int64(1), run.Model.ID)
	assert.Equal(t, int64(1), run.Plan.Model.ID)
	assert.Equal(t, int64(1), run.Apply.Model.ID)

	// Ensure foreign keys populated
	assert.Equal(t, int64(1), run.Plan.RunID)
	assert.Equal(t, int64(1), run.Apply.RunID)
	assert.Equal(t, int64(1), run.Workspace.Model.ID)
	assert.Equal(t, int64(1), run.ConfigurationVersion.Model.ID)
}

func TestRun_Update(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db, "org-123", "automatize")
	ws := createTestWorkspace(t, db, "ws-123", "default", org)
	cv := createTestConfigurationVersion(t, db, "cv-123", ws)
	run := createTestRun(t, db, "run-123", ws, cv)

	rdb := NewRunDB(db)

	_, err := rdb.Update(run.ID, func(run *otf.Run) error {
		run.Status = otf.RunPlanQueued
		run.Plan.Status = otf.PlanQueued
		return nil
	})
	require.NoError(t, err)

	got, err := rdb.Get(otf.RunGetOptions{ID: otf.String(run.ID)})
	require.NoError(t, err)

	assert.Equal(t, otf.RunPlanQueued, got.Status)
	assert.Equal(t, otf.PlanQueued, got.Plan.Status)
}

func TestRun_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db, "org-123", "automatize")
	ws := createTestWorkspace(t, db, "ws-123", "default", org)
	cv := createTestConfigurationVersion(t, db, "cv-123", ws)
	run := createTestRun(t, db, "run-123", ws, cv)

	rdb := NewRunDB(db)

	got, err := rdb.Get(otf.RunGetOptions{ID: otf.String(run.ID)})
	require.NoError(t, err)

	// Assertion won't succeed unless transitive relations are nil (resources
	// retrieved from the DB only possess immediate relations).
	run.Workspace.Organization = nil
	run.ConfigurationVersion.Workspace = nil
	assert.Equal(t, run, got)
}

func TestRun_List(t *testing.T) {
	tests := []struct {
		name string
		opts otf.RunListOptions
		want func(*testing.T, *otf.RunList, ...*otf.Run)
	}{
		{
			name: "default",
			opts: otf.RunListOptions{},
			want: func(t *testing.T, l *otf.RunList, created ...*otf.Run) {
				assert.Equal(t, 2, len(l.Items))
				for _, c := range created {
					// Assertion won't succeed unless transitive relations are
					// nil (resources retrieved from the DB only possess
					// immediate relations).
					c.Workspace.Organization = nil
					c.ConfigurationVersion.Workspace = nil

					assert.Contains(t, l.Items, c)
				}
			},
		},
		{
			name: "filter by workspace",
			opts: otf.RunListOptions{WorkspaceID: otf.String("ws-123")},
			want: func(t *testing.T, l *otf.RunList, created ...*otf.Run) {
				assert.Equal(t, 1, len(l.Items))
			},
		},
		{
			name: "filter by status - hit",
			opts: otf.RunListOptions{Statuses: []otf.RunStatus{otf.RunPending}},
			want: func(t *testing.T, l *otf.RunList, created ...*otf.Run) {
				assert.Equal(t, 2, len(l.Items))
			},
		},
		{
			name: "filter by status - miss",
			opts: otf.RunListOptions{Statuses: []otf.RunStatus{otf.RunApplied}},
			want: func(t *testing.T, l *otf.RunList, created ...*otf.Run) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestDB(t)
			org := createTestOrganization(t, db, "org-123", "automatize")

			ws1 := createTestWorkspace(t, db, "ws-123", "dev", org)
			cv1 := createTestConfigurationVersion(t, db, "cv-123", ws1)
			run1 := createTestRun(t, db, "run-1", ws1, cv1)

			ws2 := createTestWorkspace(t, db, "ws-345", "prod", org)
			cv2 := createTestConfigurationVersion(t, db, "cv-345", ws2)
			run2 := createTestRun(t, db, "run-2", ws2, cv2)

			rdb := NewRunDB(db)

			results, err := rdb.List(tt.opts)
			require.NoError(t, err)

			tt.want(t, results, run1, run2)
		})
	}
}
