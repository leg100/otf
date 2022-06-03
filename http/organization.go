package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/dto"
)

func (s *Server) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	opts := dto.OrganizationCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	org, err := s.OrganizationService().Create(r.Context(), otf.OrganizationCreateOptions{
		Name:            opts.Name,
		SessionRemember: opts.SessionRemember,
		SessionTimeout:  opts.SessionTimeout,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, org, withCode(http.StatusCreated))
}

func (s *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org, err := s.OrganizationService().Get(r.Context(), vars["name"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, org)
}

func (s *Server) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	org, err := s.OrganizationService().List(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, org)
}

func (s *Server) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	opts := dto.OrganizationUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	org, err := s.OrganizationService().Update(r.Context(), name, &otf.OrganizationUpdateOptions{
		Name:            opts.Name,
		SessionRemember: opts.SessionRemember,
		SessionTimeout:  opts.SessionTimeout,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, org)
}

func (s *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := s.OrganizationService().Delete(r.Context(), name); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	entitlements, err := s.OrganizationService().GetEntitlements(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, entitlements)
}
