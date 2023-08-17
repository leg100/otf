package organization

import (
	"context"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/api"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
)

type tfe struct {
	Service
	*api.Responder
}

// Implements TFC organizations API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/organizations
func (a *tfe) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/organizations", a.createOrganization).Methods("POST")
	r.HandleFunc("/organizations", a.listOrganizations).Methods("GET")
	r.HandleFunc("/organizations/{name}", a.getOrganization).Methods("GET")
	r.HandleFunc("/organizations/{name}", a.updateOrganization).Methods("PATCH")
	r.HandleFunc("/organizations/{name}", a.deleteOrganization).Methods("DELETE")
	r.HandleFunc("/organizations/{name}/entitlement-set", a.getEntitlements).Methods("GET")
}

func (a *tfe) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts types.OrganizationCreateOptions
	if err := api.Unmarshal(r.Body, &opts); err != nil {
		api.Error(w, err)
		return
	}

	org, err := a.CreateOrganization(r.Context(), CreateOptions{
		Name:                       opts.Name,
		Email:                      opts.Email,
		CollaboratorAuthPolicy:     (*string)(opts.CollaboratorAuthPolicy),
		CostEstimationEnabled:      opts.CostEstimationEnabled,
		SessionRemember:            opts.SessionRemember,
		SessionTimeout:             opts.SessionTimeout,
		AllowForceDeleteWorkspaces: opts.AllowForceDeleteWorkspaces,
	})
	if err != nil {
		api.Error(w, err)
		return
	}

	a.Respond(w, r, org, http.StatusCreated)
}

func (a *tfe) getOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		api.Error(w, err)
		return
	}

	org, err := a.GetOrganization(r.Context(), name)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.Respond(w, r, org, http.StatusOK)
}

func (a *tfe) listOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts ListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		api.Error(w, err)
		return
	}

	page, err := a.ListOrganizations(r.Context(), opts)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.RespondWithPage(w, r, page.Items, page.Pagination)
}

func (a *tfe) updateOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		api.Error(w, err)
		return
	}
	var opts types.OrganizationUpdateOptions
	if err := api.Unmarshal(r.Body, &opts); err != nil {
		api.Error(w, err)
		return
	}

	org, err := a.UpdateOrganization(r.Context(), name, UpdateOptions{
		Name:                   opts.Name,
		Email:                  opts.Email,
		CollaboratorAuthPolicy: (*string)(opts.CollaboratorAuthPolicy),
		CostEstimationEnabled:  opts.CostEstimationEnabled,
		SessionRemember:        opts.SessionRemember,
		SessionTimeout:         opts.SessionTimeout,
	})
	if err != nil {
		api.Error(w, err)
		return
	}

	a.Respond(w, r, org, http.StatusOK)
}

func (a *tfe) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		api.Error(w, err)
		return
	}

	if err := a.DeleteOrganization(r.Context(), name); err != nil {
		api.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) getEntitlements(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		api.Error(w, err)
		return
	}

	entitlements, err := a.GetEntitlements(r.Context(), name)
	if err != nil {
		api.Error(w, err)
		return
	}

	a.Respond(w, r, (*types.Entitlements)(&entitlements), http.StatusOK)
}

func (a *tfe) include(ctx context.Context, v any) (any, error) {
	dst := reflect.Indirect(reflect.ValueOf(v))

	// v must be a struct with a field named Organization of kind string
	if dst.Kind() != reflect.Struct {
		return nil, nil
	}
	name := dst.FieldByName("Organization")
	if !name.IsValid() {
		return nil, nil
	}
	if name.Kind() != reflect.String {
		return nil, nil
	}
	org, err := a.GetOrganization(ctx, name.String())
	if err != nil {
		return nil, err
	}
	return a.toOrganization(org), nil
}

func (a *tfe) toOrganization(from *Organization) *types.Organization {
	to := &types.Organization{
		Name:                       from.Name,
		CreatedAt:                  from.CreatedAt,
		ExternalID:                 from.ID,
		Permissions:                &types.DefaultOrganizationPermissions,
		SessionRemember:            from.SessionRemember,
		SessionTimeout:             from.SessionTimeout,
		AllowForceDeleteWorkspaces: from.AllowForceDeleteWorkspaces,
		CostEstimationEnabled:      from.CostEstimationEnabled,
	}
	if from.Email != nil {
		to.Email = *from.Email
	}
	if from.CollaboratorAuthPolicy != nil {
		to.CollaboratorAuthPolicy = types.AuthPolicyType(*from.CollaboratorAuthPolicy)
	}
	return to
}
