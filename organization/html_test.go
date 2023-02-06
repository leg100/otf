package organization

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrganizationHandler(t *testing.T) {
	app := newFakeWebApp(t, &fakeOrganizationHandlerApp{})
	r := httptest.NewRequest("GET", "/organization/new", nil)
	w := httptest.NewRecorder()
	app.newOrganization(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestCreateOrganizationHandler(t *testing.T) {
	app := newFakeWebApp(t, &fakeOrganizationHandlerApp{})

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
	app := newFakeWebApp(t, &fakeOrganizationHandlerApp{orgs: orgs})

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

type fakeOrganizationHandlerApp struct {
	orgs []*otf.Organization
	otf.Application
}

func (f *fakeOrganizationHandlerApp) CreateOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return otf.NewOrganization(opts)
}

func (f *fakeOrganizationHandlerApp) ListOrganizations(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	return &otf.OrganizationList{
		Items:      f.orgs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.orgs)),
	}, nil
}

// TODO: do we need this?
func (f *fakeOrganizationHandlerApp) DeleteSession(context.Context, string) error {
	return nil
}
