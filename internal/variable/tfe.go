package variable

import (
	"errors"
	"net/http"

	"github.com/leg100/otf/internal"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
)

type tfe struct {
	Service
	*tfeapi.Responder
}

// Implements TFC workspace variables and variable set APIs:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspace-variables
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/variable-sets
func (a *tfe) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/vars", a.create).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/vars", a.list).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.get).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.update).Methods("PATCH")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.delete).Methods("DELETE")

	r.HandleFunc("/organizations/{organization_name}/varsets", a.createVariableSet).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/varsets", a.listVariableSets).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}", a.getVariableSet).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}", a.updateVariableSet).Methods("PATCH")
	r.HandleFunc("/varsets/{varset_id}", a.deleteVariableSet).Methods("DELETE")
}

func (a *tfe) create(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.VariableCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	variable, err := a.CreateVariable(r.Context(), workspaceID, CreateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		variableError(w, err)
		return
	}
	a.Respond(w, r, a.convert(variable), http.StatusCreated)
}

func (a *tfe) get(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	variable, err := a.GetVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(variable), http.StatusOK)
}

func (a *tfe) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	variables, err := a.ListVariables(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	to := make([]*types.Variable, len(variables))
	for i, from := range variables {
		to[i] = a.convert(from)
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) update(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.VariableUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		variableError(w, err)
		return
	}
	updated, err := a.UpdateVariable(r.Context(), variableID, UpdateVariableOptions{
		Key:         opts.Key,
		Value:       opts.Value,
		Description: opts.Description,
		Category:    (*VariableCategory)(opts.Category),
		Sensitive:   opts.Sensitive,
		HCL:         opts.HCL,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(updated), http.StatusOK)
}

func (a *tfe) delete(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	_, err = a.DeleteVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
}

func (a *tfe) createVariableSet(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("organization", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params types.VariableSetCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	variable, err := a.CreateVariableSet(r.Context(), workspaceID, CreateVariableSetOptions{
		Name:        params.Name,
		Description: params.Description,
		Global:      params.Global,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, variable, http.StatusCreated)
}

func (a *tfe) updateVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params types.VariableSetUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	set, err := a.UpdateVariableSet(r.Context(), setID, UpdateVariableSetOptions{
		Name:        params.Name,
		Description: params.Description,
		Global:      params.Global,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, set, http.StatusOK)
}

func (a *tfe) listVariableSets(w http.ResponseWriter, r *http.Request) {
	organization, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	sets, err := a.ListVariableSets(r.Context(), organization)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	to := make([]*types.Variable, len(sets))
	for i, from := range sets {
		to[i] = a.convert(from)
	}
	a.Respond(w, r, sets, http.StatusOK)
}

func (a *tfe) getVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.GetVariableSet(r.Context(), setID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, set, http.StatusOK)
}

func (a *tfe) deleteVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.DeleteVariableSet(r.Context(), setID); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convert(from *Variable) *types.Variable {
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

func variableError(w http.ResponseWriter, err error) {
	var isUnprocessableError bool
	if errors.Is(err, ErrVariableDescriptionMaxExceeded) {
		isUnprocessableError = true
	}
	if errors.Is(err, ErrVariableKeyMaxExceeded) {
		isUnprocessableError = true
	}
	if errors.Is(err, ErrVariableValueMaxExceeded) {
		isUnprocessableError = true
	}
	if isUnprocessableError {
		tfeapi.Error(w, &internal.HTTPError{
			Message: err.Error(),
			Code:    http.StatusUnprocessableEntity,
		})
	} else {
		tfeapi.Error(w, err)
	}
}
