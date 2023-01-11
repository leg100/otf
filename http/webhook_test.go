package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
)

func TestWebhookHandler(t *testing.T) {
	got := make(chan cloud.VCSEvent, 1)
	want := cloud.VCSPushEvent{}
	handler := webhookHandler{
		events: got,
		Logger: logr.Discard(),
		Application: &fakeWebhookHandlerApp{
			hook: otf.NewTestWebhook(cloud.NewTestRepo(), cloud.Config{
				Cloud: &fakeCloud{event: &want},
			}),
		},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/?webhook_id=158c758a-7090-11ed-a843-d398c839c7ad", nil)
	handler.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code)

	assert.Equal(t, &want, <-got)
}

type fakeWebhookHandlerApp struct {
	hook *otf.Webhook

	otf.Application
}

func (f *fakeWebhookHandlerApp) DB() otf.DB {
	return &fakeWebhookHandlerDB{hook: f.hook}
}

type fakeWebhookHandlerDB struct {
	hook *otf.Webhook

	otf.DB
}

func (f *fakeWebhookHandlerDB) GetWebhook(ctx context.Context, id uuid.UUID) (*otf.Webhook, error) {
	return f.hook, nil
}

type fakeCloud struct {
	event cloud.VCSEvent

	cloud.Cloud
}

func (f *fakeCloud) HandleEvent(w http.ResponseWriter, r *http.Request, opts cloud.HandleEventOptions) cloud.VCSEvent {
	return f.event
}
