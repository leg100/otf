package module

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListModules(t *testing.T) {
	h := newTestWebHandlers(t, withMod(&Module{}))

	q := "/?organization_name=acme-corp"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	h.list(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestGetModule(t *testing.T) {
	mod := Module{
		Connection: &otf.Connection{},
		Status:     ModuleStatusSetupComplete,
		Latest:     &ModuleVersion{},
		Versions: map[string]*ModuleVersion{
			"1.0.0": {},
		},
	}
	tarball, err := os.ReadFile("./testdata/module.tar.gz")
	require.NoError(t, err)
	h := newTestWebHandlers(t, withMod(&mod), withTarball(tarball))

	q := "/?module_id=mod-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	h.get(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestNewModule_Connect(t *testing.T) {
	h := newTestWebHandlers(t, withVCSProviders(
		&otf.VCSProvider{},
		&otf.VCSProvider{},
	))

	q := "/?organization_name=acme-corp"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	h.newModuleConnect(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestNewModule_Repo(t *testing.T) {
	h := newTestWebHandlers(t,
		withVCSProviders(&otf.VCSProvider{}),
		withRepos(
			cloud.NewTestModuleRepo("aws", "vpc"),
			cloud.NewTestModuleRepo("aws", "s3"),
		),
	)

	q := "/?organization_name=acme-corp&vcs_provider_id=vcs-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	h.newModuleRepo(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestNewModule_Confirm(t *testing.T) {
	h := newTestWebHandlers(t, withVCSProviders(&otf.VCSProvider{}))

	q := "/?organization_name=acme-corp&vcs_provider_id=vcs-123&identifier=leg100/terraform-otf-test"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	h.newModuleConfirm(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestWeb_Publish(t *testing.T) {
	mod := Module{}
	h := newTestWebHandlers(t, withMod(&mod))

	q := "/?organization_name=acme-corp&vcs_provider_id=vcs-123&identifier=leg100/terraform-otf-test"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	h.publish(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Module(mod.ID), redirect.Path)
	}
}

func TestNewModule_Delete(t *testing.T) {
	mod := Module{Organization: "acme-corp"}
	h := newTestWebHandlers(t, withMod(&mod))

	q := "/?module_id=mod-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	h.delete(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Modules("acme-corp"), redirect.Path)
	}
}

func newTestWebHandlers(t *testing.T, opts ...testWebOption) *webHandlers {
	renderer, err := html.NewViewEngine(false)
	require.NoError(t, err)

	var svc fakeWebServices
	for _, fn := range opts {
		fn(&svc)
	}

	return &webHandlers{
		Renderer:           renderer,
		VCSProviderService: &svc,
		svc:                &svc,
	}
}

type testWebOption func(*fakeWebServices)

func withMod(mod *Module) testWebOption {
	return func(svc *fakeWebServices) {
		svc.mod = mod
	}
}

func withTarball(tarball []byte) testWebOption {
	return func(svc *fakeWebServices) {
		svc.tarball = tarball
	}
}

func withVCSProviders(vcsprovs ...*otf.VCSProvider) testWebOption {
	return func(svc *fakeWebServices) {
		svc.vcsprovs = vcsprovs
	}
}

func withRepos(repos ...string) testWebOption {
	return func(svc *fakeWebServices) {
		svc.repos = repos
	}
}

type fakeWebServices struct {
	mod      *Module
	tarball  []byte
	vcsprovs []*otf.VCSProvider
	repos    []string

	service

	otf.VCSProviderService
}

func (f *fakeWebServices) PublishModule(context.Context, PublishModuleOptions) (*Module, error) {
	return f.mod, nil
}

func (f *fakeWebServices) GetModuleByID(context.Context, string) (*Module, error) {
	return f.mod, nil
}

func (f *fakeWebServices) DeleteModule(context.Context, string) (*Module, error) {
	return f.mod, nil
}

func (f *fakeWebServices) ListModules(context.Context, ListModulesOptions) ([]*Module, error) {
	return []*Module{f.mod}, nil
}

func (f *fakeWebServices) GetVCSProvider(context.Context, string) (*otf.VCSProvider, error) {
	return f.vcsprovs[0], nil
}

func (f *fakeWebServices) ListVCSProviders(context.Context, string) ([]*otf.VCSProvider, error) {
	return f.vcsprovs, nil
}

func (f *fakeWebServices) GetVCSClient(ctx context.Context, providerID string) (cloud.Client, error) {
	return &fakeModulesCloudClient{repos: f.repos}, nil
}

func (f *fakeWebServices) downloadVersion(context.Context, string) ([]byte, error) {
	return f.tarball, nil
}

type fakeModulesCloudClient struct {
	repos []string

	cloud.Client
}

func (f *fakeModulesCloudClient) ListRepositories(ctx context.Context, opts cloud.ListRepositoriesOptions) ([]string, error) {
	return f.repos, nil
}
