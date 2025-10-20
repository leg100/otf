package variable

import (
	"net/http"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
)

type tfe struct {
	*tfeapi.Responder
	*Service
}

// Implements TFC workspace variables and variable set APIs:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspace-variables
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/variable-sets
func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/workspaces/{workspace_id}/vars", a.createWorkspaceVariable).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/vars", a.list).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.get).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.update).Methods("PATCH")
	r.HandleFunc("/workspaces/{workspace_id}/vars/{variable_id}", a.delete).Methods("DELETE")

	r.HandleFunc("/organizations/{organization_name}/varsets", a.createVariableSet).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/varsets", a.listVariableSets).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/varsets", a.listWorkspaceVariableSets).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}", a.getVariableSet).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}", a.updateVariableSet).Methods("PATCH")
	r.HandleFunc("/varsets/{varset_id}", a.deleteVariableSet).Methods("DELETE")

	r.HandleFunc("/varsets/{varset_id}/relationships/vars", a.listVariableSetVariables).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}/relationships/vars", a.addVariableToSet).Methods("POST")
	r.HandleFunc("/varsets/{varset_id}/relationships/vars/{variable_id}", a.getVariableSetVariable).Methods("GET")
	r.HandleFunc("/varsets/{varset_id}/relationships/vars/{variable_id}", a.updateVariableSetVariable).Methods("PATCH")
	r.HandleFunc("/varsets/{varset_id}/relationships/vars/{variable_id}", a.deleteVariableFromSet).Methods("DELETE")
	r.HandleFunc("/varsets/{varset_id}/relationships/workspaces", a.applySetToWorkspaces).Methods("POST")
	r.HandleFunc("/varsets/{varset_id}/relationships/workspaces", a.deleteSetFromWorkspaces).Methods("DELETE")
}

func (a *tfe) createWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts TFEVariableCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}
	v, err := a.Service.CreateWorkspaceVariable(r.Context(), workspaceID, CreateVariableOptions{
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
	a.Respond(w, r, a.convertWorkspaceVariable(v, workspaceID), http.StatusCreated)
}

func (a *tfe) get(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	wv, err := a.GetWorkspaceVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertWorkspaceVariable(wv.Variable, wv.WorkspaceID), http.StatusOK)
}

