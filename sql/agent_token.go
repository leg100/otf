package sql

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateAgentToken inserts an agent token, associating it with an organization
func (db *DB) CreateAgentToken(ctx context.Context, token *otf.AgentToken) error {
	_, err := db.InsertAgentToken(ctx, pggen.InsertAgentTokenParams{
		TokenID:          String(token.ID()),
		Token:            String(*token.Token()),
		Description:      String(token.Description()),
		OrganizationName: String(token.Organization()),
		CreatedAt:        Timestamptz(token.CreatedAt()),
	})
	return err
}

func (db *DB) ListAgentTokens(ctx context.Context, organizationName string) ([]*otf.AgentToken, error) {
	rows, err := db.FindAgentTokens(ctx, String(organizationName))
	if err != nil {
		return nil, Error(err)
	}
	var unmarshalled []*otf.AgentToken
	for _, r := range rows {
		unmarshalled = append(unmarshalled, otf.UnmarshalAgentTokenResult(otf.AgentTokenRow(r)))
	}
	return unmarshalled, nil
}

func (db *DB) GetAgentTokenByID(ctx context.Context, id string) (*otf.AgentToken, error) {
	r, err := db.FindAgentTokenByID(ctx, String(id))
	if err != nil {
		return nil, Error(err)
	}
	return otf.UnmarshalAgentTokenResult(otf.AgentTokenRow(r)), nil
}

func (db *DB) GetAgentTokenByToken(ctx context.Context, token string) (*otf.AgentToken, error) {
	r, err := db.FindAgentTokenByToken(ctx, String(token))
	if err != nil {
		return nil, Error(err)
	}
	return otf.UnmarshalAgentTokenResult(otf.AgentTokenRow(r)), nil
}

// DeleteAgentToken deletes an agent token.
func (db *DB) DeleteAgentToken(ctx context.Context, id string) error {
	_, err := db.DeleteAgentTokenByID(ctx, String(id))
	if err != nil {
		return Error(err)
	}
	return nil
}
