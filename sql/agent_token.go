package sql

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateAgentToken inserts an agent token, associating it with an organization
func (db *DB) CreateAgentToken(ctx context.Context, token *otf.Token) error {
	_, err := db.InsertToken(ctx, pggen.InsertTokenParams{
		TokenID:     String(token.ID()),
		Token:       String(token.Token()),
		Description: String(token.Description()),
		UserID:      String(token.UserID()),
		CreatedAt:   Timestamptz(token.CreatedAt()),
	})
	return err
}

func (db *DB) GetAgentToken(ctx context.Context, token string) (*otf.AgentToken, error) {
	r, err := db.FindAgentToken(ctx, String(token))
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalAgentTokenDBResult(r)
}

// DeleteAgentToken deletes an agent token.
func (db *DB) DeleteAgentToken(ctx context.Context, id string) error {
	_, err := db.DeleteTokenByID(ctx, String(id))
	if err != nil {
		return databaseError(err)
	}
	return nil
}
