package repo

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestWebhookHandler(t *testing.T) {
	broker := &fakeBroker{}
	want := pubsub.Event{Type: pubsub.EventVCS, Payload: cloud.VCSEvent{}, Local: true}
	f := newTestFactory(t, cloud.VCSEvent{})
	hook := newTestHook(t, f, "vcs-123", internal.String("123"))
	handler := handler{
		Logger:        logr.Discard(),
		handlerBroker: broker,
		handlerDB:     &fakeHandlerDB{hook: hook},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/?webhook_id=158c758a-7090-11ed-a843-d398c839c7ad", nil)
	handler.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)

	assert.Equal(t, want, broker.got)
}

type (
	fakeHandlerDB struct {
		hook *hook
	}
	fakeBroker struct {
		got cloud.VCSEvent
	}
)

func (db *fakeHandlerDB) getHookByID(context.Context, uuid.UUID) (*hook, error) {
	return db.hook, nil
}

func (f *fakeBroker) publish(got cloud.VCSEvent) { f.got = got }
