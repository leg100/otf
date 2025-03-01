package vcsprovider

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	gogithub "github.com/google/go-github/v65/github"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/github"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestVCSProvider_newPersonalToken(t *testing.T) {
	svc := &webHandlers{}

	for _, kind := range []string{"github", "gitlab"} {
		t.Run(kind, func(t *testing.T) {
			q := "/?organization_name=acme-corp&kind=" + kind
			r := httptest.NewRequest("GET", q, nil)
			w := httptest.NewRecorder()
			svc.newPersonalToken(w, r)
			assert.Equal(t, 200, w.Code, w.Body.String())
		})
	}
}

func TestVCSProvider_newGithubApp(t *testing.T) {
	svc := &webHandlers{
		githubApps: &fakeGithubAppService{
			app: &github.App{},
			installs: []*github.Installation{{
				Installation: &gogithub.Installation{ID: internal.Int64(123)},
			}},
		},
	}

	q := "/?organization_name=acme-corp&"
	r := httptest.NewRequest("GET", q, nil)
	w := httptest.NewRecorder()
	svc.newGithubApp(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestCreateVCSProviderHandler(t *testing.T) {
	svc := &webHandlers{
		githubApps: &fakeGithubAppService{},
		client:     &fakeService{provider: &VCSProvider{Organization: "acme-corp"}},
	}

	r := httptest.NewRequest("POST", "/organization/acme-corp/vcs-providers/create", strings.NewReader(url.Values{
		"organization_name": {"acme-corp"},
		"token":             {"secret-token"},
		"name":              {"my-new-vcs-provider"},
		"kind":              {"fake-cloud"},
	}.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	svc.create(w, r)

	testutils.AssertRedirect(t, w, "/app/organizations/acme-corp/vcs-providers")
}

func TestListVCSProvidersHandler(t *testing.T) {
	svc := &webHandlers{
		githubApps: &fakeGithubAppService{},
		client:     &fakeService{provider: &VCSProvider{Organization: "acme-corp"}},
	}

	r := httptest.NewRequest("GET", "/?organization_name=acme-corp", nil)
	w := httptest.NewRecorder()
	svc.list(w, r)

	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestDeleteVCSProvidersHandler(t *testing.T) {
	svc := &webHandlers{
		client: &fakeService{provider: &VCSProvider{Organization: "acme"}},
	}

	r := httptest.NewRequest("POST", "/?", strings.NewReader(url.Values{
		"vcs_provider_id": {"fake-id"},
	}.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	svc.delete(w, r)

	testutils.AssertRedirect(t, w, "/app/organizations/acme/vcs-providers")
}

type fakeService struct {
	provider *VCSProvider

	Service
}

func (f *fakeService) Create(ctx context.Context, opts CreateOptions) (*VCSProvider, error) {
	return f.provider, nil
}

func (f *fakeService) List(context.Context, string) ([]*VCSProvider, error) {
	return []*VCSProvider{f.provider}, nil
}

func (f *fakeService) Delete(context.Context, resource.ID) (*VCSProvider, error) {
	return f.provider, nil
}

type fakeGithubAppService struct {
	app      *github.App
	installs []*github.Installation

	*github.Service
}

func (f *fakeGithubAppService) GetApp(context.Context) (*github.App, error) {
	return f.app, nil
}

func (f *fakeGithubAppService) ListInstallations(context.Context) ([]*github.Installation, error) {
	return f.installs, nil
}
