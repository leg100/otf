package api

import (
	"github.com/leg100/otf/api/types"
	"github.com/leg100/otf/variable"
)

// writeResponse encodes v as json:api and writes it to the body of the http response.
func (m *jsonapiMarshaler) toVariable(from *variable.Variable) *types.Variable {
	to := &types.Variable{
		ID:          from.ID,
		Key:         from.Key,
		Value:       from.Value,
		Description: from.Description,
		Category:    string(from.Category),
		Sensitive:   from.Sensitive,
		HCL:         from.HCL,
		Workspace: &types.Workspace{
			ID: from.WorkspaceID,
		},
	}
	if to.Sensitive {
		to.Value = "" // scrub sensitive values
	}
	return to
}

func (m *jsonapiMarshaler) toVariableList(from []*variable.Variable) (to []*types.Variable) {
	for _, v := range from {
		to = append(to, m.toVariable(v))
	}
	return
}
