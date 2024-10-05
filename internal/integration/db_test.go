package integration

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
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

	for i := 0; i < 100; i++ {
		func() {
			err := db.WaitAndLock(ctx, 123, func(context.Context) error { return nil })
			require.NoError(t, err)
		}()
	}
}

// TestTx tests database transactions.
func TestTx(t *testing.T) {
	integrationTest(t)

	ctx := context.Background()
	db, err := sql.New(ctx, logr.Discard(), sql.NewTestDB(t))
	require.NoError(t, err)
	t.Cleanup(db.Close)

	org, err := organization.NewOrganization(organization.CreateOptions{
		Name: internal.String("acmeco"),
	})
	require.NoError(t, err)

	err = db.Tx(ctx, func(txCtx context.Context, q *sqlc.Queries) error {
		err := q.InsertOrganization(txCtx, sqlc.InsertOrganizationParams{
			ID:                         sql.String(org.ID),
			CreatedAt:                  sql.Timestamptz(org.CreatedAt),
			UpdatedAt:                  sql.Timestamptz(org.UpdatedAt),
			Name:                       sql.String(org.Name),
			Email:                      sql.StringPtr(org.Email),
			CollaboratorAuthPolicy:     sql.StringPtr(org.CollaboratorAuthPolicy),
			CostEstimationEnabled:      sql.Bool(org.CostEstimationEnabled),
			SessionRemember:            sql.Int4Ptr(org.SessionRemember),
			SessionTimeout:             sql.Int4Ptr(org.SessionTimeout),
			AllowForceDeleteWorkspaces: sql.Bool(org.AllowForceDeleteWorkspaces),
		})
		if err != nil {
			return err
		}

		// this should succeed because it is using the same querier from the
		// same tx
		_, err = q.FindOrganizationByID(txCtx, sql.String(org.ID))
		assert.NoError(t, err)

		// this should succeed because it is using the same ctx from the same tx
		_, err = db.Querier(txCtx).FindOrganizationByID(txCtx, sql.String(org.ID))
		assert.NoError(t, err)

		err = db.Tx(txCtx, func(ctx context.Context, q *sqlc.Queries) error {
			// this should succeed because it is using a child tx via the
			// querier
			_, err = q.FindOrganizationByID(ctx, sql.String(org.ID))
			assert.NoError(t, err)

			// this should succeed because it is using a child tx via the
			// context
			_, err = db.Querier(ctx).FindOrganizationByID(ctx, sql.String(org.ID))
			assert.NoError(t, err)

			return nil
		})
		require.NoError(t, err)

		// this should fail because it is using a different ctx
		_, err = db.Querier(ctx).FindOrganizationByID(txCtx, sql.String(org.ID))
		assert.ErrorIs(t, err, pgx.ErrNoRows)

		return nil
	})
	require.NoError(t, err)
}
