package auth

import (
	"testing"

	"github.com/leg100/otf/rbac"
	"github.com/stretchr/testify/assert"
)

func TestSiteAdminCanAccessOrganization(t *testing.T) {
	u := User{
		ID: SiteAdminID,
	}
	assert.True(t, u.CanAccessOrganization(rbac.ListRunsAction, "acme-corp"))
}

func TestOwnerCanAccessOrganization(t *testing.T) {
	u := User{
		Teams: []*Team{
			{
				Name:         "owners",
				Organization: "acme-corp",
			},
		},
	}
	assert.True(t, u.CanAccessOrganization(rbac.ListRunsAction, "acme-corp"))
}

func TestUser_Organizations(t *testing.T) {
	u := User{
		Teams: []*Team{
			{
				Name:         "owners",
				Organization: "acme-corp",
			},
			{
				Name:         "owners",
				Organization: "big-tabacco",
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
	want := []string{
		"acme-corp",
		"big-tabacco",
		"big-pharma",
	}
	assert.Equal(t, want, u.Organizations())
}
