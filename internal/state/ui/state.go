package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/state"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
)

func addStateHandlers(r *mux.Router, h *Handlers) {
	r.HandleFunc("/workspaces/{workspace_id}/state", h.getState).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/state-versions", h.listStateVersions).Methods("GET")
	r.HandleFunc("/state-versions/{state_version_id}", h.getStateVersion).Methods("GET")
	r.HandleFunc("/state-versions/{state_version_id}/rollback", h.rollbackStateVersion).Methods("POST")
	r.HandleFunc("/state-versions/{state_version_id}/delete", h.deleteStateVersion).Methods("POST")
	r.HandleFunc("/state-versions/{state_version_id}/diff", h.diffStateVersion).Methods("GET")
}

func (h *Handlers) getState(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	// ignore errors and instead render unpopulated template
	f := &state.File{}
	sv, err := h.State.GetCurrentStateVersion(r.Context(), id)
	if err == nil {
		f, _ = sv.File()
	}

	helpers.Render(getState(f), w, r)
}

func (h *Handlers) listStateVersions(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		resource.PageOptions
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.Workspaces.GetWorkspace(r.Context(), params.WorkspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	page, err := h.State.ListStateVersions(r.Context(), params.WorkspaceID, params.PageOptions)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	var currentID *resource.TfeID
	current, err := h.State.GetCurrentStateVersion(r.Context(), params.WorkspaceID)
	if err == nil {
		currentID = &current.ID
	} else if !errors.Is(err, internal.ErrResourceNotFound) {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.RenderPage(
		h.templates.listStateVersions(listStateVersionsProps{
			page:      page,
			currentID: currentID,
		}),
		"State Versions",
		w, r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "State versions"},
		),
	)
}

func (h *Handlers) rollbackStateVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("state_version_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	sv, err := h.State.RollbackStateVersion(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.StateVersions(sv.WorkspaceID), http.StatusFound)
}

func (h *Handlers) deleteStateVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("state_version_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	sv, err := h.State.GetStateVersion(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}
	workspaceID := sv.WorkspaceID

	if err := h.State.DeleteStateVersion(r.Context(), id); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	http.Redirect(w, r, paths.StateVersions(workspaceID), http.StatusFound)
}

func (h *Handlers) getStateVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("state_version_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	sv, err := h.State.GetStateVersion(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	ws, err := h.Workspaces.GetWorkspace(r.Context(), sv.WorkspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, sv.State, "", "  "); err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.RenderPage(
		h.templates.getStateVersion(pretty.String()),
		fmt.Sprintf("State version (serial %d)", sv.Serial),
		w, r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "State versions", Link: paths.StateVersions(sv.WorkspaceID)},
			helpers.Breadcrumb{Name: sv.ID.String()},
		),
	)
}

func (h *Handlers) diffStateVersion(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("state_version_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	sv, err := h.State.GetStateVersion(r.Context(), id)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	ws, err := h.Workspaces.GetWorkspace(r.Context(), sv.WorkspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	toFile, err := sv.File()
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	var fromFile *state.File
	var prev *state.Version
	prev, err = h.State.GetPreviousStateVersion(r.Context(), sv)
	if err == nil {
		fromFile, _ = prev.File()
	} else if !errors.Is(err, internal.ErrResourceNotFound) {
		helpers.Error(r, w, err.Error())
		return
	}

	diff := state.Diff(fromFile, toFile)

	var title string
	if prev != nil {
		title = fmt.Sprintf("State diff: serial %d → %d", prev.Serial, sv.Serial)
	} else {
		title = fmt.Sprintf("State diff: initial (serial %d)", sv.Serial)
	}

	helpers.RenderPage(
		h.templates.diffStateVersion(diffStateVersionProps{
			sv:   sv,
			prev: prev,
			diff: diff,
		}),
		title,
		w, r,
		helpers.WithWorkspace(ws, h.Authorizer),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "State versions", Link: paths.StateVersions(sv.WorkspaceID)},
			helpers.Breadcrumb{Name: fmt.Sprintf("Diff (%s)", sv.ID.String())},
		),
	)
}
