package sql

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

var (
	_ otf.TokenStore = (*TokenDB)(nil)
)

type TokenDB struct {
	*pgxpool.Pool
}

func NewTokenDB(conn *pgxpool.Pool) *TokenDB {
	return &TokenDB{
		Pool: conn,
	}
}

// CreateToken inserts the token, associating it with the user.
func (db TokenDB) CreateToken(ctx context.Context, token *otf.Token) error {
	q := pggen.NewQuerier(db.Pool)

	_, err := q.InsertToken(ctx, pggen.InsertTokenParams{
		TokenID:     token.ID(),
		Token:       token.Token(),
		Description: token.Description(),
		UserID:      token.UserID(),
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteToken deletes a user's token from the DB.
func (db TokenDB) DeleteToken(ctx context.Context, id string) error {
	q := pggen.NewQuerier(db.Pool)

	result, err := q.DeleteTokenByID(ctx, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
