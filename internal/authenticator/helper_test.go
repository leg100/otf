package authenticator

import (
	"context"
	"crypto/rsa"
	"net/http"
	"testing"

	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/tokens"
	"golang.org/x/oauth2"
)

type (
	fakeTokensService struct {
		tokens.TokensService
	}

	fakeOAuthClient struct {
		user  cloud.User
		token *oauth2.Token
	}

	fakeCloudClient struct {
		user cloud.User
		cloud.Client
	}

	fakeTokenHandler struct {
		username string
	}
)

func (f fakeTokenHandler) getUsername(ctx context.Context, token *oauth2.Token) (string, error) {
	return f.username, nil
}

func (fakeTokensService) StartSession(w http.ResponseWriter, r *http.Request, opts tokens.StartSessionOptions) error {
	w.Header().Set("username", *opts.Username)
	return nil
}

func (f *fakeOAuthClient) CallbackHandler(*http.Request) (*oauth2.Token, error) {
	return f.token, nil
}

func (f *fakeOAuthClient) NewClient(context.Context, *oauth2.Token) (cloud.Client, error) {
	return &fakeCloudClient{user: f.user}, nil
}

func (f *fakeCloudClient) GetCurrentUser(context.Context) (cloud.User, error) {
	return f.user, nil
}

func fakeOAuthToken(t *testing.T, username, aud string, key *rsa.PrivateKey) *oauth2.Token {
	idtoken := fakeIDToken(t, username, aud, "", key)
	return (&oauth2.Token{}).WithExtra(map[string]any{"id_token": idtoken})
}
