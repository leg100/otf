package organization

import (
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeb(t *testing.T) {
	t.Run("new", func(t *testing.T) {
		wd, _ := os.Getwd()
		t.Log(wd)

		app := newFakeWeb(t, &fakeApp{})

		r := httptest.NewRequest("GET", "/?", nil)
		w := httptest.NewRecorder()
		app.new(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("create", func(t *testing.T) {
		app := newFakeWeb(t, &fakeApp{})

		form := strings.NewReader(url.Values{
			"name": {"my-new-org"},
		}.Encode())

		r := httptest.NewRequest("POST", "/organization/create", form)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		app.create(w, r)

		if assert.Equal(t, 302, w.Code) {
			redirect, err := w.Result().Location()
			require.NoError(t, err)
			assert.Equal(t, "/organizations/my-new-org", redirect.Path)
		}
	})

	t.Run("list", func(t *testing.T) {
		orgs := []*Organization{
			NewTestOrganization(t),
			NewTestOrganization(t),
			NewTestOrganization(t),
			NewTestOrganization(t),
			NewTestOrganization(t),
		}
		app := newFakeWeb(t, &fakeApp{orgs: orgs})

		t.Run("first page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=1&page[size]=2", nil)
			w := httptest.NewRecorder()
			app.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.NotContains(t, w.Body.String(), "Previous Page")
			assert.Contains(t, w.Body.String(), "Next Page")
		})

		t.Run("second page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=2&page[size]=2", nil)
			w := httptest.NewRecorder()
			app.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.Contains(t, w.Body.String(), "Previous Page")
			assert.Contains(t, w.Body.String(), "Next Page")
		})

		t.Run("last page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=3&page[size]=2", nil)
			w := httptest.NewRecorder()
			app.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.Contains(t, w.Body.String(), "Previous Page")
			assert.NotContains(t, w.Body.String(), "Next Page")
		})
	})
}

type fakeWeb struct {
	app *fakeApp
	otf.Renderer
}

func newFakeWeb(t *testing.T, app *fakeApp) *web {
	renderer, err := html.NewViewEngine(false)
	require.NoError(t, err)
	return &web{
		app:      app,
		Renderer: renderer,
	}
}
