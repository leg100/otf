package auth

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
)

func TestTokenWeb(t *testing.T) {
	user := NewUser(uuid.NewString())

	t.Run("new", func(t *testing.T) {
		web := newTestTokenHandlers(t, "acme-org")
		q := "/?"
		r := httptest.NewRequest("GET", q, nil)
		w := httptest.NewRecorder()

		web.newTokenHandler(w, r)

		if !assert.Equal(t, 200, w.Code) {
			t.Log(t, w.Body.String())
		}
	})

	t.Run("create", func(t *testing.T) {
		web := newTestTokenHandlers(t, "acme-org")
		q := "/?"
		r := httptest.NewRequest("GET", q, nil)
		r = r.WithContext(otf.AddSubjectToContext(context.Background(), user))
		w := httptest.NewRecorder()

		web.createTokenHandler(w, r)

		if assert.Equal(t, 302, w.Code) {
			redirect, _ := w.Result().Location()
			assert.Equal(t, paths.Tokens(), redirect.Path)
		}
	})

	t.Run("list", func(t *testing.T) {
		web := newTestTokenHandlers(t, "acme-org")
		q := "/?"
		r := httptest.NewRequest("GET", q, nil)
		r = r.WithContext(otf.AddSubjectToContext(context.Background(), user))
		w := httptest.NewRecorder()

		web.tokensHandler(w, r)

		if !assert.Equal(t, 200, w.Code) {
			t.Log(t, w.Body.String())
		}
	})

	t.Run("delete", func(t *testing.T) {
		web := newTestTokenHandlers(t, "acme-org")
		q := "/?id=token-123"
		r := httptest.NewRequest("POST", q, nil)
		r = r.WithContext(otf.AddSubjectToContext(context.Background(), user))
		w := httptest.NewRecorder()

		web.deleteTokenHandler(w, r)

		if assert.Equal(t, 302, w.Code) {
			redirect, _ := w.Result().Location()
			assert.Equal(t, paths.Tokens(), redirect.Path)
		}
	})
}

func newTestTokenHandlers(t *testing.T, org string) *webHandlers {
	return newFakeWeb(t, &fakeService{
		userToken: NewTestToken(t, org),
	})
}
