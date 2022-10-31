package html

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
	app := newFakeWebApp(t, &fakeCreateOrganizationApp{})
	r := httptest.NewRequest("GET", "/organization/new", nil)
	w := httptest.NewRecorder()
	app.newOrganization(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestCreateOrganizationHandler(t *testing.T) {
	app := newFakeWebApp(t, &fakeCreateOrganizationApp{})

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

type fakeCreateOrganizationApp struct {
	otf.Application
}

func (f *fakeCreateOrganizationApp) CreateOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return otf.NewOrganization(opts)
}

func (f *fakeCreateOrganizationApp) DeleteSession(context.Context, string) error {
	return nil
}
