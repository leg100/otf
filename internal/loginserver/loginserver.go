// Package loginserver implements a "terraform login protocol" server:
//
// https://developer.hashicorp.com/terraform/internals/v1.3.x/login-protocol#client
package loginserver

import (
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/tokens"
)

const (
	// OAuth2 client ID - purely advisory according to:
	// https://developer.hashicorp.com/terraform/internals/v1.3.x/login-protocol#client
	ClientID = "terraform"

	AuthRoute  = "/app/oauth2/auth"
	TokenRoute = "/oauth2/token"
)

var Discovery = DiscoverySpec{
	Client:     ClientID,
	GrantTypes: []string{"authz_code"},
	Authz:      AuthRoute,
	Token:      TokenRoute,
	Ports:      []int{10000, 10010},
}

type (
	server struct {
		secret []byte // for encrypting auth code

		html.Renderer        // render consent UI
		tokens.TokensService // for creating user API token
	}

	// Options for server constructor
	Options struct {
		Secret []byte // for encrypting auth code

		html.Renderer
		tokens.TokensService
	}

	authcode struct {
		CodeChallenge       string `json:"code_challenge"`
		CodeChallengeMethod string `json:"code_challenge_method"`
		Username            string `json:"username"`
	}

	DiscoverySpec struct {
		Client     string   `json:"client"`
		GrantTypes []string `json:"grant_types"`
		Authz      string   `json:"authz"`
		Token      string   `json:"token"`
		Ports      []int    `json:"ports"`
	}
)

func NewServer(opts Options) (*server, error) {
	return &server{
		secret:        opts.Secret,
		Renderer:      opts.Renderer,
		TokensService: opts.TokensService,
	}, nil
}

func (s *server) AddHandlers(r *mux.Router) {
	// authenticated
	r.HandleFunc(AuthRoute, s.authHandler).Methods("GET", "POST")
	// unauthenticated
	r.HandleFunc(TokenRoute, s.tokenHandler).Methods("POST")
}
