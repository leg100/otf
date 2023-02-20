package organization

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type api struct {
	app application
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/organizations
//
func (h *api) addHandlers(r *mux.Router) {
	r.HandleFunc("/organizations", h.ListOrganizations)
	r.HandleFunc("/organizations", h.CreateOrganization)
	r.HandleFunc("/organizations/{name}", h.GetOrganization)
	r.HandleFunc("/organizations/{name}", h.UpdateOrganization)
	r.HandleFunc("/organizations/{name}", h.DeleteOrganization)
	r.HandleFunc("/organizations/{name}/entitlement-set", h.GetEntitlements)
}

func (h *api) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	opts := jsonapi.OrganizationCreateOptions{}
	if err := jsonapi.UnmarshalPayload(r.Body, &opts); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	org, err := h.app.create(r.Context(), otf.OrganizationCreateOptions{
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

func (h *api) GetOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	org, err := h.app.get(r.Context(), name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jsonapi.WriteResponse(w, r, org)
}

func (h *api) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts ListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	list, err := h.app.list(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	jsonapi.WriteResponse(w, r, list)
}

func (h *api) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
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
	org, err := h.app.update(r.Context(), name, UpdateOptions{
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

func (h *api) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := h.app.delete(r.Context(), name); err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *api) GetEntitlements(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	entitlements, err := h.app.getEntitlements(r.Context(), name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	response := jsonapi.Entitlements(entitlements)
	jsonapi.WriteResponse(w, r, &response)
}
