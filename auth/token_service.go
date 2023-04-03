package auth

import (
	"context"

	"github.com/leg100/otf"
)

type tokenService interface {
	// CreateToken creates a user token.
	CreateToken(ctx context.Context, opts *TokenCreateOptions) (*Token, error)
	// ListTokens lists API tokens for a user
	ListTokens(ctx context.Context) ([]*Token, error)
	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, tokenID string) error
}

// CreateToken creates a user token. Only users can create a user token, and
// they can only create a token for themselves.
func (a *service) CreateToken(ctx context.Context, opts *TokenCreateOptions) (*Token, error) {
	user, err := UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	token, err := NewToken(user.Username, opts.Description)
	if err != nil {
		a.Error(err, "constructing token", "user", user)
		return nil, err
	}

	if err := a.db.CreateToken(ctx, token); err != nil {
		a.Error(err, "creating token", "user", user)
		return nil, err
	}

	a.V(1).Info("created token", "user", user)

	return token, nil
}

func (a *service) ListTokens(ctx context.Context) ([]*Token, error) {
	user, err := UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return a.db.ListTokens(ctx, user.Username)
}

func (a *service) DeleteToken(ctx context.Context, tokenID string) error {
	user, err := UserFromContext(ctx)
	if err != nil {
		return err
	}

	token, err := a.db.GetToken(ctx, tokenID)
	if err != nil {
		a.Error(err, "retrieving token", "user", user)
		return err
	}

	if user.Username != token.Username {
		return otf.ErrAccessNotPermitted
	}

	if err := a.db.DeleteToken(ctx, tokenID); err != nil {
		a.Error(err, "deleting token", "user", user)
		return err
	}

	a.V(1).Info("deleted token", "username", user)

	return nil
}
