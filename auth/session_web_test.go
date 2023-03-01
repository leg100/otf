package auth

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/stretchr/testify/assert"
)

func TestSessionHandlers(t *testing.T) {
	user := NewUser(uuid.NewString())
	active := newTestSession(t, user.id, nil)
	other := newTestSession(t, user.id, nil)

	h := newFakeWeb(t, &fakeService{
		sessionService: &fakeSessionService{
			sessions: []*Session{active, other},
		},
	})

	t.Run("list sessions", func(t *testing.T) {
		// add user and active session to request
		r := httptest.NewRequest("GET", "/sessions", nil)
		r = r.WithContext(otf.AddSubjectToContext(context.Background(), user))
		r = r.WithContext(addSessionCtx(r.Context(), active))

		w := httptest.NewRecorder()
		h.sessionsHandler(w, r)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("revoke session", func(t *testing.T) {
		form := strings.NewReader(url.Values{
			"token": {"asklfdkljfj"},
		}.Encode())

		r := httptest.NewRequest("POST", "/sessions/delete", form)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		h.revokeSessionHandler(w, r)

		assert.Equal(t, 302, w.Code)
	})
}
