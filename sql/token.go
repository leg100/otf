package sql

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateToken inserts the token, associating it with the user.
func (db *DB) CreateToken(ctx context.Context, token *otf.Token) error {
	_, err := db.InsertToken(ctx, pggen.InsertTokenParams{
		TokenID:     String(token.ID()),
		Token:       String(token.Token()),
		Description: String(token.Description()),
		UserID:      String(token.UserID()),
		CreatedAt:   Timestamptz(token.CreatedAt()),
	})
	return err
}

// DeleteToken deletes a user's token from the DB.
func (db *DB) DeleteToken(ctx context.Context, id string) error {
	_, err := db.DeleteTokenByID(ctx, String(id))
	if err != nil {
		return databaseError(err)
	}
	return nil
}
