package api

import (
	"net/http"

	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/variable"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
)

// Implements TFC workspace variables API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspace-variables#update-variables
func (a *api) addVariableHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/vars", a.create).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/vars", a.list).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.get).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.update).Methods("PATCH")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.delete).Methods("DELETE")
}

func (a *api) create(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}
	var opts types.VariableCreateOptions
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}
	variable, err := a.CreateVariable(r.Context(), workspaceID, variable.CreateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*variable.VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, variable, withCode(http.StatusCreated))
}

func (a *api) get(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		Error(w, err)
		return
	}
	variable, err := a.GetVariable(r.Context(), variableID)
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, variable)
}

func (a *api) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}
	variables, err := a.ListVariables(r.Context(), workspaceID)
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, variables)
}

func (a *api) update(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		Error(w, err)
		return
	}
	var opts types.VariableUpdateOptions
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}
	updated, err := a.UpdateVariable(r.Context(), variableID, variable.UpdateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*variable.VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		Error(w, err)
		return
	}
	a.writeResponse(w, r, updated)
}

func (a *api) delete(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		Error(w, err)
		return
	}
	_, err = a.DeleteVariable(r.Context(), variableID)
	if err != nil {
		Error(w, err)
		return
	}
}
