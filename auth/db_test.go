package auth

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/sql"
)

func newTestDB(t *testing.T) *pgdb {
	return newDB(sql.NewTestDB(t), logr.Discard())
}
