package organization

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/antchfx/htmlquery"
	internal "github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/leg100/otf/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeb_ListHandler(t *testing.T) {
	t.Run("pagination", func(t *testing.T) {
		orgs := []*Organization{
			NewTestOrganization(t),
			NewTestOrganization(t),
			NewTestOrganization(t),
			NewTestOrganization(t),
			NewTestOrganization(t),
		}
		svc := newFakeWeb(t, &fakeService{orgs: orgs}, false)

		t.Run("first page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=1&page[size]=2", nil)
			r = r.WithContext(internal.AddSubjectToContext(context.Background(), &internal.Superuser{}))
			w := httptest.NewRecorder()
			svc.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.NotContains(t, w.Body.String(), "Previous Page")
			assert.Contains(t, w.Body.String(), "Next Page")
		})

		t.Run("second page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=2&page[size]=2", nil)
			r = r.WithContext(internal.AddSubjectToContext(context.Background(), &internal.Superuser{}))
			w := httptest.NewRecorder()
			svc.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.Contains(t, w.Body.String(), "Previous Page")
			assert.Contains(t, w.Body.String(), "Next Page")
		})

		t.Run("last page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=3&page[size]=2", nil)
			r = r.WithContext(internal.AddSubjectToContext(context.Background(), &internal.Superuser{}))
			w := httptest.NewRecorder()
			svc.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.Contains(t, w.Body.String(), "Previous Page")
			assert.NotContains(t, w.Body.String(), "Next Page")
		})
	})

	t.Run("new organization button", func(t *testing.T) {
		tests := []struct {
			name     string
			subject  internal.Subject
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
			{"site admin", &internal.Superuser{}, true, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := newFakeWeb(t, &fakeService{}, tt.restrict)
				r := httptest.NewRequest("GET", "/?", nil)
				r = r.WithContext(internal.AddSubjectToContext(context.Background(), tt.subject))
				w := httptest.NewRecorder()
				svc.list(w, r)
				assert.Equal(t, 200, w.Code)
				doc, err := htmlquery.Parse(w.Body)
				require.NoError(t, err)
				button := htmlquery.FindOne(doc, `//button`)
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

func TestWeb_DeleteHandler(t *testing.T) {
	svc := newFakeWeb(t, &fakeService{
		orgs: []*Organization{NewTestOrganization(t)},
	}, false)

	r := httptest.NewRequest("POST", "/?name=acme-corp", nil)
	w := httptest.NewRecorder()
	svc.delete(w, r)
	testutils.AssertRedirect(t, w, paths.Organizations())
}
