package state

import "encoding/json"

// File is the terraform state file contents
type File struct {
	Version int
	Serial  int64
	Lineage string
	Outputs map[string]FileOutput
}

// FileOutput is an output in the terraform state file
type FileOutput struct {
	Value     json.RawMessage
	Sensitive bool
}
