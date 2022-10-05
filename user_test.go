package otf

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSynchroniseOrganizations(t *testing.T) {
	app := &fakeSynchroniseOrganizationsApp{}
	err := SynchroniseOrganizations(context.Background(), app, NewUser("fake-user"), "org-1", "org-2")
	require.NoError(t, err)

	if assert.Equal(t, 3, len(app.synced)) {
		assert.Equal(t, "org-1", app.synced[0].Name())
		assert.Equal(t, "org-2", app.synced[1].Name())
		assert.Equal(t, "fake-user", app.synced[2].Name())
	}
}

type fakeSynchroniseOrganizationsApp struct {
	// list of synchronised organizations
	synced []*Organization
	Application
}

func (f *fakeSynchroniseOrganizationsApp) EnsureCreatedOrganization(ctx context.Context, opts OrganizationCreateOptions) (*Organization, error) {
	return NewOrganization(opts)
}

func (f *fakeSynchroniseOrganizationsApp) SyncOrganizationMemberships(ctx context.Context, u *User, orgs []*Organization) (*User, error) {
	f.synced = orgs
	return nil, nil
}
