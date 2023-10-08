package github

import (
	"context"
	"net/http/httptest"
	"testing"

	gogithub "github.com/google/go-github/v55/github"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestWebHandlers_new(t *testing.T) {
	h := &webHandlers{
		Renderer:        testutils.NewRenderer(t),
		HostnameService: internal.NewHostnameService("example.com"),
	}

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	h.new(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

func TestWebHandlers_get(t *testing.T) {
	h := &webHandlers{
		Renderer:        testutils.NewRenderer(t),
		HostnameService: internal.NewHostnameService("example.com"),
		svc: &fakeService{
			app: &App{},
			installs: []*Installation{{
				Installation: &gogithub.Installation{ID: internal.Int64(123)},
			}},
		},
	}

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	h.get(w, r)
	assert.Equal(t, 200, w.Code, w.Body.String())
}

type fakeService struct {
	app      *App
	installs []*Installation
	GithubAppService
}

func (f *fakeService) GetGithubApp(context.Context) (*App, error) {
	return f.app, nil
}

func (f *fakeService) ListInstallations(context.Context) ([]*Installation, error) {
	return f.installs, nil
}
