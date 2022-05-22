package http

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
)

func (s *Server) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	opts := otf.OrganizationCreateOptions{}

	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := opts.Valid(); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.OrganizationService().Create(r.Context(), opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, OrganizationJSONAPIObject(obj), WithCode(http.StatusCreated))
}

func (s *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.OrganizationService().Get(context.Background(), vars["name"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, OrganizationJSONAPIObject(obj))
}

func (s *Server) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationListOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.OrganizationService().List(context.Background(), opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, OrganizationListJSONAPIObject(obj))
}

func (s *Server) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	opts := otf.OrganizationUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.OrganizationService().Update(context.Background(), name, &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, OrganizationJSONAPIObject(obj))
}

func (s *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	if err := s.OrganizationService().Delete(context.Background(), name); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	obj, err := s.OrganizationService().GetEntitlements(context.Background(), name)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

// OrganizationListJSONAPIObject converts a OrganizationList to a struct that
// can be marshalled into a JSON-API object
func OrganizationListJSONAPIObject(ol *otf.OrganizationList) *OrganizationList {
	obj := &OrganizationList{
		Pagination: ol.Pagination,
	}
	for _, item := range ol.Items {
		obj.Items = append(obj.Items, OrganizationJSONAPIObject(item))
	}

	return obj
}
