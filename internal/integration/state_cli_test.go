package integration

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/leg100/otf/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_StateCLI demonstrates managing state via the CLI
func TestIntegration_StateCLI(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t)

	t.Run("list", func(t *testing.T) {
		ws := daemon.createWorkspace(t, ctx, nil)
		sv1 := daemon.createStateVersion(t, ctx, ws)
		sv2 := daemon.createStateVersion(t, ctx, ws)
		sv3 := daemon.createStateVersion(t, ctx, ws) // current

		out := daemon.otfCLI(t, ctx, "state", "list",
			"--organization", ws.Organization.String(),
			"--workspace", ws.Name,
		)

		want := fmt.Sprintf("%s (current)\n%s\n%s\n", sv3, sv2, sv1)
		assert.Equal(t, want, out)
	})

	t.Run("delete", func(t *testing.T) {
		ws := daemon.createWorkspace(t, ctx, nil)
		sv := daemon.createStateVersion(t, ctx, ws)
		// because deleting the 'current' state version is not allowed, create
		// another version which becomes the current state version, thereby
		// permitting the test to delete the previous version.
		_ = daemon.createStateVersion(t, ctx, ws)

		got := daemon.otfCLI(t, ctx, "state", "delete", sv.ID.String())

		want := fmt.Sprintf("Deleted state version: %s\n", sv.ID)
		assert.Equal(t, want, got)
	})

	t.Run("download", func(t *testing.T) {
		sv := daemon.createStateVersion(t, ctx, nil)
		want := unmarshalState(t, sv.State)

		out := daemon.otfCLI(t, ctx, "state", "download", sv.ID.String())
		got := unmarshalState(t, []byte(out))

		assert.Equal(t, want, got)
	})

	t.Run("rollback", func(t *testing.T) {
		rollbackTo := daemon.createStateVersion(t, ctx, nil)
		current := daemon.createStateVersion(t, ctx, nil)

		gotOut := daemon.otfCLI(t, ctx, "state", "rollback", rollbackTo.ID.String())
		require.Equal(t, "Successfully rolled back state\n", gotOut)

		// should be new current state
		newCurrent := daemon.getCurrentState(t, ctx, rollbackTo.WorkspaceID)
		assert.NotEqual(t, newCurrent.ID, current.ID)
	})
}

func unmarshalState(t *testing.T, contents []byte) *state.File {
	f := &state.File{}
	err := json.Unmarshal(contents, &f)
	require.NoError(t, err)
	return f
}
