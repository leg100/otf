package otf

import (
	"github.com/leg100/otf/sql/pggen"
)

type StateVersionOutput struct {
	id        string
	Name      string
	Sensitive bool
	Type      string
	Value     string
}

func (svo *StateVersionOutput) ID() string     { return svo.id }
func (svo *StateVersionOutput) String() string { return svo.id }

// UnmarshalStateVersionOutputRow unmarshals a database row into a state version
// output.
func UnmarshalStateVersionOutputRow(row pggen.StateVersionOutputs) *StateVersionOutput {
	return &StateVersionOutput{
		id:        row.StateVersionOutputID.String,
		Sensitive: row.Sensitive,
		Type:      row.Type.String,
		Value:     row.Value.String,
		Name:      row.Name.String,
	}
}
