package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Get(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	svc := testutil.NewAuthService(t, db)

	t.Run("get", func(t *testing.T) {
		org1 := testutil.CreateOrganization(t, db)
		org2 := testutil.CreateOrganization(t, db)

		user := testutil.CreateUser(t, db,
			auth.WithOrganizations(org1.Name(), org2.Name()))

		got, err := svc.GetUser(ctx, otf.UserSpec{UserID: otf.String(user.ID())})
		require.NoError(t, err)
		assert.Len(t, got.Organizations(), 2)
	})

	t.Run("add organization membership", func(t *testing.T) {
		org := testutil.CreateOrganization(t, db)
		user := testutil.CreateUser(t, db)

		err := svc.AddOrganizationMembership(ctx, user.ID(), org.Name())
		require.NoError(t, err)

		got, err := svc.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
		require.NoError(t, err)

		assert.Contains(t, got.Organizations(), org.Name())
	})

	t.Run("remove organization membership", func(t *testing.T) {
		org := testutil.CreateOrganization(t, db)
		user := testutil.CreateUser(t, db)

		err := svc.RemoveOrganizationMembership(ctx, user.ID(), org.Name())
		require.NoError(t, err)

		got, err := svc.GetUser(ctx, otf.UserSpec{Username: otf.String(user.Username())})
		require.NoError(t, err)

		assert.NotContains(t, got.Organizations(), org)
	})
}
