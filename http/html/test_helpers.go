package html

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertRedirect(t *testing.T, w *httptest.ResponseRecorder, path string) {
	if assert.Equal(t, 302, w.Code, w.Body.String()) {
		redirect, _ := w.Result().Location()
		assert.Equal(t, path, redirect.Path)
	}
}
