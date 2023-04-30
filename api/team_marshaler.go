package api

import (
	"github.com/leg100/otf/api/types"
	"github.com/leg100/otf/auth"
)

func (m *jsonapiMarshaler) toTeam(from *auth.Team) *types.Team {
	return &types.Team{
		ID:   from.ID,
		Name: from.Name,
	}
}
