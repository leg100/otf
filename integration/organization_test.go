package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_Create(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	svc := testutil.NewOrganizationService(t, db)

	t.Run("create", func(t *testing.T) {
		name := uuid.NewString()

		t.Cleanup(func() {
			svc.DeleteOrganization(ctx, name)
		})

		_, err := svc.CreateOrganization(ctx, organization.OrganizationCreateOptions{
			Name: otf.String(uuid.NewString()),
		})
		require.NoError(t, err)

		t.Run("duplicate error", func(t *testing.T) {
			_, err := svc.CreateOrganization(ctx, organization.OrganizationCreateOptions{
				Name: otf.String(uuid.NewString()),
			})
			require.Equal(t, otf.ErrResourceAlreadyExists, err)
		})
	})

	t.Run("update", func(t *testing.T) {
		org := testutil.CreateOrganization(t, db)

		// update org name
		want := uuid.NewString()
		org, err := svc.UpdateOrganization(ctx, org.Name(), organization.UpdateOptions{
			Name: &want,
		})
		require.NoError(t, err)

		assert.Equal(t, want, org.Name())
	})

	t.Run("get", func(t *testing.T) {
		want := testutil.CreateOrganization(t, db)

		got, err := svc.GetOrganization(ctx, want.Name())
		require.NoError(t, err)

		assert.Equal(t, want, got)
	})

	t.Run("list by user", func(t *testing.T) {
		org1 := testutil.CreateOrganization(t, db)
		org2 := testutil.CreateOrganization(t, db)
		user := testutil.CreateUser(t, db,
			auth.WithOrganizations(org1.Name(), org2.Name()))

		// make call as the user to filter orgs by user
		ctx = otf.AddSubjectToContext(ctx, user)
		got, err := svc.ListOrganizations(ctx, organization.ListOptions{})
		require.NoError(t, err)

		assert.Contains(t, got, org1)
		assert.Contains(t, got, org2)
	})
}
