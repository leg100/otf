package sql

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/inmem"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v4"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

func NewTestDB(t *testing.T, overrides ...newTestDBOption) *DB {
	urlStr := os.Getenv(TestDatabaseURL)
	if urlStr == "" {
		t.Fatalf("%s must be set", TestDatabaseURL)
	}

	u, err := url.Parse(urlStr)
	require.NoError(t, err)

	require.Equal(t, "postgres", u.Scheme)

	opts := Options{
		Logger:       logr.Discard(),
		ConnString:   u.String(),
		Cache:        nil,
		CloudService: inmem.NewTestCloudService(),
	}

	for _, or := range overrides {
		or(&opts)
	}

	db, err := New(context.Background(), opts)
	require.NoError(t, err)

	t.Cleanup(func() { db.Close() })

	return db
}

type newTestDBOption func(*Options)

func overrideCleanupInterval(d time.Duration) newTestDBOption {
	return func(o *Options) {
		o.CleanupInterval = d
	}
}
