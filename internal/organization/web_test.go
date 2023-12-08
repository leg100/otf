package organization

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWeb_NewHandler(t *testing.T) {
	svc := &web{
		svc:      &fakeWebService{},
		Renderer: testutils.NewRenderer(t),
	}

	r := httptest.NewRequest("GET", "/?", nil)
	w := httptest.NewRecorder()
	svc.new(w, r)
	assert.Equal(t, 200, w.Code)
}

func TestWeb_CreateHandler(t *testing.T) {
	svc := &web{
		svc:      &fakeWebService{},
		Renderer: testutils.NewRenderer(t),
	}

	form := strings.NewReader(url.Values{
		"name": {"my-new-org"},
	}.Encode())

	r := httptest.NewRequest("POST", "/organization/create", form)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	svc.create(w, r)

	if assert.Equal(t, 302, w.Code) {
		redirect, err := w.Result().Location()
		require.NoError(t, err)
		assert.Equal(t, "/app/organizations/my-new-org", redirect.Path)
	}
}

func TestWeb_ListHandler(t *testing.T) {
	t.Run("pagination", func(t *testing.T) {
		orgs := make([]*Organization, 201)
		for i := 1; i <= 201; i++ {
			orgs[i-1] = &Organization{Name: uuid.NewString()}
		}
		svc := &web{
			svc:      &fakeWebService{orgs: orgs},
			Renderer: testutils.NewRenderer(t),
		}

		t.Run("first page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=1", nil)
			r = r.WithContext(internal.AddSubjectToContext(context.Background(), &internal.Superuser{}))
			w := httptest.NewRecorder()
			svc.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.NotContains(t, w.Body.String(), "Previous Page")
			assert.Contains(t, w.Body.String(), "Next Page")
		})

		t.Run("second page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=2", nil)
			r = r.WithContext(internal.AddSubjectToContext(context.Background(), &internal.Superuser{}))
			w := httptest.NewRecorder()
			svc.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.Contains(t, w.Body.String(), "Previous Page")
			assert.Contains(t, w.Body.String(), "Next Page")
		})

		t.Run("last page", func(t *testing.T) {
			r := httptest.NewRequest("GET", "/?page[number]=3", nil)
			r = r.WithContext(internal.AddSubjectToContext(context.Background(), &internal.Superuser{}))
			w := httptest.NewRecorder()
			svc.list(w, r)
			assert.Equal(t, 200, w.Code)
			assert.Contains(t, w.Body.String(), "Previous Page")
			assert.NotContains(t, w.Body.String(), "Next Page", w.Body.String())
		})
	})

	t.Run("new organization button", func(t *testing.T) {
		tests := []struct {
			name     string
			subject  internal.Subject
			restrict bool
			want     bool
		}{
			// unrestricted creation of organizations, so button should be
			// enabled, even to unprivileged users
			{"enabled", &unprivilegedSubject{}, false, true},
			// restricted creation of organizations, so button is disabled for
			// unprivileged users
			{"disabled", &unprivilegedSubject{}, true, false},
			// restricted creation of organizations, but button still enabled
			// for site admins
			{"site admin", &internal.Superuser{}, true, true},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				svc := &web{
					svc:              &fakeWebService{},
					Renderer:         testutils.NewRenderer(t),
					RestrictCreation: tt.restrict,
				}
				r := httptest.NewRequest("GET", "/?", nil)
				r = r.WithContext(internal.AddSubjectToContext(context.Background(), tt.subject))
				w := httptest.NewRecorder()
				svc.list(w, r)
				assert.Equal(t, 200, w.Code)
				doc, err := htmlquery.Parse(w.Body)
				require.NoError(t, err)
				button := htmlquery.FindOne(doc, `//button`)
				if assert.NotNil(t, button) {
					// if want button enabled, then the button should not contain a
					// 'disabled' attribute
					if tt.want {
						assert.NotContains(t, testutils.AttrMap(button), "disabled")
					} else {
						assert.Contains(t, testutils.AttrMap(button), "disabled")
					}
				}
			})
		}
	})
}

func TestWeb_DeleteHandler(t *testing.T) {
	svc := &web{
		svc: &fakeWebService{
			orgs: []*Organization{{Name: uuid.NewString()}},
		},
		Renderer: testutils.NewRenderer(t),
	}

	r := httptest.NewRequest("POST", "/?name=acme-corp", nil)
	w := httptest.NewRecorder()
	svc.delete(w, r)
	testutils.AssertRedirect(t, w, paths.Organizations())
}

type (
	fakeWebService struct {
		orgs []*Organization

		webService
	}

	unprivilegedSubject struct {
		internal.Subject
	}
)

func (f *fakeWebService) Create(ctx context.Context, opts CreateOptions) (*Organization, error) {
	return NewOrganization(opts)
}

func (f *fakeWebService) List(ctx context.Context, opts ListOptions) (*resource.Page[*Organization], error) {
	return resource.NewPage(f.orgs, opts.PageOptions, nil), nil
}

func (f *fakeWebService) Delete(context.Context, string) error {
	return nil
}

func (s *unprivilegedSubject) CanAccessSite(_ rbac.Action) bool {
	return false
}
