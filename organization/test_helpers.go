package organization

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/pubsub"
	"github.com/stretchr/testify/require"
)

func NewTestService(t *testing.T, db otf.DB, opts *Options) *service {
	if opts != nil {
		return NewService(*opts)
	}
	return NewService(Options{
		Logger: logr.Discard(),
		DB:     db,
		Broker: pubsub.NewTestBroker(t, db),
	})
}

func NewTestOrganization(t *testing.T) *Organization {
	org, err := NewOrganization(OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)
	return org
}

// CreateTestOrganization creates an organization in the database for testing
// purposes.
func CreateTestOrganization(t *testing.T, db otf.DB) *Organization {
	ctx := context.Background()
	orgdb := &pgdb{db}

	org, err := NewOrganization(OrganizationCreateOptions{
		Name: otf.String(uuid.NewString()),
	})
	require.NoError(t, err)

	err = orgdb.create(ctx, org)
	require.NoError(t, err)

	t.Cleanup(func() {
		orgdb.delete(ctx, org.Name)
	})
	return org
}

func createTestOrganization(t *testing.T, db *pgdb) *Organization {
	ctx := context.Background()
	org := NewTestOrganization(t)
	err := db.create(ctx, org)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.delete(ctx, org.Name)
	})
	return org
}

type fakeService struct {
	orgs []*Organization

	Service
}

func (f *fakeService) CreateOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error) {
	return NewOrganization(opts)
}

func (f *fakeService) ListOrganizations(ctx context.Context, opts OrganizationListOptions) (*OrganizationList, error) {
	return &OrganizationList{
		Items:      f.orgs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.orgs)),
	}, nil
}

type fakeBroker struct {
	pubsub.Broker
}

func (f *fakeBroker) Publish(otf.Event)                           {}
func (f *fakeBroker) Register(table string, getter pubsub.Getter) {}
