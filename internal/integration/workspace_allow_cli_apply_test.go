package integration

import (
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_AllowCLIApply demonstrates overriding the default terraform
// behaviour of prohibiting running terraform apply from the CLI on a
// workspace connected to a VCS repository.
func TestIntegration_AllowCLIApply(t *testing.T) {
	integrationTest(t)

	repo := vcs.NewTestRepo()
	daemon, org, ctx := setup(t, nil,
		github.WithRepo(repo),
		github.WithArchive(testutils.ReadFile(t, "../testdata/github.tar.gz")),
	)

	vcsProvider := daemon.createVCSProvider(t, ctx, org)
	ws, err := daemon.CreateWorkspace(ctx, workspace.CreateOptions{
		Name:         internal.String("connected-workspace"),
		Organization: internal.String(org.Name),
		ConnectOptions: &workspace.ConnectOptions{
			RepoPath:      &repo,
			VCSProviderID: &vcsProvider.ID,
		},
	})
	require.NoError(t, err)

	// by default, terraform apply should fail
	config := newRootModule(t, daemon.Hostname(), ws.Organization, ws.Name)
	daemon.tfcli(t, ctx, "init", config)
	out, err := daemon.tfcliWithError(t, ctx, "apply", config, "-auto-approve")
	require.Error(t, err, out)
	assert.Contains(t, out, "Apply not allowed for workspaces with a VCS connection")

	_, err = daemon.UpdateWorkspace(ctx, ws.ID, workspace.UpdateOptions{
		ConnectOptions: &workspace.ConnectOptions{
			AllowCLIApply: internal.Bool(true),
		},
	})
	require.NoError(t, err)

	// terraform apply should now be possible from CLI
	daemon.tfcli(t, ctx, "init", config)
	_, err = daemon.tfcliWithError(t, ctx, "apply", config, "-auto-approve")
	require.NoError(t, err, out)
}
