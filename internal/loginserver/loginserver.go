// Package loginserver implements a "terraform login protocol" server:
//
// https://developer.hashicorp.com/terraform/internals/v1.3.x/login-protocol#client
package loginserver

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/user"
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
		secret []byte           // for encrypting auth code
		users  userTokensClient // for creating user API token

		html.Renderer // render consent UI
	}

	userTokensClient interface {
		CreateUserToken(ctx context.Context, opts user.CreateUserTokenOptions) (*user.UserToken, []byte, error)
	}

	// Options for server constructor
	Options struct {
		Secret      []byte // for encrypting auth code
		UserService *user.Service

		html.Renderer
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

func NewServer(opts Options) *server {
	return &server{
		Renderer: opts.Renderer,
		secret:   opts.Secret,
		users:    opts.UserService,
	}
}

func (s *server) AddHandlers(r *mux.Router) {
	// authenticated
	r.HandleFunc(AuthRoute, s.authHandler).Methods("GET", "POST")
	// unauthenticated
	r.HandleFunc(TokenRoute, s.tokenHandler).Methods("POST")
}
