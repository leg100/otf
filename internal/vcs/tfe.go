package vcs

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/leg100/otf/internal"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/tfeapi"
)

type tfe struct {
	*Service
	*tfeapi.Responder
}

func (a *tfe) addHandlers(r *mux.Router) {
	r = r.PathPrefix(tfeapi.APIPrefixV2).Subrouter()

	r.HandleFunc("/organizations/{organization_name}/oauth-clients", a.createOAuthClient).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/oauth-clients", a.listOAuthClients).Methods("GET")
	r.HandleFunc("/oauth-clients/{oauth_client_id}", a.getOAuthClient).Methods("GET")
	r.HandleFunc("/oauth-clients/{oauth_client_id}", a.deleteOAuthClient).Methods("DELETE")
}

func (a *tfe) createOAuthClient(w http.ResponseWriter, r *http.Request) {
	var pathParams struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&pathParams, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	var params TFEOAuthClientCreateOptions
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

	kind, err := a.Service.GetKindByTFEServiceProviderType(*params.ServiceProvider)
	if err != nil {
		tfeapi.Error(w, fmt.Errorf("service-provider=%s is unsupported", string(*params.ServiceProvider)))
		return
	}

	// default parameters
	if params.Name == nil {
		params.Name = new("")
	}

	oauthClient, err := a.Create(r.Context(), CreateOptions{
		Name:                   *params.Name,
		Organization:           pathParams.Organization,
		Token:                  params.OAuthToken,
		KindID:                 kind.ID,
		BaseURL:                params.HTTPURL,
		apiURL:                 params.APIURL,
		tfeServiceProviderType: params.ServiceProvider,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(oauthClient), http.StatusCreated)
}

func (a *tfe) listOAuthClients(w http.ResponseWriter, r *http.Request) {
	var params struct {
		Organization organization.Name `schema:"organization_name"`
	}
	if err := decode.All(&params, r); err != nil {
		tfeapi.Error(w, err)
		return
	}

	providers, err := a.List(r.Context(), params.Organization)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	to := make([]*TFEOAuthClient, len(providers))
	for i, from := range providers {
		to[i] = a.convert(from)
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) getOAuthClient(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("oauth_client_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	oauthClient, err := a.Get(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(oauthClient), http.StatusOK)
}

func (a *tfe) deleteOAuthClient(w http.ResponseWriter, r *http.Request) {
	id, err := decode.ID("oauth_client_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if _, err = a.Delete(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convert(from *Provider) *TFEOAuthClient {
	to := &TFEOAuthClient{
		ID:              from.ID,
		CreatedAt:       from.CreatedAt,
		ServiceProvider: from.serviceProviderType,
		// OTF has no corresponding concept of an OAuthToken, so just use the
		// VCS provider ID (the go-tfe integration tests we use expect
		// at least an ID).
		OAuthTokens: []*TFEOAuthToken{
			{ID: from.ID},
		},
		Organization: &organization.TFEOrganization{Name: from.Organization},
		APIURL:       from.apiURL.String(),
		HTTPURL:      from.BaseURL.String(),
	}
	// an empty name in OTF is equivalent to a nil name in TFE
	if from.Name != "" {
		to.Name = &from.Name
	}
	return to
}
