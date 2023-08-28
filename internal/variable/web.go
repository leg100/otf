package variable

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
)

type (
	web struct {
		html.Renderer
		workspace.Service

		svc Service
	}
	workspaceInfo struct {
		ID   string
		Name string
	}
)

func (h *web) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/variables", h.listWorkspaceVariables)
	r.HandleFunc("/workspaces/{workspace_id}/variables/new", h.newWorkspaceVariable)
	r.HandleFunc("/workspaces/{workspace_id}/variables/create", h.createWorkspaceVariable)
	r.HandleFunc("/variables/{variable_id}/edit", h.editWorkspaceVariable)
	r.HandleFunc("/variables/{variable_id}/update", h.updateWorkspaceVariable)
	r.HandleFunc("/variables/{variable_id}/delete", h.deleteWorkspaceVariable)

	r.HandleFunc("/organizations/{organization_name}/variable-sets", h.listVariableSets)
	r.HandleFunc("/organizations/{organization_name}/variable-sets/new", h.newVariableSet)
	r.HandleFunc("/organizations/{organization_name}/variable-sets/create", h.createVariableSet)
	r.HandleFunc("/variable-sets/{variable_set_id}/edit", h.editVariableSet)
	r.HandleFunc("/variable-sets/{variable_set_id}/update", h.updateVariableSet)
	r.HandleFunc("/variable-sets/{variable_set_id}/delete", h.deleteVariableSet)
}

func (h *web) newWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ws, err := h.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	policy, err := h.GetPolicy(r.Context(), ws.ID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	subject, err := internal.SubjectFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("variable_new.tmpl", w, struct {
		workspace.WorkspacePage
		Variable   *Variable
		EditMode   bool
		FormAction string
		CanAccess  bool
	}{
		WorkspacePage: workspace.NewPage(r, "new variable", ws),
		Variable:      &Variable{},
		EditMode:      false,
		FormAction:    paths.CreateVariable(workspaceID),
		CanAccess:     subject.CanAccessWorkspace(rbac.CreateWorkspaceVariableAction, policy),
	})
}

