package api

import (
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/auth"
)

func (m *jsonapiMarshaler) toTeam(from *auth.Team) *types.Team {
	return &types.Team{
		ID:   from.ID,
		Name: from.Name,
	}
}
