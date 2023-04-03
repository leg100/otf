package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/orgcreator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	t.Run("new user", func(t *testing.T) {
		svc := setup(t, nil)

		err := svc.Sync(ctx, cloud.User{
			Name: "bobby",
			Teams: []cloud.Team{
				{Name: "owners", Organization: "acme-corp"},
				{Name: "devs", Organization: "big-pharma"},
			},
		})
		require.NoError(t, err)

		t.Run("made owner of acme-corp", func(t *testing.T) {
			_, err = svc.GetOrganization(ctx, "acme-corp")
			assert.NoError(t, err)
			isOwner(t, svc, "bobby", "bobby")
		})

		t.Run("made owner of personal org", func(t *testing.T) {
			_, err = svc.GetOrganization(ctx, "bobby")
			assert.NoError(t, err)
			isOwner(t, svc, "bobby", "bobby")
		})

		t.Run("should not have created big-pharma", func(t *testing.T) {
			_, err = svc.GetOrganization(ctx, "big-pharma")
			assert.True(t, errors.Is(err, otf.ErrResourceNotFound))
		})
	})

	t.Run("existing user", func(t *testing.T) {
		svc := setup(t, nil)

		// create existing user:
		// 1) member of a existing of an org
		existing := svc.createTeam(t, ctx, nil)
		user, err := svc.CreateUser(ctx, "bobby", auth.WithTeams(existing))
		require.NoError(t, err)
		// 2) owner of personal org
		userCtx := otf.AddSubjectToContext(ctx, user)
		_, err = svc.CreateOrganization(userCtx, orgcreator.OrganizationCreateOptions{
			Name: otf.String("bobby"),
		})
		require.NoError(t, err)

		err = svc.Sync(ctx, cloud.User{
			Name: "bobby",
			Teams: []cloud.Team{
				// new org
				{Name: "owners", Organization: "acme-corp"},
			},
		})
		require.NoError(t, err)

		t.Run("made owner of acme-corp", func(t *testing.T) {
			_, err = svc.GetOrganization(ctx, "acme-corp")
			assert.NoError(t, err)
			isOwner(t, svc, "bobby", "bobby")
		})

		t.Run("removed from existing team", func(t *testing.T) {
			members, err := svc.ListTeamMembers(ctx, existing.ID)
			require.NoError(t, err)
			assert.Empty(t, members)
		})
	})
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
