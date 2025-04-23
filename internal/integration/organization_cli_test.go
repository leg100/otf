package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestOrganizationCLI tests managing organizations via the CLI
func TestOrganizationCLI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t)

	// create organization
	out := daemon.otfCLI(t, ctx, "organizations", "new", "acme-corp")
	require.Equal(t, "Successfully created organization acme-corp\n", out)

	// delete organization
	out = daemon.otfCLI(t, ctx, "organizations", "delete", "acme-corp")
	require.Equal(t, "Successfully deleted organization acme-corp\n", string(out))
}
