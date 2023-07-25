package api

import (
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/variable"
)

func (m *jsonapiMarshaler) toVariable(from *variable.Variable) *types.Variable {
	to := &types.Variable{
		ID:          from.ID,
		Key:         from.Key,
		Value:       from.Value,
		Description: from.Description,
		Category:    string(from.Category),
		Sensitive:   from.Sensitive,
		HCL:         from.HCL,
		VersionID:   from.VersionID,
		Workspace: &types.Workspace{
			ID: from.WorkspaceID,
		},
	}
	if to.Sensitive {
		to.Value = "" // scrub sensitive values
	}
	return to
}
