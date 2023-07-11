package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcsprovider"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/api/types"
	otfhttp "github.com/leg100/otf/internal/http"
	"github.com/leg100/otf/internal/http/decode"
)

const (
	GithubAPIURL  = "https://api.github.com"
	GithubHTTPURL = "https://github.com"
)

func (a *api) addOAuthClientHandlers(r *mux.Router) {
	r = otfhttp.APIRouter(r)

	r.HandleFunc("/organizations/{organization_name}/oauth-clients", a.createOAuthClient).Methods("POST")
	r.HandleFunc("/organizations/{organization_name}/oauth-clients", a.listOAuthClients).Methods("GET")
	r.HandleFunc("/oauth-clients/{oauth_client_id}", a.getOAuthClient).Methods("GET")
	r.HandleFunc("/oauth-clients/{oauth_client_id}", a.deleteOAuthClient).Methods("DELETE")
}

func (a *api) createOAuthClient(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}

	var params types.OAuthClientCreateOptions
	if err := unmarshal(r.Body, &params); err != nil {
		Error(w, err)
		return
	}

	// required parameters
	if params.OAuthToken == nil {
		Error(w, &internal.MissingParameterError{Parameter: "oauth-token-string"})
		return
	}
	if params.ServiceProvider == nil {
		Error(w, &internal.MissingParameterError{Parameter: "service-provider"})
		return
	}
	if params.APIURL == nil {
		Error(w, &internal.MissingParameterError{Parameter: "api-url"})
		return
	}
	if params.HTTPURL == nil {
		Error(w, &internal.MissingParameterError{Parameter: "http-url"})
		return
	}

	// unsupported parameters
	if params.PrivateKey != nil {
		Error(w, errors.New("private-key parameter is unsupported"))
		return
	}
	if params.RSAPublicKey != nil {
		Error(w, errors.New("rsa-public-key parameter is unsupported"))
		return
	}
	if params.Secret != nil {
		Error(w, errors.New("secret parameter is unsupported"))
		return
	}
	if *params.ServiceProvider != types.ServiceProviderGithub {
		Error(w, fmt.Errorf("service-provider=%s is unsupported", string(*params.ServiceProvider)))
		return
	}
	if *params.APIURL != GithubAPIURL {
		Error(w, fmt.Errorf("only api-url=%s is supported", GithubAPIURL))
		return
	}
	if *params.HTTPURL != GithubHTTPURL {
		Error(w, fmt.Errorf("only http-url=%s is supported", string(*params.HTTPURL)))
		return
	}

	// default parameters
	if params.Name == nil {
		params.Name = internal.String("")
	}

	oauthClient, err := a.CreateVCSProvider(r.Context(), vcsprovider.CreateOptions{
		Name:         *params.Name,
		Organization: org,
		Token:        *params.OAuthToken,
		Cloud:        "github",
	})
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, oauthClient, withCode(http.StatusCreated))
}

func (a *api) listOAuthClients(w http.ResponseWriter, r *http.Request) {
	org, err := decode.Param("organization_name", r)
	if err != nil {
		Error(w, err)
		return
	}

	oauthClients, err := a.ListVCSProviders(r.Context(), org)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, oauthClients)
}

func (a *api) getOAuthClient(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("oauth_client_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	oauthClient, err := a.GetVCSProvider(r.Context(), id)
	if err != nil {
		Error(w, err)
		return
	}

	a.writeResponse(w, r, oauthClient)
}

func (a *api) deleteOAuthClient(w http.ResponseWriter, r *http.Request) {
	id, err := decode.Param("oauth_client_id", r)
	if err != nil {
		Error(w, err)
		return
	}

	if _, err = a.DeleteVCSProvider(r.Context(), id); err != nil {
		Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
