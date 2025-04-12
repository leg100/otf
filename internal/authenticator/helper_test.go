package authenticator

import (
	"context"
	"net/http"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/user"
	"golang.org/x/oauth2"
)

type fakeTokenHandler struct{}

func (f fakeTokenHandler) getUsername(ctx context.Context, token *oauth2.Token) (string, error) {
	return "", nil
}

type fakeTokensService struct{}

func (*fakeTokensService) StartSession(w http.ResponseWriter, r *http.Request, userID resource.TfeID) error {
	w.Header().Set("user-id", userID.String())
	return nil
}

type fakeUserService struct {
	userID resource.TfeID
}

func (f *fakeUserService) GetUser(ctx context.Context, spec user.UserSpec) (*user.User, error) {
	return &user.User{ID: f.userID}, nil
}

func (f *fakeUserService) Create(ctx context.Context, username user.Username, opts ...user.NewUserOption) (*user.User, error) {
	return nil, nil
}
