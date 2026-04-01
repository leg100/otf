package user

import (
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/team"
	"github.com/stretchr/testify/assert"
)

func TestSiteAdminCanAccessOrganization(t *testing.T) {
	u := User{
		ID: SiteAdminID,
	}
	assert.True(t, u.CanAccess(authz.ListRunsAction, authz.Request{}))
}

func TestOwnerCanAccessOrganization(t *testing.T) {
	org := organization.NewTestName(t)
	u := User{
		Teams: []*team.Team{
			{
				Name:         team.Owners,
				Organization: org,
			},
		},
	}
	assert.True(t, u.CanAccess(authz.ListRunsAction, authz.Request{ID: org}))
}

func TestUser_Organizations(t *testing.T) {
	org1 := organization.NewTestName(t)
	org2 := organization.NewTestName(t)
	org3 := organization.NewTestName(t)

	u := User{
		Teams: []*team.Team{
			{
				Name:         team.Owners,
				Organization: org1,
			},
			{
				Name:         team.Owners,
				Organization: org2,
			},
			{
				Name:         team.Owners,
				Organization: org3,
			},
			{
				Name:         "engineers",
				Organization: org1,
			},
		},
	}
	want := u.Organizations()
	assert.Equal(t, 3, len(want), want)
	assert.Contains(t, want, org1)
	assert.Contains(t, want, org2)
	assert.Contains(t, want, org3)
}
