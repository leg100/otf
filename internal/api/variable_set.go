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
	r.HandleFunc("/organizations/{organization_name}/varsets", a.listVariableSets).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}", a.getVariableSet).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}", a.updateVariableSet).Methods("PATCH")
	r.HandleFunc("/varsets/{varset_id}", a.deleteVariableSet).Methods("DELETE")
}

func (a *api) createVariableSet(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("organization", r)
	if err != nil {
		Error(w, err)
		return
	}
	var params types.VariableSetCreateOptions
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}
	variable, err := a.CreateVariableSet(r.Context(), workspaceID, variable.CreateVariableSetOptions{
		Name:        params.Name,
		Description: params.Description,
		Global:      params.Global,
	})
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, variable, withCode(http.StatusCreated))
}

func (a *api) updateVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		Error(w, err)
		return
	}
	var params types.VariableSetUpdateOptions
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}
	set, err := a.UpdateVariableSet(r.Context(), setID, variable.UpdateVariableSetOptions{
		Name:        params.Name,
		Description: params.Description,
		Global:      params.Global,
	})
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, set)
}

func (a *api) listVariableSets(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}

	variables, err := a.ListVariableSets(r.Context(), organization)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, variables)
}

func (a *api) getVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	set, err := a.GetVariableSet(r.Context(), setID)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, set)
}

func (a *api) deleteVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	if err := a.DeleteVariableSet(r.Context(), setID); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
