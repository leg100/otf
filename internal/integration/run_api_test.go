package integration

import (
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/runstatus"
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
	repo := vcs.NewRandomRepo()
	daemon, org, ctx := setup(t, withGithubOptions(
		github.WithRepo(repo),
		github.WithCommit("0335fb07bb0244b7a169ee89d15c7703e4aaf7de"),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
	))
	_, token := daemon.createToken(t, ctx, nil)

	tfeClient, err := tfe.NewClient(&tfe.Config{
		Address:           daemon.System.URL("/"),
		Token:             string(token),
		RetryServerErrors: true,
	})
	require.NoError(t, err)

	// pull config from workspace's vcs repo
	t.Run("create run using config from repo", func(t *testing.T) {
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

		created, err := tfeClient.Runs.Create(ctx, tfe.RunCreateOptions{
			// no config version ID specified
			Workspace: &tfe.Workspace{
				ID: ws.ID.String(),
			},
		})
		require.NoError(t, err)

		// wait for run to reach planned status
		planned := daemon.waitRunStatus(t, ctx, testutils.ParseID(t, created.ID), runstatus.Planned)

		// run should have planned two resources (defined in the config from the
		// github repo)
		assert.Equal(t, 2, planned.Plan.ResourceReport.Additions)
	})
}
