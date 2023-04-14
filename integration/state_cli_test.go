package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_StateCLI demonstrates managing state via the CLI
func TestIntegration_StateCLI(t *testing.T) {
	t.Parallel()

	daemon := setup(t, nil)
	_, ctx := daemon.createUserCtx(t, ctx)

	t.Run("rollback", func(t *testing.T) {
		rollbackTo := daemon.createStateVersion(t, ctx, nil)
		current := daemon.createStateVersion(t, ctx, nil)

		gotOut := daemon.otfcli(t, ctx, "state", "rollback", rollbackTo.ID)
		require.Equal(t, "Successfully rolled back state\n", gotOut)

		// should be new current state
		newCurrent := daemon.getCurrentState(t, ctx, rollbackTo.WorkspaceID)
		assert.NotEqual(t, newCurrent.ID, current.ID)
	})
}
