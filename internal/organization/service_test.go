package organization

import (
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/stretchr/testify/assert"
)

func TestAuthorize(t *testing.T) {
	tests := []struct {
		name     string
		subject  authz.Subject
		restrict bool
		want     error
	}{
		{
			"anyone can create an organization",
			&unprivUser{},
			false,
			nil,
		},
		{
			"site admin can create an organization when creation is restricted",
			&siteAdmin{},
			true,
			nil,
		},
		{
			"normal users cannot create an organization when creation is restricted",
			&unprivUser{},
			true,
			internal.ErrAccessNotPermitted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				Logger:                       logr.Discard(),
				RestrictOrganizationCreation: tt.restrict,
			}
			err := svc.restrictOrganizationCreation(tt.subject)
			assert.Equal(t, tt.want, err)
		})
	}
}

type unprivUser struct {
	authz.Subject
}

func (s *unprivUser) IsSiteAdmin() bool { return false }

type siteAdmin struct {
	authz.Subject
}

func (s *siteAdmin) IsSiteAdmin() bool { return true }
