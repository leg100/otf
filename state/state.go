// Package state manages terraform state.
package state

const (
	DefaultStateVersion = 4
)

// State is terraform state i.e. the state file
type State struct {
	Version int
	Serial  int64
	Lineage string
	Outputs map[string]StateOutput
}

// StateOutput is a terraform state output.
type StateOutput struct {
	Name      string
	Value     string
	Type      string
	Sensitive bool
}
