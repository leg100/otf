package ots

import "encoding/json"

// State represents the schema of terraform state.
type State struct {
	Version int
	Serial  int64
	Lineage string
	Outputs map[string]StateOutput
}

// StateOutput represents an output in terraform state.
type StateOutput struct {
	Value string
	Type  string
}

// Parse unmarshals terraform state from a raw byte slice into a State object.
func Parse(data []byte) (*State, error) {
	state := State{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
