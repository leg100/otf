package api

import (
	"context"
	"errors"
	"net/http"
	"reflect"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/workspace/execution"
)

//lint:ignore ST1005 go-tfe integration tests expect capitalized error string
var errTFEAgentPoolSpecifiedWithNonAgentExecutionMode = errors.New("Default agent pool must not be specified unless using 'agent' execution mode")

// Implements TFC organizations API:
//
// https://developer.hashicorp.com/terraform/cloud-docs/api-docs/organizations
type tfe struct {
	*tfeapi.Responder
	client tfeClient
}

type tfeClient interface {
	CreateOrganization(ctx context.Context, opts organization.CreateOptions) (*organization.Organization, error)
	UpdateOrganization(ctx context.Context, name organization.Name, opts organization.UpdateOptions) (*organization.Organization, error)
	GetOrganization(ctx context.Context, name organization.Name) (*organization.Organization, error)
	ListOrganizations(ctx context.Context, opts organization.ListOptions) (*resource.Page[*organization.Organization], error)
	DeleteOrganization(ctx context.Context, name organization.Name) error

	CreateOrganizationToken(ctx context.Context, opts organization.CreateOrganizationTokenOptions) (*organization.OrganizationToken, []byte, error)
	GetOrganizationToken(ctx context.Context, organization organization.Name) (*organization.OrganizationToken, error)
	ListOrganizationTokens(ctx context.Context, org organization.Name) ([]*organization.OrganizationToken, error)
	DeleteOrganizationToken(ctx context.Context, org organization.Name) error

	GetOrganizationEntitlements(ctx context.Context, organization organization.Name) (organization.Entitlements, error)
}

func NewTFEAPI(
	client tfeClient,
	responder *tfeapi.Responder,
) *tfe {
	api := &tfe{
		Responder: responder,
		client:    client,
	}

	// Fetch organization when API calls request organization be included in the
	// response
	responder.Register(tfeapi.IncludeOrganization, api.include)

	return api
}

func (a *tfe) AddHandlers(r *mux.Router) {
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
	var opts organization.TFEOrganizationCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	org, err := a.client.CreateOrganization(r.Context(), organization.CreateOptions{
		Name:                       opts.Name,
		Email:                      opts.Email,
		CollaboratorAuthPolicy:     (*string)(opts.CollaboratorAuthPolicy),
		CostEstimationEnabled:      opts.CostEstimationEnabled,
		SessionRemember:            opts.SessionRemember,
		SessionTimeout:             opts.SessionTimeout,
		AllowForceDeleteWorkspaces: opts.AllowForceDeleteWorkspaces,
		DefaultExecutionMode:       opts.DefaultExecutionMode,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.toOrganization(org), http.StatusCreated)
}

func (a *tfe) getOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	org, err := a.client.GetOrganization(r.Context(), params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.toOrganization(org), http.StatusOK)
}

func (a *tfe) listOrganizations(w http.ResponseWriter, r *http.Request) {
	var opts organization.ListOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		tfeapi.Error(w, err)
		return
	}

	page, err := a.client.ListOrganizations(r.Context(), opts)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	items := make([]*organization.TFEOrganization, len(page.Items))
	for i, from := range page.Items {
		items[i] = a.toOrganization(from)
	}
	a.RespondWithPage(w, r, items, page.Pagination)
}

