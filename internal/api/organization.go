package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
)

// Implements TFC organizations API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/organizations
func (a *api) addOrganizationHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/organizations", a.createOrganization).Methods("POST")
	r.HandleFunc("/organizations", a.listOrganizations).Methods("GET")
	r.HandleFunc("/organizations/{name}", a.getOrganization).Methods("GET")
	r.HandleFunc("/organizations/{name}", a.updateOrganization).Methods("PATCH")
	r.HandleFunc("/organizations/{name}", a.deleteOrganization).Methods("DELETE")
	r.HandleFunc("/organizations/{name}/entitlement-set", a.getEntitlements).Methods("GET")
}

func (a *api) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts types.OrganizationCreateOptions
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}

	org, err := a.CreateOrganization(r.Context(), organization.OrganizationCreateOptions{
		Name:                       opts.Name,
		Email:                      opts.Email,
		CollaboratorAuthPolicy:     (*string)(opts.CollaboratorAuthPolicy),
		SessionRemember:            opts.SessionRemember,
		SessionTimeout:             opts.SessionTimeout,
		AllowForceDeleteWorkspaces: opts.AllowForceDeleteWorkspaces,
	})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, org, withCode(http.StatusCreated))
}

func (a *api) getOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		Error(w, err)
		return
	}

	org, err := a.GetOrganization(r.Context(), name)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, org)
}

func (a *api) listOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts organization.ListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		Error(w, err)
		return
	}

	list, err := a.ListOrganizations(r.Context(), opts)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, list)
}

func (a *api) updateOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		Error(w, err)
		return
	}
	var opts types.OrganizationUpdateOptions
	if err := unmarshal(r.Body, &opts); err != nil {
		Error(w, err)
		return
	}

	org, err := a.UpdateOrganization(r.Context(), name, organization.OrganizationUpdateOptions{
		Name:                   opts.Name,
		Email:                  opts.Email,
		CollaboratorAuthPolicy: (*string)(opts.CollaboratorAuthPolicy),
		SessionRemember:        opts.SessionRemember,
		SessionTimeout:         opts.SessionTimeout,
	})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, org)
}

func (a *api) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		Error(w, err)
		return
	}

	if err := a.DeleteOrganization(r.Context(), name); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *api) getEntitlements(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		Error(w, err)
		return
	}

	entitlements, err := a.GetEntitlements(r.Context(), name)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, entitlements)
}
