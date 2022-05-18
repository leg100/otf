package otf

import "github.com/leg100/otf/sql/pggen"

func unmarshalTokenDBType(typ pggen.Tokens) (*Token, error) {
	token := Token{
		ID: typ.UserID,
		Timestamps: Timestamps{
			CreatedAt: typ.CreatedAt,
			UpdatedAt: typ.UpdatedAt,
		},
		Token:       typ.TokenID,
		Description: typ.Description,
		UserID:      typ.UserID,
	}

	return &token, nil
}
