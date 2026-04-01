package state

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type API struct {
	*tfeapi.Responder
	Client apiClient
	TFEAPI *tfe
}

type apiClient interface {
	CreateStateVersion(ctx context.Context, opts CreateStateVersionOptions) (*Version, error)
	GetCurrentStateVersion(ctx context.Context, workspaceID resource.ID) (*Version, error)
	ListStateVersions(ctx context.Context, workspaceID resource.ID, opts resource.PageOptions) (*resource.Page[*Version], error)
	GetStateVersion(ctx context.Context, id resource.ID) (*Version, error)
	RollbackStateVersion(ctx context.Context, id resource.ID) (*Version, error)
	DeleteStateVersion(ctx context.Context, id resource.ID) error
	DownloadState(ctx context.Context, svID resource.ID) ([]byte, error)
}

func (a *API) AddHandlers(r *mux.Router) {
	// proxy this endpoint to the tfeapi endpoint because the behaviour is
	// identical (although it returns a tfe struct the only user of this
	// endpoint, the agent, ignores the return value).
	r.HandleFunc("/workspaces/{workspace_id}/state-versions", a.TFEAPI.createVersion).Methods("POST")

	r.HandleFunc("/workspaces/{workspace_id}/current-state-version", a.getCurrentVersion).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/state-versions", a.listVersions).Methods("GET")

	r.HandleFunc("/state-versions/{id}/download", a.downloadState).Methods("GET")
	r.HandleFunc("/state-versions/{id}/rollback", a.rollbackVersion).Methods("PATCH")
	r.HandleFunc("/state-versions/{id}", a.deleteVersion).Methods("DELETE")
}

func (a *API) listVersions(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		resource.PageOptions
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	page, err := a.Client.ListStateVersions(r.Context(), params.WorkspaceID, params.PageOptions)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.RespondWithPage(w, r, page.Items, page.Pagination)
}

func (a *API) getCurrentVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	sv, err := a.Client.GetCurrentStateVersion(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, sv, http.StatusOK)
}

func (a *API) deleteVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.Client.DeleteStateVersion(r.Context(), versionID); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) rollbackVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	sv, err := a.Client.RollbackStateVersion(r.Context(), versionID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, sv, http.StatusOK)
}

func (a *API) downloadState(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	resp, err := a.Client.DownloadState(r.Context(), versionID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(resp)
}
