package user

import (
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/team"
	"github.com/stretchr/testify/assert"
)

func TestSiteAdminCanAccessOrganization(t *testing.T) {
	u := User{
		ID: SiteAdminID,
	}
	assert.True(t, u.CanAccess(authz.ListRunsAction, &authz.AccessRequest{Organization: "acme-corp"}))
}

func TestOwnerCanAccessOrganization(t *testing.T) {
	u := User{
		Teams: []*team.Team{
			{
				Name:         "owners",
				Organization: "acme-corp",
			},
		},
	}
	assert.True(t, u.CanAccess(authz.ListRunsAction, &authz.AccessRequest{Organization: "acme-corp"}))
}

func TestUser_Organizations(t *testing.T) {
	u := User{
		Teams: []*team.Team{
			{
				Name:         "owners",
				Organization: "acme-corp",
			},
			{
				Name:         "owners",
				Organization: "big-tobacco",
			},
			{
				Name:         "owners",
				Organization: "big-pharma",
			},
			{
				Name:         "engineers",
				Organization: "acme-corp",
			},
		},
	}
	want := u.Organizations()
	assert.Equal(t, 3, len(want), want)
	assert.Contains(t, want, "acme-corp")
	assert.Contains(t, want, "big-tobacco")
	assert.Contains(t, want, "big-pharma")
}
