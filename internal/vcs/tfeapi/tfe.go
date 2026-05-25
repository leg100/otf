package tfeapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/leg100/otf/internal"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/vcs"
)

type TFEAPI struct {
	Client    tfeClient
	Responder *tfeapi.Responder
}

type tfeClient interface {
	CreateVCSProvider(ctx context.Context, opts vcs.CreateOptions) (*vcs.Provider, error)
	GetVCSProvider(ctx context.Context, id resource.TfeID) (*vcs.Provider, error)
	UpdateVCSProvider(ctx context.Context, id resource.TfeID, opts vcs.UpdateOptions) (*vcs.Provider, error)
	ListVCSProviders(ctx context.Context, organization organization.Name) ([]*vcs.Provider, error)
	DeleteVCSProvider(ctx context.Context, id resource.TfeID) (*vcs.Provider, error)

	GetKind(id vcs.KindID) (vcs.Kind, error)
	GetKinds() []vcs.Kind
	GetKindByTFEServiceProviderType(sp vcs.TFEServiceProviderType) (vcs.Kind, error)
}

func (a *TFEAPI) AddHandlers(r *mux.Router) {
	r.HandleFunc("/organizations/{organization_name}/oauth-clients", a.createOAuthClient).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/oauth-clients", a.listOAuthClients).Methods("GET")
	r.HandleFunc("/oauth-clients/{oauth_client_id}", a.getOAuthClient).Methods("GET")
	r.HandleFunc("/oauth-clients/{oauth_client_id}", a.deleteOAuthClient).Methods("DELETE")
}

func (a *TFEAPI) createOAuthClient(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	var params vcs.TFEOAuthClientCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	// required parameters
	if params.OAuthToken == nil {
		tfeapi.Error(w, &internal.ErrMissingParameter{Parameter: "oauth-token-string"})
		return
	}
	if params.ServiceProvider == nil || *params.ServiceProvider == "" {
		tfeapi.Error(w, &internal.ErrMissingParameter{Parameter: "service-provider"})
		return
	}
	if params.APIURL == nil {
		tfeapi.Error(w, &internal.ErrMissingParameter{Parameter: "api-url"})
		return
	}
	if params.HTTPURL == nil {
		tfeapi.Error(w, &internal.ErrMissingParameter{Parameter: "http-url"})
		return
	}

	// unsupported parameters
	if params.PrivateKey != nil {
		tfeapi.Error(w, errors.New("private-key parameter is unsupported"))
		return
	}
	if params.RSAPublicKey != nil {
		tfeapi.Error(w, errors.New("rsa-public-key parameter is unsupported"))
		return
	}
	if params.Secret != nil {
		tfeapi.Error(w, errors.New("secret parameter is unsupported"))
		return
	}

	kind, err := a.Client.GetKindByTFEServiceProviderType(*params.ServiceProvider)
	if err != nil {
		tfeapi.Error(w, fmt.Errorf("service-provider=%s is unsupported", string(*params.ServiceProvider)))
		return
	}

	// default parameters
	if params.Name == nil {
		params.Name = new("")
	}

	oauthClient, err := a.Client.CreateVCSProvider(r.Context(), vcs.CreateOptions{
		Name:                   *params.Name,
		Organization:           pathParams.Organization,
		Token:                  params.OAuthToken,
		KindID:                 kind.ID,
		BaseURL:                params.HTTPURL,
		APIURL:                 params.APIURL,
		TFEServiceProviderType: params.ServiceProvider,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Responder.Respond(w, r, a.convert(oauthClient), http.StatusCreated)
}

func (a *TFEAPI) listOAuthClients(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	providers, err := a.Client.ListVCSProviders(r.Context(), params.Organization)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	to := make([]*vcs.TFEOAuthClient, len(providers))
	for i, from := range providers {
		to[i] = a.convert(from)
	}
	a.Responder.Respond(w, r, to, http.StatusOK)
}

func (a *TFEAPI) getOAuthClient(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("oauth_client_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	oauthClient, err := a.Client.GetVCSProvider(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Responder.Respond(w, r, a.convert(oauthClient), http.StatusOK)
}

func (a *TFEAPI) deleteOAuthClient(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("oauth_client_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if _, err = a.Client.DeleteVCSProvider(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *TFEAPI) convert(from *vcs.Provider) *vcs.TFEOAuthClient {
	to := &vcs.TFEOAuthClient{
		ID:              from.ID,
		CreatedAt:       from.CreatedAt,
		ServiceProvider: from.ServiceProviderType,
		// OTF has no corresponding concept of an OAuthToken, so just use the
		// VCS provider ID (the go-tfe integration tests we use expect
		// at least an ID).
		OAuthTokens: []*vcs.TFEOAuthToken{
			{ID: from.ID},
		},
		Organization: &organization.TFEOrganization{Name: from.Organization},
		APIURL:       from.APIURL.String(),
		HTTPURL:      from.BaseURL.String(),
	}
	// an empty name in OTF is equivalent to a nil name in TFE
	if from.Name != "" {
		to.Name = &from.Name
	}
	return to
}
