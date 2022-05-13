package sql

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var (
	_ otf.TokenStore = (*TokenDB)(nil)
)

type TokenDB struct {
	*pgx.Conn
}

func NewTokenDB(conn *pgx.Conn) *TokenDB {
	return &TokenDB{
		Conn: conn,
	}
}

// CreateToken inserts the token, associating it with the user.
func (db TokenDB) CreateToken(ctx context.Context, token *otf.Token) error {
	q := NewQuerier(db.Conn)

	result, err := q.InsertToken(ctx, InsertTokenParams{
		TokenID:     token.ID,
		Token:       token.Token,
		Description: token.Description,
		UserID:      token.UserID,
	})
	if err != nil {
		return err
	}
	token.CreatedAt = result.CreatedAt
	token.UpdatedAt = result.UpdatedAt

	return nil
}

// DeleteToken deletes a user's token from the DB.
func (db TokenDB) DeleteToken(ctx context.Context, id string) error {
	q := NewQuerier(db.Conn)

	result, err := q.DeleteTokenByID(ctx, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
