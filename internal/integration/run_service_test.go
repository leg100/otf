package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/resource"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/runstatus"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunService(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, _, ctx := setup(t, disableScheduler())
		cv := svc.createConfigurationVersion(t, ctx, nil, nil)

		run, err := svc.Runs.Create(ctx, cv.WorkspaceID, otfrun.CreateOptions{})
		require.NoError(t, err)

		assertRunCreatedByCurrentUser(t, ctx, run)
	})

	t.Run("create run using config from repo", func(t *testing.T) {
		// setup daemon along with fake github repo
		repo := vcs.NewRandomRepo()
		daemon, _, ctx := setup(t, withGithubOptions(
			github.WithRepo(repo),
			github.WithCommit("0335fb07bb0244b7a169ee89d15c7703e4aaf7de"),
			github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
		))
		org := daemon.createOrganization(t, ctx)
		vcsProvider := daemon.createVCSProvider(t, ctx, org, nil)
		ws, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
			Name:         new("connected-workspace"),
			Organization: &org.Name,
			ConnectOptions: &workspace.ConnectOptions{
				RepoPath:      &repo,
				VCSProviderID: &vcsProvider.ID,
			},
		})
		require.NoError(t, err)

		_, err = daemon.Runs.Create(ctx, ws.ID, otfrun.CreateOptions{})
		require.NoError(t, err)
	})

	t.Run("enqueue plan", func(t *testing.T) {
		svc, _, ctx := setup(t, disableScheduler())

		tests := []struct {
			name      string
			planOnly  bool
			latestRun bool
			lock      bool
		}{
			{
				"normal run - lock and make current run",
				false,
				true,
				true,
			},
			{
				"plan-only run - dont lock and dont make latest run",
				true,
				false,
				false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				run := svc.createRun(t, ctx, nil, nil, &otfrun.CreateOptions{
					PlanOnly: &tt.planOnly,
				})

				got, err := svc.Runs.EnqueuePlan(ctx, run.ID)
				require.NoError(t, err)

				ws := svc.getWorkspace(t, ctx, run.WorkspaceID)
				if tt.latestRun {
					assert.Equal(t, run.ID, ws.LatestRun.ID)
				} else {
					assert.Nil(t, ws.LatestRun)
				}
				if tt.lock {
					assert.Equal(t, run.ID, ws.Lock)
				} else {
					assert.Nil(t, ws.Lock)
				}

				assert.Equal(t, runstatus.PlanQueued, got.Status)
				timestamp, err := got.StatusTimestamp(runstatus.PlanQueued)
				assert.NoError(t, err)
				assert.True(t, timestamp.After(got.CreatedAt))
			})
		}
	})

	t.Run("cancel pending run", func(t *testing.T) {
		svc, _, ctx := setup(t, disableScheduler())
		run := svc.createRun(t, ctx, nil, nil, nil)

		err := svc.Runs.Cancel(ctx, run.ID)
		require.NoError(t, err)

		got, err := svc.Runs.Get(ctx, run.ID)
		require.NoError(t, err)

		assert.Equal(t, runstatus.Canceled, got.Status)
		canceled, err := got.StatusTimestamp(runstatus.Canceled)
		assert.NoError(t, err)
		assert.True(t, canceled.After(got.CreatedAt))
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t, disableScheduler())
		want := svc.createRun(t, ctx, nil, nil, nil)

		got, err := svc.Runs.Get(ctx, want.ID)
		require.NoError(t, err)

		assertEqualRuns(t, want, got)
		assertRunCreatedByCurrentUser(t, ctx, got)
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t, disableScheduler())

		ws1 := svc.createWorkspace(t, ctx, nil)
		ws2 := svc.createWorkspace(t, ctx, nil)
		cv1 := svc.createConfigurationVersion(t, ctx, ws1, nil)
		cv2, err := svc.Configs.Create(ctx, ws2.ID, configversion.CreateOptions{
			Speculative: new(true),
		})
		require.NoError(t, err)

		run1 := svc.createRun(t, ctx, ws1, cv1, nil)
		run2 := svc.createRun(t, ctx, ws1, cv1, nil)
		run3 := svc.createRun(t, ctx, ws2, cv2, nil)
		run4 := svc.createRun(t, ctx, ws2, cv2, nil)

		tests := []struct {
			name string
			opts otfrun.ListOptions
			want func(*testing.T, *resource.Page[*otfrun.Run])
		}{
			{
				name: "unfiltered",
				opts: otfrun.ListOptions{},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					// may match runs in the db belonging to organizations outside
					// of this test
					assert.GreaterOrEqual(t, len(l.Items), 4)
					assertContainsRun(t, l.Items, run1)
					assertContainsRun(t, l.Items, run2)
					assertContainsRun(t, l.Items, run3)
					assertContainsRun(t, l.Items, run4)
				},
			},
			{
				name: "by organization name",
				opts: otfrun.ListOptions{Organization: &ws1.Organization},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					assert.Equal(t, 2, len(l.Items))
					assertContainsRun(t, l.Items, run1)
					assertContainsRun(t, l.Items, run2)
				},
			},
			{
				name: "by workspace id",
				opts: otfrun.ListOptions{WorkspaceID: &ws1.ID},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					assert.Equal(t, 2, len(l.Items))
					assertContainsRun(t, l.Items, run1)
					assertContainsRun(t, l.Items, run2)
				},
			},
			{
				name: "by workspace name and organization",
				opts: otfrun.ListOptions{WorkspaceName: new(ws1.Name), Organization: &ws1.Organization},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					assert.Equal(t, 2, len(l.Items))
					assertContainsRun(t, l.Items, run1)
					assertContainsRun(t, l.Items, run2)
				},
			},
			{
				name: "by pending status",
				opts: otfrun.ListOptions{Organization: &ws1.Organization, Statuses: []runstatus.Status{runstatus.Pending}},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					assert.Equal(t, 2, len(l.Items))
					assertContainsRun(t, l.Items, run1)
					assertContainsRun(t, l.Items, run2)
				},
			},
			{
				name: "by statuses - no match",
				opts: otfrun.ListOptions{Organization: &ws1.Organization, Statuses: []runstatus.Status{runstatus.Planned}},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "filter out speculative runs in org1",
				opts: otfrun.ListOptions{Organization: &ws1.Organization, PlanOnly: new(false)},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					// org1 has no speculative runs, so should return both runs
					assert.Equal(t, 2, len(l.Items))
					assert.Equal(t, 2, l.TotalCount)
				},
			},
			{
				name: "filter out speculative runs in org2",
				opts: otfrun.ListOptions{Organization: &ws2.Organization, PlanOnly: new(false)},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					// org2 only has speculative runs, so should return zero
					assert.Equal(t, 0, len(l.Items))
					assert.Equal(t, 0, l.TotalCount)
				},
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// call endpoint using admin to avoid authz errors (particularly
				// when listing runs across a site).
				got, err := svc.Runs.List(adminCtx, tt.opts)
				require.NoError(t, err)

				tt.want(t, got)
			})
		}
	})
}

