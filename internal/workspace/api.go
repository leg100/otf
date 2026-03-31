package workspace

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type API struct {
	*tfeapi.Responder
	Client apiClient
}

type apiClient interface {
	GetWorkspace(context.Context, resource.ID) (*Workspace, error)
	WatchWorkspaces(ctx context.Context) (<-chan pubsub.Event[*Event], func())
	ListWorkspaces(ctx context.Context, opts ListOptions) (*resource.Page[*Workspace], error)
	CreateWorkspace(ctx context.Context, opts CreateOptions) (*Workspace, error)
	GetWorkspaceByName(ctx context.Context, organization resource.ID, name string) (*Workspace, error)
	GetWorkspacePolicy(ctx context.Context, workspaceID resource.TfeID) (Policy, error)
	UpdateWorkspace(ctx context.Context, workspaceID resource.TfeID, opts UpdateOptions) (*Workspace, error)
	DeleteWorkspace(ctx context.Context, workspaceID resource.TfeID) (*Workspace, error)
	Lock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID) (*Workspace, error)
	Unlock(ctx context.Context, workspaceID resource.TfeID, runID *resource.TfeID, force bool) (*Workspace, error)
}

func (a *API) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/workspaces", a.listWorkspaces).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", a.getWorkspaceByName).Methods("GET")

	r.HandleFunc("/workspaces/{workspace_id}", a.updateWorkspace).Methods("PATCH")
	r.HandleFunc("/workspaces/{workspace_id}", a.getWorkspace).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/actions/lock", a.lockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/actions/unlock", a.unlockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/actions/force-unlock", a.forceUnlockWorkspace).Methods("POST")
}

func (a *API) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.Client.GetWorkspace(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, ws, http.StatusOK)
}

func (a *API) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byWorkspaceName
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.Client.GetWorkspaceByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, ws, http.StatusOK)
}

func (a *API) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var params ListOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.Client.ListWorkspaces(r.Context(), params)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.RespondWithPage(w, r, page.Items, page.Pagination)
}

func (a *API) updateWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	var params UpdateOptions
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.Client.UpdateWorkspace(r.Context(), id, params)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, ws, http.StatusOK)
}

func (a *API) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.Client.Lock(r.Context(), id, nil)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, ws, http.StatusOK)
}

func (a *API) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	a.unlock(w, r, false)
}

func (a *API) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	a.unlock(w, r, true)
}

func (a *API) unlock(w http.ResponseWriter, r *http.Request, force bool) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.Client.Unlock(r.Context(), id, nil, force)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, ws, http.StatusOK)
}
