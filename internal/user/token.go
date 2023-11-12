package user

import (
	"context"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/tokens"
)

const UserTokenKind tokens.Kind = "user_token"

type (
	// UserToken provides information about an API token for a user.
	UserToken struct {
		ID          string
		CreatedAt   time.Time
		Description string
		Username    string // Token belongs to a user
	}

	// CreateUserTokenOptions are options for creating a user token via the service
	// endpoint
	CreateUserTokenOptions struct {
		Description string
	}

	userTokenService interface {
		// CreateUserToken creates a user token.
		CreateUserToken(ctx context.Context, opts CreateUserTokenOptions) (*UserToken, []byte, error)
		// ListUserTokens lists API tokens for a user
		ListUserTokens(ctx context.Context) ([]*UserToken, error)
		// DeleteUserToken deletes a user token.
		DeleteUserToken(ctx context.Context, tokenID string) error
	}

	userTokenFactory struct {
		tokens.TokensService
	}
)

func (f *userTokenFactory) NewUserToken(username string, opts CreateUserTokenOptions) (*UserToken, []byte, error) {
	ut := UserToken{
		ID:          internal.NewID("ut"),
		CreatedAt:   internal.CurrentTimestamp(nil),
		Description: opts.Description,
		Username:    username,
	}
	token, err := f.NewToken(tokens.NewTokenOptions{
		Subject: ut.ID,
		Kind:    UserTokenKind,
	})
	if err != nil {
		return nil, nil, err
	}
	return &ut, token, nil
}
