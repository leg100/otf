package vcsprovider

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/inmem"
	"github.com/leg100/otf/organization"
	"github.com/stretchr/testify/require"
)

type testServiceOption func(*service)

func WithCloudService(cloudService CloudService) testServiceOption {
	return func(svc *service) {
		svc.factory.CloudService = cloudService
		svc.web.CloudService = cloudService
		svc.db.factory.CloudService = cloudService
	}
}

func NewTestService(t *testing.T, db otf.DB, opts ...testServiceOption) *service {
	service := NewService(Options{
		DB:           db,
		Logger:       logr.Discard(),
		CloudService: inmem.NewTestCloudService(),
	})
	service.organization = otf.NewAllowAllAuthorizer()
	for _, fn := range opts {
		fn(service)
	}
	return service
}

func NewTestVCSProvider(t *testing.T, org *organization.Organization) *VCSProvider {
	var organizationName string
	if org == nil {
		organizationName = uuid.NewString()
	} else {
		organizationName = org.Name
	}
	factory := &factory{inmem.NewTestCloudService()}
	provider, err := factory.new(createOptions{
		Organization: organizationName,
		// unit tests require a legitimate cloud name to avoid invalid foreign
		// key error upon insert/update
		Cloud: "github",
		Name:  uuid.NewString(),
		Token: uuid.NewString(),
	})
	require.NoError(t, err)
	return provider
}

func CreateTestVCSProvider(t *testing.T, db otf.DB, org *organization.Organization, opts ...testServiceOption) *VCSProvider {
	ctx := context.Background()
	svc := NewTestService(t, db, opts...)

	vcsprov, err := svc.create(ctx, createOptions{
		Organization: org.Name,
		// unit tests require a legitimate cloud name to avoid invalid foreign
		// key error upon insert/update
		Cloud: "github",
		Name:  uuid.NewString(),
		Token: uuid.NewString(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		svc.delete(ctx, vcsprov.ID)
	})
	return vcsprov
}

func createTestVCSProvider(t *testing.T, db *pgdb, org *organization.Organization) *VCSProvider {
	provider := NewTestVCSProvider(t, org)
	ctx := context.Background()

	err := db.create(ctx, provider)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, provider.ID)
	})
	return provider
}
