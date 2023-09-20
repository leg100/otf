package vcsprovider

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/leg100/otf/internal"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/cloud"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tfeapi/types"
)

const (
	GithubAPIURL  = "https://api.github.com"
	GithubHTTPURL = "https://github.com"
)

type tfe struct {
	Service
	*tfeapi.Responder
}

func (a *tfe) addHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/oauth-clients", a.createOAuthClient).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/oauth-clients", a.listOAuthClients).Methods("GET")
	r.HandleFunc("/oauth-clients/{oauth_client_id}", a.getOAuthClient).Methods("GET")
	r.HandleFunc("/oauth-clients/{oauth_client_id}", a.deleteOAuthClient).Methods("DELETE")
}

func (a *tfe) createOAuthClient(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	var params types.OAuthClientCreateOptions
	if err := tfeapi.Unmarshal(r.Body, &params); err != nil {
		tfeapi.Error(w, err)
		return
	}

	// required parameters
	if params.OAuthToken == nil {
		tfeapi.Error(w, &internal.MissingParameterError{Parameter: "oauth-token-string"})
		return
	}
	if params.ServiceProvider == nil {
		tfeapi.Error(w, &internal.MissingParameterError{Parameter: "service-provider"})
		return
	}
	if params.APIURL == nil {
		tfeapi.Error(w, &internal.MissingParameterError{Parameter: "api-url"})
		return
	}
	if params.HTTPURL == nil {
		tfeapi.Error(w, &internal.MissingParameterError{Parameter: "http-url"})
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
	if *params.ServiceProvider != types.ServiceProviderGithub {
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

	oauthClient, err := a.CreateVCSProvider(r.Context(), CreateOptions{
		Name:         *params.Name,
		Organization: org,
		Token:        params.OAuthToken,
		Kind:         cloud.GithubKind,
	})
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(oauthClient), http.StatusCreated)
}

func (a *tfe) listOAuthClients(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	providers, err := a.ListVCSProviders(r.Context(), org)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	// convert items
	to := make([]*types.OAuthClient, len(providers))
	for i, from := range providers {
		to[i] = a.convert(from)
	}
	a.Respond(w, r, to, http.StatusOK)
}

func (a *tfe) getOAuthClient(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("oauth_client_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	oauthClient, err := a.GetVCSProvider(r.Context(), id)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	a.Respond(w, r, a.convert(oauthClient), http.StatusOK)
}

func (a *tfe) deleteOAuthClient(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("oauth_client_id", r)
	if err != nil {
		tfeapi.Error(w, err)
		return
	}

	if _, err = a.DeleteVCSProvider(r.Context(), id); err != nil {
		tfeapi.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *tfe) convert(from *VCSProvider) *types.OAuthClient {
	to := &types.OAuthClient{
		ID:        from.ID,
		CreatedAt: from.CreatedAt,
		// Only github via github.com is supported currently, so hardcode these values.
		ServiceProvider: types.ServiceProviderGithub,
		APIURL:          GithubAPIURL,
		HTTPURL:         GithubHTTPURL,
		// OTF has no corresponding concept of an OAuthToken, so just use the
		// VCS provider ID (the go-tfe integration tests we use expect
		// at least an ID).
		OAuthTokens: []*types.OAuthToken{
			{ID: from.ID},
		},
		Organization: &types.Organization{Name: from.Organization},
	}
	// an empty name in otf is equivalent to a nil name in tfe
	if from.Name != "" {
		to.Name = &from.Name
	}
	return to
}
