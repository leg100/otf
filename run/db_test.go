package run

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/configversion"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	ctx := context.Background()
	db := &pgdb{sql.NewTestDB(t)}

	t.Run("create", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		cv := configversion.CreateTestConfigurationVersion(t, db, ws, configversion.ConfigurationVersionCreateOptions{})

		run := NewRun(cv, ws, RunCreateOptions{})
		err := db.CreateRun(ctx, run)
		require.NoError(t, err)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		cv := configversion.CreateTestConfigurationVersion(t, db, ws, configversion.ConfigurationVersionCreateOptions{})

		run := CreateTestRun(t, db, ws, cv, RunCreateOptions{})
		got, err := db.UpdateStatus(ctx, run.ID, func(run *Run) error {
			return run.EnqueuePlan()
		})
		require.NoError(t, err)
		assert.Equal(t, otf.RunPlanQueued, got.Status)
		timestamp, err := got.StatusTimestamp(otf.RunPlanQueued)
		assert.NoError(t, err)
		assert.True(t, timestamp.After(got.CreatedAt))
	})

	t.Run("cancel run", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		cv := configversion.CreateTestConfigurationVersion(t, db, ws, configversion.ConfigurationVersionCreateOptions{})
		run := CreateTestRun(t, db, ws, cv, RunCreateOptions{})
		got, err := db.UpdateStatus(ctx, run.ID, func(run *Run) error {
			_, err := run.Cancel()
			return err
		})
		require.NoError(t, err)
		assert.Equal(t, otf.RunCanceled, got.Status)
		canceled, err := got.StatusTimestamp(otf.RunCanceled)
		assert.NoError(t, err)
		assert.True(t, canceled.After(got.CreatedAt))

		// force cancel available after a cool down period following cancelation
		assert.True(t, got.ForceCancelAvailableAt.After(canceled))
	})

	t.Run("get", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		cv := configversion.CreateTestConfigurationVersion(t, db, ws, configversion.ConfigurationVersionCreateOptions{})

		want := CreateTestRun(t, db, ws, cv, RunCreateOptions{})

		got, err := db.GetRun(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want.ForceCancelAvailableAt, got.ForceCancelAvailableAt)
		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		org1 := organization.CreateTestOrganization(t, db)
		org2 := organization.CreateTestOrganization(t, db)
		ws1 := workspace.CreateTestWorkspace(t, db, org1.Name)
		ws2 := workspace.CreateTestWorkspace(t, db, org2.Name)
		cv1 := configversion.CreateTestConfigurationVersion(t, db, ws1, configversion.ConfigurationVersionCreateOptions{})
		cv2 := configversion.CreateTestConfigurationVersion(t, db, ws1, configversion.ConfigurationVersionCreateOptions{
			Speculative: otf.Bool(true),
		})

		run1 := CreateTestRun(t, db, ws1, cv1, RunCreateOptions{})
		run2 := CreateTestRun(t, db, ws1, cv1, RunCreateOptions{})
		run3 := CreateTestRun(t, db, ws2, cv2, RunCreateOptions{})
		run4 := CreateTestRun(t, db, ws2, cv2, RunCreateOptions{})

		tests := []struct {
			name string
			opts RunListOptions
			want func(*testing.T, *RunList)
		}{
			{
				name: "unfiltered",
				opts: RunListOptions{},
				want: func(t *testing.T, l *RunList) {
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
				opts: RunListOptions{Organization: otf.String(org1.Name)},
				want: func(t *testing.T, l *RunList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by workspace id",
				opts: RunListOptions{WorkspaceID: otf.String(ws1.ID)},
				want: func(t *testing.T, l *RunList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by workspace name and organization",
				opts: RunListOptions{WorkspaceName: otf.String(ws1.Name), Organization: otf.String(org1.Name)},
				want: func(t *testing.T, l *RunList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by pending status",
				opts: RunListOptions{Organization: otf.String(org1.Name), Statuses: []otf.RunStatus{otf.RunPending}},
				want: func(t *testing.T, l *RunList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by statuses - no match",
				opts: RunListOptions{Organization: otf.String(org1.Name), Statuses: []otf.RunStatus{otf.RunPlanned}},
				want: func(t *testing.T, l *RunList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "filter out speculative runs in org1",
				opts: RunListOptions{Organization: otf.String(org1.Name), Speculative: otf.Bool(false)},
				want: func(t *testing.T, l *RunList) {
					// org1 has no speculative runs, so should return both runs
					assert.Equal(t, 2, len(l.Items))
					assert.Equal(t, 2, l.TotalCount())
				},
			},
			{
				name: "filter out speculative runs in org2",
				opts: RunListOptions{Organization: otf.String(org2.Name), Speculative: otf.Bool(false)},
				want: func(t *testing.T, l *RunList) {
					// org2 only has speculative runs, so should return zero
					assert.Equal(t, 0, len(l.Items))
					assert.Equal(t, 0, l.TotalCount())
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := db.ListRuns(ctx, tt.opts)
				require.NoError(t, err)

				tt.want(t, got)
			})
		}
	})

	t.Run("create plan report", func(t *testing.T) {
		org := organization.CreateTestOrganization(t, db)
		ws := workspace.CreateTestWorkspace(t, db, org.Name)
		cv := configversion.CreateTestConfigurationVersion(t, db, ws, configversion.ConfigurationVersionCreateOptions{})
		run := CreateTestRun(t, db, ws, cv, RunCreateOptions{})

		report := ResourceReport{
			Additions:    5,
			Changes:      2,
			Destructions: 99,
		}

		err := db.CreatePlanReport(ctx, run.ID, report)
		require.NoError(t, err)

		run, err = db.GetRun(ctx, run.ID)
		require.NoError(t, err)

		assert.NotNil(t, run.Plan.ResourceReport)
		assert.Equal(t, &report, run.Plan.ResourceReport)
	})
}
