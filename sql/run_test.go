package sql

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_Create(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)

	rdb := NewRunDB(db)

	run, err := rdb.Create(newTestRun(ws, cv))
	require.NoError(t, err)

	// Ensure primary keys populated
	assert.NotEmpty(t, run.ID)
	assert.NotEmpty(t, run.Plan.ID)
	assert.NotEmpty(t, run.Apply.ID)

	// Ensure foreign keys populated
	assert.NotEmpty(t, run.Plan.RunID)
	assert.NotEmpty(t, run.Apply.RunID)
	assert.NotEmpty(t, run.Workspace.ID)
	assert.NotEmpty(t, run.ConfigurationVersion.ID)
}

func TestRun_Update(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)
	run := createTestRun(t, db, ws, cv)

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
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)
	run := createTestRun(t, db, ws, cv)

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
	db := newTestDB(t)
	org := createTestOrganization(t, db)

	ws1 := createTestWorkspace(t, db, org)
	cv1 := createTestConfigurationVersion(t, db, ws1)
	run1 := createTestRun(t, db, ws1, cv1)

	ws2 := createTestWorkspace(t, db, org)
	cv2 := createTestConfigurationVersion(t, db, ws2)
	run2 := createTestRun(t, db, ws2, cv2)

	rdb := NewRunDB(db)

	tests := []struct {
		name string
		opts otf.RunListOptions
		want func(*testing.T, *otf.RunList, ...*otf.Run)
	}{
		{
			name: "default",
			opts: otf.RunListOptions{},
			want: func(t *testing.T, l *otf.RunList, created ...*otf.Run) {
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
			opts: otf.RunListOptions{WorkspaceID: otf.String(ws1.ID)},
			want: func(t *testing.T, l *otf.RunList, created ...*otf.Run) {
				assert.Equal(t, 1, len(l.Items))
			},
		},
		{
			name: "filter by status - hit",
			opts: otf.RunListOptions{WorkspaceID: otf.String(ws1.ID), Statuses: []otf.RunStatus{otf.RunPending}},
			want: func(t *testing.T, l *otf.RunList, created ...*otf.Run) {
				assert.Equal(t, 1, len(l.Items))
			},
		},
		{
			name: "filter by status - miss",
			opts: otf.RunListOptions{WorkspaceID: otf.String(ws1.ID), Statuses: []otf.RunStatus{otf.RunApplied}},
			want: func(t *testing.T, l *otf.RunList, created ...*otf.Run) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := rdb.List(tt.opts)
			require.NoError(t, err)

			tt.want(t, results, run1, run2)
		})
	}
}
