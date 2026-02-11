package tfeapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/assert"
)

func TestResponder_Respond_AllowMemberNameWithInvalidCharacter(t *testing.T) {
	type resource struct {
		ID       string `jsonapi:"primary,valid-resources"`
		AtSymbol string `jsonapi:"attribute" json:"@timestamp"`
	}

	resp := NewResponder(logr.Discard())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/v2/test", nil)

	resp.Respond(w, r, &resource{ID: "test-123", AtSymbol: "value"}, http.StatusOK)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/vnd.api+json", w.Header().Get("Content-type"))

	want := `{"data":{"id":"test-123","type":"valid-resources","attributes":{"@timestamp":"value"}}}`
	assert.Equal(t, want, w.Body.String())
}
