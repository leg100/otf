package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
)

func (s *Server) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	opts := tfe.OrganizationCreateOptions{}

	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := opts.Valid(); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.OrganizationService.Create(&opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.OrganizationJSONAPIObject(obj), WithCode(http.StatusCreated))
}

func (s *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.OrganizationService.Get(vars["name"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.OrganizationJSONAPIObject(obj))
}

func (s *Server) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts tfe.OrganizationListOptions

	if err := DecodeQuery(&opts, r.URL.Query()); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.OrganizationService.List(opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.OrganizationListJSONAPIObject(obj))
}

func (s *Server) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	opts := tfe.OrganizationUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	obj, err := s.OrganizationService.Update(name, &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, s.OrganizationJSONAPIObject(obj))
}

func (s *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	if err := s.OrganizationService.Delete(name); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	obj, err := s.OrganizationService.GetEntitlements(name)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj.Entitlements)
}

// OrganizationJSONAPIObject converts a Organization to a struct
// that can be marshalled into a JSON-API object
func (s *Server) OrganizationJSONAPIObject(org *otf.Organization) *tfe.Organization {
	obj := &tfe.Organization{
		Name:                   org.Name,
		CollaboratorAuthPolicy: org.CollaboratorAuthPolicy,
		CostEstimationEnabled:  org.CostEstimationEnabled,
		CreatedAt:              org.CreatedAt,
		Email:                  org.Email,
		ExternalID:             org.ID,
		OwnersTeamSAMLRoleID:   org.OwnersTeamSAMLRoleID,
		Permissions:            org.Permissions,
		SAMLEnabled:            org.SAMLEnabled,
		SessionRemember:        org.SessionRemember,
		SessionTimeout:         org.SessionTimeout,
		TrialExpiresAt:         org.TrialExpiresAt,
		TwoFactorConformant:    org.TwoFactorConformant,
	}

	return obj
}

// OrganizationListJSONAPIObject converts a OrganizationList to
// a struct that can be marshalled into a JSON-API object
func (s *Server) OrganizationListJSONAPIObject(cvl *otf.OrganizationList) *tfe.OrganizationList {
	obj := &tfe.OrganizationList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, s.OrganizationJSONAPIObject(item))
	}

	return obj
}
