package sql

import (
	"context"
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

	_, err := db.RunStore().Create(newTestRun(ws, cv))
	require.NoError(t, err)
}

func TestRun_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)

	want := createTestRun(t, db, ws, cv)

	tests := []struct {
		name string
		opts otf.RunGetOptions
	}{
		{
			name: "by id",
			opts: otf.RunGetOptions{ID: &want.ID},
		},
		{
			name: "by plan id",
			opts: otf.RunGetOptions{PlanID: &want.Plan.ID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.RunStore().Get(tt.opts)
			require.NoError(t, err)

			assert.Equal(t, want, got)
		})
	}
}

func TestRun_CreatePlanReport(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)
	run := createTestRun(t, db, ws, cv)

	report := otf.ResourceReport{
		ResourceAdditions:    5,
		ResourceChanges:      2,
		ResourceDestructions: 99,
	}

	err := db.RunStore().CreatePlanReport(run.Plan.ID, report)
	require.NoError(t, err)

	run, err = db.RunStore().Get(otf.RunGetOptions{ID: &run.ID})
	require.NoError(t, err)

	assert.NotNil(t, run.Plan.ResourceReport)
}

func TestRun_Unmarshal(t *testing.T) {
	testdb := newTestDB(t)
	org := createTestOrganization(t, testdb)
	ws := createTestWorkspace(t, testdb, org)
	cv := createTestConfigurationVersion(t, testdb, ws)
	run := createTestRun(t, testdb, ws, cv)

	conn := testdb.(db).Conn
	q := NewQuerier(conn)
	row, err := q.FindRunByID(context.Background(), run.ID)
	require.NoError(t, err)

	got, err := otf.UnmarshalRunFromDB(row)
	require.NoError(t, err)

	assert.Equal(t, run.ID, got.ID)
	assert.Equal(t, run.StatusTimestamps, got.StatusTimestamps)
	assert.Equal(t, run.Workspace, got.Workspace)
	assert.Equal(t, run.Workspace.Organization, got.Workspace.Organization)
	assert.Equal(t, run, got)
}
