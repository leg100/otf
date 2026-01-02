package repohooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/vcs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_repohookHandler(t *testing.T) {
	hook, err := newRepohook(newRepohookOptions{
		vcsProviderID:   testutils.ParseID(t, "vcs-123"),
		vcsKindID:       vcs.KindID("test"),
		HostnameService: internal.NewHostnameService("fakehost.org"),
	})
	require.NoError(t, err)

	broker := &fakeBroker{}
	handler := newHandler(
		logr.Discard(),
		broker,
		&fakeVCSKindDB{
			handler: func(*http.Request, string) (*vcs.EventPayload, error) {
				return &vcs.EventPayload{}, nil
			},
		},
		&fakeHandlerDB{
			hook: hook,
		},
	)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/?webhook_id=158c758a-7090-11ed-a843-d398c839c7ad", nil)
	handler.repohookHandler(w, r)
	assert.Equal(t, 200, w.Code, "response body: %s", w.Body.String())

	want := vcs.Event{
		EventHeader: vcs.EventHeader{
			VCSProviderID: testutils.ParseID(t, "vcs-123"),
		},
		EventPayload: vcs.EventPayload{Repo: hook.repoPath},
	}
	assert.Equal(t, want, broker.got)
}

type (
	fakeHandlerDB struct {
		hook *hook
	}
	fakeVCSKindDB struct {
		handler func(*http.Request, string) (*vcs.EventPayload, error)
	}
	fakeBroker struct {
		got vcs.Event
	}
)

func (db *fakeHandlerDB) getHookByID(context.Context, uuid.UUID) (*hook, error) {
	return db.hook, nil
}

func (f *fakeBroker) Publish(got vcs.Event) { f.got = got }

func (db *fakeVCSKindDB) GetKind(id vcs.KindID) (vcs.Kind, error) {
	return vcs.Kind{
		EventHandler: db.handler,
	}, nil
}
