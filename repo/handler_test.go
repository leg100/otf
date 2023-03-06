package repo

import (
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
)

func TestWebhookHandler(t *testing.T) {
	got := make(chan cloud.VCSEvent, 1)
	want := cloud.VCSPushEvent{}
	f := newTestFactory(t, want)
	hook := newTestHook(t, f, otf.String("123"))
	handler := handler{
		events: got,
		Logger: logr.Discard(),
		db:     &fakeHandlerDB{hook: hook},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/?webhook_id=158c758a-7090-11ed-a843-d398c839c7ad", nil)
	handler.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)

	assert.Equal(t, want, <-got)
}
