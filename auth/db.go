package auth

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// pgdb is a registry session database on postgres
type pgdb struct {
	otf.DB // provides access to generated SQL queries
	logr.Logger
}

func newDB(database otf.DB, logger logr.Logger) *pgdb {
	return &pgdb{database, logger}
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(*pgdb) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
		return callback(newDB(tx, db.Logger))
	})
}

// Registry sessions database

func (db *pgdb) createRegistrySession(ctx context.Context, session *registrySession) error {
	_, err := db.InsertRegistrySession(ctx, pggen.InsertRegistrySessionParams{
		Token:            sql.String(session.Token()),
		Expiry:           sql.Timestamptz(session.Expiry()),
		OrganizationName: sql.String(session.Organization()),
	})
	return sql.Error(err)
}

func (db *pgdb) getRegistrySession(ctx context.Context, token string) (*registrySession, error) {
	row, err := db.FindRegistrySession(ctx, sql.String(token))
	if err != nil {
		return nil, sql.Error(err)
	}
	return registrySessionRow(row).toRegistrySession(), nil
}

// CreateAgentToken inserts an agent token, associating it with an organization
func (db *pgdb) CreateAgentToken(ctx context.Context, token *AgentToken) error {
	_, err := db.InsertAgentToken(ctx, pggen.InsertAgentTokenParams{
		TokenID:          sql.String(token.ID()),
		Token:            sql.String(token.Token()),
		Description:      sql.String(token.Description()),
		OrganizationName: sql.String(token.Organization()),
		CreatedAt:        sql.Timestamptz(token.CreatedAt()),
	})
	return err
}

func (db *pgdb) listAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error) {
	rows, err := db.FindAgentTokens(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	var unmarshalled []*AgentToken
	for _, r := range rows {
		unmarshalled = append(unmarshalled, agentTokenRow(r).toAgentToken())
	}
	return unmarshalled, nil
}

// deleteAgentToken deletes an agent token.
func (db *pgdb) deleteAgentToken(ctx context.Context, id string) error {
	_, err := db.DeleteAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) GetAgentTokenByID(ctx context.Context, id string) (*AgentToken, error) {
	r, err := db.FindAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return agentTokenRow(r).toAgentToken(), nil
}

func (db *pgdb) GetAgentTokenByToken(ctx context.Context, token string) (*AgentToken, error) {
	r, err := db.FindAgentTokenByToken(ctx, sql.String(token))
	if err != nil {
		return nil, sql.Error(err)
	}
	return agentTokenRow(r).toAgentToken(), nil
}
