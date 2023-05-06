package sql

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/iancoleman/strcase"
	internal "github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jackc/pgx/v4"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

// NewTestDB creates a logical database in postgres for a test and returns a
// connection string for connecting to the database. The database is dropped
// upon test completion.
func NewTestDB(t *testing.T) string {
	t.Helper()

	connstr, ok := os.LookupEnv(TestDatabaseURL)
	if !ok {
		t.Skip("Export valid OTF_TEST_DATABASE_URL before running this test")
	}

	ctx := context.Background()

	// connect and create database
	conn, err := pgx.Connect(ctx, connstr)
	require.NoError(t, err)

	// generate a safe, unique logical database name
	logical := t.Name()
	logical = strcase.ToSnake(logical)
	logical = strings.ReplaceAll(logical, "/", "_")
	// NOTE: maximum size of a postgres name is 31
	// 21 + "_" + 8 = 30
	if len(logical) > 22 {
		logical = logical[:22]
	}
	logical = logical + "_" + strings.ToLower(internal.GenerateRandomString(8))

	_, err = conn.Exec(ctx, "CREATE DATABASE "+logical)
	require.NoError(t, err, "unable to create database")
	t.Cleanup(func() {
		_, err := conn.Exec(ctx, "DROP DATABASE "+logical)
		assert.NoError(t, err, "unable to drop database %s", logical)
		err = conn.Close(ctx)
		assert.NoError(t, err, "unable to close connection")
	})

	// modify connection string to use new logical database
	u, err := url.Parse(connstr)
	require.NoError(t, err)
	u.Path = "/" + logical

	return u.String()
}
