package otf

import (
	"github.com/leg100/otf/sql/pggen"
)

func UnmarshalSessionDBType(typ pggen.Sessions) (*Session, error) {
	session := Session{
		Token: typ.Token,
		Timestamps: Timestamps{
			CreatedAt: typ.CreatedAt,
			UpdatedAt: typ.UpdatedAt,
		},
		Expiry: typ.Expiry,
		UserID: typ.UserID,
		SessionData: SessionData{
			Address: typ.Address,
		},
	}

	return &session, nil
}
