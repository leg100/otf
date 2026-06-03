package authn

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/authz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errInvalidToken = errors.New("test error: invalid token")

type fakeAuthenticator struct {
	subject authz.Subject
	err     error
}

func (f *fakeAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) (authz.Subject, error) {
	return f.subject, f.err
}

type fakeRoute struct {
	path string
	err  error
}

func (f *fakeRoute) IsPath(path string) bool { return f.path == path }

func (f *fakeRoute) HandleError(err error, w http.ResponseWriter, r *http.Request) {
	f.err = err
}

func TestAuth(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		route          AuthenticationRoute
		authenticators []Authenticator
		wantCode       int
		wantError      error
		wantSubject    authz.Subject
	}{
		{
			name:      "no authenticators",
			path:      "/app/protected",
			route:     &fakeRoute{path: "/app/protected"},
			wantError: errLoginNeeded,
		},
		{
			name:     "skip unprotected path",
			path:     "/app/unprotected",
			route:    &fakeRoute{path: "/app/protected"},
			wantCode: 200,
		},
		{
			name: "authenticate protected path successfully",
			path: "/app/protected",
			authenticators: []Authenticator{&fakeAuthenticator{
				subject: &authz.Superuser{},
			}},
			route:       &fakeRoute{path: "/app/protected"},
			wantCode:    200,
			wantSubject: &authz.Superuser{},
		},
		{
			name: "disallow protected path",
			path: "/app/protected",
			authenticators: []Authenticator{&fakeAuthenticator{
				err: errInvalidToken,
			}},
			route:     &fakeRoute{path: "/app/protected"},
			wantError: errInvalidToken,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			var downstreamCtx context.Context
			emptyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				downstreamCtx = r.Context()
			})

			mw := &Middleware{Route: tt.route, Authenticators: tt.authenticators}
			mw.Authenticate(emptyHandler).ServeHTTP(w, r)

			if tt.wantCode != 0 {
				assert.Equal(t, tt.wantCode, w.Code)
			}
			if tt.wantError != nil {
				assert.Equal(t, tt.wantError, mw.Route.(*fakeRoute).err)
			}
			if tt.wantSubject != nil {
				subj, err := authz.SubjectFromContext(downstreamCtx)
				require.NoError(t, err)
				assert.Equal(t, tt.wantSubject, subj)
			}
		})
	}
}
