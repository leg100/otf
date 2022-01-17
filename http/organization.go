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
	Name                  string                       `jsonapi:"primary,organizations"`
	CostEstimationEnabled bool                         `jsonapi:"attr,cost-estimation-enabled"`
	CreatedAt             time.Time                    `jsonapi:"attr,created-at,iso8601"`
	ExternalID            string                       `jsonapi:"attr,external-id"`
	OwnersTeamSAMLRoleID  string                       `jsonapi:"attr,owners-team-saml-role-id"`
	Permissions           *otf.OrganizationPermissions `jsonapi:"attr,permissions"`
	SAMLEnabled           bool                         `jsonapi:"attr,saml-enabled"`
	SessionRemember       int                          `jsonapi:"attr,session-remember"`
	SessionTimeout        int                          `jsonapi:"attr,session-timeout"`
	TrialExpiresAt        time.Time                    `jsonapi:"attr,trial-expires-at,iso8601"`
	TwoFactorConformant   bool                         `jsonapi:"attr,two-factor-conformant"`
}

// OrganizationList represents a list of organizations.
type OrganizationList struct {
	*otf.Pagination
	Items []*Organization
}

// ToDomain converts http organization obj to a domain organization obj.
func (o *Organization) ToDomain() *otf.Organization {
	return &otf.Organization{
		ID:              o.ExternalID,
		Name:            o.Name,
		SessionRemember: o.SessionRemember,
		SessionTimeout:  o.SessionTimeout,
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

	obj, err := s.OrganizationService().Create(r.Context(), opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, OrganizationJSONAPIObject(obj), WithCode(http.StatusCreated))
}

func (s *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.OrganizationService().Get(vars["name"])
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

	obj, err := s.OrganizationService().List(opts)
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

	obj, err := s.OrganizationService().Update(name, &opts)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, OrganizationJSONAPIObject(obj))
}

func (s *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	if err := s.OrganizationService().Delete(name); err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	obj, err := s.OrganizationService().GetEntitlements(name)
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	WriteResponse(w, r, obj)
}

// OrganizationJSONAPIObject converts a Organization to a struct
// that can be marshalled into a JSON-API object
func OrganizationJSONAPIObject(org *otf.Organization) *Organization {
	obj := &Organization{
		Name:            org.Name,
		CreatedAt:       org.CreatedAt,
		ExternalID:      org.ID,
		Permissions:     &otf.DefaultOrganizationPermissions,
		SessionRemember: org.SessionRemember,
		SessionTimeout:  org.SessionTimeout,
	}

	return obj
}

// OrganizationListJSONAPIObject converts a OrganizationList to
// a struct that can be marshalled into a JSON-API object
func OrganizationListJSONAPIObject(cvl *otf.OrganizationList) *OrganizationList {
	obj := &OrganizationList{
		Pagination: cvl.Pagination,
	}
	for _, item := range cvl.Items {
		obj.Items = append(obj.Items, OrganizationJSONAPIObject(item))
	}

	return obj
}
