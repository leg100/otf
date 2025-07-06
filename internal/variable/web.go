package variable

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
)

type (
	web struct {
		workspaces webWorkspaceClient

		variables  webVariablesClient
		authorizer webAuthorizer
	}

	// webVariablesClient provides web handlers with access to variables
	webVariablesClient interface {
		CreateWorkspaceVariable(ctx context.Context, workspaceID resource.TfeID, opts CreateVariableOptions) (*Variable, error)
		GetWorkspaceVariable(ctx context.Context, variableID resource.TfeID) (*WorkspaceVariable, error)
		ListWorkspaceVariables(ctx context.Context, workspaceID resource.TfeID) ([]*Variable, error)
		listWorkspaceVariableSets(ctx context.Context, workspaceID resource.TfeID) ([]*VariableSet, error)
		UpdateWorkspaceVariable(ctx context.Context, variableID resource.TfeID, opts UpdateVariableOptions) (*WorkspaceVariable, error)
		DeleteWorkspaceVariable(ctx context.Context, variableID resource.TfeID) (*WorkspaceVariable, error)

		CreateVariableSet(ctx context.Context, organization organization.Name, opts CreateVariableSetOptions) (*VariableSet, error)
		updateVariableSet(ctx context.Context, setID resource.TfeID, opts UpdateVariableSetOptions) (*VariableSet, error)
		getVariableSet(ctx context.Context, setID resource.TfeID) (*VariableSet, error)
		getVariableSetByVariableID(ctx context.Context, variableID resource.TfeID) (*VariableSet, error)
		listVariableSets(ctx context.Context, organization organization.Name) ([]*VariableSet, error)
		deleteVariableSet(ctx context.Context, setID resource.TfeID) (*VariableSet, error)
		CreateVariableSetVariable(ctx context.Context, setID resource.TfeID, opts CreateVariableOptions) (*Variable, error)
		updateVariableSetVariable(ctx context.Context, variableID resource.TfeID, opts UpdateVariableOptions) (*VariableSet, error)
		deleteVariableSetVariable(ctx context.Context, variableID resource.TfeID) (*VariableSet, error)
	}

	// webWorkspaceClient provides web handlers with access to workspaces
	webWorkspaceClient interface {
		Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
		List(ctx context.Context, opts workspace.ListOptions) (*resource.Page[*workspace.Workspace], error)
	}

	webAuthorizer interface {
		CanAccess(context.Context, authz.Action, resource.ID) bool
	}

	workspaceInfo struct {
		ID   resource.TfeID `json:"id"`
		Name string         `json:"name"`
	}

	createVariableParams struct {
		Key         *string `schema:"key,required"`
		Value       *string
		Description *string
		Category    *VariableCategory `schema:"category,required"`
		Sensitive   bool
		HCL         bool
	}

	updateVariableParams struct {
		Key         *string
		Value       *string
		Description *string
		Category    *VariableCategory
		Sensitive   *bool
		HCL         bool
		VariableID  resource.TfeID `schema:"variable_id,required"`
	}
)

func (h *web) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/variables", h.listWorkspaceVariables).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/variables/new", h.newWorkspaceVariable).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/variables/create", h.createWorkspaceVariable).Methods("POST")
	r.HandleFunc("/variables/{variable_id}/edit", h.editWorkspaceVariable).Methods("GET")
	r.HandleFunc("/variables/{variable_id}/update", h.updateWorkspaceVariable).Methods("POST")
	r.HandleFunc("/variables/{variable_id}/delete", h.deleteWorkspaceVariable).Methods("POST")

	r.HandleFunc("/organizations/{organization_name}/variable-sets", h.listVariableSets).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/variable-sets/new", h.newVariableSet).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/variable-sets/create", h.createVariableSet).Methods("POST")
	r.HandleFunc("/variable-sets/{variable_set_id}/edit", h.editVariableSet).Methods("GET")
	r.HandleFunc("/variable-sets/{variable_set_id}/update", h.updateVariableSet).Methods("POST")
	r.HandleFunc("/variable-sets/{variable_set_id}/delete", h.deleteVariableSet).Methods("POST")

	r.HandleFunc("/variable-sets/{variable_set_id}/variable-set-variables/new", h.newVariableSetVariable).Methods("GET")
	r.HandleFunc("/variable-sets/{variable_set_id}/variable-set-variables/create", h.createVariableSetVariable).Methods("POST")
	r.HandleFunc("/variable-set-variables/{variable_id}/edit", h.editVariableSetVariable).Methods("GET")
	r.HandleFunc("/variable-set-variables/{variable_id}/update", h.updateVariableSetVariable).Methods("POST")
	r.HandleFunc("/variable-set-variables/{variable_id}/delete", h.deleteVariableSetVariable).Methods("POST")
}

