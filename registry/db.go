package registry

import (
	"context"
	"time"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// db is a database of registry sessions
type db interface {
	otf.Database

	create(context.Context, *Session) error
	get(ctx context.Context, token string) (*Session, error)
}

// pgdb is a registry session database on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func newDB(ctx context.Context, database otf.Database, cleanupInterval time.Duration) *pgdb {
	db := &pgdb{database}

	if cleanupInterval == 0 {
		cleanupInterval = defaultExpiry
	}
	// purge expired registry sessions
	go db.startExpirer(ctx, cleanupInterval)

	return db
}

func (db *pgdb) create(ctx context.Context, session *Session) error {
	_, err := db.InsertRegistrySession(ctx, pggen.InsertRegistrySessionParams{
		Token:            sql.String(session.Token()),
		Expiry:           sql.Timestamptz(session.Expiry()),
		OrganizationName: sql.String(session.Organization()),
	})
	return sql.Error(err)
}

func (db *pgdb) get(ctx context.Context, token string) (*Session, error) {
	row, err := db.FindRegistrySession(ctx, sql.String(token))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(row).toSession(), nil
}

func (db *pgdb) startExpirer(ctx context.Context, interval time.Duration) {
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

type pgRow struct {
	Token            pgtype.Text        `json:"token"`
	Expiry           pgtype.Timestamptz `json:"expiry"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

func (result pgRow) toSession() *Session {
	return &Session{
		token:        result.Token.String,
		expiry:       result.Expiry.Time.UTC(),
		organization: result.OrganizationName.String,
	}
}
