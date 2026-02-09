package ui

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization_ListHandler(t *testing.T) {
	t.Run("new organization button", func(t *testing.T) {
		tests := []struct {
			name     string
			subject  authz.Subject
			restrict bool
			want     bool
		}{
			// unrestricted creation of organizations, so button should be
			// enabled, even to unprivileged users
			{"enabled", &unprivilegedSubject{}, false, true},
			// restricted creation of organizations, so button is disabled for
			// unprivileged users
			{"disabled", &unprivilegedSubject{}, true, false},
			// restricted creation of organizations, but button still enabled
			// for site admins
			{"site admin", &authz.Superuser{}, true, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &Handlers{
					Organizations:                &fakeOrganizationService{},
					RestrictOrganizationCreation: tt.restrict,
				}
				r := httptest.NewRequest("GET", "/?", nil)
				r = r.WithContext(authz.AddSubjectToContext(context.Background(), tt.subject))
				w := httptest.NewRecorder()
				svc.listOrganizations(w, r)
				assert.Equal(t, 200, w.Code, w.Body.String())
				doc, err := htmlquery.Parse(w.Body)
				require.NoError(t, err)
				button := htmlquery.FindOne(doc, `//button[@id='new-organization-button']`)
				if assert.NotNil(t, button) {
					// if want button enabled, then the button should not contain a
					// 'disabled' attribute
					if tt.want {
						assert.NotContains(t, testutils.AttrMap(button), "disabled")
					} else {
						assert.Contains(t, testutils.AttrMap(button), "disabled")
					}
				}
			})
		}
	})
}

type (
	fakeOrganizationService struct {
		*organization.Service
		orgs []*organization.Organization
	}

	unprivilegedSubject struct {
		authz.Subject
	}
)

func (f *fakeOrganizationService) List(ctx context.Context, opts organization.ListOptions) (*resource.Page[*organization.Organization], error) {
	return resource.NewPage(f.orgs, opts.PageOptions, nil), nil
}

func (s *unprivilegedSubject) CanAccess(authz.Action, authz.Request) bool {
	return false
}
