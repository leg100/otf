package html

import (
	"fmt"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/leg100/otf"
)

const (
	FlashSuccessType = "success"
	FlashErrorType   = "error"
)

type sessions struct {
	*scs.SessionManager
}

type FlashType string

type Flash struct {
	Type    FlashType
	Message string
}

func (s *sessions) FlashSuccess(r *http.Request, msg ...string) {
	s.flash(r, FlashSuccessType, msg...)
}

func (s *sessions) FlashError(r *http.Request, msg ...string) {
	s.flash(r, FlashErrorType, msg...)
}

func (s *sessions) flash(r *http.Request, t FlashType, msg ...string) {
	s.Put(r.Context(), otf.FlashSessionKey, Flash{
		Type:    t,
		Message: fmt.Sprint(convertStringSliceToInterfaceSlice(msg)...),
	})
}

func convertStringSliceToInterfaceSlice(ss []string) (is []interface{}) {
	for _, s := range ss {
		is = append(is, interface{}(s))
	}
	return
}

// PopFlashMessages retrieves all flash messages from the current session. The
// messages are thereafter discarded.
func (s *sessions) PopAllFlash(r *http.Request) (msgs []Flash) {
	if msg := s.Pop(r.Context(), otf.FlashSessionKey); msg != nil {
		msgs = append(msgs, msg.(Flash))
	}
	return
}

func (s *sessions) CurrentUser(r *http.Request) string {
	return s.GetString(r.Context(), otf.UsernameSessionKey)
}
