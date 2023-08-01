package workspace

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
)

func (h *webHandlers) addTagHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/create-remote-state-consumer", h.createRemoteStateConsumer).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/delete-remote-state-consumer", h.deleteRemoteStateConsumer).Methods("POST")
}

func (h *webHandlers) createRemoteStateConsumer(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID  *string `schema:"workspace_id,required"`
		ConsumerName *string `schema:"consumer_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// we only have the *name* of the workspace consuming state but we need the
	// *ID*. The only way to get this is to first look up the organization of
	// the workspace owning the state.
	ws, err := h.svc.GetWorkspace(r.Context(), *params.WorkspaceID)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	consumer, err := h.svc.GetWorkspaceByName(r.Context(), ws.Organization, *params.ConsumerName)
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.svc.AddRemoteStateConsumers(r.Context(), *params.WorkspaceID, []string{
		consumer.ID,
	})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "created tag: "+*params.ConsumerName)
	http.Redirect(w, r, paths.Workspace(*params.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) deleteRemoteStateConsumer(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID *string `schema:"workspace_id,required"`
		TagName     *string `schema:"tag_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		h.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err := h.svc.RemoveTags(r.Context(), *params.WorkspaceID, []TagSpec{{Name: *params.TagName}})
	if err != nil {
		h.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	html.FlashSuccess(w, "removed tag: "+*params.TagName)
	http.Redirect(w, r, paths.Workspace(*params.WorkspaceID), http.StatusFound)
}
