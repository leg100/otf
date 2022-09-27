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
	org, err := s.Application.CreateOrganization(r.Context(), otf.OrganizationCreateOptions{
		Name:            opts.Name,
		SessionRemember: opts.SessionRemember,
		SessionTimeout:  opts.SessionTimeout,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Organization{org}, withCode(http.StatusCreated))
}

func (s *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	org, err := s.Application.GetOrganization(r.Context(), vars["name"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Organization{org})
}

func (s *Server) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts otf.OrganizationListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	list, err := s.Application.ListOrganizations(r.Context(), opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &OrganizationList{list})
}

func (s *Server) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	opts := dto.OrganizationUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	org, err := s.Application.UpdateOrganization(r.Context(), name, &otf.OrganizationUpdateOptions{
		Name:            opts.Name,
		SessionRemember: opts.SessionRemember,
		SessionTimeout:  opts.SessionTimeout,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Organization{org})
}

func (s *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := s.Application.DeleteOrganization(r.Context(), name); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	entitlements, err := s.Application.GetEntitlements(r.Context(), name)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, &Entitlements{entitlements})
}

type Organization struct {
	*otf.Organization
}

// ToJSONAPI assembles a JSONAPI DTO
func (org *Organization) ToJSONAPI() any {
	return &dto.Organization{
		Name:            org.Name(),
		CreatedAt:       org.CreatedAt(),
		ExternalID:      org.ID(),
		Permissions:     &dto.DefaultOrganizationPermissions,
		SessionRemember: org.SessionRemember(),
		SessionTimeout:  org.SessionTimeout(),
	}
}

type OrganizationList struct {
	*otf.OrganizationList
}

// ToJSONAPI assembles a JSON-API DTO.
func (l *OrganizationList) ToJSONAPI() any {
	obj := &dto.OrganizationList{
		Pagination: l.Pagination.ToJSONAPI(),
	}
	for _, item := range l.Items {
		obj.Items = append(obj.Items, (&Organization{item}).ToJSONAPI().(*dto.Organization))
	}
	return obj
}

type Entitlements struct {
	*otf.Entitlements
}

// ToJSONAPI assembles a JSONAPI DTO
func (e *Entitlements) ToJSONAPI() any {
	return (*dto.Entitlements)(e.Entitlements)
}
