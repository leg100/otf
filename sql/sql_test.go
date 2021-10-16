package sql

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPG(t *testing.T) {
	db, err := sqlx.Connect("postgres", "postgres:///postgres?host=/var/run/postgresql")
	require.NoError(t, err)

	dbname := "db_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	t.Log("dbname: ", dbname)

	_, err = db.Exec(fmt.Sprint("CREATE DATABASE ", dbname))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err = db.Exec(fmt.Sprint("DROP DATABASE ", dbname))
		require.NoError(t, err)
	})

	//_, err = db.Exec(testCreateOrganizations) assert.NoError(t, err)

	connStr := fmt.Sprintf("postgres:///%s?host=/var/run/postgresql", dbname)

	newdb, err := New(logr.Discard(), connStr)
	assert.NoError(t, err)

	t.Cleanup(func() {
		err := newdb.Close()
		assert.NoError(t, err)
	})
}
