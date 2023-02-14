package organization

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeWeb struct {
	app *fakeApp
	otf.Renderer
}

func newFakeWeb(t *testing.T, app *fakeApp) *web {
	renderer, err := html.NewViewEngine(true)
	require.NoError(t, err)
	return &web{
		app:      app,
		Renderer: renderer,
	}
}

func TestNewOrganizationHandler(t *testing.T) {
	app := newFakeWebApp(t, &fakeApp{})
	r := httptest.NewRequest("GET", "/organization/new", nil)
	w := httptest.NewRecorder()
	app.newOrganization(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestCreateOrganizationHandler(t *testing.T) {
	app := newFakeWebApp(t, &fakeApp{})

	form := strings.NewReader(url.Values{
		"name": {"my-new-org"},
	}.Encode())

	r := httptest.NewRequest("POST", "/organization/create", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	app.createOrganization(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/organizations/my-new-org", redirect.Path)
	}
}

func TestListOrganizationsHandler(t *testing.T) {
	orgs := []*otf.Organization{
		otf.NewTestOrganization(t),
		otf.NewTestOrganization(t),
		otf.NewTestOrganization(t),
		otf.NewTestOrganization(t),
		otf.NewTestOrganization(t),
	}
	app := newFakeWebApp(t, &fakeApp{orgs: orgs})

	t.Run("first page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?page[number]=1&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listOrganizations(w, r)
		assert.Equal(t, 200, w.Code)
		assert.NotContains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("second page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?page[number]=2&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listOrganizations(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.Contains(t, w.Body.String(), "Next Page")
	})

	t.Run("last page", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/?page[number]=3&page[size]=2", nil)
		w := httptest.NewRecorder()
		app.listOrganizations(w, r)
		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Body.String(), "Previous Page")
		assert.NotContains(t, w.Body.String(), "Next Page")
	})
}
