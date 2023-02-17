package testutil

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_Get(t *testing.T) {
	ctx := context.Background()
	db := sql.NewTestDB(t)
	orgService := NewOrganizationService(t, db)
	service := newAuthService(t, db)

	org1 := createOrganization(t, orgService)
	org2 := createOrganization(t, orgService)

	user := createUser(t, service,
		auth.WithOrganizations(org1.Name(), org2.Name()))

	got, err := service.GetUser(ctx, otf.UserSpec{UserID: otf.String(user.ID())})
	require.NoError(t, err)

	assert.Equal(t, got.ID(), user.ID())
	assert.Equal(t, got.Username(), user.Username())
	assert.Len(t, got.Organizations(), 2)
}

func newAuthService(t *testing.T, db otf.DB) *auth.Service {
	service, err := auth.NewService(context.Background(), auth.Options{
		Authorizer: &AllowAllAuthorizer{User: &otf.Superuser{"bob"}},
		DB:         db,
		Logger:     logr.Discard(),
	})
	require.NoError(t, err)
	return service
}

func createUser(t *testing.T, service *auth.Service, opts ...auth.NewUserOption) *auth.User {
	ctx := context.Background()

	user, err := service.CreateUser(ctx, uuid.NewString())
	require.NoError(t, err)

	t.Cleanup(func() {
		service.DeleteUser(ctx, user.Username())
	})
	return user
}
