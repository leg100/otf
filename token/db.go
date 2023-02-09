package token

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type db interface {
	// CreateToken creates a user token.
	CreateToken(ctx context.Context, token *Token) error
	// ListTokens lists user tokens.
	ListTokens(ctx context.Context, userID string) ([]*Token, error)
	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, id string) error
}

// DB is a database of API tokens
type DB struct {
	otf.Database // provides access to generated SQL queries
}

func newPGDB(db otf.Database) *DB {
	return &DB{db}
}

// CreateToken inserts the token, associating it with the user.
func (db *DB) CreateToken(ctx context.Context, token *Token) error {
	_, err := db.InsertToken(ctx, pggen.InsertTokenParams{
		TokenID:     sql.String(token.ID()),
		Token:       sql.String(token.Token()),
		Description: sql.String(token.Description()),
		UserID:      sql.String(token.UserID()),
		CreatedAt:   sql.Timestamptz(token.CreatedAt()),
	})
	return err
}

func (db *DB) ListTokens(ctx context.Context, userID string) ([]*Token, error) {
	result, err := db.FindTokensByUserID(ctx, sql.String(userID))
	if err != nil {
		return nil, err
	}
	var tokens []*Token
	for _, row := range result {
		tokens = append(tokens, unmarshalRow(row))
	}
	return tokens, nil
}

// DeleteToken deletes a user's token from the DB.
func (db *DB) DeleteToken(ctx context.Context, id string) error {
	_, err := db.DeleteTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func unmarshalRow(result pggen.FindTokensByUserIDRow) *Token {
	return &Token{
		id:          result.TokenID.String,
		createdAt:   result.CreatedAt.Time,
		token:       result.Token.String,
		description: result.Description.String,
		userID:      result.UserID.String,
	}
}
