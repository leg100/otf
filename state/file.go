package state

import "encoding/json"

const (
	DefaultStateVersion = 4
)

// file is the terraform state file contents
type file struct {
	Version int
	Serial  int64
	Lineage string
	Outputs map[string]fileOutput
}

// fileOutput is an output in the terraform state file
type fileOutput struct {
	Value     json.RawMessage
	Sensitive bool
}
