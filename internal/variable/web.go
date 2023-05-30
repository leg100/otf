package variable

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/workspace"
)

type web struct {
	html.Renderer
	workspace.Service

	svc Service
}

func (h *web) addHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/variables", h.list)
	r.HandleFunc("/workspaces/{workspace_id}/variables/new", h.new)
	r.HandleFunc("/workspaces/{workspace_id}/variables/create", h.create)
	r.HandleFunc("/variables/{variable_id}/edit", h.edit)
	r.HandleFunc("/variables/{variable_id}/update", h.update)
	r.HandleFunc("/variables/{variable_id}/delete", h.delete)
}

func (h *web) new(w http.ResponseWriter, r *http.Request) {
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
		CanAccess:     subject.CanAccessWorkspace(rbac.CreateVariableAction, policy),
	})
}

func (h *web) create(w http.ResponseWriter, r *http.Request) {
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

	variable, err := h.svc.CreateVariable(r.Context(), params.WorkspaceID, CreateVariableOptions{
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

func (h *web) list(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variables, err := h.svc.ListVariables(r.Context(), workspaceID)
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

	h.Render("variable_list.tmpl", w, struct {
		workspace.WorkspacePage
		Variables            []*Variable
		Policy               internal.WorkspacePolicy
		CreateVariableAction rbac.Action
		DeleteVariableAction rbac.Action
	}{
		WorkspacePage:        workspace.NewPage(r, "variables", ws),
		Variables:            variables,
		Policy:               policy,
		CreateVariableAction: rbac.CreateVariableAction,
		DeleteVariableAction: rbac.DeleteVariableAction,
	})
}

func (h *web) edit(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.svc.GetVariable(r.Context(), variableID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ws, err := h.GetWorkspace(r.Context(), variable.WorkspaceID)
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

	h.Render("variable_edit.tmpl", w, struct {
		workspace.WorkspacePage
		Variable   *Variable
		EditMode   bool
		FormAction string
		CanAccess  bool
	}{
		WorkspacePage: workspace.NewPage(r, "edit | "+variable.ID, ws),
		Variable:      variable,
		EditMode:      true,
		FormAction:    paths.UpdateVariable(variable.ID),
		CanAccess:     subject.CanAccessWorkspace(rbac.CreateVariableAction, policy),
	})
}

func (h *web) update(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.svc.GetVariable(r.Context(), variableID)
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

	variable, err = h.svc.UpdateVariable(r.Context(), variableID, UpdateVariableOptions{
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

func (h *web) updateSensitive(w http.ResponseWriter, r *http.Request, variable *Variable) {
	value, err := decode.Param("value", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err = h.svc.UpdateVariable(r.Context(), variable.ID, UpdateVariableOptions{
		Value: internal.String(value),
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "updated variable: "+variable.Key)
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID), http.StatusFound)
}

func (h *web) delete(w http.ResponseWriter, r *http.Request) {
	variableID, err := decode.Param("variable_id", r)
	if err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	variable, err := h.svc.DeleteVariable(r.Context(), variableID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "deleted variable: "+variable.Key)
	http.Redirect(w, r, paths.Variables(variable.WorkspaceID), http.StatusFound)
}
