package html

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlash(t *testing.T) {
	w := httptest.NewRecorder()
	FlashSuccess(w, "great news")
	cookies := w.Result().Cookies()
	if assert.Equal(t, 1, len(cookies)) {
		t.Run("pop flash", func(t *testing.T) {
			w = httptest.NewRecorder()
			r := fakeRequest(cookies[0])
			got := popFlashFunc(w, r)()
			if assert.NotNil(t, got) {
				assert.Equal(t, "great news", got.Message)
			}
			cookies = w.Result().Cookies()
			if assert.Equal(t, 1, len(cookies)) {
				assert.Equal(t, -1, cookies[0].MaxAge)
			}
		})
	}
}

func fakeRequest(cookie *http.Cookie) *http.Request {
	r := &http.Request{}
	r.Header = make(map[string][]string)
	r.AddCookie(cookie)
	return r
}
