package otf

import "github.com/leg100/otf/sql/pggen"

type StateVersionOutput struct {
	id        string
	Name      string
	Sensitive bool
	Type      string
	Value     string
}

func (svo *StateVersionOutput) ID() string     { return svo.id }
func (svo *StateVersionOutput) String() string { return svo.id }

type StateVersionOutputList []*StateVersionOutput

// UnmarshalStateVersionOutputDBType unmarshals a state version output postgres
// composite type.
func UnmarshalStateVersionOutputDBType(typ pggen.StateVersionOutputs) *StateVersionOutput {
	return &StateVersionOutput{
		id:        typ.StateVersionOutputID.String,
		Sensitive: typ.Sensitive,
		Type:      typ.Type.String,
		Value:     typ.Value.String,
		Name:      typ.Name.String,
	}
}
