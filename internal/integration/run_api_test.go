package integration

import (
	"testing"

	tfe "github.com/hashicorp/go-tfe"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/testutils"
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
	repo := cloud.NewTestRepo()
	daemon, org, ctx := setup(t, nil,
		github.WithRepo(repo),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
	)
	_, token := daemon.createToken(t, ctx, nil)

	tfeClient, err := tfe.NewClient(&tfe.Config{
		Address:           "https://" + daemon.Hostname(),
		Token:             string(token),
		RetryServerErrors: true,
	})
	require.NoError(t, err)

	// test the "magic string" behaviour specific to OTF: if
	// run.PullVCSMagicString is specified for the config version ID then the
	// config is pulled from the workspace's connected repo.
	t.Run("create run using config from repo", func(t *testing.T) {
		vcsProvider := daemon.createVCSProvider(t, ctx, org)
		ws, err := daemon.CreateWorkspace(ctx, workspace.CreateOptions{
			Name:         internal.String("connected-workspace"),
			Organization: internal.String(org.Name),
			ConnectOptions: &workspace.ConnectOptions{
				RepoPath:      repo,
				VCSProviderID: vcsProvider.ID,
			},
		})
		require.NoError(t, err)

		created, err := tfeClient.Runs.Create(ctx, tfe.RunCreateOptions{
			ConfigurationVersion: &tfe.ConfigurationVersion{
				ID: run.PullVCSMagicString,
			},
			Workspace: &tfe.Workspace{
				ID: ws.ID,
			},
		})
		require.NoError(t, err)

		// wait for run to reach planned status
		for event := range daemon.sub {
			if r, ok := event.Payload.(*run.Run); ok {
				switch r.Status {
				case internal.RunErrored:
					t.Fatal("run unexpectedly errored")
				case internal.RunPlanned:
					// run should have planned two resources (defined in the config from the
					// github repo)
					planned, err := daemon.GetRun(ctx, created.ID)
					require.NoError(t, err)

					assert.Equal(t, 2, planned.Plan.Additions)
					return // success
				}
			}
		}

	})
}
