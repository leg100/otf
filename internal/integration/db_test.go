package integration

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
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

	ctx := context.Background()
	db, err := sql.New(ctx, logr.Discard(), sql.NewTestDB(t))
	require.NoError(t, err)
	t.Cleanup(db.Close)

	org, err := organization.NewOrganization(organization.CreateOptions{
		Name: internal.String("acmeco"),
	})
	require.NoError(t, err)

	// retain reference to old context for testing below.
	oldContext := ctx

	err = db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		q := &organization.Queries{}
		// insert org using tx
		err := q.InsertOrganization(ctx, conn, organization.InsertOrganizationParams{
			ID:                         org.ID,
			CreatedAt:                  sql.Timestamptz(org.CreatedAt),
			UpdatedAt:                  sql.Timestamptz(org.UpdatedAt),
			Name:                       org.Name,
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
		// query org just created using same tx conn.
		_, err = q.FindOrganizationByID(ctx, conn, org.ID)
		assert.NoError(t, err)

		err = db.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
			// query org just created using child tx conn
			_, err = q.FindOrganizationByID(ctx, conn, org.ID)
			return err
		})
		require.NoError(t, err)

		// this should fail because it is using a different conn from the old
		// context
		_, err = q.FindOrganizationByID(ctx, db.Conn(oldContext), org.ID)
		assert.ErrorIs(t, err, pgx.ErrNoRows)

		return nil
	})
	require.NoError(t, err)
}
