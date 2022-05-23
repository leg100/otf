package http

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	opts := otf.OrganizationCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := opts.Valid(); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	obj, err := s.OrganizationService().Create(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, OrganizationDTO(obj), withCode(http.StatusCreated))
}

func (s *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	obj, err := s.OrganizationService().Get(context.Background(), vars["name"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, OrganizationDTO(obj))
}

func (s *Server) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	obj, err := s.OrganizationService().List(context.Background(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, OrganizationListDTO(obj))
}

func (s *Server) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	opts := otf.OrganizationUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	obj, err := s.OrganizationService().Update(context.Background(), name, &opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, OrganizationDTO(obj))
}

func (s *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := s.OrganizationService().Delete(context.Background(), name); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	obj, err := s.OrganizationService().GetEntitlements(context.Background(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, obj)
}

// OrganizationDTO converts an org into a DTO
func OrganizationDTO(org *otf.Organization) *dto.Organization {
	return &dto.Organization{
		Name:            org.Name(),
		CreatedAt:       org.CreatedAt,
		ExternalID:      org.ID(),
		Permissions:     &dto.DefaultOrganizationPermissions,
		SessionRemember: org.SessionRemember(),
		SessionTimeout:  org.SessionTimeout(),
	}
}

// OrganizationListDTO converts an org list into a DTO
func OrganizationListDTO(ol *otf.OrganizationList) *dto.OrganizationList {
	pagination := dto.Pagination(*ol.Pagination)
	jol := &dto.OrganizationList{
		Pagination: &pagination,
	}
	for _, item := range ol.Items {
		jol.Items = append(jol.Items, OrganizationDTO(item))
	}
	return jol
}
