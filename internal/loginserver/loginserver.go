// Package loginserver implements a "terraform login protocol" server:
//
// https://developer.hashicorp.com/terraform/internals/v1.3.x/login-protocol#client
package loginserver

import (
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal/http/html"
	"github.com/lestrrat-go/jwx/v2/jwk"
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
		key    jwk.Key // for signing access token
		secret string  // for encrypting auth code

		html.Renderer // render consent UI
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

func NewServer(secret string) (*server, error) {
	key, err := jwk.FromRaw([]byte(secret))
	if err != nil {
		return nil, err
	}
	return &server{key: key, secret: secret}, nil
}

func (s *server) AddHandlers(r *mux.Router) {
	// authenticated
	r.HandleFunc(AuthRoute, s.authHandler).Methods("GET", "POST")
	// unauthenticated
	r.HandleFunc(TokenRoute, s.tokenHandler).Methods("POST")
}
