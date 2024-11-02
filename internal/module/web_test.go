package module

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/connections"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/user"
	"github.com/leg100/otf/internal/vcs"
	"github.com/leg100/otf/internal/vcsprovider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListModules(t *testing.T) {
	h := newTestWebHandlers(t, withMod(&Module{}))

	q := "/?organization_name=acme-corp"
	r := httptest.NewRequest("GET", q, nil)
	r = r.WithContext(authz.AddSubjectToContext(r.Context(), &user.User{ID: "janitor"}))
	w := httptest.NewRecorder()
	h.list(w, r)
	if !assert.Equal(t, 200, w.Code) {
		t.Log(w.Body.String())
	}
}

func TestGetModule(t *testing.T) {
	tarball, err := os.ReadFile("./testdata/module.tar.gz")
	require.NoError(t, err)

	tests := []struct {
		name string
		mod  Module
	}{
		{
			name: "pending",
			mod: Module{
				Status: ModuleStatusPending,
			},
		},
		{
			name: "no versions",
			mod: Module{
				Status: ModuleStatusNoVersionTags,
			},
		},
		{
			name: "setup failed",
			mod: Module{
				Status: ModuleStatusSetupFailed,
			},
		},
		{
			name: "setup complete",
			mod: Module{
				Connection: &connections.Connection{},
				Status:     ModuleStatusSetupComplete,
				Versions:   []ModuleVersion{{Version: "1.0.0"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newTestWebHandlers(t, withMod(&tt.mod), withTarball(tarball), withHostname("fake-host.org"))

			q := "/?module_id=mod-123&version=1.0.0"
			r := httptest.NewRequest("GET", q, nil)
			w := httptest.NewRecorder()
			h.get(w, r)
			if !assert.Equal(t, 200, w.Code) {
				t.Log(w.Body.String())
			}
		})
	}
}

func TestNewModule_Connect(t *testing.T) {
	h := newTestWebHandlers(t, withVCSProviders(
		&vcsprovider.VCSProvider{},
		&vcsprovider.VCSProvider{},
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
		withVCSProviders(&vcsprovider.VCSProvider{}),
		withRepos(
			vcs.NewTestModuleRepo("aws", "vpc"),
			vcs.NewTestModuleRepo("aws", "s3"),
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
	h := newTestWebHandlers(t, withVCSProviders(&vcsprovider.VCSProvider{}))

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
	var svc fakeService
	for _, fn := range opts {
		fn(&svc)
	}
	return &webHandlers{
		Renderer:     testutils.NewRenderer(t),
		system:       &svc,
		client:       &svc,
		vcsproviders: &svc,
	}
}

type testWebOption func(*fakeService)

func withMod(mod *Module) testWebOption {
	return func(svc *fakeService) {
		svc.mod = mod
	}
}

func withTarball(tarball []byte) testWebOption {
	return func(svc *fakeService) {
		svc.tarball = tarball
	}
}

func withVCSProviders(vcsprovs ...*vcsprovider.VCSProvider) testWebOption {
	return func(svc *fakeService) {
		svc.vcsprovs = vcsprovs
	}
}

func withRepos(repos ...string) testWebOption {
	return func(svc *fakeService) {
		svc.repos = repos
	}
}

func withHostname(hostname string) testWebOption {
	return func(svc *fakeService) {
		svc.hostname = hostname
	}
}
