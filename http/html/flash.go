package html

import (
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

// flashStack is a simple implementation of a flash stack, storing flash
// messages in memory, and storing no more than one flash message per session.
type flashStack struct {
	// mapping session token to flash message
	stack map[string]flash
}

func newFlashStack() *flashStack {
	return &flashStack{
		stack: make(map[string]flash),
	}
}

// push a flash message for the request's session
func (s *flashStack) push(r *http.Request, f flash) {
	session := getCtxSession(r.Context())
	if session == nil {
		panic("pushing flash message: no session found")
	}
	s.stack[session.Token] = f
}

// popFunc returns a func to pop a flash message for the current session - for
// use in a go template
func (s *flashStack) popFunc(r *http.Request) func() *flash {
	return func() *flash {
		token := getCtxSession(r.Context()).Token
		f, ok := s.stack[token]
		if !ok {
			return nil
		}
		delete(s.stack, token)
		return &f
	}
}
