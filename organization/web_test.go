package organization

import (
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeb(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		orgs := []*Organization{
			NewTestOrganization(t),
			NewTestOrganization(t),
			NewTestOrganization(t),
			NewTestOrganization(t),
			NewTestOrganization(t),
		}
		svc := newFakeWeb(t, &fakeService{orgs: orgs})

		t.Run("first page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=1&page[size]=2", nil)
			w := httptest.NewRecorder()
			svc.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.NotContains(t, w.Body.String(), "Previous Page")
			assert.Contains(t, w.Body.String(), "Next Page")
		})

		t.Run("second page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=2&page[size]=2", nil)
			w := httptest.NewRecorder()
			svc.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.Contains(t, w.Body.String(), "Previous Page")
			assert.Contains(t, w.Body.String(), "Next Page")
		})

		t.Run("last page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=3&page[size]=2", nil)
			w := httptest.NewRecorder()
			svc.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.Contains(t, w.Body.String(), "Previous Page")
			assert.NotContains(t, w.Body.String(), "Next Page")
		})
	})

	t.Run("delete", func(t *testing.T) {
		svc := newFakeWeb(t, &fakeService{
			orgs: []*Organization{NewTestOrganization(t)},
		})

		r := httptest.NewRequest("POST", "/?name=acme-corp", nil)
		w := httptest.NewRecorder()
		svc.delete(w, r)
		html.AssertRedirect(t, w, paths.Organizations())
	})
}

func newFakeWeb(t *testing.T, svc *fakeService) *web {
	renderer, err := html.NewViewEngine(false)
	require.NoError(t, err)
	return &web{
		svc:      svc,
		Renderer: renderer,
	}
}