func (h *web) createWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Key         *string `schema:"key,required"`
		Value       *string
		Description *string
		Category    *VariableCategory `schema:"category,required"`
		Sensitive   bool
		HCL         bool
		WorkspaceID string `schema:"workspace_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.svc.CreateWorkspaceVariable(r.Context(), params.WorkspaceID, CreateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		Sensitive:   &params.Sensitive,
		HCL:         &params.HCL,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "added variable: "+variable.Key)
	http.Redirect(w, r, paths.Variables(params.WorkspaceID), http.StatusFound)
}

func (h *web) listWorkspaceVariables(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variables, err := h.svc.ListWorkspaceVariables(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := h.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	policy, err := h.GetPolicy(r.Context(), ws.ID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user, err := auth.UserFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("variable_list.tmpl", w, struct {
		workspace.WorkspacePage
		Variables          []*WorkspaceVariable
		Policy             internal.WorkspacePolicy
		CanCreateVariable  bool
		CanDeleteVariable  bool
		CanUpdateWorkspace bool
	}{
		WorkspacePage:      workspace.NewPage(r, "variables", ws),
		Variables:          variables,
		Policy:             policy,
		CanCreateVariable:  user.CanAccessWorkspace(rbac.CreateWorkspaceVariableAction, policy),
		CanDeleteVariable:  user.CanAccessWorkspace(rbac.DeleteWorkspaceVariableAction, policy),
		CanUpdateWorkspace: user.CanAccessWorkspace(rbac.UpdateWorkspaceAction, policy),
	})
}

func (h *web) editWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.svc.GetWorkspaceVariable(r.Context(), variableID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := h.GetWorkspace(r.Context(), variable.WorkspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("variable_edit.tmpl", w, struct {
		workspace.WorkspacePage
		Variable   *WorkspaceVariable
		EditMode   bool
		FormAction string
	}{
		WorkspacePage: workspace.NewPage(r, "edit | "+variable.ID, ws),
		Variable:      variable,
		EditMode:      true,
		FormAction:    paths.UpdateVariable(variable.ID),
	})
}

func (h *web) updateWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.svc.GetWorkspaceVariable(r.Context(), variableID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Handle updates to sensitive variables in a separate handler.
	if variable.Sensitive {
		h.updateSensitive(w, r, variable)
		return
	}

	var params struct {
		Key         *string
		Value       *string
		Description *string
		Category    *VariableCategory
		Sensitive   *bool // form checkbox can only be true/false, not nil
		HCL         *bool // form checkbox can only be true/false, not nil
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err = h.svc.UpdateWorkspaceVariable(r.Context(), variableID, UpdateVariableOptions{
		Key:         params.Key,
		Value:       params.Value,
		Description: params.Description,
		Category:    params.Category,
		Sensitive:   params.Sensitive,
		HCL:         params.HCL,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated variable: "+variable.Key)
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID), http.StatusFound)
}

func (h *web) updateSensitive(w http.ResponseWriter, r *http.Request, variable *WorkspaceVariable) {
	value, err := decode.Param("value", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err = h.svc.UpdateWorkspaceVariable(r.Context(), variable.ID, UpdateVariableOptions{
		Value: internal.String(value),
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated variable: "+variable.Key)
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID), http.StatusFound)
}

func (h *web) deleteWorkspaceVariable(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.svc.DeleteWorkspaceVariable(r.Context(), variableID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted variable: "+variable.Key)
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID), http.StatusFound)
}

func (h *web) listVariableSets(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	sets, err := h.svc.listVariableSets(r.Context(), org)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := auth.UserFromContext(r.Context())
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("variable_set_list.tmpl", w, struct {
		organization.OrganizationPage
		VariableSets []*VariableSet
		CanCreate    bool
	}{
		OrganizationPage: organization.NewPage(r, "variable sets", org),
		VariableSets:     sets,
		CanCreate:        user.CanAccessOrganization(rbac.CreateVariableSetAction, org),
	})
}

func (h *web) newVariableSet(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// retrieve names of all workspaces in org to show in dropdown widget
	workspaces, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Workspace], error) {
		return h.Service.ListWorkspaces(r.Context(), workspace.ListOptions{
			Organization: &org,
			PageOptions:  opts,
		})
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	availableWorkspaces := make([]workspaceInfo, len(workspaces))
	for i, ws := range workspaces {
		availableWorkspaces[i] = workspaceInfo{
			ID:   ws.ID,
			Name: ws.Name,
		}
	}

	h.Render("variable_set_new.tmpl", w, struct {
		organization.OrganizationPage
		VariableSet         *VariableSet
		EditMode            bool
		FormAction          string
		AvailableWorkspaces []workspaceInfo
	}{
		OrganizationPage:    organization.NewPage(r, "variable sets", org),
		VariableSet:         &VariableSet{},
		EditMode:            false,
		FormAction:          paths.CreateVariableSet(org),
		AvailableWorkspaces: availableWorkspaces,
	})
}

func (h *web) createVariableSet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name         *string `schema:"name,required"`
		Description  string
		Global       bool
		Organization string `schema:"organization_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	set, err := h.svc.createVariableSet(r.Context(), params.Organization, CreateVariableSetOptions{
		Name:        *params.Name,
		Description: params.Description,
		Global:      params.Global,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "added variable set: "+set.Name)
	http.Redirect(w, r, paths.VariableSet(set.ID), http.StatusFound)
}

func (h *web) editVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("variable_set_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	set, err := h.svc.getVariableSet(r.Context(), setID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.Render("variable_set_edit.tmpl", w, struct {
		organization.OrganizationPage
		VariableSet *VariableSet
		EditMode    bool
		FormAction  string
	}{
		OrganizationPage: organization.NewPage(r, "edit | "+set.ID, set.Organization),
		VariableSet:      set,
		EditMode:         true,
		FormAction:       paths.UpdateVariableSet(set.ID),
	})
}

func (h *web) updateVariableSet(w http.ResponseWriter, r *http.Request) {
	var params struct {
		SetID       string `schema:"variable_set_id,required"`
		Name        *string
		Description *string
		Global      *bool
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	set, err := h.svc.updateVariableSet(r.Context(), params.SetID, UpdateVariableSetOptions{
		Name:        params.Name,
		Description: params.Description,
		Global:      params.Global,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated variable set: "+set.Name)
	http.Redirect(w, r, paths.VariableSet(set.ID), http.StatusFound)
}

func (h *web) deleteVariableSet(w http.ResponseWriter, r *http.Request) {
	setID, err := decode.Param("variable_set_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	set, err := h.svc.deleteVariableSet(r.Context(), setID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted variable set: "+set.Name)
	http.Redirect(w, r, paths.VariableSets(set.Organization), http.StatusFound)
}
