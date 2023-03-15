package state

import (
	"encoding/json"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/require"
)

// newTestVersion creates a new state.Version for testing purposes
func newTestVersion(t *testing.T, ws *workspace.Workspace, outputKVs ...string) *version {
	// create empty terraform state file
	f := file{
		Version: DefaultStateVersion,
		Serial:  1,
	}
	f.Outputs = make(map[string]fileOutput, len(outputKVs))
	for i := 0; i < len(outputKVs); i += 2 {
		f.Outputs[outputKVs[i]] = fileOutput{
			Value: []byte(outputKVs[i+1]),
		}
	}

	// encode state file as json
	encoded, err := json.Marshal(f)
	require.NoError(t, err)

	// wrap it in a version and return
	version, err := newVersion(CreateStateVersionOptions{
		State:       encoded,
		WorkspaceID: otf.String(ws.ID),
	})
	require.NoError(t, err)
	return version
}
