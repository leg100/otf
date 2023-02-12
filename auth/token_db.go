package auth

import (
	"context"

	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type tokenDB interface {
	// CreateToken creates a user token.
	CreateToken(ctx context.Context, token *Token) error
	// ListTokens lists user tokens.
	ListTokens(ctx context.Context, userID string) ([]*Token, error)
	// DeleteToken deletes a user token.
	DeleteToken(ctx context.Context, id string) error
}

// CreateToken inserts the token, associating it with the user.
func (db *pgdb) CreateToken(ctx context.Context, token *Token) error {
	_, err := db.InsertToken(ctx, pggen.InsertTokenParams{
		TokenID:     sql.String(token.ID()),
		Token:       sql.String(token.Token()),
		Description: sql.String(token.Description()),
		UserID:      sql.String(token.UserID()),
		CreatedAt:   sql.Timestamptz(token.CreatedAt()),
	})
	return err
}

func (db *pgdb) ListTokens(ctx context.Context, userID string) ([]*Token, error) {
	result, err := db.FindTokensByUserID(ctx, sql.String(userID))
	if err != nil {
		return nil, err
	}
	var tokens []*Token
	for _, row := range result {
		tokens = append(tokens, &Token{
			id:          row.TokenID.String,
			createdAt:   row.CreatedAt.Time,
			token:       row.Token.String,
			description: row.Description.String,
			userID:      row.UserID.String,
		})
	}
	return tokens, nil
}

// DeleteToken deletes a user's token from the DB.
func (db *pgdb) DeleteToken(ctx context.Context, id string) error {
	_, err := db.DeleteTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
