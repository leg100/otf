package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/require"
)

// TestWaitAndLock tests acquiring a connection from a pool, obtaining a session
// lock and then releasing lock and the connection, and it does this several
// times, to demonstrate that it is returning resources and not running into
// limits.
func TestWaitAndLock(t *testing.T) {
	integrationTest(t)

	ctx := context.Background()
	db, err := sql.New(ctx, sql.Options{
		Logger:     logr.Discard(),
		ConnString: sql.NewTestDB(t),
	})
	require.NoError(t, err)
	t.Cleanup(db.Close)

	for i := 0; i < 100; i++ {
		func() {
			err := db.WaitAndLock(ctx, 123, func(context.Context) error { return nil })
			require.NoError(t, err)
		}()
	}
}
