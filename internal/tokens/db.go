package tokens

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb stores authentication resources in a postgres database
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
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
	_, err := db.Conn(ctx).InsertToken(ctx, pggen.InsertTokenParams{
		TokenID:     sql.String(token.ID),
		Description: sql.String(token.Description),
		Username:    sql.String(token.Username),
		CreatedAt:   sql.Timestamptz(token.CreatedAt),
	})
	return err
}

func (db *pgdb) listUserTokens(ctx context.Context, username string) ([]*UserToken, error) {
	result, err := db.Conn(ctx).FindTokensByUsername(ctx, sql.String(username))
	if err != nil {
		return nil, err
	}
	var tokens []*UserToken
	for _, row := range result {
		tokens = append(tokens, &UserToken{
			ID:          row.TokenID.String,
			CreatedAt:   row.CreatedAt.Time.UTC(),
			Description: row.Description.String,
			Username:    row.Username.String,
		})
	}
	return tokens, nil
}

func (db *pgdb) getUserToken(ctx context.Context, id string) (*UserToken, error) {
	row, err := db.Conn(ctx).FindTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return &UserToken{
		ID:          row.TokenID.String,
		CreatedAt:   row.CreatedAt.Time.UTC(),
		Description: row.Description.String,
		Username:    row.Username.String,
	}, nil
}

func (db *pgdb) deleteUserToken(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

//
// Team tokens
//

func (db *pgdb) createTeamToken(ctx context.Context, token *TeamToken) error {
	_, err := db.Conn(ctx).InsertTeamToken(ctx, pggen.InsertTeamTokenParams{
		TeamTokenID: sql.String(token.ID),
		Description: sql.String(token.Description),
		TeamID:      sql.String(token.Team),
		CreatedAt:   sql.Timestamptz(token.CreatedAt),
		Expiry:      sql.TimestamptzPtr(token.Expiry),
	})
	return err
}

//
// Organization tokens
//

func (db *pgdb) upsertOrganizationToken(ctx context.Context, token *OrganizationToken) error {
	_, err := db.Conn(ctx).UpsertOrganizationToken(ctx, pggen.UpsertOrganizationTokenParams{
		OrganizationTokenID: sql.String(token.ID),
		OrganizationName:    sql.String(token.Organization),
		CreatedAt:           sql.Timestamptz(token.CreatedAt),
		Expiry:              sql.TimestamptzPtr(token.Expiry),
	})
	return err
}

func (db *pgdb) getOrganizationTokenByName(ctx context.Context, organization string) (*OrganizationToken, error) {
	// query only returns 0 or 1 tokens
	result, err := db.Conn(ctx).FindOrganizationTokensByName(ctx, sql.String(organization))
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	ot := &OrganizationToken{
		ID:           result[0].OrganizationTokenID.String,
		CreatedAt:    result[0].CreatedAt.Time.UTC(),
		Organization: result[0].OrganizationName.String,
	}
	if result[0].Expiry.Status == pgtype.Present {
		ot.Expiry = internal.Time(result[0].Expiry.Time.UTC())
	}
	return ot, nil
}

func (db *pgdb) getOrganizationTokenByID(ctx context.Context, tokenID string) (*OrganizationToken, error) {
	result, err := db.Conn(ctx).FindOrganizationTokensByID(ctx, sql.String(tokenID))
	if err != nil {
		return nil, sql.Error(err)
	}
	ot := &OrganizationToken{
		ID:           result.OrganizationTokenID.String,
		CreatedAt:    result.CreatedAt.Time.UTC(),
		Organization: result.OrganizationName.String,
	}
	if result.Expiry.Status == pgtype.Present {
		ot.Expiry = internal.Time(result.Expiry.Time.UTC())
	}
	return ot, nil
}

func (db *pgdb) deleteOrganizationToken(ctx context.Context, organization string) error {
	_, err := db.Conn(ctx).DeleteOrganiationTokenByName(ctx, sql.String(organization))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

//
// Agent tokens
//

func (db *pgdb) createAgentToken(ctx context.Context, token *AgentToken) error {
	_, err := db.Conn(ctx).InsertAgentToken(ctx, pggen.InsertAgentTokenParams{
		TokenID:          sql.String(token.ID),
		Description:      sql.String(token.Description),
		OrganizationName: sql.String(token.Organization),
		CreatedAt:        sql.Timestamptz(token.CreatedAt.UTC()),
	})
	return err
}

func (db *pgdb) getAgentTokenByID(ctx context.Context, id string) (*AgentToken, error) {
	r, err := db.Conn(ctx).FindAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return agentTokenRow(r).toAgentToken(), nil
}

func (db *pgdb) listAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error) {
	rows, err := db.Conn(ctx).FindAgentTokens(ctx, sql.String(organization))
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
	_, err := db.Conn(ctx).DeleteAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (row agentTokenRow) toAgentToken() *AgentToken {
	return &AgentToken{
		ID:           row.TokenID.String,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		Description:  row.Description.String,
		Organization: row.OrganizationName.String,
	}
}
