package api

import (
	"net/http"

	"github.com/DataDog/jsonapi"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
)

// Stub implementation of the TFC Organization Memberships API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/organization-memberships
//
// Note: this is only implemented insofar as to satisfy some of the API tests we
// run from the `go-tfe` project.

func (a *api) addOrganizationMembershipHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/organization-memberships", a.inviteUser).Methods("POST")
	r.HandleFunc("/organization-memberships/{id}", a.deleteMembership).Methods("DELETE")
}

func (a *api) inviteUser(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}
	var params types.OrganizationMembershipCreateOptions
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}

	membership := &types.OrganizationMembership{
		ID: internal.NewID("ou"),
		User: &types.User{
			ID: internal.NewID("user"),
		},
		Organization: &types.Organization{
			Name: org,
		},
	}

	b, err := jsonapi.Marshal(membership)
	if err != nil {
		Error(w, err)
		return
	}
	w.Header().Set("Content-type", mediaType)
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}

func (a *api) deleteMembership(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
