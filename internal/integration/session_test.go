package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSession(t *testing.T) {
	integrationTest(t)

	t.Run("start", func(t *testing.T) {
		svc, _, ctx := setup(t, nil)
		want := userFromContext(t, ctx)
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/?", nil)
		err := svc.Tokens.StartSession(w, r, tokens.StartSessionOptions{
			Username: &want.Username,
		})
		require.NoError(t, err)
		cookies := w.Result().Cookies()
		require.Equal(t, 1, len(cookies))

		t.Run("authenticate", func(t *testing.T) {
			upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got := userFromContext(t, ctx)
				assert.Equal(t, want.Username, got.Username)
			})
			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/app/protected", nil)
			r.AddCookie(cookies[0])
			svc.Tokens.Middleware()(upstream).ServeHTTP(w, r)
			assert.Equal(t, 200, w.Code)
		})
	})
}
