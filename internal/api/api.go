// Package api provides http handlers for the API.
package api

import (
	"io"

	"github.com/DataDog/jsonapi"
)

func Unmarshal(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return jsonapi.Unmarshal(b, v)
}

//func (a *api) AddHandlers(r *mux.Router) {
//	a.addOrganizationHandlers(r)
//	a.addRunHandlers(r)
//	a.addWorkspaceHandlers(r)
//	a.addStateHandlers(r)
//	a.addTagHandlers(r)
//	a.addConfigHandlers(r)
//	a.addTeamMembershipHandlers(r)
//	a.addVariableHandlers(r)
//	a.addTokenHandlers(r)
//	a.addNotificationHandlers(r)
//	a.addOrganizationMembershipHandlers(r)
//	a.addOAuthClientHandlers(r)
//}
