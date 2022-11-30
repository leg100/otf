package http

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestWebhookHandler(t *testing.T) {
	want := &otf.Webhook{}
	srv := &Server{
		Application: &fakeWebhookHandlerApp{hook: want},
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/?webhook_id=158c758a-7090-11ed-a843-d398c839c7ad", nil)
	srv.webhookHandler(w, r)
	assert.Equal(t, 200, w.Code)
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
