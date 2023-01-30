package state

import "testing"

// newTestVersion creates a new state.Version for testing purposes
func newTestVersion(t *testing.T, outputs ...StateOutput) *State {
	state := State{
		Version: DefaultStateVersion,
		Serial:  1,
	}
	state.Outputs = make(map[string]StateOutput, len(outputs))
	for _, out := range outputs {
		state.Outputs[out.Name] = out
	}
	return &state
}
