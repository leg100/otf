package repo

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	internal "github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
)

func TestWebhookHandler(t *testing.T) {
	publisher := &fakePublisher{}
	want := internal.Event{Type: internal.EventVCS, Payload: cloud.VCSPushEvent{}, Local: true}
	f := newTestFactory(t, cloud.VCSPushEvent{})
	hook := newTestHook(t, f, internal.String("123"))
	handler := handler{
		Publisher: publisher,
		Logger:    logr.Discard(),
		db:        &fakeHandlerDB{hook: hook},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/?webhook_id=158c758a-7090-11ed-a843-d398c839c7ad", nil)
	handler.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)

	assert.Equal(t, want, publisher.got)
}

type fakeHandlerDB struct {
	hook *hook
}

func (db *fakeHandlerDB) getHookByID(context.Context, uuid.UUID) (*hook, error) {
	return db.hook, nil
}

type fakePublisher struct {
	got internal.Event
}

func (f *fakePublisher) Publish(got internal.Event) { f.got = got }
