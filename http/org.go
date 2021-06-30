package http

import (
	"fmt"
	"net/http"

	"github.com/google/jsonapi"
	"github.com/gorilla/mux"
	"github.com/leg100/go-tfe"
)

func (h *Server) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts tfe.OrganizationListOptions
	if err := decoder.Decode(&opts, r.URL.Query()); err != nil {
		ErrUnprocessable(w, fmt.Errorf("unable to decode query string: %w", err))
		return
	}

	SanitizeListOptions(&opts.ListOptions)

	ListObjects(w, r, func() (interface{}, error) {
		return h.OrganizationService.ListOrganizations(opts)
	})
}

func (h *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	GetObject(w, r, func() (interface{}, error) {
		return h.OrganizationService.GetOrganization(vars["name"])
	})
}

func (h *Server) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	CreateObject(w, r, &tfe.OrganizationCreateOptions{}, func(opts interface{}) (interface{}, error) {
		return h.OrganizationService.CreateOrganization(opts.(*tfe.OrganizationCreateOptions))
	})
}

func (h *Server) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	UpdateObject(w, r, &tfe.OrganizationUpdateOptions{}, func(opts interface{}) (interface{}, error) {
		return h.OrganizationService.UpdateOrganization(name, opts.(*tfe.OrganizationUpdateOptions))
	})
}

func (h *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	if err := h.OrganizationService.DeleteOrganization(name); err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Server) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	entitlements, err := h.OrganizationService.GetEntitlements(name)
	if err != nil {
		ErrNotFound(w)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", jsonapi.MediaType)
	if err := jsonapi.MarshalPayload(w, entitlements); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
