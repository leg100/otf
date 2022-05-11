package otf

import "context"

// TokenStore is a persistence store for user authentication tokens.
type TokenStore interface {
	// CreateToken creates a user token.
	CreateToken(ctx context.Context, token *Token) error

	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, id string) error
}
