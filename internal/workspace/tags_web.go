package workspace

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/resource"
)

func (h *webHandlers) addTagHandlers(r *mux.Router) {
	r = html.UIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/create-tag", h.createTag).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/delete-tag", h.deleteTag).Methods("POST")
}

func (h *webHandlers) createTag(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID *resource.TfeID `schema:"workspace_id,required"`
		TagName     *string         `schema:"tag_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err := h.client.AddTags(r.Context(), *params.WorkspaceID, []TagSpec{{Name: *params.TagName}})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "created tag: "+*params.TagName)
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
}

func (h *webHandlers) deleteTag(w http.ResponseWriter, r *http.Request) {
	var params struct {
		WorkspaceID *resource.TfeID `schema:"workspace_id,required"`
		TagName     *string         `schema:"tag_name,required"`
	}
	if err := decode.All(&params, r); err != nil {
		html.Error(r, w, err.Error(), html.WithStatus(http.StatusUnprocessableEntity))
		return
	}

	err := h.client.RemoveTags(r.Context(), *params.WorkspaceID, []TagSpec{{Name: *params.TagName}})
	if err != nil {
		html.Error(r, w, err.Error())
		return
	}

	html.FlashSuccess(w, "removed tag: "+*params.TagName)
	http.Redirect(w, r, paths.Workspace(params.WorkspaceID), http.StatusFound)
}
