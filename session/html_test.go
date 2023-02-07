package session

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestSessionHandlers(t *testing.T) {
	user := otf.NewTestUser(t)
	active := otf.NewTestSession(t, user.ID())
	other := otf.NewTestSession(t, user.ID())

	app := newFakeWebApp(t, &fakeSessionHandlerApp{sessions: []*otf.Session{active, other}})

	t.Run("list sessions", func(t *testing.T) {
		// add user and active session to request
		r := httptest.NewRequest("GET", "/sessions", nil)
		r = r.WithContext(otf.AddSubjectToContext(context.Background(), user))
		r = r.WithContext(addSessionToContext(r.Context(), active))

		w := httptest.NewRecorder()
		app.sessionsHandler(w, r)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("revoke session", func(t *testing.T) {
		form := strings.NewReader(url.Values{
			"token": {"asklfdkljfj"},
		}.Encode())

		r := httptest.NewRequest("POST", "/sessions/delete", form)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		app.revokeSessionHandler(w, r)

		assert.Equal(t, 302, w.Code)
	})
}

type fakeSessionHandlerApp struct {
	sessions []*otf.Session
	otf.Application
}

func (f *fakeSessionHandlerApp) ListSessions(context.Context, string) ([]*otf.Session, error) {
	return f.sessions, nil
}

func (f *fakeSessionHandlerApp) DeleteSession(context.Context, string) error {
	return nil
}
