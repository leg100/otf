package sql

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateToken inserts the token, associating it with the user.
func (db *DB) CreateToken(ctx context.Context, token *otf.Token) error {
	_, err := db.InsertToken(ctx, pggen.InsertTokenParams{
		TokenID:     pgtype.Text{String: token.ID(), Status: pgtype.Present},
		Token:       pgtype.Text{String: token.Token(), Status: pgtype.Present},
		Description: pgtype.Text{String: token.Description(), Status: pgtype.Present},
		UserID:      pgtype.Text{String: token.UserID(), Status: pgtype.Present},
		CreatedAt:   token.CreatedAt(),
	})
	return err
}

// DeleteToken deletes a user's token from the DB.
func (db *DB) DeleteToken(ctx context.Context, id string) error {
	result, err := db.DeleteTokenByID(ctx, pgtype.Text{String: id, Status: pgtype.Present})
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}
	return nil
}
