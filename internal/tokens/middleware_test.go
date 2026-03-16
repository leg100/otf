package tokens

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	t.Run("skip non-protected path", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/login", nil)
		w := httptest.NewRecorder()
		mw := &Middleware{}
		mw.Authenticate(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 200, w.Code)
	})

	t.Run("no authenticators", func(t *testing.T) {
		r := httptest.NewRequest("GET", "/api/v2/protected", nil)
		w := httptest.NewRecorder()
		mw := &Middleware{}
		mw.Authenticate(emptyHandler).ServeHTTP(w, r)
		assert.Equal(t, 401, w.Code)
	})
}

var emptyHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// implicitly responds with 200 OK
})
