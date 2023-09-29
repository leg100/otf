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
	"github.com/stretchr/testify/require"
)

func TestWebhookHandler(t *testing.T) {
	hook, err := newHook(newHookOptions{
		vcsProviderID:   "vcs-123",
		cloud:           cloud.GithubKind,
		HostnameService: internal.NewHostnameService("fakehost.org"),
	})
	require.NoError(t, err)

	broker := &fakeBroker{}
	handler := newHandler(
		logr.Discard(),
		broker,
		&fakeHandlerDB{
			hook: hook,
		})
	handler.cloudHandlers.Set(cloud.GithubKind, func(http.ResponseWriter, *http.Request, string) *cloud.VCSEvent {
		return &cloud.VCSEvent{}
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/?webhook_id=158c758a-7090-11ed-a843-d398c839c7ad", nil)
	handler.ServeHTTP(w, r)
	assert.Equal(t, 200, w.Code, "response body: %s", w.Body.String())

	want := cloud.VCSEvent{RepoID: hook.id, VCSProviderID: "vcs-123", RepoPath: hook.identifier}
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

func (f *fakeBroker) Publish(got cloud.VCSEvent) { f.got = got }
