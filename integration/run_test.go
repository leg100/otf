package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCreate(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	svc := testutil.NewRunService(db)
	org := testutil.CreateOrganization(t, db)
	ws := testutil.CreateWorkspace(t, db, org.Name())

	t.Run("create", func(t *testing.T) {
		cv := testutil.CreateConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})

		var got *run.Run
		t.Cleanup(func() {
			svc.Delete(ctx, got.ID())
		})

		var err error
		got, err = svc.Create(ctx, ws.ID(), run.RunCreateOptions{
			ConfigurationVersionID: otf.String(cv.ID()),
		})
		require.NoError(t, err)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		cv := testutil.CreateConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})
		run := testutil.CreateRun(t, db, ws, cv)

		got, err := svc.EnqueuePlan(ctx, run.ID())
		require.NoError(t, err)
		assert.Equal(t, otf.RunPlanQueued, got.Status())

		timestamp, err := got.StatusTimestamp(otf.RunPlanQueued)
		assert.NoError(t, err)
		assert.True(t, timestamp.After(got.CreatedAt()))
	})

	t.Run("get run", func(t *testing.T) {
		cv := testutil.CreateConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})
		run := testutil.CreateRun(t, db, ws, cv)

		_, err := svc.Get(ctx, run.ID())
		require.NoError(t, err)
	})
}

func TestRun_List(t *testing.T) {
	db := NewTestDB(t)
	org1 := CreateTestOrganization(t, db)
	org2 := CreateTestOrganization(t, db)
	ws1 := CreateTestWorkspace(t, db, org1)
	ws2 := CreateTestWorkspace(t, db, org2)
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
	db := NewTestDB(t)
	org := CreateTestOrganization(t, db)
	ws := CreateTestWorkspace(t, db, org)
	cv := createTestConfigurationVersion(t, db, ws, otf.ConfigurationVersionCreateOptions{})
	run := createTestRun(t, db, ws, cv)

	report := run.ResourceReport{
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
