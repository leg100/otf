package sql

import (
	"github.com/alexedwards/scs/postgresstore"
	"github.com/jmoiron/sqlx"
)

type SessionDB struct {
	*postgresstore.PostgresStore
}

func NewSessionDB(db *sqlx.DB) *SessionDB {
	return &SessionDB{
		PostgresStore: postgresstore.New(db.DB),
	}
}
