package html

import (
	"context"
	"net/http"
)

const (
	FlashSuccess flashType = "success"
	FlashError   flashType = "error"
)

// flash represents a flash message indicating success/failure to a user
type flash struct {
	Type    flashType
	Message string
}

func flashSuccess(msg string) flash {
	return flash{Type: FlashSuccess, Message: msg}
}

func flashError(msg string) flash {
	return flash{Type: FlashError, Message: msg}
}

type flashType string

// flashStore stores flash messages in memory, storing no more than one flash
// message per session.
type flashStore struct {
	// mapping session token to flash message
	db map[string]flash
	// func for retrieving session token from context
	getToken func(context.Context) string
}

func newFlashStore() *flashStore {
	return &flashStore{
		db:       make(map[string]flash),
		getToken: getCtxToken,
	}
}

// push a flash message for the request's session
func (s *flashStore) push(r *http.Request, f flash) {
	token := s.getToken(r.Context())
	s.db[token] = f
}

// popFunc returns a func to pop a flash message for the current session - for
// use in a go template
func (s *flashStore) popFunc(r *http.Request) func() *flash {
	return func() *flash {
		token := s.getToken(r.Context())
		f, ok := s.db[token]
		if !ok {
			return nil
		}
		delete(s.db, token)
		return &f
	}
}
