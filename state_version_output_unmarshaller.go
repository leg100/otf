package otf

import "github.com/leg100/otf/sql/pggen"

func UnmarshalStateVersionOutputDBType(typ pggen.StateVersionOutputs) (*StateVersionOutput, error) {
	out := StateVersionOutput{
		id:             typ.StateVersionOutputID,
		Sensitive:      typ.Sensitive,
		Type:           typ.Type,
		Value:          typ.Value,
		Name:           typ.Name,
		StateVersionID: typ.StateVersionID,
	}

	return &out, nil
}
