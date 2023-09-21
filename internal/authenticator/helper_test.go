package authenticator

import (
	"context"
	"net/http"

	"github.com/leg100/otf/internal/tokens"
	"golang.org/x/oauth2"
)

type (
	fakeTokensService struct {
		tokens.TokensService
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
