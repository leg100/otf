package ots

import "encoding/json"

type State struct {
	Version int
	Serial  int64
	Lineage string
	Outputs map[string]StateOutput
}

type StateOutput struct {
	Value string
	Type  string
}

func Parse(data []byte) (*State, error) {
	state := State{}
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}
