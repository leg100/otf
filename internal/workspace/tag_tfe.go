package workspace

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
)

const (
	addTags tagOperation = iota
	removeTags
)

type tagOperation int

func (a *tfe) addTagHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.addTags).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.removeTags).Methods("DELETE")
	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.getTags).Methods("GET")

	r.HandleFunc("/organizations/{organization_name}/tags", a.listTags).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/tags", a.deleteTags).Methods("DELETE")
	r.HandleFunc("/tags/{tag_id}/relationships/workspaces", a.tagWorkspaces).Methods("POST")
}

func (a *tfe) listTags(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization resource.OrganizationName `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params ListTagsOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.ListTags(r.Context(), pathParams.Organization, params)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert to
	to := make([]*TFEOrganizationTag, len(page.Items))
	for i, from := range page.Items {
		to[i] = a.toTag(from)
	}
	a.RespondWithPage(w, r, to, page.Pagination)
}

func (a *tfe) deleteTags(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization resource.OrganizationName `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []struct {
		ID resource.TfeID `jsonapi:"primary,tags"`
	}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	tagIDs := make([]resource.TfeID, len(params))
	for i, p := range params {
		tagIDs[i] = p.ID
	}

	if err := a.DeleteTags(r.Context(), pathParams.Organization, tagIDs); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) tagWorkspaces(w http.ResponseWriter, r *http.Request) {
	tagID, err := decode.ID("tag_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []*TFEWorkspace
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	workspaceIDs := make([]resource.TfeID, len(params))
	for i, p := range params {
		workspaceIDs[i] = p.ID
	}

	if err := a.TagWorkspaces(r.Context(), tagID, workspaceIDs); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) addTags(w http.ResponseWriter, r *http.Request) {
	a.alterWorkspaceTags(w, r, addTags)
}

func (a *tfe) removeTags(w http.ResponseWriter, r *http.Request) {
	a.alterWorkspaceTags(w, r, removeTags)
}

func (a *tfe) alterWorkspaceTags(w http.ResponseWriter, r *http.Request, op tagOperation) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []*TFETag
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert from json:api structs to tag specs
	specs := make([]TagSpec, len(params))
	for i, tag := range params {
		specs[i] = TagSpec{ID: tag.ID, Name: tag.Name}
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
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) getTags(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.ID("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params ListWorkspaceTagsOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.ListWorkspaceTags(r.Context(), workspaceID, params)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert to
	to := make([]*TFEOrganizationTag, len(page.Items))
	for i, from := range page.Items {
		to[i] = a.toTag(from)
	}
	a.RespondWithPage(w, r, to, page.Pagination)
}

func (a *tfe) toTag(from *Tag) *TFEOrganizationTag {
	return &TFEOrganizationTag{
		ID:            from.ID,
		Name:          from.Name,
		InstanceCount: from.InstanceCount,
		Organization: &organization.TFEOrganization{
			Name: from.Organization,
		},
	}
}
