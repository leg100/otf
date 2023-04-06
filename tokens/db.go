package tokens

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type (
	// pgdb stores authentication resources in a postgres database
	pgdb struct {
		otf.DB // provides access to generated SQL queries
		logr.Logger
	}
	agentTokenRow struct {
		TokenID          pgtype.Text        `json:"token_id"`
		CreatedAt        pgtype.Timestamptz `json:"created_at"`
		Description      pgtype.Text        `json:"description"`
		OrganizationName pgtype.Text        `json:"organization_name"`
	}
)

func newDB(database otf.DB, logger logr.Logger) *pgdb {
	return &pgdb{database, logger}
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		return callback(newDB(tx, db.Logger))
	})
}

func (row agentTokenRow) toAgentToken() *AgentToken {
	return &AgentToken{
		ID:           row.TokenID.String,
		CreatedAt:    row.CreatedAt.Time,
		Description:  row.Description.String,
		Organization: row.OrganizationName.String,
	}
}

// CreateToken inserts the token, associating it with the user.
func (db *pgdb) CreateToken(ctx context.Context, token *Token) error {
	_, err := db.InsertToken(ctx, pggen.InsertTokenParams{
		TokenID:     sql.String(token.ID),
		Description: sql.String(token.Description),
		Username:    sql.String(token.Username),
		CreatedAt:   sql.Timestamptz(token.CreatedAt),
	})
	return err
}

func (db *pgdb) ListTokens(ctx context.Context, username string) ([]*Token, error) {
	result, err := db.FindTokensByUsername(ctx, sql.String(username))
	if err != nil {
		return nil, err
	}
	var tokens []*Token
	for _, row := range result {
		tokens = append(tokens, &Token{
			ID:          row.TokenID.String,
			CreatedAt:   row.CreatedAt.Time,
			Description: row.Description.String,
			Username:    row.Username.String,
		})
	}
	return tokens, nil
}

func (db *pgdb) GetToken(ctx context.Context, id string) (*Token, error) {
	row, err := db.FindTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return &Token{
		ID:          row.TokenID.String,
		CreatedAt:   row.CreatedAt.Time,
		Description: row.Description.String,
		Username:    row.Username.String,
	}, nil
}

// DeleteToken deletes a user's token from the DB.
func (db *pgdb) DeleteToken(ctx context.Context, id string) error {
	_, err := db.DeleteTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
