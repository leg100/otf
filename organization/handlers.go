package organization

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type handlers struct {
	app appService
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/organizations
//
func (h *handlers) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations", h.ListOrganizations)
	r.HandleFunc("/organizations", h.CreateOrganization)
	r.HandleFunc("/organizations/{name}", h.GetOrganization)
	r.HandleFunc("/organizations/{name}", h.UpdateOrganization)
	r.HandleFunc("/organizations/{name}", h.DeleteOrganization)
	r.HandleFunc("/organizations/{name}/entitlement-set", h.GetEntitlements)
}

func (h *handlers) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	opts := jsonapi.OrganizationCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	org, err := h.app.createOrganization(r.Context(), OrganizationCreateOptions{
		Name:            opts.Name,
		SessionRemember: opts.SessionRemember,
		SessionTimeout:  opts.SessionTimeout,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, org, jsonapi.WithCode(http.StatusCreated))
}

func (h *handlers) GetOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	org, err := h.app.GetOrganization(r.Context(), name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, org)
}

func (h *handlers) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts OrganizationListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	list, err := h.app.ListOrganizations(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, list)
}

func (h *handlers) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	opts := jsonapi.OrganizationUpdateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}
	org, err := h.app.UpdateOrganization(r.Context(), name, &OrganizationUpdateOptions{
		Name:            opts.Name,
		SessionRemember: opts.SessionRemember,
		SessionTimeout:  opts.SessionTimeout,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, org)
}

func (h *handlers) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := h.app.DeleteOrganization(r.Context(), name); err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handlers) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	entitlements, err := h.app.GetEntitlements(r.Context(), name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &Entitlements{entitlements})
}
