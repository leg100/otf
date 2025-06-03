package vcsprovider

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/leg100/otf/internal"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/vcs"
)

const (
	GithubAPIURL  = "https://api.github.com"
	GithubHTTPURL = "https://github.com"
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
	if params.ServiceProvider == nil {
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
	if *params.ServiceProvider != ServiceProviderGithub {
		tfeapi.Error(w, fmt.Errorf("service-provider=%s is unsupported", string(*params.ServiceProvider)))
		return
	}
	if *params.APIURL != GithubAPIURL {
		tfeapi.Error(w, fmt.Errorf("only api-url=%s is supported", GithubAPIURL))
		return
	}
	if *params.HTTPURL != GithubHTTPURL {
		tfeapi.Error(w, fmt.Errorf("only http-url=%s is supported", string(*params.HTTPURL)))
		return
	}

	// default parameters
	if params.Name == nil {
		params.Name = internal.String("")
	}

	oauthClient, err := a.Create(r.Context(), CreateOptions{
		Name:         *params.Name,
		Organization: pathParams.Organization,
		Token:        params.OAuthToken,
		Kind:         vcs.GithubKind,
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

func (a *tfe) convert(from *VCSProvider) *TFEOAuthClient {
	to := &TFEOAuthClient{
		ID:        from.ID,
		CreatedAt: from.CreatedAt,
		// Only github via github.com is supported currently, so hardcode these values.
		ServiceProvider: ServiceProviderGithub,
		APIURL:          GithubAPIURL,
		HTTPURL:         GithubHTTPURL,
		// OTF has no corresponding concept of an OAuthToken, so just use the
		// VCS provider ID (the go-tfe integration tests we use expect
		// at least an ID).
		OAuthTokens: []*TFEOAuthToken{
			{ID: from.ID},
		},
		Organization: &organization.TFEOrganization{Name: from.Organization},
	}
	// an empty name in otf is equivalent to a nil name in tfe
	if from.Name != "" {
		to.Name = &from.Name
	}
	return to
}
