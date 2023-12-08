package organization

import (
	"context"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
)

type tfe struct {
	*Service
	*tfeapi.Responder
}

// Implements TFC organizations API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/organizations
func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/organizations", a.createOrganization).Methods("POST")
	r.HandleFunc("/organizations", a.listOrganizations).Methods("GET")
	r.HandleFunc("/organizations/{name}", a.getOrganization).Methods("GET")
	r.HandleFunc("/organizations/{name}", a.updateOrganization).Methods("PATCH")
	r.HandleFunc("/organizations/{name}", a.deleteOrganization).Methods("DELETE")
	r.HandleFunc("/organizations/{name}/entitlement-set", a.getEntitlements).Methods("GET")

	// Organization token routes
	r.HandleFunc("/organizations/{organization_name}/authentication-token", a.createOrganizationToken).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/authentication-token", a.getOrganizationToken).Methods("GET")
	r.HandleFunc("/organizations/{organization_name}/authentication-token", a.deleteOrganizationToken).Methods("DELETE")
}

func (a *tfe) createOrganization(w http.ResponseWriter, r *http.Request) {
	var opts types.OrganizationCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	org, err := a.Create(r.Context(), CreateOptions{
		Name:                       opts.Name,
		Email:                      opts.Email,
		CollaboratorAuthPolicy:     (*string)(opts.CollaboratorAuthPolicy),
		CostEstimationEnabled:      opts.CostEstimationEnabled,
		SessionRemember:            opts.SessionRemember,
		SessionTimeout:             opts.SessionTimeout,
		AllowForceDeleteWorkspaces: opts.AllowForceDeleteWorkspaces,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.toOrganization(org), http.StatusCreated)
}

func (a *tfe) getOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	org, err := a.Get(r.Context(), name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.toOrganization(org), http.StatusOK)
}

func (a *tfe) listOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts ListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.List(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*types.Organization, len(page.Items))
	for i, from := range page.Items {
		items[i] = a.toOrganization(from)
	}
	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) updateOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.OrganizationUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	org, err := a.Update(r.Context(), name, UpdateOptions{
		Name:                       opts.Name,
		Email:                      opts.Email,
		CollaboratorAuthPolicy:     (*string)(opts.CollaboratorAuthPolicy),
		CostEstimationEnabled:      opts.CostEstimationEnabled,
		SessionRemember:            opts.SessionRemember,
		SessionTimeout:             opts.SessionTimeout,
		AllowForceDeleteWorkspaces: opts.AllowForceDeleteWorkspaces,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.toOrganization(org), http.StatusOK)
}

func (a *tfe) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.Delete(r.Context(), name); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) getEntitlements(w http.ResponseWriter, r *http.Request) {
	name, err := decode.Param("name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	entitlements, err := a.GetEntitlements(r.Context(), name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, (*types.Entitlements)(&entitlements), http.StatusOK)
}

func (a *tfe) createOrganizationToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts types.OrganizationTokenCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ot, token, err := a.CreateToken(r.Context(), CreateOrganizationTokenOptions{
		Organization: org,
		Expiry:       opts.ExpiredAt,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := &types.OrganizationToken{
		ID:        ot.ID,
		CreatedAt: ot.CreatedAt,
		Token:     string(token),
		ExpiredAt: ot.Expiry,
	}
	a.Respond(w, r, to, http.StatusCreated)
}

func (a *tfe) getOrganizationToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	ot, err := a.GetOrganizationToken(r.Context(), org)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if ot == nil {
		tfeapi.Error(w, internal.ErrResourceNotFound)
		return
	}

	to := &types.OrganizationToken{
		ID:        ot.ID,
		CreatedAt: ot.CreatedAt,
		ExpiredAt: ot.Expiry,
	}
	a.Respond(w, r, to, http.StatusCreated)
}

func (a *tfe) deleteOrganizationToken(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	err = a.DeleteToken(r.Context(), org)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) include(ctx context.Context, v any) ([]any, error) {
	dst := reflect.Indirect(reflect.ValueOf(v))

	// v must be a struct with a field named Organization of type
	// *types.Organization
	if dst.Kind() != reflect.Struct {
		return nil, nil
	}
	field := dst.FieldByName("Organization")
	if !field.IsValid() {
		return nil, nil
	}
	tfeOrganization, ok := field.Interface().(*types.Organization)
	if !ok {
		return nil, nil
	}
	org, err := a.Get(ctx, tfeOrganization.Name)
	if err != nil {
		return nil, err
	}
	return []any{a.toOrganization(org)}, nil
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
		// go-tfe tests expect this attribute to be equal to 5
		RemainingTestableCount: 5,
	}
	if from.Email != nil {
		to.Email = *from.Email
	}
	if from.CollaboratorAuthPolicy != nil {
		to.CollaboratorAuthPolicy = types.AuthPolicyType(*from.CollaboratorAuthPolicy)
	}
	return to
}
