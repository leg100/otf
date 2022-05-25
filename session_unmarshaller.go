package otf

import (
	"github.com/leg100/otf/sql/pggen"
)

func UnmarshalSessionDBType(typ pggen.Sessions) (*Session, error) {
	session := Session{
		Token:     typ.Token,
		createdAt: typ.CreatedAt,
		Expiry:    typ.Expiry,
		UserID:    typ.UserID,
		SessionData: SessionData{
			Address: typ.Address,
		},
	}
	return &session, nil
}
