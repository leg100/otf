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

func TestNewVCSProviderHandler(t *testing.T) {
	app := newFakeWebApp(t, &fakeVCSProviderApp{})
	r := httptest.NewRequest("GET", "/organization/acme-corp/vcs-providers/new", nil)
	w := httptest.NewRecorder()
	app.newOrganization(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestCreateVCSProviderHandler(t *testing.T) {
	app := newFakeWebApp(t, &fakeVCSProviderApp{})
	app.cloudConfigs = map[string]cloudConfig{
		"fake-cloud": {},
	}

	form := strings.NewReader(url.Values{
		"organization_name": {"acme-corp"},
		"token":             {"secret-token"},
		"name":              {"my-new-vcs-provider"},
		"cloud_name":        {"fake-cloud"},
	}.Encode())

	r := httptest.NewRequest("POST", "/organization/acme-corp/vcs-providers/create", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	app.createVCSProvider(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/organizations/acme-corp/vcs-providers", redirect.Path)
	} else {
		t.Log(w.Body.String())
	}
}

func TestListVCSProvidersHandler(t *testing.T) {
	org := otf.NewTestOrganization(t)
	app := newFakeWebApp(t, &fakeVCSProviderApp{providers: []*otf.VCSProvider{
		otf.NewTestVCSProvider(t, org, fakeCloud{}),
		otf.NewTestVCSProvider(t, org, fakeCloud{}),
		otf.NewTestVCSProvider(t, org, fakeCloud{}),
	}})

	r := httptest.NewRequest("GET", "/organization/acme-corp/vcs-providers", nil)
	w := httptest.NewRecorder()
	app.listVCSProviders(w, r)

	assert.Equal(t, 200, w.Code)
}

func TestDeleteVCSProvidersHandler(t *testing.T) {
	app := newFakeWebApp(t, &fakeVCSProviderApp{})

	form := strings.NewReader(url.Values{
		"organization_name": {"acme-corp"},
		"id":                {"fake-id"},
	}.Encode())

	r := httptest.NewRequest("POST", "/organization/acme-corp/vcs-providers/delete", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	app.deleteVCSProvider(w, r)

	assert.Equal(t, 302, w.Code)
}

type fakeVCSProviderApp struct {
	otf.Application
	providers []*otf.VCSProvider
}

func (f *fakeVCSProviderApp) CreateVCSProvider(ctx context.Context, opts otf.VCSProviderCreateOptions) (*otf.VCSProvider, error) {
	return otf.NewVCSProvider(opts), nil
}

func (f *fakeVCSProviderApp) ListVCSProviders(context.Context, string) ([]*otf.VCSProvider, error) {
	return f.providers, nil
}

func (f *fakeVCSProviderApp) DeleteVCSProvider(context.Context, string, string) error {
	return nil
}
