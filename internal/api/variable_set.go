package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/variable"
)

// Implements TFC variable set API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/variable-sets
func (a *api) addVariableSetHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/varsets", a.createVariableSet).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/varsets", a.list).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/varsets/{variable_id}", a.get).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/varsets/{variable_id}", a.update).Methods("PATCH")
	r.HandleFunc("/organizations/{organization_name}/varsets/{variable_id}", a.delete).Methods("DELETE")
}

func (a *api) createVariableSet(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("organization", r)
	if err != nil {
		Error(w, err)
		return
	}
	var params types.VariableSetVariableCreateOptions
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}
	variable, err := a.CreateVariable(r.Context(), workspaceID, variable.CreateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    (*variable.VariableCategory)(params.Category),
		Sensitive:   params.Sensitive,
		HCL:         params.HCL,
	})
	if err != nil {
		variableError(w, err)
		return
	}
	a.writeResponse(w, r, variable, withCode(http.StatusCreated))
}
