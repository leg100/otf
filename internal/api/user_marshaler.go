package api

import (
	"github.com/leg100/otf/api/types"
	"github.com/leg100/otf/auth"
)

func (m *jsonapiMarshaler) toUser(from *auth.User) *types.User {
	return &types.User{
		ID:       from.ID,
		Username: from.Username,
	}
}
