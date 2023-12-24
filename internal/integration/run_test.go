package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/configversion"
	"github.com/leg100/otf/internal/daemon"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/resource"
	otfrun "github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	integrationTest(t)

	t.Run("create", func(t *testing.T) {
		svc, _, ctx := setup(t, &config{Config: daemon.Config{DisableScheduler: true}})
		cv := svc.createConfigurationVersion(t, ctx, nil, nil)

		run, err := svc.Runs.Create(ctx, cv.WorkspaceID, otfrun.CreateOptions{})
		require.NoError(t, err)

		user, err := user.UserFromContext(ctx)
		require.NoError(t, err)
		assert.NotNil(t, run.CreatedBy)
		assert.Equal(t, user.Username, *run.CreatedBy)
	})

	t.Run("create run using config from repo", func(t *testing.T) {
		// setup daemon along with fake github repo
		repo := vcs.NewTestRepo()
		daemon, _, ctx := setup(t, nil,
			github.WithRepo(repo),
			github.WithCommit("0335fb07bb0244b7a169ee89d15c7703e4aaf7de"),
			github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
		)
		org := daemon.createOrganization(t, ctx)
		vcsProvider := daemon.createVCSProvider(t, ctx, org)
		ws, err := daemon.Workspaces.Create(ctx, workspace.CreateOptions{
			Name:         "connected-workspace",
			Organization: org.Name,
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
		svc, _, ctx := setup(t, &config{Config: daemon.Config{DisableScheduler: true}})
		run := svc.createRun(t, ctx, nil, nil)

		got, err := svc.Runs.EnqueuePlan(ctx, run.ID)
		require.NoError(t, err)

		assert.Equal(t, otfrun.RunPlanQueued, got.Status)
		timestamp, err := got.StatusTimestamp(otfrun.RunPlanQueued)
		assert.NoError(t, err)
		assert.True(t, timestamp.After(got.CreatedAt))
	})

	t.Run("cancel pending run", func(t *testing.T) {
		svc, _, ctx := setup(t, &config{Config: daemon.Config{DisableScheduler: true}})
		run := svc.createRun(t, ctx, nil, nil)

		err := svc.Runs.Cancel(ctx, run.ID)
		require.NoError(t, err)

		got, err := svc.Runs.Get(ctx, run.ID)
		require.NoError(t, err)

		assert.Equal(t, otfrun.RunCanceled, got.Status)
		canceled, err := got.StatusTimestamp(otfrun.RunCanceled)
		assert.NoError(t, err)
		assert.True(t, canceled.After(got.CreatedAt))
	})

	t.Run("get", func(t *testing.T) {
		svc, _, ctx := setup(t, &config{Config: daemon.Config{DisableScheduler: true}})
		want := svc.createRun(t, ctx, nil, nil)

		got, err := svc.Runs.Get(ctx, want.ID)
		require.NoError(t, err)

		assert.Equal(t, want, got)

		user, err := user.UserFromContext(ctx)
		require.NoError(t, err)
		assert.NotNil(t, got.CreatedBy)
		assert.Equal(t, user.Username, *got.CreatedBy)
	})

	t.Run("list", func(t *testing.T) {
		svc, _, ctx := setup(t, &config{Config: daemon.Config{DisableScheduler: true}})

		ws1 := svc.createWorkspace(t, ctx, nil)
		ws2 := svc.createWorkspace(t, ctx, nil)
		cv1 := svc.createConfigurationVersion(t, ctx, ws1, nil)
		cv2, err := svc.Configs.Create(ctx, ws2.ID, configversion.CreateOptions{
			Speculative: internal.Bool(true),
		})
		require.NoError(t, err)

		run1 := svc.createRun(t, ctx, ws1, cv1)
		run2 := svc.createRun(t, ctx, ws1, cv1)
		run3 := svc.createRun(t, ctx, ws2, cv2)
		run4 := svc.createRun(t, ctx, ws2, cv2)

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
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
					assert.Contains(t, l.Items, run3)
					assert.Contains(t, l.Items, run4)
				},
			},
			{
				name: "by organization name",
				opts: otfrun.ListOptions{Organization: internal.String(ws1.Organization)},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by workspace id",
				opts: otfrun.ListOptions{WorkspaceID: internal.String(ws1.ID)},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by workspace name and organization",
				opts: otfrun.ListOptions{WorkspaceName: internal.String(ws1.Name), Organization: internal.String(ws1.Organization)},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by pending status",
				opts: otfrun.ListOptions{Organization: internal.String(ws1.Organization), Statuses: []otfrun.Status{otfrun.RunPending}},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					assert.Equal(t, 2, len(l.Items))
					assert.Contains(t, l.Items, run1)
					assert.Contains(t, l.Items, run2)
				},
			},
			{
				name: "by statuses - no match",
				opts: otfrun.ListOptions{Organization: internal.String(ws1.Organization), Statuses: []otfrun.Status{otfrun.RunPlanned}},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					assert.Equal(t, 0, len(l.Items))
				},
			},
			{
				name: "filter out speculative runs in org1",
				opts: otfrun.ListOptions{Organization: internal.String(ws1.Organization), PlanOnly: internal.Bool(false)},
				want: func(t *testing.T, l *resource.Page[*otfrun.Run]) {
					// org1 has no speculative runs, so should return both runs
					assert.Equal(t, 2, len(l.Items))
					assert.Equal(t, 2, l.TotalCount)
				},
			},
			{
				name: "filter out speculative runs in org2",
				opts: otfrun.ListOptions{Organization: internal.String(ws2.Organization), PlanOnly: internal.Bool(false)},
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
