package workspace

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	api struct {
		*Service
		*tfeapi.Responder
	}
)

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfhttp.APIBasePath).Subrouter()

	r.HandleFunc("/organizations/{organization_name}/workspaces", a.listWorkspaces).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/workspaces/{workspace_name}", a.getWorkspaceByName).Methods("GET")

	r.HandleFunc("/workspaces/{workspace_id}", a.updateWorkspace).Methods("PATCH")
	r.HandleFunc("/workspaces/{workspace_id}", a.getWorkspace).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/actions/lock", a.lockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/actions/unlock", a.unlockWorkspace).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/actions/force-unlock", a.forceUnlockWorkspace).Methods("POST")
}

func (a *api) getWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.Get(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, ws, http.StatusOK)
}

func (a *api) getWorkspaceByName(w http.ResponseWriter, r *http.Request) {
	var params byWorkspaceName
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.GetByName(r.Context(), params.Organization, params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, ws, http.StatusOK)
}

func (a *api) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	var params ListOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.List(r.Context(), params)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.RespondWithPage(w, r, page.Items, page.Pagination)
}

func (a *api) updateWorkspace(w http.ResponseWriter, r *http.Request) {
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

	ws, err := a.Update(r.Context(), id, params)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, ws, http.StatusOK)
}

func (a *api) lockWorkspace(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.Lock(r.Context(), id, nil)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, ws, http.StatusOK)
}

func (a *api) unlockWorkspace(w http.ResponseWriter, r *http.Request) {
	a.unlock(w, r, false)
}

func (a *api) forceUnlockWorkspace(w http.ResponseWriter, r *http.Request) {
	a.unlock(w, r, true)
}

func (a *api) unlock(w http.ResponseWriter, r *http.Request, force bool) {
	id, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ws, err := a.Unlock(r.Context(), id, nil, force)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, ws, http.StatusOK)
}
