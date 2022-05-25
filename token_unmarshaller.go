package otf

import "github.com/leg100/otf/sql/pggen"

func unmarshalTokenDBType(typ pggen.Tokens) (*Token, error) {
	token := Token{
		id:          typ.UserID,
		createdAt:   typ.CreatedAt,
		token:       typ.TokenID,
		description: typ.Description,
		userID:      typ.UserID,
	}
	return &token, nil
}
