package integration

import (
	"encoding/json"
	"testing"

	"github.com/leg100/otf/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_StateCLI demonstrates managing state via the CLI
func TestIntegration_StateCLI(t *testing.T) {
	t.Parallel()

	daemon := setup(t, nil)
	_, ctx := daemon.createUserCtx(t, ctx)

	t.Run("download", func(t *testing.T) {
		sv := daemon.createStateVersion(t, ctx, nil)
		want := unmarshalState(t, sv.State)

		out := daemon.otfcli(t, ctx, "state", "download", sv.ID)
		got := unmarshalState(t, []byte(out))

		assert.Equal(t, want, got)
	})

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

func unmarshalState(t *testing.T, contents []byte) *state.File {
	f := &state.File{}
	err := json.Unmarshal(contents, &f)
	require.NoError(t, err)
	return f
}
