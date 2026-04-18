package ui

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run/trigger"
	"github.com/leg100/otf/internal/ui/helpers"
	"github.com/leg100/otf/internal/ui/paths"
	"github.com/leg100/otf/internal/workspace"
)

type Handlers struct {
	authorizer authz.Interface
	logger     logr.Logger
	client     Client
}

type Client interface {
	CreateRunTrigger(ctx context.Context, workspaceID, sourceableWorkspaceID resource.TfeID) (*trigger.Trigger, error)
	ListRunTriggers(ctx context.Context, opts trigger.ListOptions) ([]*trigger.Trigger, error)
	GetRunTrigger(ctx context.Context, triggerID resource.TfeID) (*trigger.Trigger, error)
	DeleteRunTrigger(ctx context.Context, triggerID resource.TfeID) error
	ListWorkspaces(ctx context.Context, opts workspace.ListOptions) (*resource.Page[*workspace.Workspace], error)
	GetWorkspace(context.Context, resource.TfeID) (*workspace.Workspace, error)
}

func NewHandlers(
	logger logr.Logger,
	client Client,
	authorizer authz.Interface,
) *Handlers {
	return &Handlers{
		logger:     logger,
		client:     client,
		authorizer: authorizer,
	}
}

func (h *Handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/workspaces/{workspace_id}/edit-triggers", h.editTriggers).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/create-trigger", h.createTrigger).Methods("POST")
	r.HandleFunc("/triggers/{trigger_id}/delete", h.deleteTrigger).Methods("POST")
}

type connection struct {
	ws      *workspace.Workspace
	trigger *trigger.Trigger
}

func (h *Handlers) editTriggers(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	ws, err := h.client.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	workspaces, err := resource.ListAll(func(opts resource.PageOptions) (*resource.Page[*workspace.Workspace], error) {
		return h.client.ListWorkspaces(r.Context(), workspace.ListOptions{
			PageOptions:  opts,
			Organization: &ws.Organization,
		})
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	triggers, err := h.client.ListRunTriggers(r.Context(), trigger.ListOptions{
		WorkspaceID: workspaceID,
		Direction:   trigger.Inbound,
	})
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	// Build list of connected and unconnected workspaces
	var (
		connected   []connection
		unconnected []*workspace.Workspace
	)
	for _, ws := range workspaces {
		// Skip current workspace
		if ws.ID == workspaceID {
			continue
		}
		var trigger *trigger.Trigger
		for _, t := range triggers {
			if ws.ID == t.SourceableWorkspaceID {
				trigger = t
				break
			}
		}
		if trigger != nil {
			connected = append(connected, connection{
				ws:      ws,
				trigger: trigger,
			})
		} else {
			unconnected = append(unconnected, ws)
		}
	}

	props := editTriggersProps{
		ws:          ws,
		connected:   connected,
		unconnected: unconnected,
	}
	helpers.RenderPage(
		editTriggers(props),
		"edit run triggers | "+ws.ID.String(),
		w,
		r,
		helpers.WithWorkspace(ws, h.authorizer),
		helpers.WithSideMenu(helpers.WorkspaceSettingsMenu(ws.ID)),
		helpers.WithBreadcrumbs(
			helpers.Breadcrumb{Name: "Run Triggers"},
		),
	)
}

func (h *Handlers) createTrigger(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID           resource.TfeID `schema:"workspace_id,required"`
		TriggeringWorkspaceID resource.TfeID `schema:"triggering_workspace_id,required"`
	}
	if err := decode.All(&params, r); err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	trigger, err := h.client.CreateRunTrigger(r.Context(), params.WorkspaceID, params.TriggeringWorkspaceID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "created trigger: "+trigger.ID.String())
	http.Redirect(w, r, paths.EditTriggersWorkspace(params.WorkspaceID), http.StatusFound)
}

func (h *Handlers) deleteTrigger(w http.ResponseWriter, r *http.Request) {
	triggerID, err := decode.ID("trigger_id", r)
	if err != nil {
		helpers.Error(r, w, err.Error(), helpers.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err = h.client.DeleteRunTrigger(r.Context(), triggerID)
	if err != nil {
		helpers.Error(r, w, err.Error())
		return
	}

	helpers.FlashSuccess(w, "deleted trigger: "+triggerID.String())
	http.Redirect(w, r, r.Referer(), http.StatusFound)
}
