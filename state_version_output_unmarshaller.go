package otf

import "github.com/leg100/otf/sql/pggen"

func UnmarshalStateVersionOutputDBType(typ pggen.StateVersionOutputs) (*StateVersionOutput, error) {
	out := StateVersionOutput{
		ID:        typ.StateVersionOutputID,
		Sensitive: typ.Sensitive,
		Timestamps: Timestamps{
			CreatedAt: typ.CreatedAt.Local(),
			UpdatedAt: typ.UpdatedAt.Local(),
		},
		Type:           typ.Type,
		Value:          typ.Value,
		Name:           typ.Name,
		StateVersionID: typ.StateVersionID,
	}

	return &out, nil
}
