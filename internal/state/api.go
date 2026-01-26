package state

import (
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

type api struct {
	*Service
	*tfeapi.Responder

	tfeapi *tfe
}

func (a *api) addHandlers(r *mux.Router) {
	r = r.PathPrefix(otfhttp.APIBasePath).Subrouter()

	// proxy this endpoint to the tfeapi endpoint because the behaviour is
	// identical (although it returns a tfe struct the only user of this
	// endpoint, the agent, ignores the return value).
	r.HandleFunc("/workspaces/{workspace_id}/state-versions", a.tfeapi.createVersion).Methods("POST")

	r.HandleFunc("/workspaces/{workspace_id}/current-state-version", a.getCurrentVersion).Methods("GET")
	r.HandleFunc("/workspaces/{workspace_id}/state-versions", a.listVersions).Methods("GET")

	r.HandleFunc("/state-versions/{id}/download", a.downloadState).Methods("GET")
	r.HandleFunc("/state-versions/{id}/rollback", a.rollbackVersion).Methods("PATCH")
	r.HandleFunc("/state-versions/{id}", a.deleteVersion).Methods("DELETE")
}

func (a *api) listVersions(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID resource.TfeID `schema:"workspace_id,required"`
		resource.PageOptions
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	page, err := a.List(r.Context(), params.WorkspaceID, params.PageOptions)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.RespondWithPage(w, r, page.Items, page.Pagination)
}

func (a *api) getCurrentVersion(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	sv, err := a.GetCurrent(r.Context(), workspaceID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, sv, http.StatusOK)
}

func (a *api) deleteVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if err := a.Delete(r.Context(), versionID); err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *api) rollbackVersion(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	sv, err := a.Rollback(r.Context(), versionID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	a.Respond(w, r, sv, http.StatusOK)
}

func (a *api) downloadState(w http.ResponseWriter, r *http.Request) {
	versionID, err := decode.ID("id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	resp, err := a.Download(r.Context(), versionID)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	w.Write(resp)
}
