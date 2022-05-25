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

	run := newTestRun(ws, cv)
	err := db.RunStore().Create(run)
	require.NoError(t, err)
}

func TestRun_Timestamps(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)

	run := newTestRun(ws, cv)
	err := db.RunStore().Create(run)
	require.NoError(t, err)

	got, err := db.RunStore().Get(otf.RunGetOptions{ID: otf.String(run.ID())})
	require.NoError(t, err)

	assert.Equal(t, run.CreatedAt(), got.CreatedAt())
	assert.Equal(t, run.CreatedAt().UTC(), got.CreatedAt().UTC())
	assert.True(t, run.CreatedAt().Equal(got.CreatedAt()))
}

func TestRun_UpdateStatus(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)

	tests := []struct {
		name   string
		update func(run *otf.Run) error
		want   func(*testing.T, *otf.Run)
	}{
		{
			name: "enqueue plan",
			update: func(run *otf.Run) error {
				return run.EnqueuePlan()
			},
			want: func(t *testing.T, got *otf.Run) {
				assert.Equal(t, otf.RunPlanQueued, got.Status())
				timestamp, err := got.StatusTimestamp(otf.RunPlanQueued)
				assert.NoError(t, err)
				assert.True(t, timestamp.After(got.CreatedAt()))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := createTestRun(t, db, ws, cv)

			got, err := db.RunStore().UpdateStatus(otf.RunGetOptions{ID: otf.String(run.ID())}, tt.update)
			require.NoError(t, err)

			tt.want(t, got)
		})
	}
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
			opts: otf.RunGetOptions{ID: otf.String(want.ID())},
		},
		{
			name: "by plan id",
			opts: otf.RunGetOptions{PlanID: otf.String(want.Plan.ID())},
		},
		{
			name: "by apply id",
			opts: otf.RunGetOptions{ApplyID: otf.String(want.Apply.ID())},
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

func TestRun_List(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)

	run1 := createTestRun(t, db, ws, cv)
	run2 := createTestRun(t, db, ws, cv)
	run3 := createTestRun(t, db, ws, cv)

	tests := []struct {
		name string
		opts otf.RunListOptions
		want func(*testing.T, *otf.RunList)
	}{
		{
			name: "by workspace id",
			opts: otf.RunListOptions{WorkspaceID: otf.String(ws.ID())},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 3, len(l.Items))
				assert.Contains(t, l.Items, run1)
				assert.Contains(t, l.Items, run2)
				assert.Contains(t, l.Items, run3)
			},
		},
		{
			name: "by workspace name and organization",
			opts: otf.RunListOptions{WorkspaceName: otf.String(ws.Name()), OrganizationName: otf.String(org.Name())},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 3, len(l.Items))
				assert.Contains(t, l.Items, run1)
				assert.Contains(t, l.Items, run2)
				assert.Contains(t, l.Items, run3)
			},
		},
		{
			name: "by statuses",
			opts: otf.RunListOptions{WorkspaceID: otf.String(ws.ID()), Statuses: []otf.RunStatus{otf.RunPending}},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 3, len(l.Items))
				assert.Contains(t, l.Items, run1)
				assert.Contains(t, l.Items, run2)
				assert.Contains(t, l.Items, run3)
			},
		},
		{
			name: "by statuses - no match",
			opts: otf.RunListOptions{WorkspaceID: otf.String(ws.ID()), Statuses: []otf.RunStatus{otf.RunPlanned}},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.RunStore().List(tt.opts)
			require.NoError(t, err)

			tt.want(t, got)
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
		Additions:    5,
		Changes:      2,
		Destructions: 99,
	}

	err := db.RunStore().CreatePlanReport(run.ID(), report)
	require.NoError(t, err)

	run, err = db.RunStore().Get(otf.RunGetOptions{ID: otf.String(run.ID())})
	require.NoError(t, err)

	assert.NotNil(t, run.Plan.ResourceReport)
	assert.Equal(t, &report, run.Plan.ResourceReport)
}
