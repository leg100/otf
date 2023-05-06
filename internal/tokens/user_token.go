package tokens

import (
	"context"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/auth"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

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

	// NewUserTokenOptions are options for constructing a user token via the
	// constructor.
	NewUserTokenOptions struct {
		CreateUserTokenOptions
		Username string
		key      jwk.Key
	}

	userTokenService interface {
		// CreateUserToken creates a user token.
		CreateUserToken(ctx context.Context, opts CreateUserTokenOptions) (*UserToken, []byte, error)
		// ListUserTokens lists API tokens for a user
		ListUserTokens(ctx context.Context) ([]*UserToken, error)
		// DeleteUserToken deletes a user token.
		DeleteUserToken(ctx context.Context, tokenID string) error
	}
)

func NewUserToken(opts NewUserTokenOptions) (*UserToken, []byte, error) {
	ut := UserToken{
		ID:          internal.NewID("ut"),
		CreatedAt:   internal.CurrentTimestamp(),
		Description: opts.Description,
		Username:    opts.Username,
	}
	token, err := newToken(newTokenOptions{
		key:     opts.key,
		subject: ut.ID,
		kind:    userTokenKind,
	})
	if err != nil {
		return nil, nil, err
	}
	return &ut, token, nil
}

// CreateUserToken creates a user token. Only users can create a user token, and
// they can only create a token for themselves.
func (a *service) CreateUserToken(ctx context.Context, opts CreateUserTokenOptions) (*UserToken, []byte, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	ut, token, err := NewUserToken(NewUserTokenOptions{
		CreateUserTokenOptions: opts,
		Username:               user.Username,
		key:                    a.key,
	})
	if err != nil {
		a.Error(err, "constructing token", "user", user)
		return nil, nil, err
	}

	if err := a.db.createUserToken(ctx, ut); err != nil {
		a.Error(err, "creating token", "user", user)
		return nil, nil, err
	}

	a.V(1).Info("created token", "user", user)

	return ut, token, nil
}

func (a *service) ListUserTokens(ctx context.Context) ([]*UserToken, error) {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return a.db.listUserTokens(ctx, user.Username)
}

func (a *service) DeleteUserToken(ctx context.Context, tokenID string) error {
	user, err := auth.UserFromContext(ctx)
	if err != nil {
		return err
	}

	token, err := a.db.getUserToken(ctx, tokenID)
	if err != nil {
		a.Error(err, "retrieving token", "user", user)
		return err
	}

	if user.Username != token.Username {
		return internal.ErrAccessNotPermitted
	}

	if err := a.db.deleteUserToken(ctx, tokenID); err != nil {
		a.Error(err, "deleting token", "user", user)
		return err
	}

	a.V(1).Info("deleted token", "username", user)

	return nil
}
