package auth

import (
	"context"

	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

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
