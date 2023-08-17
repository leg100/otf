package workspace

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
)

const (
	addTags tagOperation = iota
	removeTags
)

type tagOperation int

func (a *tfe) addTagHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.addTags).Methods("POST")
	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.removeTags).Methods("DELETE")
	r.HandleFunc("/workspaces/{workspace_id}/relationships/tags", a.getTags).Methods("GET")

	r.HandleFunc("/organizations/{organization_name}/tags", a.listTags).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/tags", a.deleteTags).Methods("DELETE")
	r.HandleFunc("/tags/{tag_id}/relationships/workspaces", a.tagWorkspaces).Methods("POST")
}

func (a *tfe) listTags(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params ListTagsOptions
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.ListTags(r.Context(), org, params)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*types.OrganizationTag, len(page.Items))
	for i, from := range page.Items {
		items[i] = a.toTag(from)
	}
	a.RespondWithPage(w, r, page.Items, page.Pagination)
}

func (a *tfe) deleteTags(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []struct {
		ID string `jsonapi:"primary,tags"`
	}
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var tagIDs []string
	for _, p := range params {
		tagIDs = append(tagIDs, p.ID)
	}

	if err := a.DeleteTags(r.Context(), org, tagIDs); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) tagWorkspaces(w http.ResponseWriter, r *http.Request) {
	tagID, err := decode.Param("tag_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []*types.Workspace
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var workspaceIDs []string
	for _, p := range params {
		workspaceIDs = append(workspaceIDs, p.ID)
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
	workspaceID, err := decode.Param("workspace_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var params []*types.Tag
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
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
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) getTags(w http.ResponseWriter, r *http.Request) {
	workspaceID, err := decode.Param("workspace_id", r)
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

	// convert items
	items := make([]*types.OrganizationTag, len(page.Items))
	for i, from := range page.Items {
		items[i] = a.toTag(from)
	}
	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) toTag(from *Tag) *types.OrganizationTag {
	return &types.OrganizationTag{
		ID:            from.ID,
		Name:          from.Name,
		InstanceCount: from.InstanceCount,
		Organization: &types.Organization{
			Name: from.Organization,
		},
	}
}

func toTagSpecs(from []*types.Tag) (to []TagSpec) {
	for _, tag := range from {
		to = append(to, TagSpec{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}
	return
}
