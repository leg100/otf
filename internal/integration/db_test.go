package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWaitAndLock tests acquiring a connection from a pool, obtaining a session
// lock and then releasing lock and the connection, and it does this several
// times, to demonstrate that it is returning resources and not running into
// limits.
func TestWaitAndLock(t *testing.T) {
	integrationTest(t)

	ctx := context.Background()
	db, err := sql.New(ctx, logr.Discard(), sql.NewTestDB(t))
	require.NoError(t, err)
	t.Cleanup(db.Close)

	for range 100 {
		func() {
			err := db.WaitAndLock(ctx, 123, func(context.Context) error { return nil })
			require.NoError(t, err)
		}()
	}
}

// TestTx tests database transactions.
func TestTx(t *testing.T) {
	integrationTest(t)

	daemon, _, ctx := setup(t)

	// retain reference to old context for testing below.
	oldContext := ctx

	err := daemon.DB.Tx(ctx, func(ctx context.Context) error {
		// insert org using tx
		org := daemon.createOrganization(t, ctx)

		// query org just created using same tx conn.
		_, err := daemon.Organizations.Get(ctx, org.Name)
		assert.NoError(t, err)

		err = daemon.Tx(ctx, func(ctx context.Context) error {
			// query org just created using child tx conn
			_, err := daemon.Organizations.Get(ctx, org.Name)
			return err
		})
		require.NoError(t, err)

		// this should fail because it is using a different conn from the old
		// context
		_, err = daemon.Organizations.Get(oldContext, org.Name)
		assert.ErrorIs(t, err, internal.ErrResourceNotFound)

		return nil
	})
	require.NoError(t, err)
}
