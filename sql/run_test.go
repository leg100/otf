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

	run := newTestRun(ws, cv)
	err := db.CreateRun(context.Background(), run)
	require.NoError(t, err)
}

func TestRun_Timestamps(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws)

	run := newTestRun(ws, cv)
	err := db.CreateRun(context.Background(), run)
	require.NoError(t, err)

	got, err := db.GetRun(context.Background(), otf.RunGetOptions{ID: otf.String(run.ID())})
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

			got, err := db.UpdateStatus(context.Background(), otf.RunGetOptions{ID: otf.String(run.ID())}, tt.update)
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
			opts: otf.RunGetOptions{PlanID: otf.String(want.Plan().ID())},
		},
		{
			name: "by apply id",
			opts: otf.RunGetOptions{ApplyID: otf.String(want.Apply().ID())},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetRun(context.Background(), tt.opts)
			require.NoError(t, err)

			assert.Equal(t, want.Apply().PhaseStatusTimestamps(), got.Apply().PhaseStatusTimestamps())
			assert.Equal(t, want.Plan().PhaseStatusTimestamps(), got.Plan().PhaseStatusTimestamps())
			assert.Equal(t, want, got)
		})
	}

	t.Run("include workspace", func(t *testing.T) {
		got, err := db.GetRun(context.Background(), otf.RunGetOptions{
			ID:      otf.String(want.ID()),
			Include: otf.String("workspace"),
		})
		require.NoError(t, err)
		assert.Equal(t, ws, got.Workspace())
	})
}

func TestRun_List(t *testing.T) {
	db := newTestDB(t)
	org1 := createTestOrganization(t, db)
	org2 := createTestOrganization(t, db)
	ws1 := createTestWorkspace(t, db, org1)
	ws2 := createTestWorkspace(t, db, org2)
	cv1 := createTestConfigurationVersion(t, db, ws1)
	cv2 := createTestConfigurationVersion(t, db, ws2)

	run1 := createTestRun(t, db, ws1, cv1)
	run2 := createTestRun(t, db, ws1, cv1)
	run3 := createTestRun(t, db, ws2, cv2)
	run4 := createTestRun(t, db, ws2, cv2)

	tests := []struct {
		name string
		opts otf.RunListOptions
		want func(*testing.T, *otf.RunList)
	}{
		{
			name: "unfiltered",
			opts: otf.RunListOptions{},
			want: func(t *testing.T, l *otf.RunList) {
				// may match runs in the db belonging to organizations outside
				// of this test
				assert.GreaterOrEqual(t, len(l.Items), 4)
				assert.Contains(t, l.Items, run1)
				assert.Contains(t, l.Items, run2)
				assert.Contains(t, l.Items, run3)
				assert.Contains(t, l.Items, run4)
			},
		},
		{
			name: "by organization name",
			opts: otf.RunListOptions{OrganizationName: otf.String(org1.Name())},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Contains(t, l.Items, run1)
				assert.Contains(t, l.Items, run2)
			},
		},
		{
			name: "by workspace id",
			opts: otf.RunListOptions{WorkspaceID: otf.String(ws1.ID())},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Contains(t, l.Items, run1)
				assert.Contains(t, l.Items, run2)
			},
		},
		{
			name: "by workspace name and organization",
			opts: otf.RunListOptions{WorkspaceName: otf.String(ws1.Name()), OrganizationName: otf.String(org1.Name())},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Contains(t, l.Items, run1)
				assert.Contains(t, l.Items, run2)
			},
		},
		{
			name: "by pending status",
			opts: otf.RunListOptions{OrganizationName: otf.String(org1.Name()), Statuses: []otf.RunStatus{otf.RunPending}},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Contains(t, l.Items, run1)
				assert.Contains(t, l.Items, run2)
			},
		},
		{
			name: "by statuses - no match",
			opts: otf.RunListOptions{OrganizationName: otf.String(org1.Name()), Statuses: []otf.RunStatus{otf.RunPlanned}},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
		{
			name: "include workspace",
			opts: otf.RunListOptions{
				OrganizationName: otf.String(org1.Name()),
				WorkspaceName:    otf.String(ws1.Name()),
				Include:          otf.String("workspace"),
			},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Equal(t, ws1, l.Items[0].Workspace())
				assert.Equal(t, ws1, l.Items[1].Workspace())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.ListRuns(context.Background(), tt.opts)
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

	err := db.CreatePlanReport(context.Background(), run.Plan().ID(), report)
	require.NoError(t, err)

	run, err = db.GetRun(context.Background(), otf.RunGetOptions{ID: otf.String(run.ID())})
	require.NoError(t, err)

	assert.NotNil(t, run.Plan().ResourceReport)
	assert.Equal(t, &report, run.Plan().ResourceReport)
}
