package organization

import (
	"net/http"

	"github.com/gorilla/mux"
	otfhttp "github.com/leg100/otf/http"
	"github.com/leg100/otf/http/decode"
	"github.com/leg100/otf/http/jsonapi"
)

type api struct {
	svc Service

	*jsonapiMarshaler
}

// Implements TFC state versions API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/organizations
func (h *api) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

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

	org, err := h.svc.CreateOrganization(r.Context(), OrganizationCreateOptions{
		Name:            opts.Name,
		SessionRemember: opts.SessionRemember,
		SessionTimeout:  opts.SessionTimeout,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	h.writeResponse(w, r, org, jsonapi.WithCode(http.StatusCreated))
}

func (h *api) GetOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	org, err := h.svc.GetOrganization(r.Context(), name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	h.writeResponse(w, r, org)
}

func (h *api) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts OrganizationListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	list, err := h.svc.ListOrganizations(r.Context(), opts)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	h.writeResponse(w, r, list)
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
	org, err := h.svc.UpdateOrganization(r.Context(), name, OrganizationUpdateOptions{
		Name:            opts.Name,
		SessionRemember: opts.SessionRemember,
		SessionTimeout:  opts.SessionTimeout,
	})
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	h.writeResponse(w, r, org)
}

func (h *api) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		jsonapi.Error(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := h.svc.DeleteOrganization(r.Context(), name); err != nil {
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

	entitlements, err := h.svc.getEntitlements(r.Context(), name)
	if err != nil {
		jsonapi.Error(w, http.StatusNotFound, err)
		return
	}

	h.writeResponse(w, r, entitlements)
}

// writeResponse encodes v as json:api and writes it to the body of the http response.
func (h *api) writeResponse(w http.ResponseWriter, r *http.Request, v any, opts ...func(http.ResponseWriter)) {
	var payload any

	switch v := v.(type) {
	case *OrganizationList:
		payload = h.toList(v)
	case *Organization:
		payload = h.toOrganization(v)
	case Entitlements:
		payload = jsonapi.Entitlements(v)
	}
	jsonapi.WriteResponse(w, r, payload, opts...)
}