func (a *tfe) updateOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var tfeOpts organization.TFEOrganizationUpdateOptions
	if err := tfeapi.Unmarshal(r.Body, &tfeOpts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	opts := organization.UpdateOptions{
		Name:                       tfeOpts.Name,
		Email:                      tfeOpts.Email,
		CollaboratorAuthPolicy:     (*string)(tfeOpts.CollaboratorAuthPolicy),
		CostEstimationEnabled:      tfeOpts.CostEstimationEnabled,
		SessionRemember:            tfeOpts.SessionRemember,
		SessionTimeout:             tfeOpts.SessionTimeout,
		AllowForceDeleteWorkspaces: tfeOpts.AllowForceDeleteWorkspaces,
		DefaultExecutionMode:       tfeOpts.DefaultExecutionMode,
	}
	if tfeOpts.DefaultAgentPool != nil {
		opts.DefaultAgentPoolID = &tfeOpts.DefaultAgentPool.ID
	}
	org, err := a.client.UpdateOrganization(r.Context(), params.Name, opts)
	if err != nil {
		if errors.Is(err, execution.ErrNonAgentExecutionModeWithPool) {
			// Return TFE specific error message to satisfy go-tfe integration tests.
			tfeapi.Error(w, errTFEAgentPoolSpecifiedWithNonAgentExecutionMode, tfeapi.WithStatus(http.StatusUnprocessableEntity))
			return
		}
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.toOrganization(org), http.StatusOK)
}

func (a *tfe) deleteOrganization(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	if err := a.client.DeleteOrganization(r.Context(), params.Name); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) getEntitlements(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	entitlements, err := a.client.GetOrganizationEntitlements(r.Context(), params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, (*organization.TFEEntitlements)(&entitlements), http.StatusOK)
}

func (a *tfe) createOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}
	var opts organization.TFEOrganizationTokenCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &opts); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ot, token, err := a.client.CreateOrganizationToken(r.Context(), organization.CreateOrganizationTokenOptions{
		Organization: params.Name,
		Expiry:       opts.ExpiredAt,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	to := &organization.TFEOrganizationToken{
		ID:        ot.ID,
		CreatedAt: ot.CreatedAt,
		Token:     string(token),
		ExpiredAt: ot.Expiry,
	}
	a.Respond(w, r, to, http.StatusCreated)
}

func (a *tfe) getOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	ot, err := a.client.GetOrganizationToken(r.Context(), params.Name)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}
	if ot == nil {
		tfeapi.Error(w, internal.ErrResourceNotFound)
		return
	}

	to := &organization.TFEOrganizationToken{
		ID:        ot.ID,
		CreatedAt: ot.CreatedAt,
		ExpiredAt: ot.Expiry,
	}
	a.Respond(w, r, to, http.StatusCreated)
}

func (a *tfe) deleteOrganizationToken(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Name organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	err := a.client.DeleteOrganizationToken(r.Context(), params.Name)
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
	tfeOrganization, ok := field.Interface().(*organization.TFEOrganization)
	if !ok {
		return nil, nil
	}
	org, err := a.client.GetOrganization(ctx, tfeOrganization.Name)
	if err != nil {
		return nil, err
	}
	return []any{a.toOrganization(org)}, nil
}

func (a *tfe) toOrganization(from *organization.Organization) *organization.TFEOrganization {
	to := &organization.TFEOrganization{
		Name:                       from.Name,
		CreatedAt:                  from.CreatedAt,
		ExternalID:                 from.ID,
		Permissions:                &organization.DefaultOrganizationPermissions,
		SessionRemember:            from.SessionRemember,
		SessionTimeout:             from.SessionTimeout,
		AllowForceDeleteWorkspaces: from.AllowForceDeleteWorkspaces,
		CostEstimationEnabled:      from.CostEstimationEnabled,
		// go-tfe tests expect this attribute to be equal to 5
		RemainingTestableCount: 5,
		DefaultExecutionMode:   from.DefaultMode.Kind(),
	}
	if from.Email != nil {
		to.Email = *from.Email
	}
	if from.CollaboratorAuthPolicy != nil {
		to.CollaboratorAuthPolicy = organization.TFEAuthPolicyType(*from.CollaboratorAuthPolicy)
	}
	if from.DefaultMode.AgentPoolID() != nil {
		to.DefaultAgentPool = &organization.TFEAgentPool{
			ID: *from.DefaultMode.AgentPoolID(),
		}
	}
	return to
}
