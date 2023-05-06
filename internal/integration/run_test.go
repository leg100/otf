package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/run"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := internal.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	t.Run("create", func(t *testing.T) {
		svc := setup(t, &config{Config: daemon.Config{DisableScheduler: true}})
		cv := svc.createConfigurationVersion(t, ctx, nil)

		_, err := svc.CreateRun(ctx, cv.WorkspaceID, run.RunCreateOptions{})
		require.NoError(t, err)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		svc := setup(t, &config{Config: daemon.Config{DisableScheduler: true}})
		run := svc.createRun(t, ctx, nil, nil)

		got, err := svc.EnqueuePlan(ctx, run.ID)
		require.NoError(t, err)

		assert.Equal(t, internal.RunPlanQueued, got.Status)
		timestamp, err := got.StatusTimestamp(internal.RunPlanQueued)
		assert.NoError(t, err)
		assert.True(t, timestamp.After(got.CreatedAt))
	})

	t.Run("cancel run", func(t *testing.T) {
		svc := setup(t, &config{Config: daemon.Config{DisableScheduler: true}})
		run := svc.createRun(t, ctx, nil, nil)

		got, err := svc.Cancel(ctx, run.ID)
		require.NoError(t, err)

		assert.Equal(t, internal.RunCanceled, got.Status)
		canceled, err := got.StatusTimestamp(internal.RunCanceled)
		assert.NoError(t, err)
		assert.True(t, canceled.After(got.CreatedAt))

		// force cancel available after a cool down period following cancelation
		assert.True(t, got.ForceCancelAvailableAt.After(canceled))
	})

	t.Run("get", func(t *testing.T) {
		svc := setup(t, &config{Config: daemon.Config{DisableScheduler: true}})
		want := svc.createRun(t, ctx, nil, nil)

		got, err := svc.GetRun(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list", func(t *testing.T) {
		svc := setup(t, &config{Config: daemon.Config{DisableScheduler: true}})

		ws1 := svc.createWorkspace(t, ctx, nil)
		ws2 := svc.createWorkspace(t, ctx, nil)
		cv1 := svc.createConfigurationVersion(t, ctx, ws1)
		cv2, err := svc.CreateConfigurationVersion(ctx, ws2.ID, configversion.ConfigurationVersionCreateOptions{
			Speculative: internal.Bool(true),
		})
		require.NoError(t, err)

		run1 := svc.createRun(t, ctx, ws1, cv1)
		run2 := svc.createRun(t, ctx, ws1, cv1)
		run3 := svc.createRun(t, ctx, ws2, cv2)
		run4 := svc.createRun(t, ctx, ws2, cv2)

		tests := []struct {
			name string
			opts run.RunListOptions
			want func(*testing.T, *run.RunList)
		}{
			{
				name: "unfiltered",
				opts: run.RunListOptions{},
				want: func(t *testing.T, l *run.RunList) {
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
				opts: run.RunListOptions{Organization: internal.String(ws1.Organization)},
				want: func(t *testing.T, l *run.RunList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by workspace id",
				opts: run.RunListOptions{WorkspaceID: internal.String(ws1.ID)},
				want: func(t *testing.T, l *run.RunList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by workspace name and organization",
				opts: run.RunListOptions{WorkspaceName: internal.String(ws1.Name), Organization: internal.String(ws1.Organization)},
				want: func(t *testing.T, l *run.RunList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by pending status",
				opts: run.RunListOptions{Organization: internal.String(ws1.Organization), Statuses: []internal.RunStatus{internal.RunPending}},
				want: func(t *testing.T, l *run.RunList) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by statuses - no match",
				opts: run.RunListOptions{Organization: internal.String(ws1.Organization), Statuses: []internal.RunStatus{internal.RunPlanned}},
				want: func(t *testing.T, l *run.RunList) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "filter out speculative runs in org1",
				opts: run.RunListOptions{Organization: internal.String(ws1.Organization), Speculative: internal.Bool(false)},
				want: func(t *testing.T, l *run.RunList) {
					// org1 has no speculative runs, so should return both runs
					assert.Equal(t, 2, len(l.Items))
					assert.Equal(t, 2, l.TotalCount())
				},
			},
			{
				name: "filter out speculative runs in org2",
				opts: run.RunListOptions{Organization: internal.String(ws2.Organization), Speculative: internal.Bool(false)},
				want: func(t *testing.T, l *run.RunList) {
					// org2 only has speculative runs, so should return zero
					assert.Equal(t, 0, len(l.Items))
					assert.Equal(t, 0, l.TotalCount())
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := svc.ListRuns(ctx, tt.opts)
				require.NoError(t, err)

				tt.want(t, got)
			})
		}
	})
}
