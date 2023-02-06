package agenttoken

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is a state/state-version database on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func newPGDB(db otf.Database) *pgdb {
	return &pgdb{db}
}

// CreateAgentToken inserts an agent token, associating it with an organization
func (db *pgdb) CreateAgentToken(ctx context.Context, token *AgentToken) error {
	_, err := db.InsertAgentToken(ctx, pggen.InsertAgentTokenParams{
		TokenID:          sql.String(token.ID()),
		Token:            sql.String(*token.Token()),
		Description:      sql.String(token.Description()),
		OrganizationName: sql.String(token.Organization()),
		CreatedAt:        sql.Timestamptz(token.CreatedAt()),
	})
	return err
}

func (db *pgdb) ListAgentTokens(ctx context.Context, organizationName string) ([]*AgentToken, error) {
	rows, err := db.FindAgentTokens(ctx, sql.String(organizationName))
	if err != nil {
		return nil, sql.Error(err)
	}
	var unmarshalled []*AgentToken
	for _, r := range rows {
		unmarshalled = append(unmarshalled, UnmarshalAgentTokenResult(AgentTokenRow(r)))
	}
	return unmarshalled, nil
}

func (db *pgdb) GetAgentTokenByID(ctx context.Context, id string) (*AgentToken, error) {
	r, err := db.FindAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalAgentTokenResult(AgentTokenRow(r)), nil
}

func (db *pgdb) GetAgentTokenByToken(ctx context.Context, token string) (*AgentToken, error) {
	r, err := db.FindAgentTokenByToken(ctx, sql.String(token))
	if err != nil {
		return nil, sql.Error(err)
	}
	return UnmarshalAgentTokenResult(AgentTokenRow(r)), nil
}

// DeleteAgentToken deletes an agent token.
func (db *pgdb) DeleteAgentToken(ctx context.Context, id string) error {
	_, err := db.DeleteAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
