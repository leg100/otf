package api

import (
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/auth"
)

func (m *jsonapiMarshaler) toUser(from *auth.User) *types.User {
	return &types.User{
		ID:       from.ID,
		Username: from.Username,
	}
}
