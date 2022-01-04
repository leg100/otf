package html

import (
	"html/template"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/leg100/otf"
)

// Wrap upstream session manager

type sessions struct {
	scs.SessionManager
}

func (s *sessions) currentUser(r *http.Request) string {
	return s.GetString(r.Context(), otf.UsernameSessionKey)
}

func (s *sessions) popFlashMessage(r *http.Request) template.HTML {
	msg := s.PopString(r.Context(), otf.FlashSessionKey)
	return template.HTML(msg)
}
