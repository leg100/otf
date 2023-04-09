package tokens

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	t.Run("skip non-protected path", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/login", nil)
		w := httptest.NewRecorder()
		fakeTokenMiddleware(t, "secret")(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("missing token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		w := httptest.NewRecorder()
		fakeTokenMiddleware(t, "secret")(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 401, w.Code)
	})

	t.Run("valid site token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add("Authorization", "Bearer secret")
		w := httptest.NewRecorder()
		fakeSiteTokenMiddleware(t, "secret")(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("invalid site token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add("Authorization", "Bearer incorrect")
		w := httptest.NewRecorder()
		fakeSiteTokenMiddleware(t, "secret")(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 401, w.Code)
	})

	t.Run("valid user token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		token := newTestJWT(t, "secret", userTokenKind, time.Hour)
		r.Header.Add("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		fakeTokenMiddleware(t, "secret")(wantSubjectHandler(t, &auth.User{})).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("valid run token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		token := newTestJWT(t, "secret", runTokenKind, time.Hour, "organization", "acme-corp")
		r.Header.Add("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		fakeTokenMiddleware(t, "secret")(wantSubjectHandler(t, &RunToken{})).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code, w.Body.String())
	})

	t.Run("valid agent token", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		token := newTestJWT(t, "secret", agentTokenKind, time.Hour)
		r.Header.Add("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		fakeTokenMiddleware(t, "secret")(wantSubjectHandler(t, &AgentToken{})).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("invalid jwt", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		token := newTestJWT(t, "wrong-secret", agentTokenKind, time.Hour)
		r.Header.Add("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		fakeTokenMiddleware(t, "secret")(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 401, w.Code)
	})

	t.Run("valid user session", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/app/protected", nil)
		token := newTestJWT(t, "secret", userTokenKind, time.Hour)
		r.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
		w := httptest.NewRecorder()
		fakeTokenMiddleware(t, "secret")(wantSubjectHandler(t, &auth.User{})).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("expired user session", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/app/protected", nil)
		token := newTestJWT(t, "secret", userTokenKind, -time.Hour)
		r.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
		w := httptest.NewRecorder()
		fakeTokenMiddleware(t, "secret")(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 302, w.Code)
	})

	t.Run("missing session cookie", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/app/protected", nil)
		w := httptest.NewRecorder()
		fakeTokenMiddleware(t, "secret")(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 302, w.Code)
	})

	t.Run("valid iap token", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add(googleIAPHeader, newIAPToken(t, "https://example.com"))
		fakeIAPMiddleware(t, "")(wantSubjectHandler(t, &auth.User{})).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("valid iap token for ui path", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/app/protected", nil)
		r.Header.Add(googleIAPHeader, newIAPToken(t, "https://example.com"))
		fakeIAPMiddleware(t, "")(wantSubjectHandler(t, &auth.User{})).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("valid iap audience", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add(googleIAPHeader, newIAPToken(t, "https://example.com"))
		fakeIAPMiddleware(t, "https://example.com")(wantSubjectHandler(t, &auth.User{})).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("invalid iap audience", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		r.Header.Add(googleIAPHeader, newIAPToken(t, "https://example.com"))
		fakeIAPMiddleware(t, "https://invalid.com")(wantSubjectHandler(t, &auth.User{})).ServeHTTP(w, r)
		assert.Equal(t, 401, w.Code)
	})
}
