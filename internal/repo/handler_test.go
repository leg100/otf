package repo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/stretchr/testify/assert"
)

func TestWebhookHandler(t *testing.T) {
	broker := &fakeBroker{}
	f := newTestFactory(t, cloud.VCSEvent{})
	hook := newTestHook(t, f, "vcs-123", internal.String("123"))
	hook.cloudHandler = fakeCloudHandler{}
	want := cloud.VCSEvent{RepoID: hook.id, VCSProviderID: "vcs-123", RepoPath: hook.identifier}
	handler := handler{
		Logger:        logr.Discard(),
		handlerBroker: broker,
		handlerDB:     &fakeHandlerDB{hook: hook},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/?webhook_id=158c758a-7090-11ed-a843-d398c839c7ad", nil)
	handler.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code, "response body: %s", w.Body.String())
	assert.Equal(t, want, broker.got)
}

type (
	fakeHandlerDB struct {
		hook *Hook
	}
	fakeBroker struct {
		got cloud.VCSEvent
	}
	fakeCloudHandler struct{}
)

func (db *fakeHandlerDB) getHookByID(context.Context, uuid.UUID) (*Hook, error) {
	return db.hook, nil
}

func (f *fakeBroker) publish(got cloud.VCSEvent) { f.got = got }

func (fakeCloudHandler) HandleEvent(w http.ResponseWriter, r *http.Request, secret string) *cloud.VCSEvent {
	return &cloud.VCSEvent{}
}