func (a *tfe) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	variables, err := a.ListWorkspaceVariables(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := make([]*TFEWorkspaceVariable, len(variables))
	for i, from := range variables {
		to[i] = a.convertWorkspaceVariable(from, workspaceID)
	}

	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) update(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts TFEVariableUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		variableError(w, err)
		return
	}
	updated, err := a.UpdateWorkspaceVariable(r.Context(), variableID, UpdateVariableOptions{
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

	a.Respond(w, r, a.convertWorkspaceVariable(updated.Variable, updated.WorkspaceID), http.StatusOK)
}

func (a *tfe) delete(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	_, err = a.DeleteWorkspaceVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
}

func (a *tfe) createVariableSet(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFEVariableSetCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	set, err := a.Service.CreateVariableSet(r.Context(), pathParams.Organization, CreateVariableSetOptions{
		Name:        params.Name,
		Description: params.Description,
		Global:      params.Global,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convertVariableSet(set), http.StatusCreated)
}

func (a *tfe) updateVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params TFEVariableSetUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	set, err := a.Service.UpdateVariableSet(r.Context(), setID, UpdateVariableSetOptions{
		Name:        params.Name,
		Description: params.Description,
		Global:      params.Global,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, a.convertVariableSet(set), http.StatusOK)
}

func (a *tfe) listVariableSets(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	sets, err := a.Service.ListVariableSets(r.Context(), pathParams.Organization)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := make([]*TFEVariableSet, len(sets))
	for i, from := range sets {
		to[i] = a.convertVariableSet(from)
	}

	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) listWorkspaceVariableSets(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	sets, err := a.Service.ListWorkspaceVariableSets(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := make([]*TFEVariableSet, len(sets))
	for i, from := range sets {
		to[i] = a.convertVariableSet(from)
	}

	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) getVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Service.GetVariableSet(r.Context(), setID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertVariableSet(set), http.StatusOK)
}

func (a *tfe) deleteVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if _, err := a.Service.DeleteVariableSet(r.Context(), setID); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) listVariableSetVariables(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Service.GetVariableSet(r.Context(), setID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := make([]*TFEVariableSetVariable, len(set.Variables))
	for i, from := range set.Variables {
		to[i] = a.convertVariableSetVariable(from, setID)
	}

	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) addVariableToSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts TFEVariableCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	v, err := a.Service.CreateVariableSetVariable(r.Context(), setID, CreateVariableOptions{
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

	a.Respond(w, r, a.convertVariableSetVariable(v, setID), http.StatusOK)
}

func (a *tfe) updateVariableSetVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts TFEVariableUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Service.UpdateVariableSetVariable(r.Context(), variableID, UpdateVariableOptions{
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

	v := set.GetVariableByID(variableID)
	a.Respond(w, r, a.convertVariableSetVariable(v, set.ID), http.StatusOK)
}

func (a *tfe) getVariableSetVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Service.GetVariableSetByVariableID(r.Context(), variableID)
	if err != nil {
		variableError(w, err)
		return
	}

	v := set.GetVariableByID(variableID)
	a.Respond(w, r, a.convertVariableSetVariable(v, set.ID), http.StatusOK)
}

func (a *tfe) deleteVariableFromSet(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	_, err = a.Service.DeleteVariableSetVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
}

func (a *tfe) applySetToWorkspaces(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []*TFEWorkspace
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	workspaceIDs := make([]resource.TfeID, len(params))
	for i, ws := range params {
		workspaceIDs[i] = ws.ID
	}

	err = a.Service.applySetToWorkspaces(r.Context(), setID, workspaceIDs)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) deleteSetFromWorkspaces(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []*TFEWorkspace
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	workspaceIDs := make([]resource.TfeID, len(params))
	for i, ws := range params {
		workspaceIDs[i] = ws.ID
	}

	err = a.Service.deleteSetFromWorkspaces(r.Context(), setID, workspaceIDs)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convertWorkspaceVariable(from *Variable, workspaceID resource.TfeID) *TFEWorkspaceVariable {
	return &TFEWorkspaceVariable{
		TFEVariable: a.convertVariable(from, true),
		Workspace: &workspace.TFEWorkspace{
			ID: workspaceID,
		},
	}
}

func (a *tfe) convertVariableSet(from *VariableSet) *TFEVariableSet {
	to := &TFEVariableSet{
		ID:          from.ID,
		Name:        from.Name,
		Description: from.Description,
		Global:      from.Global,
		Organization: &organization.TFEOrganization{
			Name: from.Organization,
		},
	}
	to.Variables = make([]*TFEVariableSetVariable, len(from.Variables))
	for i, v := range from.Variables {
		to.Variables[i] = &TFEVariableSetVariable{
			TFEVariable: a.convertVariable(v, true),
			VariableSet: &TFEVariableSet{
				ID: v.ID,
			},
		}
	}
	to.Workspaces = make([]*workspace.TFEWorkspace, len(from.Workspaces))
	for i, workspaceID := range from.Workspaces {
		to.Workspaces[i] = &workspace.TFEWorkspace{
			ID: workspaceID,
		}
	}
	return to
}

func (a *tfe) convertVariableSetVariable(from *Variable, setID resource.TfeID) *TFEVariableSetVariable {
	return &TFEVariableSetVariable{
		TFEVariable: a.convertVariable(from, true),
		VariableSet: &TFEVariableSet{ID: setID},
	}
}

func (a *tfe) convertVariable(from *Variable, scrubSensitiveValue bool) *TFEVariable {
	to := &TFEVariable{
		ID:          from.ID,
		Key:         from.Key,
		Value:       from.Value,
		Description: from.Description,
		Category:    string(from.Category),
		Sensitive:   from.Sensitive,
		HCL:         from.HCL,
		VersionID:   from.VersionID,
	}
	if to.Sensitive && scrubSensitiveValue {
		to.Value = "" // scrub sensitive values
	}
	return to
}

func variableError(w http.ResponseWriter, err error) {
	if internal.ErrorIs(err,
		ErrVariableDescriptionMaxExceeded,
		ErrVariableKeyMaxExceeded,
		ErrVariableValueMaxExceeded,
	) {
		tfeapi.Error(w, err, tfeapi.WithStatus(http.StatusUnprocessableEntity))
		return
	}
	tfeapi.Error(w, err)
}
