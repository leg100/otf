package sql

import (
	"context"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) CreateRegistrySession(ctx context.Context, session *otf.RegistrySession) error {
	_, err := db.InsertRegistrySession(ctx, pggen.InsertRegistrySessionParams{
		Token:            String(session.Token()),
		Expiry:           Timestamptz(session.Expiry()),
		OrganizationName: String(session.Organization()),
	})
	return databaseError(err)
}

func (db *DB) GetRegistrySession(ctx context.Context, token string) (*otf.RegistrySession, error) {
	row, err := db.FindRegistrySession(ctx, String(token))
	if err != nil {
		return nil, databaseError(err)
	}
	return otf.UnmarshalRegistrySessionRow(otf.RegistrySessionRow(row)), nil
}

func (db *DB) startRegistrySessionExpirer(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			// TODO: log errors
			_, _ = db.DeleteExpiredRegistrySessions(ctx)
		case <-ctx.Done():
			return
		}
	}
}