func (h *web) newWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.workspaces.Get(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.Render(newWorkspaceVariable(ws), w, r)
}

func (h *web) createWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	var params struct {
		createVariableParams
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.variables.CreateWorkspaceVariable(r.Context(), params.WorkspaceID, CreateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		Sensitive:   &params.Sensitive,
		HCL:         &params.HCL,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.NewVariable(params.WorkspaceID), http.StatusFound)
		return
	}

	html.FlashSuccess(w, "added variable: "+variable.Key)
	http.Redirect(w, r, paths.Variables(params.WorkspaceID), http.StatusFound)
}

func (h *web) listWorkspaceVariables(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.workspaces.Get(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	variables, err := h.variables.ListWorkspaceVariables(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sets, err := h.variables.listWorkspaceVariableSets(r.Context(), workspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	merged := mergeVariables(sets, variables, nil)

	// TODO: hide delete button for set variables
	var rows []variableRow
	for _, wv := range variables {
		rows = append(rows, variableRow{Variable: wv})
	}
	for _, set := range sets {
		for _, sv := range set.Variables {
			rows = append(rows, variableRow{
				Variable: sv,
				set:      set,
				overwritten: !slices.ContainsFunc(merged, func(v *Variable) bool {
					return v.ID == sv.ID
				}),
			})
		}
	}
	slices.SortFunc(rows, func(a, b variableRow) int {
		switch {
		case a.Key > b.Key:
			return 1
		case a.Key < b.Key:
			return -1
		default:
			// key is equal; if variable belongs to a set then it comes
			// afterwards.
			if a.set != nil && b.set == nil {
				return 1
			} else if a.set == nil && b.set != nil {
				return -1
			} else if a.set != nil && b.set != nil {
				// both belong to set; sort by set's name
				if a.set.Name > b.set.Name {
					return 1
				} else if a.set.Name < b.set.Name {
					return -1
				}
			}
			return 0
		}
	})

	props := listWorkspaceVariablesProps{
		ws:                 ws,
		rows:               rows,
		canCreateVariable:  h.authorizer.CanAccess(r.Context(), authz.CreateWorkspaceVariableAction, ws.ID),
		canDeleteVariable:  h.authorizer.CanAccess(r.Context(), authz.DeleteWorkspaceVariableAction, ws.ID),
		canUpdateWorkspace: h.authorizer.CanAccess(r.Context(), authz.UpdateWorkspaceAction, ws.ID),
	}
	html.Render(listWorkspaceVariables(props), w, r)
}

func (h *web) editWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	wv, err := h.variables.GetWorkspaceVariable(r.Context(), variableID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := h.workspaces.Get(r.Context(), wv.WorkspaceID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	props := editWorkspaceVariableProps{
		ws:       ws,
		variable: wv.Variable,
	}
	html.Render(editWorkspaceVariable(props), w, r)
}

func (h *web) updateWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	var params updateVariableParams
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	wv, err := h.variables.UpdateWorkspaceVariable(r.Context(), params.VariableID, UpdateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		Sensitive:   params.Sensitive,
		HCL:         &params.HCL,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.EditVariable(params.VariableID), http.StatusFound)
		return
	}

	html.FlashSuccess(w, "updated variable: "+wv.Key)
	http.Redirect(w, r, paths.Variables(wv.WorkspaceID), http.StatusFound)
}

func (h *web) deleteWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	wv, err := h.variables.DeleteWorkspaceVariable(r.Context(), variableID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted variable: "+wv.Key)
	http.Redirect(w, r, paths.Variables(wv.WorkspaceID), http.StatusFound)
}

func (h *web) listVariableSets(w http.ResponseWriter, r *http.Request) {
	var params ListOptions
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	sets, err := h.variables.listVariableSets(r.Context(), params.Organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	props := listVariableSetsProps{
		organization:         params.Organization,
		sets:                 resource.NewPage(sets, params.PageOptions, nil),
		canCreateVariableSet: h.authorizer.CanAccess(r.Context(), authz.CreateVariableSetAction, params.Organization),
	}
	html.Render(listVariableSets(props), w, r)
}

func (h *web) newVariableSet(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// retrieve names of all workspaces in org to show in dropdown widget
	availableWorkspaces, err := h.getAvailableWorkspaces(r.Context(), pathParams.Organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	props := newVariableSetProps{
		organization:        pathParams.Organization,
		availableWorkspaces: availableWorkspaces,
	}
	html.Render(newVariableSet(props), w, r)
}

func (h *web) createVariableSet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name           *string `schema:"name,required"`
		Description    string
		Global         bool
		Organization   organization.Name `schema:"organization_name,required"`
		WorkspacesJSON string            `schema:"workspaces"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var workspaces []workspaceInfo
	if err := json.Unmarshal([]byte(params.WorkspacesJSON), &workspaces); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	workspaceIDs := make([]resource.TfeID, len(workspaces))
	for i, ws := range workspaces {
		workspaceIDs[i] = ws.ID
	}

	set, err := h.variables.CreateVariableSet(r.Context(), params.Organization, CreateVariableSetOptions{
		Name:        *params.Name,
		Description: params.Description,
		Global:      params.Global,
		Workspaces:  workspaceIDs,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.NewVariableSet(params.Organization), http.StatusFound)
		return
	}

	html.FlashSuccess(w, "added variable set: "+set.Name)
	http.Redirect(w, r, paths.EditVariableSet(set.ID), http.StatusFound)
}

func (h *web) editVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("variable_set_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	set, err := h.variables.getVariableSet(r.Context(), setID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// retrieve names of workspaces in org to show in dropdown widget
	availableWorkspaces, err := h.getAvailableWorkspaces(r.Context(), set.Organization)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// create list of existing workspaces and remove them from available
	// workspaces
	existingWorkspaces := make([]workspaceInfo, len(set.Workspaces))
	for i, existing := range set.Workspaces {
		for j, avail := range availableWorkspaces {
			if avail.ID == existing {
				// add to existing
				existingWorkspaces[i] = workspaceInfo{Name: avail.Name, ID: avail.ID}
				// remove from available
				availableWorkspaces = append(availableWorkspaces[:j], availableWorkspaces[j+1:]...)
				break
			}
		}
	}
	rows := make([]variableRow, len(set.Variables))
	for i, sv := range set.Variables {
		rows[i] = variableRow{
			Variable: sv,
			set:      set,
		}
	}

	props := editVariableSetProps{
		set:                 set,
		rows:                rows,
		availableWorkspaces: availableWorkspaces,
		existingWorkspaces:  existingWorkspaces,
		canDeleteVariable:   h.authorizer.CanAccess(r.Context(), authz.DeleteWorkspaceVariableAction, set.Organization),
	}
	html.Render(editVariableSet(props), w, r)
}

func (h *web) updateVariableSet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		SetID          resource.TfeID `schema:"variable_set_id,required"`
		Name           *string
		Description    *string
		Global         *bool
		WorkspacesJSON string `schema:"workspaces"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	var workspaces []workspaceInfo
	if err := json.Unmarshal([]byte(params.WorkspacesJSON), &workspaces); err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	workspaceIDs := make([]resource.TfeID, len(workspaces))
	for i, ws := range workspaces {
		workspaceIDs[i] = ws.ID
	}

	set, err := h.variables.updateVariableSet(r.Context(), params.SetID, UpdateVariableSetOptions{
		Name:        params.Name,
		Description: params.Description,
		Global:      params.Global,
		Workspaces:  workspaceIDs,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.EditVariableSet(params.SetID), http.StatusFound)
		return
	}

	html.FlashSuccess(w, "updated variable set: "+set.Name)
	http.Redirect(w, r, paths.EditVariableSet(set.ID), http.StatusFound)
}

func (h *web) deleteVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("variable_set_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	set, err := h.variables.deleteVariableSet(r.Context(), setID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted variable set: "+set.Name)
	http.Redirect(w, r, paths.VariableSets(set.Organization), http.StatusFound)
}

func (h *web) newVariableSetVariable(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.ID("variable_set_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	set, err := h.variables.getVariableSet(r.Context(), setID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.Render(newVSV(set), w, r)
}

func (h *web) createVariableSetVariable(w http.ResponseWriter, r *http.Request) {
	var params struct {
		createVariableParams
		SetID resource.TfeID `schema:"variable_set_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.variables.CreateVariableSetVariable(r.Context(), params.SetID, CreateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		Sensitive:   &params.Sensitive,
		HCL:         &params.HCL,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.NewVariableSetVariable(params.SetID), http.StatusFound)
		return
	}

	html.FlashSuccess(w, "added variable: "+variable.Key)
	http.Redirect(w, r, paths.EditVariableSet(params.SetID), http.StatusFound)
}

func (h *web) editVariableSetVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	set, err := h.variables.getVariableSetByVariableID(r.Context(), variableID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	v := set.getVariable(variableID)

	html.Render(editVSV(editVSVProps{set: set, variable: v}), w, r)
}

func (h *web) updateVariableSetVariable(w http.ResponseWriter, r *http.Request) {
	var params updateVariableParams
	if err := decode.All(&params, r); err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	set, err := h.variables.updateVariableSetVariable(r.Context(), params.VariableID, UpdateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		Sensitive:   params.Sensitive,
		HCL:         &params.HCL,
	})
	if err != nil {
		html.FlashError(w, err.Error())
		http.Redirect(w, r, paths.EditVariableSetVariable(params.VariableID), http.StatusFound)
		return
	}
	v := set.getVariable(params.VariableID)

	html.FlashSuccess(w, "updated variable: "+v.Key)
	http.Redirect(w, r, paths.EditVariableSet(set.ID), http.StatusFound)
}

func (h *web) deleteVariableSetVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.ID("variable_id", r)
	if err != nil {
		html.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	set, err := h.variables.deleteVariableSetVariable(r.Context(), variableID)
	if err != nil {
		html.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	v := set.getVariable(variableID)

	html.FlashSuccess(w, "deleted variable: "+v.Key)
	http.Redirect(w, r, paths.EditVariableSet(set.ID), http.StatusFound)
}

func (h *web) getAvailableWorkspaces(ctx context.Context, org organization.Name) ([]workspaceInfo, error) {
	// retrieve names of all workspaces in org to show in dropdown widget
	workspaces, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Workspace], error) {
		return h.workspaces.List(ctx, workspace.ListOptions{
			Organization: &org,
			PageOptions:  opts,
		})
	})
	if err != nil {
		return nil, err
	}

	availableWorkspaces := make([]workspaceInfo, len(workspaces))
	for i, ws := range workspaces {
		availableWorkspaces[i] = workspaceInfo{
			ID:   ws.ID,
			Name: ws.Name,
		}
	}
	return availableWorkspaces, nil
}
