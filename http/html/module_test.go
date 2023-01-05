package html

import (
	"context"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListModules(t *testing.T) {
	org := otf.NewTestOrganization(t)
	mod := otf.NewTestModule(org)
	app := newFakeWebApp(t, &fakeModulesApp{mod: mod})

	q := "/?organization_name=acme-corp"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.listModules(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestGetModule(t *testing.T) {
	org := otf.NewTestOrganization(t)
	mod := otf.NewTestModule(org,
		otf.WithModuleRepo(),
		otf.WithModuleStatus(otf.ModuleStatusSetupComplete),
		otf.WithModuleVersion("1.0.0", otf.ModuleVersionStatusOk),
	)
	tarball, err := os.ReadFile("./testdata/module.tar.gz")
	require.NoError(t, err)
	app := newFakeWebApp(t, &fakeModulesApp{
		mod:     mod,
		tarball: tarball,
	})

	q := "/?module_id=mod-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.getModule(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestNewModule_Connect(t *testing.T) {
	org := otf.NewTestOrganization(t)
	provider := otf.NewTestVCSProvider(t, org)
	app := newFakeWebApp(t, &fakeModulesApp{
		provider: provider,
	})

	q := "/?organization_name=acme-corp"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.newModuleConnect(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestNewModule_Repo(t *testing.T) {
	org := otf.NewTestOrganization(t)
	provider := otf.NewTestVCSProvider(t, org)
	app := newFakeWebApp(t, &fakeModulesApp{
		provider: provider,
		repos: []*otf.Repo{
			otf.NewTestModuleRepo("aws", "vpc"),
			otf.NewTestModuleRepo("aws", "s3"),
		},
	})

	q := "/?organization_name=acme-corp&vcs_provider_id=vcs-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.newModuleRepo(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestNewModule_Confirm(t *testing.T) {
	org := otf.NewTestOrganization(t)
	provider := otf.NewTestVCSProvider(t, org)
	app := newFakeWebApp(t, &fakeModulesApp{
		provider: provider,
	})

	q := "/?organization_name=acme-corp&vcs_provider_id=vcs-123&identifier=leg100/terraform-otf-test"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.newModuleConfirm(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestNewModule_Create(t *testing.T) {
	org := otf.NewTestOrganization(t)
	provider := otf.NewTestVCSProvider(t, org)
	mod := otf.NewTestModule(org)
	app := newFakeWebApp(t, &fakeModulesApp{
		org:      org,
		provider: provider,
		mod:      mod,
	})

	q := "/?organization_name=acme-corp&vcs_provider_id=vcs-123&identifier=leg100/terraform-otf-test"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.createModule(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Module(mod.ID()), redirect.Path)
	}
}

func TestNewModule_Delete(t *testing.T) {
	org := otf.NewTestOrganization(t)
	mod := otf.NewTestModule(org)
	app := newFakeWebApp(t, &fakeModulesApp{
		org: org,
		mod: mod,
	})

	q := "/?module_id=mod-123"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	app.deleteModule(w, r)
	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, paths.Modules(org.Name()), redirect.Path)
	}
}

type fakeModulesApp struct {
	org      *otf.Organization
	mod      *otf.Module
	provider *otf.VCSProvider
	tarball  []byte
	repos    []*otf.Repo

	otf.Application
}

func (f *fakeModulesApp) PublishModule(context.Context, otf.PublishModuleOptions) (*otf.Module, error) {
	return f.mod, nil
}

func (f *fakeModulesApp) GetOrganization(context.Context, string) (*otf.Organization, error) {
	return f.org, nil
}

func (f *fakeModulesApp) GetModuleByID(context.Context, string) (*otf.Module, error) {
	return f.mod, nil
}

func (f *fakeModulesApp) DeleteModule(context.Context, string) (*otf.Module, error) {
	return f.mod, nil
}

func (f *fakeModulesApp) ListModules(context.Context, otf.ListModulesOptions) ([]*otf.Module, error) {
	return []*otf.Module{f.mod}, nil
}

func (f *fakeModulesApp) DownloadModuleVersion(context.Context, otf.DownloadModuleOptions) ([]byte, error) {
	return f.tarball, nil
}

func (f *fakeModulesApp) GetVCSProvider(context.Context, string) (*otf.VCSProvider, error) {
	return f.provider, nil
}

func (f *fakeModulesApp) ListVCSProviders(context.Context, string) ([]*otf.VCSProvider, error) {
	return []*otf.VCSProvider{f.provider}, nil
}

func (f *fakeModulesApp) ListRepositories(ctx context.Context, providerID string, opts otf.ListOptions) (*otf.RepoList, error) {
	return &otf.RepoList{
		Items:      f.repos,
		Pagination: otf.NewPagination(opts, len(f.repos)),
	}, nil
}
