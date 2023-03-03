package auth

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

type tokenService interface {
	// CreateToken creates a user token.
	CreateToken(ctx context.Context, userID string, opts *otf.TokenCreateOptions) (*otf.Token, error)
	// ListTokens lists API tokens for a user
	ListTokens(ctx context.Context, userID string) ([]*otf.Token, error)
	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, userID string, tokenID string) error
}

// CreateToken creates a user token. Only users can create a user token, and
// they can only create a token for themselves.
func (a *Service) CreateToken(ctx context.Context, userID string, opts *otf.TokenCreateOptions) (*otf.Token, error) {
	subject, err := otf.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if subject.ID != userID {
		return nil, fmt.Errorf("cannot create a token for a different user")
	}

	token, err := otf.NewToken(userID, opts.Description)
	if err != nil {
		a.Error(err, "constructing token", "user", subject)
		return nil, err
	}

	if err := a.db.CreateToken(ctx, token); err != nil {
		a.Error(err, "creating token", "user", subject)
		return nil, err
	}

	a.V(1).Info("created token", "user", subject)

	return token, nil
}

func (a *Service) ListTokens(ctx context.Context, userID string) ([]*otf.Token, error) {
	return a.db.ListTokens(ctx, userID)
}

func (a *Service) DeleteToken(ctx context.Context, userID string, tokenID string) error {
	subject, err := otf.UserFromContext(ctx)
	if err != nil {
		return err
	}
	if subject.ID != userID {
		return fmt.Errorf("cannot delete a token for a different user")
	}

	if err := a.db.DeleteToken(ctx, tokenID); err != nil {
		a.Error(err, "deleting token", "user", subject)
		return err
	}

	a.V(1).Info("deleted token", "username", subject)

	return nil
}
