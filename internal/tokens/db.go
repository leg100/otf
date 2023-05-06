package tokens

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

type (
	// pgdb stores authentication resources in a postgres database
	pgdb struct {
		otf.DB // provides access to generated SQL queries
	}

	agentTokenRow struct {
		TokenID          pgtype.Text        `json:"token_id"`
		CreatedAt        pgtype.Timestamptz `json:"created_at"`
		Description      pgtype.Text        `json:"description"`
		OrganizationName pgtype.Text        `json:"organization_name"`
	}
)

//
// User tokens
//

func (db *pgdb) createUserToken(ctx context.Context, token *UserToken) error {
	_, err := db.InsertToken(ctx, pggen.InsertTokenParams{
		TokenID:     sql.String(token.ID),
		Description: sql.String(token.Description),
		Username:    sql.String(token.Username),
		CreatedAt:   sql.Timestamptz(token.CreatedAt),
	})
	return err
}

func (db *pgdb) listUserTokens(ctx context.Context, username string) ([]*UserToken, error) {
	result, err := db.FindTokensByUsername(ctx, sql.String(username))
	if err != nil {
		return nil, err
	}
	var tokens []*UserToken
	for _, row := range result {
		tokens = append(tokens, &UserToken{
			ID:          row.TokenID.String,
			CreatedAt:   row.CreatedAt.Time,
			Description: row.Description.String,
			Username:    row.Username.String,
		})
	}
	return tokens, nil
}

func (db *pgdb) getUserToken(ctx context.Context, id string) (*UserToken, error) {
	row, err := db.FindTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return &UserToken{
		ID:          row.TokenID.String,
		CreatedAt:   row.CreatedAt.Time,
		Description: row.Description.String,
		Username:    row.Username.String,
	}, nil
}

func (db *pgdb) deleteUserToken(ctx context.Context, id string) error {
	_, err := db.DeleteTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

//
// Agent tokens
//

func (db *pgdb) createAgentToken(ctx context.Context, token *AgentToken) error {
	_, err := db.InsertAgentToken(ctx, pggen.InsertAgentTokenParams{
		TokenID:          sql.String(token.ID),
		Description:      sql.String(token.Description),
		OrganizationName: sql.String(token.Organization),
		CreatedAt:        sql.Timestamptz(token.CreatedAt),
	})
	return err
}

func (db *pgdb) getAgentTokenByID(ctx context.Context, id string) (*AgentToken, error) {
	r, err := db.FindAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return agentTokenRow(r).toAgentToken(), nil
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

func (db *pgdb) deleteAgentToken(ctx context.Context, id string) error {
	_, err := db.DeleteAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (row agentTokenRow) toAgentToken() *AgentToken {
	return &AgentToken{
		ID:           row.TokenID.String,
		CreatedAt:    row.CreatedAt.Time,
		Description:  row.Description.String,
		Organization: row.OrganizationName.String,
	}
}
