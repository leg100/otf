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
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})

	run := otf.NewRun(cv, ws, otf.RunCreateOptions{})
	err := db.CreateRun(context.Background(), run)
	require.NoError(t, err)
}

func TestRun_UpdateStatus(t *testing.T) {
	ctx := context.Background()

	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})

	t.Run("update status", func(t *testing.T) {
		run := createTestRun(t, db, ws, cv)
		got, err := db.UpdateStatus(ctx, run.ID(), func(run *otf.Run) error {
			return run.EnqueuePlan(ctx, &otf.FakeWorkspaceLockService{})
		})
		require.NoError(t, err)
		assert.Equal(t, otf.RunPlanQueued, got.Status())
		timestamp, err := got.StatusTimestamp(otf.RunPlanQueued)
		assert.NoError(t, err)
		assert.True(t, timestamp.After(got.CreatedAt()))
	})

	t.Run("update status", func(t *testing.T) {
		run := createTestRun(t, db, ws, cv)
		got, err := db.UpdateStatus(ctx, run.ID(), func(run *otf.Run) error {
			_, err := run.Cancel()
			return err
		})
		require.NoError(t, err)
		assert.Equal(t, otf.RunCanceled, got.Status())
		canceled, err := got.StatusTimestamp(otf.RunCanceled)
		assert.NoError(t, err)
		assert.True(t, canceled.After(got.CreatedAt()))

		// force cancel available after a cool down period following cancelation
		assert.True(t, got.ForceCancelAvailableAt().After(canceled))
	})
}

func TestRun_Get(t *testing.T) {
	db := newTestDB(t)
	org := createTestOrganization(t, db)
	ws := createTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})

	want := createTestRun(t, db, ws, cv)

	got, err := db.GetRun(context.Background(), want.ID())
	require.NoError(t, err)

	assert.Equal(t, want.ForceCancelAvailableAt(), got.ForceCancelAvailableAt())
	assert.Equal(t, want, got)
}

func TestRun_List(t *testing.T) {
	db := newTestDB(t)
	org1 := createTestOrganization(t, db)
	org2 := createTestOrganization(t, db)
	ws1 := createTestWorkspace(t, db, org1)
	ws2 := createTestWorkspace(t, db, org2)
	cv1 := createTestConfigurationVersion(t, db, ws1, otf.ConfigurationVersionCreateOptions{})
	cv2 := createTestConfigurationVersion(t, db, ws2, otf.ConfigurationVersionCreateOptions{
		Speculative: otf.Bool(true),
	})

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
			opts: otf.RunListOptions{Organization: otf.String(org1.Name())},
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
			opts: otf.RunListOptions{WorkspaceName: otf.String(ws1.Name()), Organization: otf.String(org1.Name())},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Contains(t, l.Items, run1)
				assert.Contains(t, l.Items, run2)
			},
		},
		{
			name: "by pending status",
			opts: otf.RunListOptions{Organization: otf.String(org1.Name()), Statuses: []otf.RunStatus{otf.RunPending}},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 2, len(l.Items))
				assert.Contains(t, l.Items, run1)
				assert.Contains(t, l.Items, run2)
			},
		},
		{
			name: "by statuses - no match",
			opts: otf.RunListOptions{Organization: otf.String(org1.Name()), Statuses: []otf.RunStatus{otf.RunPlanned}},
			want: func(t *testing.T, l *otf.RunList) {
				assert.Equal(t, 0, len(l.Items))
			},
		},
		{
			name: "filter out speculative runs in org1",
			opts: otf.RunListOptions{Organization: otf.String(org1.Name()), Speculative: otf.Bool(false)},
			want: func(t *testing.T, l *otf.RunList) {
				// org1 has no speculative runs, so should return both runs
				assert.Equal(t, 2, len(l.Items))
				assert.Equal(t, 2, l.TotalCount())
			},
		},
		{
			name: "filter out speculative runs in org2",
			opts: otf.RunListOptions{Organization: otf.String(org2.Name()), Speculative: otf.Bool(false)},
			want: func(t *testing.T, l *otf.RunList) {
				// org2 only has speculative runs, so should return zero
				assert.Equal(t, 0, len(l.Items))
				assert.Equal(t, 0, l.TotalCount())
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
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})
	run := createTestRun(t, db, ws, cv)

	report := otf.ResourceReport{
		Additions:    5,
		Changes:      2,
		Destructions: 99,
	}

	err := db.CreatePlanReport(context.Background(), run.ID(), report)
	require.NoError(t, err)

	run, err = db.GetRun(context.Background(), run.ID())
	require.NoError(t, err)

	assert.NotNil(t, run.Plan().ResourceReport)
	assert.Equal(t, &report, run.Plan().ResourceReport)
}
