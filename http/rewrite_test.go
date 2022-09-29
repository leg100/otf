package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewrite(t *testing.T) {
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	router := NewRouter()
	router.PathPrefix("/signed/{foo}.{bar}").Sub(func(signed *Router) {
		signed.Use(mw)
		signed.GET("/somewhere_else", func(w http.ResponseWriter, r *http.Request) {
			t.Logf("foo = %s", mux.Vars(r)["foo"])
			t.Logf("bar = %s", mux.Vars(r)["bar"])
		})
	})

	srv := httptest.NewServer(router)
	defer srv.Close()

	res, err := http.Get(srv.URL + "/signed/abc.def/somewhere_else")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
