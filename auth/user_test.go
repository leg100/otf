package auth

import (
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/stretchr/testify/assert"
)

func TestSiteAdminCanAccessOrganization(t *testing.T) {
	u := User{
		ID:            otf.SiteAdminID,
		Organizations: []string{"acme-corp"},
	}
	assert.True(t, u.CanAccessOrganization(rbac.ListRunsAction, "acme-corp"))
}

func TestOwnerCanAccessOrganization(t *testing.T) {
	u := User{
		Organizations: []string{"acme-corp"},
		Teams: []*Team{
			{
				Name:         "owners",
				Organization: "acme-corp",
			},
		},
	}
	assert.True(t, u.CanAccessOrganization(rbac.ListRunsAction, "acme-corp"))
}
