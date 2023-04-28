package api

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/tags"
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
		jsonapi.Error(w, err)
		return
	}
	var params tags.ListTagsOptions
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}

	tags, err := a.ListTags(r.Context(), org, params)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	writeTags(w, r, tags)
}

func (a *api) deleteTags(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	var params jsonapi.OrganizationTagsDeleteOptions
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}

	if err := a.DeleteTags(r.Context(), org, params.IDs); err != nil {
		jsonapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *api) tagWorkspaces(w http.ResponseWriter, r *http.Request) {
	tagID, err := decode.Param("tag_id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	var params jsonapi.AddWorkspacesToTagOptions
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}

	if err := a.TagWorkspaces(r.Context(), tagID, params.WorkspaceIDs); err != nil {
		jsonapi.Error(w, err)
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
		jsonapi.Error(w, err)
		return
	}
	var params jsonapi.Tags
	if err := jsonapi.UnmarshalPayload(r.Body, &params); err != nil {
		jsonapi.Error(w, err)
		return
	}

	// convert from json:api structs to tag specs
	var specs []tags.TagSpec
	for _, tag := range params.Tags {
		specs = append(specs, tags.TagSpec{
			ID:   otf.String(tag.ID),
			Name: otf.String(tag.Name),
		})
	}

	switch op {
	case addTags:
		err = a.AddTags(r.Context(), workspaceID, specs)
	case removeTags:
		err = a.RemoveTags(r.Context(), workspaceID, specs)
	default:
		err = errors.New("unknown tag operation")
	}
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *api) getTags(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}
	var params tags.ListWorkspaceTagsOptions
	if err := decode.All(&params, r); err != nil {
		jsonapi.Error(w, err)
		return
	}

	tags, err := a.ListWorkspaceTags(r.Context(), workspaceID, params)
	if err != nil {
		jsonapi.Error(w, err)
		return
	}

	writeTags(w, r, tags)
}

func writeTags(w http.ResponseWriter, r *http.Request, tags *tags.TagList) {
	to := jsonapi.OrganizationTagsList{
		Pagination: jsonapi.NewPagination(tags.Pagination),
	}
	for _, from := range tags.Items {
		to.Items = append(to.Items, &jsonapi.OrganizationTag{
			ID:            from.ID,
			Name:          from.Name,
			InstanceCount: from.InstanceCount,
			Organization: &jsonapi.Organization{
				Name: from.Organization,
			},
		})
	}
	jsonapi.WriteResponse(w, r, to)
}
