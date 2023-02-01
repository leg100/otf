package state

import (
	"encoding/json"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/require"
)

// newTestVersion creates a new state.Version for testing purposes
func newTestVersion(t *testing.T, ws *otf.Workspace, outputs ...StateOutput) *version {
	// create empty terraform state
	state := State{
		Version: DefaultStateVersion,
		Serial:  1,
	}
	state.Outputs = make(map[string]StateOutput, len(outputs))
	for _, out := range outputs {
		state.Outputs[out.Name] = out
	}

	// marshal it into json
	js, err := json.Marshal(state)
	require.NoError(t, err)

	// wrap it in a version and return
	version, err := newVersion(otf.CreateStateVersionOptions{
		State:       js,
		WorkspaceID: otf.String(ws.ID()),
	})
	require.NoError(t, err)
	return version
}
