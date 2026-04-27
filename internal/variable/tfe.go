package variable

import (
	"context"
	"net/http"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
)

// Implements TFC workspace variables and variable set APIs:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/workspace-variables
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/variable-sets
type TFEAPI struct {
	*tfeapi.Responder
	Client tfeClient
}

type tfeClient interface {
	CreateWorkspaceVariable(ctx context.Context, workspaceID resource.TfeID, opts CreateVariableOptions) (*Variable, error)
	ListWorkspaceVariables(ctx context.Context, workspaceID resource.TfeID) ([]*Variable, error)
	GetVariable(ctx context.Context, variableID resource.TfeID) (*Variable, error)
	UpdateVariable(ctx context.Context, variableID resource.TfeID, opts UpdateVariableOptions) (*Variable, error)
	DeleteVariable(ctx context.Context, variableID resource.TfeID) (*Variable, error)

	ListWorkspaceVariableSets(ctx context.Context, workspaceID resource.TfeID) ([]*VariableSet, error)
	ListVariableSets(ctx context.Context, organization organization.Name) ([]*VariableSet, error)
	GetVariableSet(ctx context.Context, setID resource.TfeID) (*VariableSet, error)
	CreateVariableSet(ctx context.Context, org organization.Name, opts CreateVariableSetOptions) (*VariableSet, error)
	UpdateVariableSet(ctx context.Context, setID resource.TfeID, opts UpdateVariableSetOptions) (*VariableSet, error)
	DeleteVariableSet(ctx context.Context, setID resource.TfeID) (*VariableSet, error)
	GetVariableSetByVariableID(ctx context.Context, variableID resource.TfeID) (*VariableSet, error)
	CreateVariableSetVariable(ctx context.Context, setID resource.TfeID, opts CreateVariableOptions) (*Variable, error)
	UpdateVariableSetVariable(ctx context.Context, variableID resource.TfeID, opts UpdateVariableOptions) (*VariableSet, error)
	DeleteVariableSetVariable(ctx context.Context, variableID resource.TfeID) (*VariableSet, error)
	GetWorkspace(context.Context, resource.TfeID) (*workspace.Workspace, error)
	ListWorkspaces(ctx context.Context, opts workspace.ListOptions) (*resource.Page[*workspace.Workspace], error)

	ApplySetToWorkspaces(ctx context.Context, setID resource.TfeID, workspaceIDs []resource.TfeID) error
	DeleteSetFromWorkspaces(ctx context.Context, setID resource.TfeID, workspaceIDs []resource.TfeID) error
}

func (a *TFEAPI) AddHandlers(r *mux.Router) {
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

func (a *TFEAPI) createWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
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
	v, err := a.Client.CreateWorkspaceVariable(r.Context(), workspaceID, CreateVariableOptions{
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

func (a *TFEAPI) get(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	v, err := a.Client.GetVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertWorkspaceVariable(v, v.ParentID), http.StatusOK)
}

func (a *TFEAPI) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	variables, err := a.Client.ListWorkspaceVariables(r.Context(), workspaceID)
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

func (a *TFEAPI) update(w http.ResponseWriter, r *http.Request) {
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
	updated, err := a.Client.UpdateVariable(r.Context(), variableID, UpdateVariableOptions{
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

	a.Respond(w, r, a.convertWorkspaceVariable(updated, updated.ParentID), http.StatusOK)
}

func (a *TFEAPI) delete(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	_, err = a.Client.DeleteVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
}

func (a *TFEAPI) createVariableSet(w http.ResponseWriter, r *http.Request) {
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
	set, err := a.Client.CreateVariableSet(r.Context(), pathParams.Organization, CreateVariableSetOptions{
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

func (a *TFEAPI) updateVariableSet(w http.ResponseWriter, r *http.Request) {
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
	set, err := a.Client.UpdateVariableSet(r.Context(), setID, UpdateVariableSetOptions{
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

func (a *TFEAPI) listVariableSets(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	sets, err := a.Client.ListVariableSets(r.Context(), pathParams.Organization)
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

func (a *TFEAPI) listWorkspaceVariableSets(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	sets, err := a.Client.ListWorkspaceVariableSets(r.Context(), workspaceID)
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

func (a *TFEAPI) getVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Client.GetVariableSet(r.Context(), setID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convertVariableSet(set), http.StatusOK)
}

func (a *TFEAPI) deleteVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if _, err := a.Client.DeleteVariableSet(r.Context(), setID); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *TFEAPI) listVariableSetVariables(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("varset_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Client.GetVariableSet(r.Context(), setID)
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

func (a *TFEAPI) addVariableToSet(w http.ResponseWriter, r *http.Request) {
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

	v, err := a.Client.CreateVariableSetVariable(r.Context(), setID, CreateVariableOptions{
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

func (a *TFEAPI) updateVariableSetVariable(w http.ResponseWriter, r *http.Request) {
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

	set, err := a.Client.UpdateVariableSetVariable(r.Context(), variableID, UpdateVariableOptions{
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

func (a *TFEAPI) getVariableSetVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	set, err := a.Client.GetVariableSetByVariableID(r.Context(), variableID)
	if err != nil {
		variableError(w, err)
		return
	}

	v := set.GetVariableByID(variableID)
	a.Respond(w, r, a.convertVariableSetVariable(v, set.ID), http.StatusOK)
}

func (a *TFEAPI) deleteVariableFromSet(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	_, err = a.Client.DeleteVariableSetVariable(r.Context(), variableID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
}

func (a *TFEAPI) applySetToWorkspaces(w http.ResponseWriter, r *http.Request) {
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

	err = a.Client.ApplySetToWorkspaces(r.Context(), setID, workspaceIDs)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *TFEAPI) deleteSetFromWorkspaces(w http.ResponseWriter, r *http.Request) {
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

	err = a.Client.DeleteSetFromWorkspaces(r.Context(), setID, workspaceIDs)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *TFEAPI) convertWorkspaceVariable(from *Variable, workspaceID resource.TfeID) *TFEWorkspaceVariable {
	return &TFEWorkspaceVariable{
		TFEVariable: a.convertVariable(from, true),
		Workspace: &workspace.TFEWorkspace{
			ID: workspaceID,
		},
	}
}

func (a *TFEAPI) convertVariableSet(from *VariableSet) *TFEVariableSet {
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

func (a *TFEAPI) convertVariableSetVariable(from *Variable, setID resource.TfeID) *TFEVariableSetVariable {
	return &TFEVariableSetVariable{
		TFEVariable: a.convertVariable(from, true),
		VariableSet: &TFEVariableSet{ID: setID},
	}
}

func (a *TFEAPI) convertVariable(from *Variable, scrubSensitiveValue bool) *TFEVariable {
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
