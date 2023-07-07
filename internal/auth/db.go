package auth

import (
	"github.com/go-logr/logr"
	"github.com/leg100/otf/internal/sql"
)

// pgdb stores authentication resources in a postgres database
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
	logr.Logger
}

func newDB(database *sql.DB, logger logr.Logger) *pgdb {
	return &pgdb{database, logger}
}
