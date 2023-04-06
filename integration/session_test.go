package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	t.Parallel()

	// perform all actions as superuser
	ctx := otf.AddSubjectToContext(context.Background(), &auth.SiteAdmin)

	t.Run("start", func(t *testing.T) {
		svc := setup(t, nil)
		want := svc.createUser(t, ctx)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/?", nil)
		err := svc.StartSession(w, r, tokens.StartSessionOptions{
			Username: &want.Username,
		})
		require.NoError(t, err)
		cookies := w.Result().Cookies()
		require.Equal(t, 1, len(cookies))

		t.Run("authenticate", func(t *testing.T) {
			upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got, err := auth.UserFromContext(r.Context())
				require.NoError(t, err)
				assert.Equal(t, want.Username, got.Username)
			})
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/app?", nil)
			r.AddCookie(cookies[0])
			svc.Middleware()(upstream).ServeHTTP(w, r)
			assert.Equal(t, 200, w.Code)
		})
	})
}
