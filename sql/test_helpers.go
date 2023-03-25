package sql

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v4"
)

const TestDatabaseURL = "OTF_TEST_DATABASE_URL"

func NewContainer() (*DB, *postgres.PostgresContainer, error) {
	ctx := context.Background()

	container, err := postgres.StartContainer(ctx,
		postgres.WithImage("postgres:14-alpine"),
		postgres.WithDatabase("otf"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, nil, err
	}

	connstr, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, nil, err
	}

	opts := Options{
		Logger:     logr.Discard(),
		ConnString: connstr,
	}

	db, err := New(ctx, opts)
	if err != nil {
		return nil, nil, err
	}

	return db, container, nil
}

func NewTestDB(t *testing.T) (*DB, string) {
	ctx := context.Background()
	db, container, err := NewContainer()
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
		container.Terminate(ctx)
	})
	connstr, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	return db, connstr
}