func assertEqualRuns(t *testing.T, want *otfrun.Run, got *otfrun.Run) {
	t.Helper()

	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, want.CreatedAt, got.CreatedAt)
	assert.Equal(t, want.UpdatedAt, got.UpdatedAt)
	assert.Equal(t, want.IsDestroy, got.IsDestroy)
	assert.Equal(t, want.CancelSignaledAt, got.CancelSignaledAt)
	assert.Equal(t, want.Message, got.Message)
	assert.Equal(t, want.Organization, got.Organization)
	assert.Equal(t, want.Refresh, got.Refresh)
	assert.Equal(t, want.RefreshOnly, got.RefreshOnly)
	assert.Equal(t, want.ReplaceAddrs, got.ReplaceAddrs)
	assert.Equal(t, want.PositionInQueue, got.PositionInQueue)
	assert.Equal(t, want.TargetAddrs, got.TargetAddrs)
	assert.Equal(t, want.EngineVersion, got.EngineVersion)
	assert.Equal(t, want.Engine, got.Engine)
	assert.Equal(t, want.AllowEmptyApply, got.AllowEmptyApply)
	assert.Equal(t, want.AutoApply, got.AutoApply)
	assert.Equal(t, want.PlanOnly, got.PlanOnly)
	assert.Equal(t, want.Source, got.Source)
	// Cannot compare source icons because concrete type of templ.Component is
	// templ.ComponentFunc, which is a function, and functions cannot be
	// compared.
	//
	// assert.Equal(t, want.SourceIcon, got.SourceIcon)
	assert.Equal(t, want.Status, got.Status)
	assert.Equal(t, want.WorkspaceID, got.WorkspaceID)
	assert.Equal(t, want.ConfigurationVersionID, got.ConfigurationVersionID)
	assert.Equal(t, want.ExecutionMode, got.ExecutionMode)
	assert.Equal(t, want.Variables, got.Variables)
	assert.Equal(t, want.Plan, got.Plan)
	assert.Equal(t, want.Apply, got.Apply)
	assert.Equal(t, want.StatusTimestamps, got.StatusTimestamps)
	assert.Equal(t, want.Latest, got.Latest)
	assert.Equal(t, want.IngressAttributes, got.IngressAttributes)
	assert.Equal(t, want.CreatedBy, got.CreatedBy)
	assert.Equal(t, want.CostEstimationEnabled, got.CostEstimationEnabled)
}

func assertContainsRun(t *testing.T, got []*otfrun.Run, want *otfrun.Run) {
	t.Helper()

	for _, gotRun := range got {
		if gotRun.ID == want.ID {
			assertEqualRuns(t, want, gotRun)
			return
		}
	}
	t.Errorf("run list %v does not contain run: %v", got, want)
}

func assertRunCreatedByCurrentUser(t *testing.T, ctx context.Context, got *otfrun.Run) {
	t.Helper()

	user, err := user.UserFromContext(ctx)
	require.NoError(t, err)
	assert.NotNil(t, got.CreatedBy)
	assert.Equal(t, user.Username, *got.CreatedBy)
}
