package cli

import (
	"bytes"
	"testing"

	"github.com/leg100/otf/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_State(t *testing.T) {
	t.Run("rollback", func(t *testing.T) {
		sv := &state.Version{ID: "sv-456"}
		cmd := fakeApp(withStateVersion(sv)).stateRollbackCommand()

		cmd.SetArgs([]string{"sv-123"})
		got := bytes.Buffer{}
		cmd.SetOut(&got)
		require.NoError(t, cmd.Execute())

		assert.Equal(t, "Successfully rolled back state\n", got.String())
	})
}
