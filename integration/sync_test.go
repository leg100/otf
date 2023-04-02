package integration

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	svc := setup(t, nil)
	err := svc.Sync(ctx, cloud.User{
		Name: "bobby",
		Teams: []cloud.Team{
			{Name: "owners", Organization: "acme-corp"},
		},
	})
	require.NoError(t, err)

	// should have created an organization named acme-corp
	_, err = svc.GetOrganization(ctx, "acme-corp")
	assert.NoError(t, err)

	// and made them an owner of acme-corp
	isOwner(t, svc, "bobby", "bobby")

	// should have created a personal organization
	_, err = svc.GetOrganization(ctx, "bobby")
	assert.NoError(t, err)

	// and made them an owner of their personal organization
	isOwner(t, svc, "bobby", "bobby")
}

func isOwner(t *testing.T, svc *testDaemon, organization, username string) {
	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	// and made them an owner of acme-corp
	owners, err := svc.GetTeam(ctx, organization, "owners")
	assert.NoError(t, err)
	members, err := svc.ListTeamMembers(ctx, owners.ID)
	require.NoError(t, err)
	if assert.Equal(t, 1, len(members)) {
		assert.Equal(t, members[0].Username, username)
	}
}
