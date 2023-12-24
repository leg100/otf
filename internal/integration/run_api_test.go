package integration

import (
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_RunAPI tests those parts of the run API that are not covered
// by the `go-tfe` integration tests, i.e. behaviours that are specific to
// OTF.
func TestIntegration_RunAPI(t *testing.T) {
	integrationTest(t)

	// setup daemon along with fake github repo
	repo := vcs.NewTestRepo()
	daemon, org, ctx := setup(t, nil,
		github.WithRepo(repo),
		github.WithCommit("0335fb07bb0244b7a169ee89d15c7703e4aaf7de"),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
	)
	_, token := daemon.createToken(t, ctx, nil)

	tfeClient, err := tfe.NewClient(&tfe.Config{
		Address:           "https://" + daemon.System.Hostname(),
		Token:             string(token),
		RetryServerErrors: true,
	})
	require.NoError(t, err)

	// pull config from workspace's vcs repo
	t.Run("create run using config from repo", func(t *testing.T) {
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

		sub, unsub := daemon.Runs.Watch(ctx)
		defer unsub()

		created, err := tfeClient.Runs.Create(ctx, tfe.RunCreateOptions{
			// no config version ID specified
			Workspace: &tfe.Workspace{
				ID: ws.ID,
			},
		})
		require.NoError(t, err)

		// wait for run to reach planned status
		for event := range sub {
			switch event.Payload.Status {
			case run.RunErrored:
				t.Fatal("run unexpectedly errored")
			case run.RunPlanned:
				// run should have planned two resources (defined in the config from the
				// github repo)
				planned, err := daemon.Runs.Get(ctx, created.ID)
				require.NoError(t, err)

				assert.Equal(t, 2, planned.Plan.ResourceReport.Additions)
				return // success
			}
		}
	})
}
