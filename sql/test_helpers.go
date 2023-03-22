package sql

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v4"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

func NewTestDB(t *testing.T) *DB {
	urlStr := os.Getenv(TestDatabaseURL)
	if urlStr == "" {
		t.Fatalf("%s must be set", TestDatabaseURL)
	}

	u, err := url.Parse(urlStr)
	require.NoError(t, err)

	require.Equal(t, "postgres", u.Scheme)

	opts := Options{
		Logger:     logr.Discard(),
		ConnString: u.String(),
	}

	db, err := New(context.Background(), opts)
	require.NoError(t, err)

	t.Cleanup(db.Close)

	return db
}
