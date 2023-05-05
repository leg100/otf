package authenticator

import (
	"context"
	"crypto"
	"crypto/rsa"
	"net/http"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/tokens"
	"golang.org/x/oauth2"
)

type (
	fakeAuthenticatorService struct {
		tokens.TokensService
	}

	fakeOAuthClient struct {
		user *cloud.User
		oauthClient
		token *oauth2.Token
	}

	fakeCloudClient struct {
		user *cloud.User
		cloud.Client
	}
)

func (f *fakeAuthenticatorService) StartSession(w http.ResponseWriter, r *http.Request, opts tokens.StartSessionOptions) error {
	http.Redirect(w, r, paths.Profile(), http.StatusFound)
	return nil
}

func (f *fakeOAuthClient) CallbackHandler(*http.Request) (*oauth2.Token, error) {
	return f.token, nil
}

func (f *fakeOAuthClient) NewClient(context.Context, *oauth2.Token) (cloud.Client, error) {
	return &fakeCloudClient{user: f.user}, nil
}

func (f *fakeCloudClient) GetUser(context.Context) (*cloud.User, error) {
	return f.user, nil
}

func fakeOAuthToken(t *testing.T, username, aud string, key *rsa.PrivateKey) *oauth2.Token {
	idtoken := fakeIDToken(t, username, aud, "", key)
	return (&oauth2.Token{}).WithExtra(map[string]any{"id_token": idtoken})
}

func fakeVerifier(t *testing.T, aud string, key *rsa.PrivateKey) *oidc.IDTokenVerifier {
	keySet := &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{key.Public()}}
	return oidc.NewVerifier("", keySet, &oidc.Config{
		ClientID: "otf",
	})
}
