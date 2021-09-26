package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/jsonapi"
	"github.com/leg100/otf"
)

// Organization represents a Terraform Enterprise organization.
type Organization struct {
	Name                   string                       `jsonapi:"primary,organizations"`
	CollaboratorAuthPolicy otf.AuthPolicyType           `jsonapi:"attr,collaborator-auth-policy"`
	CostEstimationEnabled  bool                         `jsonapi:"attr,cost-estimation-enabled"`
	CreatedAt              time.Time                    `jsonapi:"attr,created-at,iso8601"`
	Email                  string                       `jsonapi:"attr,email"`
	ExternalID             string                       `jsonapi:"attr,external-id"`
	OwnersTeamSAMLRoleID   string                       `jsonapi:"attr,owners-team-saml-role-id"`
	Permissions            *otf.OrganizationPermissions `jsonapi:"attr,permissions"`
	SAMLEnabled            bool                         `jsonapi:"attr,saml-enabled"`
	SessionRemember        int                          `jsonapi:"attr,session-remember"`
	SessionTimeout         int                          `jsonapi:"attr,session-timeout"`
	TrialExpiresAt         time.Time                    `jsonapi:"attr,trial-expires-at,iso8601"`
	TwoFactorConformant    bool                         `jsonapi:"attr,two-factor-conformant"`
}

// OrganizationList represents a list of organizations.
type OrganizationList struct {
	*otf.Pagination
	Items []*Organization
}

// ToDomain converts http organization obj to a domain organization obj.
func (o *Organization) ToDomain() *otf.Organization {
	return &otf.Organization{
		ID:                     o.ExternalID,
		Name:                   o.Name,
		CollaboratorAuthPolicy: o.CollaboratorAuthPolicy,
		CostEstimationEnabled:  o.CostEstimationEnabled,
		Email:                  o.Email,
		OwnersTeamSAMLRoleID:   o.OwnersTeamSAMLRoleID,
		Permissions:            o.Permissions,
		SAMLEnabled:            o.SAMLEnabled,
		SessionRemember:        o.SessionRemember,
		SessionTimeout:         o.SessionTimeout,
		TrialExpiresAt:         o.TrialExpiresAt,
		TwoFactorConformant:    o.TwoFactorConformant,
	}
}

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

	obj, err := s.OrganizationService.Create(r.Context(), opts)
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
	var opts otf.OrganizationListOptions

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

	opts := otf.OrganizationUpdateOptions{}
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

	WriteResponse(w, r, obj)
}

// OrganizationJSONAPIObject converts a Organization to a struct
// that can be marshalled into a JSON-API object
func (s *Server) OrganizationJSONAPIObject(org *otf.Organization) *Organization {
	obj := &Organization{
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
func (s *Server) OrganizationListJSONAPIObject(cvl *otf.OrganizationList) *OrganizationList {
	obj := &OrganizationList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, s.OrganizationJSONAPIObject(item))
	}

	return obj
}
