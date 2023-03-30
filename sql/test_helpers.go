package sql

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-logr/logr"
	"github.com/iancoleman/strcase"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jackc/pgx/v4"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

// NewTestDB creates a logical database in postgres for a test, dropping the
// database upon completion. The db connection is returned along with its
// connection string.
func NewTestDB(t *testing.T) (*DB, string) {
	t.Helper()

	connstr, ok := os.LookupEnv("OTF_TEST_DATABASE_URL")
	if !ok {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	// connect and create database
	conn, err := pgx.Connect(ctx, connstr)
	require.NoError(t, err)

	// generate a safe, unique logical database name
	logical := strcase.ToSnake(t.Name())
	logical = strings.ReplaceAll(logical, "/", "_")
	logical = logical + "_" + strings.ToLower(otf.GenerateRandomString(8))

	_, err = conn.Exec(ctx, "CREATE DATABASE "+logical)
	require.NoError(t, err, "unable to create database")
	t.Cleanup(func() {
		_, err := conn.Exec(ctx, "DROP DATABASE "+logical)
		assert.NoError(t, err, "unable to drop database %s")
		err = conn.Close(ctx)
		assert.NoError(t, err, "unable to close connection")
	})

	// modify connection string to use new logical database
	u, err := url.Parse(connstr)
	require.NoError(t, err)
	u.Path = "/" + logical
	logicalConnString := u.String()

	// create otf db in new logical database, establish conn pool, etc.
	db, err := New(ctx, Options{
		Logger:     logr.Discard(),
		ConnString: logicalConnString,
	})
	require.NoError(t, err)
	t.Cleanup(db.Close)

	return db, logicalConnString
}
