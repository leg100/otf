package api

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/api/types"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/workspace"
)

const (
	addTags tagOperation = iota
	removeTags
)

type tagOperation int

func (a *api) addTagHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.addTags).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.removeTags).Methods("DELETE")
	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.getTags).Methods("GET")

	r.HandleFunc("/organizations/{organization_name}/tags", a.listTags).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/tags", a.deleteTags).Methods("DELETE")
	r.HandleFunc("/tags/{tag_id}/relationships/workspaces", a.tagWorkspaces).Methods("POST")
}

func (a *api) listTags(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}
	var params workspace.ListTagsOptions
	if err := decode.All(&params, r); err != nil {
		Error(w, err)
		return
	}

	tags, err := a.ListTags(r.Context(), org, params)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, tags)
}

func (a *api) deleteTags(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}
	var params []struct {
		ID string `jsonapi:"primary,tags"`
	}
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}
	var tagIDs []string
	for _, p := range params {
		tagIDs = append(tagIDs, p.ID)
	}

	if err := a.DeleteTags(r.Context(), org, tagIDs); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *api) tagWorkspaces(w http.ResponseWriter, r *http.Request) {
	tagID, err := decode.Param("tag_id", r)
	if err != nil {
		Error(w, err)
		return
	}
	var params []*types.Workspace
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}
	var workspaceIDs []string
	for _, p := range params {
		workspaceIDs = append(workspaceIDs, p.ID)
	}

	if err := a.TagWorkspaces(r.Context(), tagID, workspaceIDs); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *api) addTags(w http.ResponseWriter, r *http.Request) {
	a.alterWorkspaceTags(w, r, addTags)
}

func (a *api) removeTags(w http.ResponseWriter, r *http.Request) {
	a.alterWorkspaceTags(w, r, removeTags)
}

func (a *api) alterWorkspaceTags(w http.ResponseWriter, r *http.Request, op tagOperation) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}
	var params []*types.Tag
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}
	// convert from json:api structs to tag specs
	specs := toTagSpecs(params)

	switch op {
	case addTags:
		err = a.AddTags(r.Context(), workspaceID, specs)
	case removeTags:
		err = a.RemoveTags(r.Context(), workspaceID, specs)
	default:
		err = errors.New("unknown tag operation")
	}
	if err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *api) getTags(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		Error(w, err)
		return
	}
	var params workspace.ListWorkspaceTagsOptions
	if err := decode.All(&params, r); err != nil {
		Error(w, err)
		return
	}

	tags, err := a.ListWorkspaceTags(r.Context(), workspaceID, params)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, tags)
}

func toTagSpecs(from []*types.Tag) (to []workspace.TagSpec) {
	for _, tag := range from {
		to = append(to, workspace.TagSpec{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}
	return
}
