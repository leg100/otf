package integration

import (
	"testing"

	"github.com/leg100/otf/internal/runner"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCluster is an end-to-end test of the clustering capabilities, i.e.
// more than one otfd, both connected to the same postgres db. The test runs two
// otfd daemons:
//
// otfd1) otfd to which the TF CLI connects
// otfd2) otfd to which the otf-agent connects
//
// This setup provides a limited demonstration that the cluster is co-ordinating
// processes successfully, e.g. relaying of logs from the agent through to the
// TF CLI
func TestCluster(t *testing.T) {
	integrationTest(t)

	// start two daemons, one for user, one for agent, both sharing a db
	connstr := withDatabase(sql.NewTestDB(t))
	otfd1, org, ctx := setup(t, connstr)
	otfd2, _, _ := setup(t, connstr)

	pool, err := otfd1.Runners.CreateAgentPool(ctx, runner.CreateAgentPoolOptions{
		Organization: org.Name,
		Name:         "pool-1",
	})
	require.NoError(t, err)

	// start agent, instructing it to connect to otfd2,
	// add --debug flag, which dumps info that this test relies upon
	otfd2.startAgent(t, ctx, org.Name, &pool.ID, "", withRunnerDebug())

	// create root module, setting otfd1 as hostname
	root := newRootModule(t, otfd1.System.Hostname(), org.Name, "dev")

	// terraform init automatically creates a workspace named dev
	otfd1.engineCLI(t, ctx, "", "init", root)

	// edit workspace to use agent
	out := otfd1.otfCLI(t, ctx, "workspaces", "edit", "dev", "--organization", org.Name.String(), "--execution-mode", "agent", "--agent-pool-id", pool.ID.String())
	assert.Equal(t, "updated workspace\n", out)

	// terraform plan
	out = otfd1.engineCLI(t, ctx, "", "plan", root)
	require.Contains(t, out, "Plan: 1 to add, 0 to change, 0 to destroy.")
	require.Contains(t, out, "External agent: true") // confirm run was handled by external agent

	// terraform apply
	out = otfd1.engineCLI(t, ctx, "", "apply", root, "-auto-approve")
	require.Contains(t, out, "Apply complete! Resources: 1 added, 0 changed, 0 destroyed.")
	require.Contains(t, out, "External agent: true") // confirm run was handled by external agent
}
